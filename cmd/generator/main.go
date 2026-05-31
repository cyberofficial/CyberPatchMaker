package main

import (
	"crypto/sha256"
	"encoding/binary"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cyberofficial/cyberpatchmaker/internal/core/config"
	"github.com/cyberofficial/cyberpatchmaker/internal/core/patcher"
	"github.com/cyberofficial/cyberpatchmaker/internal/core/version"
	"github.com/cyberofficial/cyberpatchmaker/pkg/utils"
)

func main() {
	// Define flags
	versionsDir := flag.String("versions-dir", "", "Directory containing version folders")
	newVersion := flag.String("new-version", "", "New version number to generate patches for")
	from := flag.String("from", "", "Source version number (for single patch)")
	to := flag.String("to", "", "Target version number (for single patch)")
	fromDir := flag.String("from-dir", "", "Full path to source version directory (overrides --versions-dir/--from)")
	toDir := flag.String("to-dir", "", "Full path to target version directory (overrides --versions-dir/--to)")
	output := flag.String("output", "", "Output directory for patches")
	keyFile := flag.String("key-file", "", "Specific key file to use (e.g., app.exe, game.exe)")
	compression := flag.String("compression", "zstd", "Compression algorithm (zstd, gzip, none)")
	level := flag.Int("level", 3, "Compression level (1-4 for zstd, 1-3 for gzip)")
	verify := flag.Bool("verify", true, "Verify patches after creation")
	createExe := flag.Bool("create-exe", false, "Create self-contained CLI executable")
	silent := flag.Bool("silent", false, "Enable silent mode in generated executable (auto-apply without prompts)")
	crp := flag.Bool("crp", false, "Create reverse patch (for downgrades)")
	saveScans := flag.Bool("savescans", false, "Save directory scans to cache for faster subsequent patches")
	rescan := flag.Bool("rescan", false, "Force rescan of cached versions")
	scanData := flag.String("scandata", "", "Custom directory for scan cache (default: .data)")
	jobs := flag.Int("jobs", 0, "Number of parallel workers (0 = auto-detect CPU cores, 1 = single-threaded)")
	splitSize := flag.String("splitsize", "", "Custom multi-part split size (e.g., '2G', '2GB', '500M', '500MB'). Default: 4GB")
	bypassSplitLimit := flag.Bool("bypasssplitlimit", false, "Bypass 100MB minimum split size check")
	versionFlag := flag.Bool("version", false, "Show version information")
	help := flag.Bool("help", false, "Show help message")

	flag.Parse()

	// Show version if requested
	if *versionFlag {
		fmt.Printf("CyberPatchMaker Patch Generator v%s\n", version.GetVersion())
		return
	}

	if *help {
		printHelp()
		return
	}

	// Load configuration
	cfg := config.NewManager()
	configPath := config.GetDefaultConfigPath()
	if err := cfg.Load(configPath); err != nil {
		fmt.Printf("Warning: failed to load config: %v\n", err)
	}

	// Initialize version manager
	versionMgr := version.NewManager()

	// Set worker threads for parallel operations
	workerCount := *jobs
	if workerCount == 0 {
		// Auto-detect: use number of CPU cores
		workerCount = cfg.GetConfig().WorkerThreads
	}
	if workerCount < 1 {
		workerCount = 1
	}
	versionMgr.SetWorkerThreads(workerCount)
	if workerCount > 1 {
		fmt.Printf("✓ Using %d worker threads for parallel operations\n", workerCount)
	}

	// Enable scan caching if requested
	if *saveScans {
		versionMgr.EnableScanCache(*scanData, *rescan)
		fmt.Printf("✓ Scan caching enabled (cache dir: %s)\n", versionMgr.GetScanCache().GetCacheDir())
		if *rescan {
			fmt.Println("  Force rescan: enabled")
		}
	}

	// Parse custom split size if provided
	var customMaxPartSize int64
	if *splitSize != "" {
		parsedSize, err := parseSplitSize(*splitSize)
		if err != nil {
			fmt.Printf("Error: invalid split size: %v\n", err)
			os.Exit(1)
		}

		// Check if below 100MB minimum and bypass flag not set
		const minSplitSize = 100 * 1024 * 1024 // 100MB
		if parsedSize < minSplitSize && !*bypassSplitLimit {
			fmt.Printf("\nWarning: Split size %.2f MB is below recommended minimum of 100 MB\n", float64(parsedSize)/(1024*1024))
			fmt.Printf("This may create many small parts and is not recommended unless upload space is very limited.\n")
			fmt.Print("Do you want to continue? (yes/no): ")

			var response string
			fmt.Scanln(&response)
			response = strings.ToLower(strings.TrimSpace(response))

			if response != "yes" && response != "y" {
				fmt.Println("Aborted. Use --bypasssplitlimit to skip this confirmation.")
				os.Exit(0)
			}
		}

		customMaxPartSize = parsedSize
		fmt.Printf("✓ Custom split size: %.2f GB (%.0f MB)\n", float64(customMaxPartSize)/(1024*1024*1024), float64(customMaxPartSize)/(1024*1024))
	}

	// Set output directory
	outputDir := *output
	if outputDir == "" {
		outputDir = cfg.GetConfig().DefaultPatchOutput
	}
	if outputDir == "" {
		outputDir = "patches"
	}

	// Ensure output directory exists
	if err := utils.EnsureDir(outputDir); err != nil {
		fmt.Printf("Error: failed to create output directory: %v\n", err)
		os.Exit(1)
	}

	// Handle different modes
	if *newVersion != "" && *versionsDir != "" {
		// Generate patches from all existing versions to new version
		generateAllPatches(versionMgr, *versionsDir, *newVersion, outputDir, *keyFile, *compression, *level, *verify, *createExe, *silent, *crp, customMaxPartSize)
	} else if *fromDir != "" && *toDir != "" {
		// Generate single patch using custom directory paths
		generateSinglePatchCustomPaths(versionMgr, *fromDir, *toDir, outputDir, *keyFile, *compression, *level, *verify, *createExe, *silent, *crp, customMaxPartSize)
	} else if *from != "" && *to != "" && *versionsDir != "" {
		// Generate single patch using versions-dir
		generateSinglePatch(versionMgr, *versionsDir, *from, *to, outputDir, *keyFile, *compression, *level, *verify, *createExe, *silent, *crp, customMaxPartSize)
	} else {
		fmt.Println("Error: insufficient arguments")
		printHelp()
		os.Exit(1)
	}
}

