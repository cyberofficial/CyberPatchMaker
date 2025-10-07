package main

import (
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

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
	level := flag.Int("level", 3, "Compression level (1-4 for zstd, 1-9 for gzip)")
	verify := flag.Bool("verify", true, "Verify patches after creation")
	createExe := flag.Bool("create-exe", false, "Create self-contained CLI executable")
	crp := flag.Bool("crp", false, "Create reverse patch (for downgrades)")
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
		generateAllPatches(versionMgr, *versionsDir, *newVersion, outputDir, *keyFile, *compression, *level, *verify, *createExe, *crp)
	} else if *fromDir != "" && *toDir != "" {
		// Generate single patch using custom directory paths
		generateSinglePatchCustomPaths(versionMgr, *fromDir, *toDir, outputDir, *keyFile, *compression, *level, *verify, *createExe, *crp)
	} else if *from != "" && *to != "" && *versionsDir != "" {
		// Generate single patch using versions-dir
		generateSinglePatch(versionMgr, *versionsDir, *from, *to, outputDir, *keyFile, *compression, *level, *verify, *createExe, *crp)
	} else {
		fmt.Println("Error: insufficient arguments")
		printHelp()
		os.Exit(1)
	}
}

func generateAllPatches(versionMgr *version.Manager, versionsDir, newVersion, outputDir, customKeyFile, compression string, level int, verify, createExe, crp bool) {
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
	var keyFile string
	if customKeyFile != "" {
		// Use custom key file if provided
		if utils.FileExists(filepath.Join(newVersionPath, customKeyFile)) {
			keyFile = customKeyFile
			fmt.Printf("Using custom key file: %s\n", keyFile)
		} else {
			fmt.Printf("Error: custom key file not found: %s\n", customKeyFile)
			os.Exit(1)
		}
	} else {
		// Auto-detect key file from common names
		keyFiles := []string{"program.exe", "game.exe", "app.exe", "main.exe"}
		for _, kf := range keyFiles {
			if utils.FileExists(filepath.Join(newVersionPath, kf)) {
				keyFile = kf
				break
			}
		}
		if keyFile == "" {
			fmt.Println("Error: could not find key file (program.exe, game.exe, app.exe, or main.exe)")
			fmt.Println("Hint: Use --key-file to specify a custom key file")
			os.Exit(1)
		}
		fmt.Printf("Auto-detected key file: %s\n", keyFile)
	}

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

		fromVer, err := versionMgr.RegisterVersion(fromVersion, fromPath, keyFile)
		if err != nil {
			fmt.Printf("Warning: failed to register version %s: %v\n", fromVersion, err)
			continue
		}

		// Generate patch (with reverse if requested)
		patchFile := filepath.Join(outputDir, fmt.Sprintf("%s-to-%s.patch", fromVersion, newVersion))

		if crp {
			// Generate both patches efficiently using the same scan data
			reversePatchFile := filepath.Join(outputDir, fmt.Sprintf("%s-to-%s.patch", newVersion, fromVersion))
			if err := generatePatchWithReverse(fromVer, toVer, patchFile, reversePatchFile, compression, level, verify); err != nil {
				fmt.Printf("Error: failed to generate patches from %s: %v\n", fromVersion, err)
				continue
			}

			// Create forward exe if requested
			if createExe {
				exePath := filepath.Join(outputDir, fmt.Sprintf("%s-to-%s.exe", fromVersion, newVersion))
				if err := createStandaloneCLIExe(patchFile, exePath, compression); err != nil {
					fmt.Printf("Warning: failed to create forward executable for %s: %v\n", fromVersion, err)
				} else {
					fmt.Printf("✓ Forward executable: %s\n", exePath)
				}

				// Create reverse exe
				reverseExePath := filepath.Join(outputDir, fmt.Sprintf("%s-to-%s.exe", newVersion, fromVersion))
				if err := createStandaloneCLIExe(reversePatchFile, reverseExePath, compression); err != nil {
					fmt.Printf("Warning: failed to create reverse executable to %s: %v\n", fromVersion, err)
				} else {
					fmt.Printf("✓ Reverse executable: %s\n", reverseExePath)
				}
			}

			patchCount += 2 // Count both patches
		} else {
			// Generate only forward patch
			if err := generatePatch(fromVer, toVer, patchFile, compression, level, verify); err != nil {
				fmt.Printf("Error: failed to generate patch from %s: %v\n", fromVersion, err)
				continue
			}

			// Create self-contained executable if requested
			if createExe {
				exePath := filepath.Join(outputDir, fmt.Sprintf("%s-to-%s.exe", fromVersion, newVersion))
				if err := createStandaloneCLIExe(patchFile, exePath, compression); err != nil {
					fmt.Printf("Warning: failed to create executable for %s: %v\n", fromVersion, err)
				} else {
					fmt.Printf("Created executable: %s\n", exePath)
				}
			}

			patchCount++
		}
	}

	fmt.Printf("\nSuccessfully generated %d patches\n", patchCount)
}