// detectKeyFile resolves the key file for a version directory.
// If customKeyFile is non-empty, validates it exists. Otherwise auto-detects
// from standard names: program.exe, game.exe, app.exe, main.exe.
// Returns the key file name, or empty string with nil error if none auto-detected.
func detectKeyFile(dirPath, customKeyFile string) (string, error) {
	if customKeyFile != "" {
		if utils.FileExists(filepath.Join(dirPath, customKeyFile)) {
			return customKeyFile, nil
		}
		return "", fmt.Errorf("custom key file not found: %s", customKeyFile)
	}

	candidates := []string{"program.exe", "game.exe", "app.exe", "main.exe"}
	for _, kf := range candidates {
		if utils.FileExists(filepath.Join(dirPath, kf)) {
			return kf, nil
		}
	}
	return "", nil
}

// parseSplitSize parses a size string like "2G", "2GB", "500M", "500MB" into bytes
func parseSplitSize(sizeStr string) (int64, error) {
	sizeStr = strings.ToUpper(strings.TrimSpace(sizeStr))

	// Extract numeric part and unit
	var numStr string
	var unit string

	for i, ch := range sizeStr {
		if ch >= '0' && ch <= '9' || ch == '.' {
			numStr += string(ch)
		} else {
			unit = sizeStr[i:]
			break
		}
	}

	if numStr == "" {
		return 0, fmt.Errorf("no numeric value found in '%s'", sizeStr)
	}

	// Parse the number
	var value float64
	if _, err := fmt.Sscanf(numStr, "%f", &value); err != nil {
		return 0, fmt.Errorf("invalid number '%s': %w", numStr, err)
	}

	if value <= 0 {
		return 0, fmt.Errorf("size must be positive, got %.2f", value)
	}

	// Parse the unit
	unit = strings.TrimSpace(unit)
	var multiplier int64

	switch unit {
	case "G", "GB":
		multiplier = 1024 * 1024 * 1024 // Gigabytes
	case "M", "MB":
		multiplier = 1024 * 1024 // Megabytes
	case "":
		// No unit specified - assume bytes
		return 0, fmt.Errorf("unit required (use 'G', 'GB', 'M', or 'MB')")
	default:
		return 0, fmt.Errorf("invalid unit '%s' (use 'G', 'GB', 'M', or 'MB')", unit)
	}

	result := int64(value * float64(multiplier))
	if result <= 0 {
		return 0, fmt.Errorf("calculated size overflow or invalid")
	}

	return result, nil
}