func generateSinglePatch(versionMgr *version.Manager, versionsDir, from, to, outputDir, customKeyFile, compression string, level int, verify, createExe, crp bool) {
	fmt.Printf("Generating patch from %s to %s\n", from, to)

	// Determine key file to use
	fromPath := filepath.Join(versionsDir, from)
	var keyFile string
	if customKeyFile != "" {
		// Use custom key file if provided
		if utils.FileExists(filepath.Join(fromPath, customKeyFile)) {
			keyFile = customKeyFile
			fmt.Printf("Using custom key file: %s\n", keyFile)
		} else {
			fmt.Printf("Error: custom key file not found: %s\n", customKeyFile)
			os.Exit(1)
		}
	} else {
		// Auto-detect key file from common names
		keyFiles := []string{"program.exe", "game.exe", "app.exe", "main.exe"}
		for _, kf := range keyFiles {
			if utils.FileExists(filepath.Join(fromPath, kf)) {
				keyFile = kf
				break
			}
		}
		if keyFile == "" {
			fmt.Println("Error: could not find key file (program.exe, game.exe, app.exe, or main.exe)")
			fmt.Println("Hint: Use --key-file to specify a custom key file")
			os.Exit(1)
		}
		fmt.Printf("Auto-detected key file: %s\n", keyFile)
	}

	// Register versions
	fromVer, err := versionMgr.RegisterVersion(from, fromPath, keyFile)
	if err != nil {
		fmt.Printf("Error: failed to register source version: %v\n", err)
		os.Exit(1)
	}

	toPath := filepath.Join(versionsDir, to)
	toVer, err := versionMgr.RegisterVersion(to, toPath, keyFile)
	if err != nil {
		fmt.Printf("Error: failed to register target version: %v\n", err)
		os.Exit(1)
	}

	// Generate patch (with reverse if requested)
	patchFile := filepath.Join(outputDir, fmt.Sprintf("%s-to-%s.patch", from, to))

	if crp {
		// Generate both patches efficiently using the same scan data
		fmt.Printf("\nGenerating forward and reverse patches...\n")
		reversePatchFile := filepath.Join(outputDir, fmt.Sprintf("%s-to-%s.patch", to, from))
		if err := generatePatchWithReverse(fromVer, toVer, patchFile, reversePatchFile, compression, level, verify); err != nil {
			fmt.Printf("Error: failed to generate patches: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("\n✓ Forward and reverse patches generated successfully")

		// Create executables if requested
		if createExe {
			exePath := filepath.Join(outputDir, fmt.Sprintf("%s-to-%s.exe", from, to))
			if err := createStandaloneCLIExe(patchFile, exePath, compression); err != nil {
				fmt.Printf("Error: failed to create forward executable: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("✓ Created forward executable: %s\n", exePath)

			reverseExePath := filepath.Join(outputDir, fmt.Sprintf("%s-to-%s.exe", to, from))
			if err := createStandaloneCLIExe(reversePatchFile, reverseExePath, compression); err != nil {
				fmt.Printf("Error: failed to create reverse executable: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("✓ Created reverse executable: %s\n", reverseExePath)
		}
	} else {
		// Generate only forward patch
		if err := generatePatch(fromVer, toVer, patchFile, compression, level, verify); err != nil {
			fmt.Printf("Error: failed to generate patch: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Patch generated successfully")

		// Create self-contained executable if requested
		if createExe {
			exePath := filepath.Join(outputDir, fmt.Sprintf("%s-to-%s.exe", from, to))
			if err := createStandaloneCLIExe(patchFile, exePath, compression); err != nil {
				fmt.Printf("Error: failed to create executable: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Created executable: %s\n", exePath)
		}
	}
}

// generateSinglePatchCustomPaths generates a patch using custom directory paths
// This allows versions to be on different drives or network locations
func generateSinglePatchCustomPaths(versionMgr *version.Manager, fromPath, toPath, outputDir, customKeyFile, compression string, level int, verify, createExe, crp bool) {
	// Extract version numbers from directory names
	fromVersion := extractVersionFromPath(fromPath)
	toVersion := extractVersionFromPath(toPath)

	fmt.Printf("Generating patch from %s (%s) to %s (%s)...\n", fromVersion, fromPath, toVersion, toPath)

	// Determine key file to use
	var keyFile string
	if customKeyFile != "" {
		// Use custom key file if provided
		if utils.FileExists(filepath.Join(fromPath, customKeyFile)) {
			keyFile = customKeyFile
			fmt.Printf("Using custom key file: %s\n", keyFile)
		} else {
			fmt.Printf("Error: custom key file not found: %s\n", customKeyFile)
			os.Exit(1)
		}
	} else {
		// Auto-detect key file from common names
		keyFiles := []string{"program.exe", "game.exe", "app.exe", "main.exe"}
		for _, kf := range keyFiles {
			if utils.FileExists(filepath.Join(fromPath, kf)) {
				keyFile = kf
				break
			}
		}
		if keyFile == "" {
			fmt.Println("Error: could not find key file in source directory (program.exe, game.exe, app.exe, or main.exe)")
			fmt.Println("Hint: Use --key-file to specify a custom key file")
			os.Exit(1)
		}
		fmt.Printf("Auto-detected key file: %s\n", keyFile)
	}

	// Register source version
	fmt.Printf("Registering source version %s...\n", fromVersion)
	fromVer, err := versionMgr.RegisterVersion(fromVersion, fromPath, keyFile)
	if err != nil {
		fmt.Printf("Error: failed to register source version: %v\n", err)
		os.Exit(1)
	}

	// Register target version (should use same key file)
	fmt.Printf("Registering target version %s...\n", toVersion)
	toVer, err := versionMgr.RegisterVersion(toVersion, toPath, keyFile)
	if err != nil {
		fmt.Printf("Error: failed to register target version: %v\n", err)
		os.Exit(1)
	}

	// Generate patch (with reverse if requested)
	patchFile := filepath.Join(outputDir, fmt.Sprintf("%s-to-%s.patch", fromVersion, toVersion))

	if crp {
		// Generate both patches efficiently using the same scan data
		fmt.Printf("\nGenerating forward and reverse patches...\n")
		reversePatchFile := filepath.Join(outputDir, fmt.Sprintf("%s-to-%s.patch", toVersion, fromVersion))
		if err := generatePatchWithReverse(fromVer, toVer, patchFile, reversePatchFile, compression, level, verify); err != nil {
			fmt.Printf("Error: failed to generate patches: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("\n✓ Forward and reverse patches generated successfully")

		// Create executables if requested
		if createExe {
			exePath := filepath.Join(outputDir, fmt.Sprintf("%s-to-%s.exe", fromVersion, toVersion))
			if err := createStandaloneCLIExe(patchFile, exePath, compression); err != nil {
				fmt.Printf("Error: failed to create forward executable: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("✓ Created forward executable: %s\n", exePath)

			reverseExePath := filepath.Join(outputDir, fmt.Sprintf("%s-to-%s.exe", toVersion, fromVersion))
			if err := createStandaloneCLIExe(reversePatchFile, reverseExePath, compression); err != nil {
				fmt.Printf("Error: failed to create reverse executable: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("✓ Created reverse executable: %s\n", reverseExePath)
		}
	} else {
		// Generate only forward patch
		if err := generatePatch(fromVer, toVer, patchFile, compression, level, verify); err != nil {
			fmt.Printf("Error: failed to generate patch: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Patch generated successfully: %s\n", patchFile)

		// Create self-contained executable if requested
		if createExe {
			exePath := filepath.Join(outputDir, fmt.Sprintf("%s-to-%s.exe", fromVersion, toVersion))
			if err := createStandaloneCLIExe(patchFile, exePath, compression); err != nil {
				fmt.Printf("Error: failed to create executable: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("✓ Created executable: %s\n", exePath)
		}
	}
}

// extractVersionFromPath extracts the version number from a directory path
// Example: "C:\\releases\\1.0.0" -> "1.0.0"
// Example: "/mnt/versions/v2.1.5" -> "v2.1.5"
func extractVersionFromPath(path string) string {
	// Get the directory name (last component of the path)
	return filepath.Base(path)
}

func generatePatch(fromVer, toVer *utils.Version, outputFile, compression string, level int, verify bool) error {
	// Create patch options
	options := &utils.PatchOptions{
		Compression:      compression,
		CompressionLevel: level,
		VerifyAfter:      verify,
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

	// Save patch to file
	if err := savePatch(patch, outputFile, options); err != nil {
		return fmt.Errorf("failed to save patch: %w", err)
	}

	fmt.Printf("Patch saved to: %s\n", outputFile)
	return nil
}

// generatePatchWithReverse generates both forward and reverse patches efficiently
// by reusing the same generator and scan data (no need to rescan directories)
func generatePatchWithReverse(fromVer, toVer *utils.Version, forwardFile, reverseFile, compression string, level int, verify bool) error {
	// Create patch options
	options := &utils.PatchOptions{
		Compression:      compression,
		CompressionLevel: level,
		VerifyAfter:      verify,
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

	// Save forward patch
	if err := savePatch(forwardPatch, forwardFile, options); err != nil {
		return fmt.Errorf("failed to save forward patch: %w", err)
	}
	fmt.Printf("Patch saved to: %s\n", forwardFile)

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

	// Save reverse patch
	if err := savePatch(reversePatch, reverseFile, options); err != nil {
		return fmt.Errorf("failed to save reverse patch: %w", err)
	}
	fmt.Printf("Reverse patch saved to: %s\n", reverseFile)

	return nil
}

func savePatch(patch *utils.Patch, filename string, options *utils.PatchOptions) error {
	// Marshal patch to JSON
	data, err := json.MarshalIndent(patch, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal patch: %w", err)
	}

	// Compress if needed
	if options.Compression != "none" && options.Compression != "" {
		compressedData, err := utils.CompressData(data, options.Compression, options.CompressionLevel)
		if err != nil {
			return fmt.Errorf("failed to compress patch: %w", err)
		}
		data = compressedData
	}

	// Calculate checksum
	checksum := utils.CalculateDataChecksum(data)
	patch.Header.Checksum = checksum

	// Write to file
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write patch file: %w", err)
	}

	return nil
}

// createStandaloneCLIExe creates a self-contained CLI executable by appending patch data to CLI applier
func createStandaloneCLIExe(patchPath, exePath, compression string) error {
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

	// Reserved (44 bytes) - already zeroed

	// Create output file
	outFile, err := os.Create(exePath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	// Write: applier.exe + patch data + header
	if _, err := outFile.Write(applierData); err != nil {
		return fmt.Errorf("failed to write applier data: %w", err)
	}

	if _, err := outFile.Write(patchData); err != nil {
		return fmt.Errorf("failed to write patch data: %w", err)
	}

	if _, err := outFile.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	return nil
}

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
	fmt.Println("  --crp             Create reverse patch (for downgrades)")
	fmt.Println("  --version         Show version information")
	fmt.Println("  --help            Show this help message")
	fmt.Println("\nExamples:")
	fmt.Println("  # Versions on different drives")
	fmt.Println("  patch-gen --from-dir C:\\releases\\1.0.0 --to-dir D:\\builds\\1.0.1 --output patches")
	fmt.Println("\n  # Create self-contained executable")
	fmt.Println("  patch-gen --from-dir C:\\\\v1 --to-dir C:\\\\v2 --output patches --create-exe")
	fmt.Println("\n  # Create forward and reverse patches with executables")
	fmt.Println("  patch-gen --from-dir C:\\\\v1.0.0 --to-dir C:\\\\v1.0.1 --output patches --crp --create-exe")
	fmt.Println("\n  # Versions on different network locations")
	fmt.Println("  patch-gen --from-dir \\\\\\\\server1\\\\app\\\\v1 --to-dir \\\\\\\\server2\\\\app\\\\v2 --output .")
}