func generateAllPatches(versionMgr *version.Manager, versionsDir, newVersion, outputDir, customKeyFile, compression string, level int, verify, createExe, silent, crp bool, customMaxPartSize int64) {
	fmt.Printf("Generating patches for new version %s\n", newVersion)

	// Scan for existing versions
	entries, err := os.ReadDir(versionsDir)
	if err != nil {
		fmt.Printf("Error: failed to read versions directory: %v\n", err)
		os.Exit(1)
	}

	// Register new version
	newVersionPath := filepath.Join(versionsDir, newVersion)
	if !utils.FileExists(newVersionPath) {
		fmt.Printf("Error: new version directory not found: %s\n", newVersionPath)
		os.Exit(1)
	}

	// Determine key file to use
	keyFile, err := detectKeyFile(newVersionPath, customKeyFile)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	if keyFile == "" {
		fmt.Println("Error: could not find key file (program.exe, game.exe, app.exe, or main.exe)")
		fmt.Println("Hint: Use --key-file to specify a custom key file")
		os.Exit(1)
	}
	fmt.Printf("Using key file: %s\n", keyFile)

	toVer, err := versionMgr.RegisterVersion(newVersion, newVersionPath, keyFile)
	if err != nil {
		fmt.Printf("Error: failed to register new version: %v\n", err)
		os.Exit(1)
	}

	// Generate patches from each existing version
	patchCount := 0
	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == newVersion {
			continue
		}

		fromVersion := entry.Name()
		fromPath := filepath.Join(versionsDir, fromVersion)

		fmt.Printf("\nProcessing version %s...\n", fromVersion)

		// Auto-detect key file for this source version (may differ from target)
		fromKeyFile, err := detectKeyFile(fromPath, customKeyFile)
		if err != nil {
			fmt.Printf("Warning: skipping %s - %v\n", fromVersion, err)
			continue
		}
		if fromKeyFile == "" {
			fmt.Printf("Warning: skipping %s - no key file found (tried: program.exe, game.exe, app.exe, main.exe)\n", fromVersion)
			continue
		}

		fromVer, err := versionMgr.RegisterVersion(fromVersion, fromPath, fromKeyFile)
		if err != nil {
			fmt.Printf("Warning: failed to register version %s: %v\n", fromVersion, err)
			continue
		}

		// Generate patch (with reverse if requested)
		patchFile := filepath.Join(outputDir, fmt.Sprintf("%s-to-%s.patch", fromVersion, newVersion))

		if crp {
			// Generate both patches efficiently using the same scan data
			reversePatchFile := filepath.Join(outputDir, fmt.Sprintf("%s-to-%s_rev.patch", newVersion, fromVersion))
			if err := generatePatchWithReverse(fromVer, toVer, patchFile, reversePatchFile, compression, level, verify, customMaxPartSize); err != nil {
				fmt.Printf("Error: failed to generate patches from %s: %v\n", fromVersion, err)
				continue
			}

			// Create forward exe if requested
			if createExe {
				exePath := filepath.Join(outputDir, fmt.Sprintf("%s-to-%s.exe", fromVersion, newVersion))
				if err := createStandaloneCLIExe(resolvePatchFile(patchFile), exePath, compression, silent); err != nil {
					fmt.Printf("Warning: failed to create forward executable for %s: %v\n", fromVersion, err)
				} else {
					fmt.Printf("✓ Forward executable: %s\n", exePath)
				}

				// Create reverse exe
				reverseExePath := filepath.Join(outputDir, fmt.Sprintf("%s-to-%s_rev.exe", newVersion, fromVersion))
				if err := createStandaloneCLIExe(resolvePatchFile(reversePatchFile), reverseExePath, compression, silent); err != nil {
					fmt.Printf("Warning: failed to create reverse executable to %s: %v\n", fromVersion, err)
				} else {
					fmt.Printf("✓ Reverse executable: %s\n", reverseExePath)
				}
			}

			patchCount += 2 // Count both patches
		} else {
			// Generate only forward patch
			if err := generatePatch(fromVer, toVer, patchFile, compression, level, createExe, silent, customMaxPartSize); err != nil {
				fmt.Printf("Error: failed to generate patch from %s: %v\n", fromVersion, err)
				continue
			}

			patchCount++
		}
	}

	fmt.Printf("\nSuccessfully generated %d patches\n", patchCount)
}

func generateSinglePatch(versionMgr *version.Manager, versionsDir, from, to, outputDir, customKeyFile, compression string, level int, verify, createExe, silent, crp bool, customMaxPartSize int64) {
	fmt.Printf("Generating patch from %s to %s\n", from, to)

	// Determine key file for FROM version
	fromPath := filepath.Join(versionsDir, from)
	fromKeyFile, err := detectKeyFile(fromPath, customKeyFile)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	if fromKeyFile == "" {
		fmt.Println("Error: could not find key file in source version (program.exe, game.exe, app.exe, or main.exe)")
		fmt.Println("Hint: Use --key-file to specify a custom key file")
		os.Exit(1)
	}
	fmt.Printf("Auto-detected source key file: %s\n", fromKeyFile)

	// Register source version
	fromVer, err := versionMgr.RegisterVersion(from, fromPath, fromKeyFile)
	if err != nil {
		fmt.Printf("Error: failed to register source version: %v\n", err)
		os.Exit(1)
	}

	// Determine key file for TO version (may differ from source)
	toPath := filepath.Join(versionsDir, to)
	toKeyFile, err := detectKeyFile(toPath, customKeyFile)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	if toKeyFile == "" {
		fmt.Println("Error: could not find key file in target version (program.exe, game.exe, app.exe, or main.exe)")
		fmt.Println("Hint: Use --key-file to specify a custom key file")
		os.Exit(1)
	}
	fmt.Printf("Auto-detected target key file: %s\n", toKeyFile)

	toVer, err := versionMgr.RegisterVersion(to, toPath, toKeyFile)
	if err != nil {
		fmt.Printf("Error: failed to register target version: %v\n", err)
		os.Exit(1)
	}

	// Generate patch (with reverse if requested)
	patchFile := filepath.Join(outputDir, fmt.Sprintf("%s-to-%s.patch", from, to))

	if crp {
		// Generate both patches efficiently using the same scan data
		fmt.Printf("\nGenerating forward and reverse patches...\n")
		reversePatchFile := filepath.Join(outputDir, fmt.Sprintf("%s-to-%s_rev.patch", to, from))
		if err := generatePatchWithReverse(fromVer, toVer, patchFile, reversePatchFile, compression, level, verify, customMaxPartSize); err != nil {
			fmt.Printf("Error: failed to generate patches: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("\n✓ Forward and reverse patches generated successfully")

		// Create executables if requested
		if createExe {
			exePath := filepath.Join(outputDir, fmt.Sprintf("%s-to-%s.exe", from, to))
			if err := createStandaloneCLIExe(resolvePatchFile(patchFile), exePath, compression, silent); err != nil {
				fmt.Printf("Error: failed to create forward executable: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("✓ Created forward executable: %s\n", exePath)

			reverseExePath := filepath.Join(outputDir, fmt.Sprintf("%s-to-%s_rev.exe", to, from))
			if err := createStandaloneCLIExe(resolvePatchFile(reversePatchFile), reverseExePath, compression, silent); err != nil {
				fmt.Printf("Error: failed to create reverse executable: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("✓ Created reverse executable: %s\n", reverseExePath)
		}
	} else {
		// Generate only forward patch
		if err := generatePatch(fromVer, toVer, patchFile, compression, level, createExe, silent, customMaxPartSize); err != nil {
			fmt.Printf("Error: failed to generate patch: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Patch generated successfully")
	}
}

// generateSinglePatchCustomPaths generates a patch using custom directory paths
// This allows versions to be on different drives or network locations
func generateSinglePatchCustomPaths(versionMgr *version.Manager, fromPath, toPath, outputDir, customKeyFile, compression string, level int, verify, createExe, silent, crp bool, customMaxPartSize int64) {
	// Extract version numbers from directory names
	fromVersion := extractVersionFromPath(fromPath)
	toVersion := extractVersionFromPath(toPath)

	fmt.Printf("Generating patch from %s (%s) to %s (%s)...\n", fromVersion, fromPath, toVersion, toPath)

	// Determine key file for FROM version
	fromKeyFile, err := detectKeyFile(fromPath, customKeyFile)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	if fromKeyFile == "" {
		fmt.Println("Error: could not find key file in source directory (program.exe, game.exe, app.exe, or main.exe)")
		fmt.Println("Hint: Use --key-file to specify a custom key file")
		os.Exit(1)
	}
	fmt.Printf("Auto-detected source key file: %s\n", fromKeyFile)

	// Register source version
	fmt.Printf("Registering source version %s...\n", fromVersion)
	fromVer, err := versionMgr.RegisterVersion(fromVersion, fromPath, fromKeyFile)
	if err != nil {
		fmt.Printf("Error: failed to register source version: %v\n", err)
		os.Exit(1)
	}

	// Determine key file for TO version (may differ from source)
	toKeyFile, err := detectKeyFile(toPath, customKeyFile)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	if toKeyFile == "" {
		fmt.Println("Error: could not find key file in target directory (program.exe, game.exe, app.exe, or main.exe)")
		fmt.Println("Hint: Use --key-file to specify a custom key file")
		os.Exit(1)
	}
	fmt.Printf("Auto-detected target key file: %s\n", toKeyFile)

	// Register target version
	// If from/to have the same directory name, unregister source first to avoid collision
	if fromVersion == toVersion {
		versionMgr.UnregisterVersion(fromVersion)
	}
	fmt.Printf("Registering target version %s...\n", toVersion)
	toVer, err := versionMgr.RegisterVersion(toVersion, toPath, toKeyFile)
	if err != nil {
		fmt.Printf("Error: failed to register target version: %v\n", err)
		os.Exit(1)
	}

	// Generate patch (with reverse if requested)
	patchFile := filepath.Join(outputDir, fmt.Sprintf("%s-to-%s.patch", fromVersion, toVersion))

	if crp {
		// Generate both patches efficiently using the same scan data
		fmt.Printf("\nGenerating forward and reverse patches...\n")
		reversePatchFile := filepath.Join(outputDir, fmt.Sprintf("%s-to-%s_rev.patch", toVersion, fromVersion))
		if err := generatePatchWithReverse(fromVer, toVer, patchFile, reversePatchFile, compression, level, verify, customMaxPartSize); err != nil {
			fmt.Printf("Error: failed to generate patches: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("\n✓ Forward and reverse patches generated successfully")

		// Create executables if requested
		if createExe {
			exePath := filepath.Join(outputDir, fmt.Sprintf("%s-to-%s.exe", fromVersion, toVersion))
			if err := createStandaloneCLIExe(resolvePatchFile(patchFile), exePath, compression, silent); err != nil {
				fmt.Printf("Error: failed to create forward executable: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("✓ Created forward executable: %s\n", exePath)

			reverseExePath := filepath.Join(outputDir, fmt.Sprintf("%s-to-%s_rev.exe", toVersion, fromVersion))
			if err := createStandaloneCLIExe(resolvePatchFile(reversePatchFile), reverseExePath, compression, silent); err != nil {
				fmt.Printf("Error: failed to create reverse executable: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("✓ Created reverse executable: %s\n", reverseExePath)
		}
	} else {
		// Generate only forward patch
		if err := generatePatch(fromVer, toVer, patchFile, compression, level, createExe, silent, customMaxPartSize); err != nil {
			fmt.Printf("Error: failed to generate patch: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Patch generated successfully: %s\n", patchFile)
	}
}

// extractVersionFromPath extracts the version number from a directory path
// Example: "C:\\releases\\1.0.0" -> "1.0.0"
// Example: "/mnt/versions/v2.1.5" -> "v2.1.5"
func extractVersionFromPath(path string) string {
	// Get the directory name (last component of the path)
	return filepath.Base(path)
}

func generatePatch(fromVer, toVer *utils.Version, outputFile, compression string, level int, createExe, silent bool, customMaxPartSize int64) error {
	// Create patch options
	options := &utils.PatchOptions{
		Compression:      compression,
		CompressionLevel: level,
		SkipIdentical:    true,
	}

	// Generate patch
	generator := patcher.NewGenerator()
	patch, err := generator.GeneratePatch(fromVer, toVer, options)
	if err != nil {
		return err
	}

	// Validate patch
	if err := generator.ValidatePatch(patch); err != nil {
		return fmt.Errorf("patch validation failed: %w", err)
	}

	// Check if patch needs to be split into multiple parts
	totalSize := generator.CalculatePatchSize(patch)
	var maxPartSize int64 = utils.DefaultMaxPartSize
	if customMaxPartSize > 0 {
		maxPartSize = customMaxPartSize
	}

	if totalSize > maxPartSize {
		fmt.Printf("\nPatch size (%d bytes / %.2f GB) exceeds %.2f GB limit, splitting into multiple parts...\n",
			totalSize, float64(totalSize)/(1024*1024*1024), float64(maxPartSize)/(1024*1024*1024))

		// Split patch into parts
		parts, err := generator.SplitPatchIntoParts(patch, maxPartSize)
		if err != nil {
			return fmt.Errorf("failed to split patch: %w", err)
		}

		// Save multi-part patch (pass chunk size for additional per-part chunking)
		if err := generator.SaveMultiPartPatch(parts, outputFile, compression, customMaxPartSize, level); err != nil {
			return fmt.Errorf("failed to save multi-part patch: %w", err)
		}

		fmt.Printf("✓ Multi-part patch saved: %d parts\n", len(parts))

		// If create-exe flag is set, check if part 01 can be turned into an exe
		if createExe {
			// Construct part 01 filename
			part01File := strings.TrimSuffix(outputFile, ".patch") + ".01.patch"

			// Check if part 01 exists and get its size
			if fileInfo, err := os.Stat(part01File); err == nil {
				part01Size := fileInfo.Size()
				const maxExeSize int64 = 3*1024*1024*1024 + 768*1024*1024 // 3.75 GB

				if part01Size < maxExeSize {
					fmt.Printf("\n✓ Part 01 size (%.2f GB) is under 3.75 GB limit, creating self-contained executable...\n",
						float64(part01Size)/(1024*1024*1024))

					// Extract version info from output filename
					baseName := filepath.Base(outputFile)
					exeName := strings.TrimSuffix(baseName, ".patch") + ".exe"
					exePath := filepath.Join(filepath.Dir(outputFile), exeName)

					if err := createStandaloneCLIExe(part01File, exePath, compression, silent); err != nil {
						fmt.Printf("Warning: failed to create executable from part 01: %v\n", err)
					} else {
						fmt.Printf("✓ Created self-contained executable from part 01: %s\n", exePath)
						fmt.Printf("  Note: This exe will automatically detect and use remaining parts (.02, .03, etc.)\n")
						fmt.Printf("  You can distribute: (1) exe + remaining parts, or (2) all parts together\n")
					}
				} else {
					fmt.Printf("\nℹ Part 01 size (%.2f GB) exceeds 3.75 GB Windows exe limit, skipping exe creation\n",
						float64(part01Size)/(1024*1024*1024))
					fmt.Printf("  Distribute all %d parts together (.01, .02, .03, etc.)\n", len(parts))
				}
			}
		}
	} else {
		// Save single-part patch
		if err := savePatch(patch, outputFile, options); err != nil {
			return fmt.Errorf("failed to save patch: %w", err)
		}

		fmt.Printf("Patch saved to: %s\n", outputFile)

		// Create self-contained executable if requested
		if createExe {
			baseName := filepath.Base(outputFile)
			exeName := strings.TrimSuffix(baseName, ".patch") + ".exe"
			exePath := filepath.Join(filepath.Dir(outputFile), exeName)

			if err := createStandaloneCLIExe(outputFile, exePath, compression, silent); err != nil {
				return fmt.Errorf("failed to create executable: %w", err)
			}
			fmt.Printf("✓ Created executable: %s\n", exePath)
		}
	}

	return nil
}

// generatePatchWithReverse generates both forward and reverse patches efficiently
// by reusing the same generator and scan data (no need to rescan directories)
func generatePatchWithReverse(fromVer, toVer *utils.Version, forwardFile, reverseFile, compression string, level int, verify bool, customMaxPartSize int64) error {
	// Create patch options
	options := &utils.PatchOptions{
		Compression:      compression,
		CompressionLevel: level,
		SkipIdentical:    true,
	}

	// Generate forward patch (from → to)
	generator := patcher.NewGenerator()
	forwardPatch, err := generator.GeneratePatch(fromVer, toVer, options)
	if err != nil {
		return fmt.Errorf("forward patch generation failed: %w", err)
	}

	// Validate forward patch
	if err := generator.ValidatePatch(forwardPatch); err != nil {
		return fmt.Errorf("forward patch validation failed: %w", err)
	}

	// Save forward patch (auto-splits into multi-part if needed)
	if err := savePatchWithSplitting(generator, forwardPatch, forwardFile, compression, level, customMaxPartSize); err != nil {
		return fmt.Errorf("failed to save forward patch: %w", err)
	}

	// Generate reverse patch (to → from) - REUSES SAME SCAN DATA!
	// No need to rescan directories since we already have all the data
	fmt.Printf("Generating reverse patch (reusing scan data)...\n")
	reversePatch, err := generator.GeneratePatch(toVer, fromVer, options)
	if err != nil {
		return fmt.Errorf("reverse patch generation failed: %w", err)
	}

	// Validate reverse patch
	if err := generator.ValidatePatch(reversePatch); err != nil {
		return fmt.Errorf("reverse patch validation failed: %w", err)
	}

	// Save reverse patch (auto-splits into multi-part if needed)
	if err := savePatchWithSplitting(generator, reversePatch, reverseFile, compression, level, customMaxPartSize); err != nil {
		return fmt.Errorf("failed to save reverse patch: %w", err)
	}

	return nil
}

func savePatch(patch *utils.Patch, filename string, options *utils.PatchOptions) error {
	return savePatchWithCustomSize(patch, filename, options, 0)
}

// resolvePatchFile checks if the patch was saved as multi-part (.01.patch) and returns
// the correct path. This is needed because savePatchWithSplitting may save as .01.patch
// instead of .patch when the patch was split, and the callers still hold the .patch path.
func resolvePatchFile(patchPath string) string {
	part01 := strings.TrimSuffix(patchPath, ".patch") + ".01.patch"
	if utils.FileExists(part01) {
		return part01
	}
	return patchPath
}

func savePatchWithCustomSize(patch *utils.Patch, filename string, options *utils.PatchOptions, customMaxPartSize int64) error {
	// customMaxPartSize is unused in the save path itself — it's consumed by the
	// caller (generatePatch) for multi-part splitting decisions before this is called.
	return utils.SavePatch(patch, filename, options.Compression, options.CompressionLevel)
}

// savePatchWithSplitting saves a patch, splitting into multi-part if total size exceeds maxPartSize.
func savePatchWithSplitting(generator *patcher.Generator, patch *utils.Patch, outputFile, compression string, level int, customMaxPartSize int64) error {
	var maxPartSize int64 = utils.DefaultMaxPartSize
	if customMaxPartSize > 0 {
		maxPartSize = customMaxPartSize
	}

	totalSize := generator.CalculatePatchSize(patch)
	if totalSize > maxPartSize {
		fmt.Printf("\nPatch size (%d bytes / %.2f GB) exceeds %.2f GB limit, splitting into multiple parts...\n",
			totalSize, float64(totalSize)/(1024*1024*1024), float64(maxPartSize)/(1024*1024*1024))

		parts, err := generator.SplitPatchIntoParts(patch, maxPartSize)
		if err != nil {
			return fmt.Errorf("failed to split patch: %w", err)
		}

		if err := generator.SaveMultiPartPatch(parts, outputFile, compression, customMaxPartSize, level); err != nil {
			return fmt.Errorf("failed to save multi-part patch: %w", err)
		}
		fmt.Printf("✓ Multi-part patch saved: %d parts\n", len(parts))
	} else {
		if err := utils.SavePatch(patch, outputFile, compression, level); err != nil {
			return fmt.Errorf("failed to save patch: %w", err)
		}
		fmt.Printf("Patch saved to: %s\n", outputFile)
	}
	return nil
}

// createStandaloneCLIExe creates a self-contained CLI executable by appending patch data to CLI applier
func createStandaloneCLIExe(patchPath, exePath, compression string, silent bool) error {
	// Get path to patch-apply.exe (CLI applier)
	genExe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Look for patch-apply.exe in the same directory
	applierPath := filepath.Join(filepath.Dir(genExe), "patch-apply.exe")
	if !utils.FileExists(applierPath) {
		return fmt.Errorf("CLI applier not found: %s", applierPath)
	}

	// Read the CLI applier executable
	applierData, err := os.ReadFile(applierPath)
	if err != nil {
		return fmt.Errorf("failed to read applier executable: %w", err)
	}

	// Read the patch file
	patchData, err := os.ReadFile(patchPath)
	if err != nil {
		return fmt.Errorf("failed to read patch file: %w", err)
	}

	// Look for any chunk sidecar JSON files in the same directory and build a sidecar blob.
	// Sidecar filename pattern: <base>.part<N>.chunks.json
	sidecarBlob := []byte{}
	sidecarFiles := []string{}
	base := filepath.Base(patchPath)
	if strings.HasSuffix(base, ".01.patch") {
		base = strings.TrimSuffix(base, ".01.patch")
	} else {
		base = strings.TrimSuffix(base, ".patch")
	}
	// Glob for sidecars
	globPattern := filepath.Join(filepath.Dir(patchPath), base+".part*.chunks.json")
	matches, _ := filepath.Glob(globPattern)
	if len(matches) > 0 {
		// We'll encode an index followed by pairs of (nameLen, name, size, data)
		var idxBuf []byte
		// Use a bytes buffer-like composition without extra import
		for _, f := range matches {
			data, err := os.ReadFile(f)
			if err != nil {
				return fmt.Errorf("failed to read sidecar %s: %w", f, err)
			}
			relName := filepath.Base(f)
			nameBytes := []byte(relName)
			// name length (uint16) + name + size (uint64) + data
			nb := make([]byte, 2)
			binary.LittleEndian.PutUint16(nb, uint16(len(nameBytes)))
			idxBuf = append(idxBuf, nb...)
			idxBuf = append(idxBuf, nameBytes...)
			sb := make([]byte, 8)
			binary.LittleEndian.PutUint64(sb, uint64(len(data)))
			idxBuf = append(idxBuf, sb...)
			idxBuf = append(idxBuf, data...)
			sidecarFiles = append(sidecarFiles, relName)
		}
		// Prepend total count (uint32)
		cnt := make([]byte, 4)
		binary.LittleEndian.PutUint32(cnt, uint32(len(matches)))
		sidecarBlob = append(cnt, idxBuf...)
	}

	// Calculate checksum of patch data (as bytes, not hex string)
	checksum := sha256.Sum256(patchData)

	// Create 128-byte header
	header := make([]byte, 128)

	// Magic bytes "CPMPATCH" (8 bytes)
	copy(header[0:8], []byte("CPMPATCH"))

	// Version (4 bytes, uint32)
	var version uint32 = 1
	header[8] = byte(version)
	header[9] = byte(version >> 8)
	header[10] = byte(version >> 16)
	header[11] = byte(version >> 24)

	// Stub size (8 bytes, uint64) - size of applier exe
	stubSize := uint64(len(applierData))
	for i := 0; i < 8; i++ {
		header[12+i] = byte(stubSize >> (i * 8))
	}

	// Data offset (8 bytes, uint64) - right after stub
	dataOffset := uint64(len(applierData))
	for i := 0; i < 8; i++ {
		header[20+i] = byte(dataOffset >> (i * 8))
	}

	// Data size (8 bytes, uint64)
	dataSize := uint64(len(patchData))
	for i := 0; i < 8; i++ {
		header[28+i] = byte(dataSize >> (i * 8))
	}

	// Compression type (16 bytes)
	compressionBytes := make([]byte, 16)
	copy(compressionBytes, []byte(compression))
	copy(header[36:52], compressionBytes)

	// Checksum (32 bytes, SHA-256)
	copy(header[52:84], checksum[:])

	// Flags (1 byte) - bit 0: silent mode
	var flags byte = 0
	if silent {
		flags |= 0x01 // Set bit 0 for silent mode
	}
	header[84] = flags

	// Reserved (43 bytes) - already zeroed

	// Create output file
	outFile, err := os.Create(exePath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	// Write: applier.exe + patch data + optional sidecar blob + header
	if _, err := outFile.Write(applierData); err != nil {
		return fmt.Errorf("failed to write applier data: %w", err)
	}

	if _, err := outFile.Write(patchData); err != nil {
		return fmt.Errorf("failed to write patch data: %w", err)
	}

	// Write sidecar blob (may be empty)
	if len(sidecarBlob) > 0 {
		if _, err := outFile.Write(sidecarBlob); err != nil {
			return fmt.Errorf("failed to write sidecar blob: %w", err)
		}
	}

	if _, err := outFile.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	return nil
}

// encodePatchStreaming writes the patch as JSON in a streaming fashion to avoid memory exhaustion
func printHelp() {
	fmt.Printf("CyberPatchMaker - Patch Generator v%s\n", version.GetVersion())
	fmt.Println("\nUsage:")
	fmt.Println("  Generate patches from all versions to new version:")
	fmt.Println("    patch-gen --versions-dir <dir> --new-version <version>")
	fmt.Println("\n  Generate single patch (versions in same directory):")
	fmt.Println("    patch-gen --versions-dir <dir> --from <version> --to <version>")
	fmt.Println("\n  Generate single patch (custom paths, different drives/locations):")
	fmt.Println("    patch-gen --from-dir <path> --to-dir <path>")
	fmt.Println("\nOptions:")
	fmt.Println("  --versions-dir    Directory containing version folders")
	fmt.Println("  --new-version     New version number to generate patches for")
	fmt.Println("  --from            Source version number (with --versions-dir)")
	fmt.Println("  --to              Target version number (with --versions-dir)")
	fmt.Println("  --from-dir        Full path to source version directory")
	fmt.Println("  --to-dir          Full path to target version directory")
	fmt.Println("  --output          Output directory for patches (default: patches)")
	fmt.Println("  --key-file        Specific key file to use (e.g., app_name.exe)")
	fmt.Println("  --compression     Compression algorithm: zstd, gzip, none (default: zstd)")
	fmt.Println("  --level           Compression level (default: 3)")
	fmt.Println("  --verify          Verify patches after creation (default: true)")
	fmt.Println("  --create-exe      Create self-contained CLI executable")
	fmt.Println("  --silent          Enable silent mode in generated executable (auto-apply without prompts)")
	fmt.Println("  --crp             Create reverse patch (for downgrades)")
	fmt.Println("  --savescans       Save directory scans to cache for faster subsequent patches")
	fmt.Println("  --rescan          Force rescan of cached versions (use with --savescans)")
	fmt.Println("  --scandata        Custom directory for scan cache (default: .data)")
	fmt.Println("  --jobs            Number of parallel workers (0=auto-detect CPU cores, 1=single-threaded, default: 0)")
	fmt.Println("  --splitsize       Custom multi-part split size (e.g., '2G', '2GB', '500M', '500MB', default: 4GB)")
	fmt.Println("  --bypasssplitlimit Bypass 100MB minimum split size confirmation")
	fmt.Println("  --version         Show version information")
	fmt.Println("  --help            Show this help message")
	fmt.Println("\nExamples:")
	fmt.Println("  # Versions on different drives")
	fmt.Println("  patch-gen --from-dir C:\\releases\\1.0.0 --to-dir D:\\builds\\1.0.1 --output patches")
	fmt.Println("\n  # Create self-contained executable")
	fmt.Println("  patch-gen --from-dir C:\\\\v1 --to-dir C:\\\\v2 --output patches --create-exe")
	fmt.Println("\n  # Create forward and reverse patches with executables")
	fmt.Println("  patch-gen --from-dir C:\\\\v1.0.0 --to-dir C:\\\\v1.0.1 --output patches --crp --create-exe")
	fmt.Println("\\n  # Use scan caching for faster subsequent patches")
	fmt.Println("  patch-gen --versions-dir C:\\\\versions --from 1.0.0 --to 1.0.1 --output patches --savescans")
	fmt.Println("  patch-gen --versions-dir C:\\\\versions --from 1.0.1 --to 1.0.2 --output patches --savescans")
	fmt.Println("\\n  # Use parallel workers for faster processing (large projects)")
	fmt.Println("  patch-gen --from-dir C:\\\\v1 --to-dir C:\\\\v2 --output patches --jobs 0")
	fmt.Println("  patch-gen --from-dir C:\\\\v1 --to-dir C:\\\\v2 --output patches --jobs 8")
	fmt.Println("\\n  # Force rescan of cached versions")
	fmt.Println("  patch-gen --versions-dir C:\\\\versions --from 1.0.0 --to 1.0.1 --output patches --savescans --rescan")
	fmt.Println("\\n  # Custom split size for multi-part patches")
	fmt.Println("  patch-gen --from-dir C:\\\\v1 --to-dir C:\\\\v2 --output patches --splitsize 2G")
	fmt.Println("  patch-gen --from-dir C:\\\\v1 --to-dir C:\\\\v2 --output patches --splitsize 500MB")
	fmt.Println("\\n  # Small split size (below 100MB) with bypass")
	fmt.Println("  patch-gen --from-dir C:\\\\v1 --to-dir C:\\\\v2 --output patches --splitsize 50M --bypasssplitlimit")
	fmt.Println("\n  # Versions on different network locations")
	fmt.Println("  patch-gen --from-dir \\\\\\\\server1\\\\app\\\\v1 --to-dir \\\\\\\\server2\\\\app\\\\v2 --output .")
}
