package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
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
	crp := flag.Bool("crp", false, "Create reverse patch (for downgrades)")
	// TODO: Implement silent mode flag support for CLI generator (currently only GUI supports this)
	_ = flag.Bool("silent-mode", false, "Enable simplified UI for end users (silent mode)")
	saveScans := flag.Bool("savescans", false, "Save directory scans to cache for faster subsequent patches")
	rescan := flag.Bool("rescan", false, "Force rescan of cached versions")
	scanData := flag.String("scandata", "", "Custom directory for scan cache (default: .data)")
	jobs := flag.Int("jobs", 0, "Number of parallel workers (0 = auto-detect CPU cores, 1 = single-threaded)")
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

		// Auto-detect key file for this source version (may differ from target)
		var fromKeyFile string
		if customKeyFile != "" {
			// Use custom key file if specified
			fromKeyFile = customKeyFile
		} else {
			// Auto-detect key file from common names for this version
			keyFiles := []string{"program.exe", "game.exe", "app.exe", "main.exe"}
			for _, kf := range keyFiles {
				if utils.FileExists(filepath.Join(fromPath, kf)) {
					fromKeyFile = kf
					break
				}
			}
			if fromKeyFile == "" {
				fmt.Printf("Warning: skipping %s - no key file found (tried: program.exe, game.exe, app.exe, main.exe)\n", fromVersion)
				continue
			}
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

	// Determine key file for FROM version
	fromPath := filepath.Join(versionsDir, from)
	var fromKeyFile string
	if customKeyFile != "" {
		// Use custom key file if provided
		fromKeyFile = customKeyFile
	} else {
		// Auto-detect key file from common names
		keyFiles := []string{"program.exe", "game.exe", "app.exe", "main.exe"}
		for _, kf := range keyFiles {
			if utils.FileExists(filepath.Join(fromPath, kf)) {
				fromKeyFile = kf
				break
			}
		}
		if fromKeyFile == "" {
			fmt.Println("Error: could not find key file in source version (program.exe, game.exe, app.exe, or main.exe)")
			fmt.Println("Hint: Use --key-file to specify a custom key file")
			os.Exit(1)
		}
		fmt.Printf("Auto-detected source key file: %s\n", fromKeyFile)
	}

	// Register source version
	fromVer, err := versionMgr.RegisterVersion(from, fromPath, fromKeyFile)
	if err != nil {
		fmt.Printf("Error: failed to register source version: %v\n", err)
		os.Exit(1)
	}

	// Determine key file for TO version (may differ from source)
	toPath := filepath.Join(versionsDir, to)
	var toKeyFile string
	if customKeyFile != "" {
		toKeyFile = customKeyFile
	} else {
		keyFiles := []string{"program.exe", "game.exe", "app.exe", "main.exe"}
		for _, kf := range keyFiles {
			if utils.FileExists(filepath.Join(toPath, kf)) {
				toKeyFile = kf
				break
			}
		}
		if toKeyFile == "" {
			fmt.Println("Error: could not find key file in target version (program.exe, game.exe, app.exe, or main.exe)")
			fmt.Println("Hint: Use --key-file to specify a custom key file")
			os.Exit(1)
		}
		fmt.Printf("Auto-detected target key file: %s\n", toKeyFile)
	}

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

	// Determine key file for FROM version
	var fromKeyFile string
	if customKeyFile != "" {
		// Use custom key file if provided
		fromKeyFile = customKeyFile
	} else {
		// Auto-detect key file from common names
		keyFiles := []string{"program.exe", "game.exe", "app.exe", "main.exe"}
		for _, kf := range keyFiles {
			if utils.FileExists(filepath.Join(fromPath, kf)) {
				fromKeyFile = kf
				break
			}
		}
		if fromKeyFile == "" {
			fmt.Println("Error: could not find key file in source directory (program.exe, game.exe, app.exe, or main.exe)")
			fmt.Println("Hint: Use --key-file to specify a custom key file")
			os.Exit(1)
		}
		fmt.Printf("Auto-detected source key file: %s\n", fromKeyFile)
	}

	// Register source version
	fmt.Printf("Registering source version %s...\n", fromVersion)
	fromVer, err := versionMgr.RegisterVersion(fromVersion, fromPath, fromKeyFile)
	if err != nil {
		fmt.Printf("Error: failed to register source version: %v\n", err)
		os.Exit(1)
	}

	// Determine key file for TO version (may differ from source)
	var toKeyFile string
	if customKeyFile != "" {
		toKeyFile = customKeyFile
	} else {
		keyFiles := []string{"program.exe", "game.exe", "app.exe", "main.exe"}
		for _, kf := range keyFiles {
			if utils.FileExists(filepath.Join(toPath, kf)) {
				toKeyFile = kf
				break
			}
		}
		if toKeyFile == "" {
			fmt.Println("Error: could not find key file in target directory (program.exe, game.exe, app.exe, or main.exe)")
			fmt.Println("Hint: Use --key-file to specify a custom key file")
			os.Exit(1)
		}
		fmt.Printf("Auto-detected target key file: %s\n", toKeyFile)
	}

	// Register target version
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

	// Check if patch needs to be split into multiple parts
	totalSize := generator.CalculatePatchSize(patch)
	if totalSize > utils.DefaultMaxPartSize {
		fmt.Printf("\nPatch size (%d bytes / %.2f GB) exceeds 4GB limit, splitting into multiple parts...\n",
			totalSize, float64(totalSize)/(1024*1024*1024))

		// Split patch into parts
		parts, err := generator.SplitPatchIntoParts(patch, utils.DefaultMaxPartSize)
		if err != nil {
			return fmt.Errorf("failed to split patch: %w", err)
		}

		// Save multi-part patch
		if err := generator.SaveMultiPartPatch(parts, outputFile, compression); err != nil {
			return fmt.Errorf("failed to save multi-part patch: %w", err)
		}

		fmt.Printf("✓ Multi-part patch saved: %d parts\n", len(parts))
	} else {
		// Save single-part patch
		if err := savePatch(patch, outputFile, options); err != nil {
			return fmt.Errorf("failed to save patch: %w", err)
		}

		fmt.Printf("Patch saved to: %s\n", outputFile)
	}

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
	// Create output file
	outFile, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create patch file: %w", err)
	}
	defer outFile.Close()

	// Create a pipe for streaming JSON encoding
	jsonReader, jsonWriter := io.Pipe()
	defer jsonReader.Close()

	// Start custom streaming JSON encoding in a goroutine
	encodeErr := make(chan error, 1)
	go func() {
		defer jsonWriter.Close()
		encodeErr <- encodePatchStreaming(patch, jsonWriter)
	}()

	// Set up compression if needed
	var finalReader io.Reader = jsonReader

	if options.Compression != "none" && options.Compression != "" {
		// Create a pipe for compression
		compressedReader, compressor := io.Pipe()

		// Start compression in a goroutine
		go func() {
			defer compressor.Close()
			err := utils.CompressDataStreaming(jsonReader, compressor, options.Compression, options.CompressionLevel)
			if err != nil {
				compressor.CloseWithError(err)
			}
		}()

		finalReader = compressedReader
	}

	// Create a hash writer to calculate checksum while writing
	hasher := sha256.New()
	multiWriter := io.MultiWriter(outFile, hasher)

	// Copy data through compression and hashing
	_, err = io.Copy(multiWriter, finalReader)
	if err != nil {
		return fmt.Errorf("failed to write patch data: %w", err)
	}

	// Close output file
	if err := outFile.Close(); err != nil {
		return fmt.Errorf("failed to close output file: %w", err)
	}

	// Check for encoding errors
	if err := <-encodeErr; err != nil {
		return fmt.Errorf("failed to encode patch: %w", err)
	}

	// Calculate checksum
	checksum := fmt.Sprintf("%x", hasher.Sum(nil))
	patch.Header.Checksum = checksum

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

// encodePatchStreaming writes the patch as JSON in a streaming fashion to avoid memory exhaustion
func encodePatchStreaming(patch *utils.Patch, writer io.Writer) error {
	// Create a buffered writer for better performance
	bufWriter := bufio.NewWriterSize(writer, 64*1024) // 64KB buffer
	defer bufWriter.Flush()

	// Write opening brace
	if _, err := bufWriter.WriteString("{\n"); err != nil {
		return err
	}

	// Encode header
	if err := encodeField(bufWriter, "Header", patch.Header, true); err != nil {
		return err
	}

	// Encode simple fields
	if err := encodeField(bufWriter, "FromVersion", patch.FromVersion, true); err != nil {
		return err
	}
	if err := encodeField(bufWriter, "ToVersion", patch.ToVersion, true); err != nil {
		return err
	}
	if err := encodeField(bufWriter, "FromKeyFile", patch.FromKeyFile, true); err != nil {
		return err
	}
	if err := encodeField(bufWriter, "ToKeyFile", patch.ToKeyFile, true); err != nil {
		return err
	}
	if err := encodeField(bufWriter, "RequiredFiles", patch.RequiredFiles, true); err != nil {
		return err
	}
	if err := encodeField(bufWriter, "SimpleMode", patch.SimpleMode, true); err != nil {
		return err
	}

	// Encode operations array manually to stream large data
	if _, err := bufWriter.WriteString(`  "Operations": [`); err != nil {
		return err
	}

	for i, op := range patch.Operations {
		if i > 0 {
			if _, err := bufWriter.WriteString(",\n"); err != nil {
				return err
			}
		} else {
			if _, err := bufWriter.WriteString("\n"); err != nil {
				return err
			}
		}

		if err := encodeOperation(bufWriter, op); err != nil {
			return err
		}
	}

	if _, err := bufWriter.WriteString("\n  ]\n"); err != nil {
		return err
	}

	// Write closing brace
	if _, err := bufWriter.WriteString("}\n"); err != nil {
		return err
	}

	return bufWriter.Flush()
}

// encodeField encodes a single field with proper JSON formatting
func encodeField(writer io.Writer, name string, value interface{}, addComma bool) error {
	commaStr := ","
	if !addComma {
		commaStr = ""
	}

	// For simple types, use JSON encoding
	data, err := json.MarshalIndent(value, "  ", "  ")
	if err != nil {
		return err
	}

	// Write field name and value
	fieldStr := fmt.Sprintf("  \"%s\": ", name)
	if _, err := writer.Write([]byte(fieldStr)); err != nil {
		return err
	}

	// Write the JSON data, but indent it properly
	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		if i > 0 {
			if _, err := writer.Write([]byte("\n")); err != nil {
				return err
			}
		}
		if _, err := writer.Write([]byte(line)); err != nil {
			return err
		}
	}

	if _, err := writer.Write([]byte(commaStr + "\n")); err != nil {
		return err
	}

	return nil
}

// encodeOperation encodes a single patch operation with streaming for large binary data
func encodeOperation(writer io.Writer, op utils.PatchOperation) error {
	// Write operation opening
	if _, err := writer.Write([]byte("    {\n")); err != nil {
		return err
	}

	// Encode simple fields
	if err := encodeOperationField(writer, "Type", int(op.Type), true); err != nil {
		return err
	}
	if err := encodeOperationField(writer, "FilePath", op.FilePath, true); err != nil {
		return err
	}
	if err := encodeOperationField(writer, "BinaryDiff", op.BinaryDiff, true); err != nil {
		return err
	}

	// Encode NewFile data - this is the large binary data that needs streaming
	if err := encodeOperationField(writer, "NewFile", op.NewFile, true); err != nil {
		return err
	}

	if err := encodeOperationField(writer, "OldChecksum", op.OldChecksum, true); err != nil {
		return err
	}
	if err := encodeOperationField(writer, "NewChecksum", op.NewChecksum, true); err != nil {
		return err
	}
	if err := encodeOperationField(writer, "Size", op.Size, false); err != nil {
		return err
	}

	// Write operation closing
	if _, err := writer.Write([]byte("\n    }")); err != nil {
		return err
	}

	return nil
}

// encodeOperationField encodes a single field within an operation
func encodeOperationField(writer io.Writer, name string, value interface{}, addComma bool) error {
	commaStr := ","
	if !addComma {
		commaStr = ""
	}

	fieldStr := fmt.Sprintf("      \"%s\": ", name)
	if _, err := writer.Write([]byte(fieldStr)); err != nil {
		return err
	}

	// For byte slices (binary data), encode as base64 using streaming encoder
	if byteData, ok := value.([]byte); ok {
		// Write opening quote
		if _, err := writer.Write([]byte("\"")); err != nil {
			return err
		}

		// Create base64 encoder that writes directly to output
		encoder := base64.NewEncoder(base64.StdEncoding, writer)

		// Write data in chunks to avoid memory exhaustion
		const chunkSize = 64 * 1024 // 64KB chunks
		for i := 0; i < len(byteData); i += chunkSize {
			end := i + chunkSize
			if end > len(byteData) {
				end = len(byteData)
			}
			if _, err := encoder.Write(byteData[i:end]); err != nil {
				encoder.Close()
				return err
			}
		}

		// Close encoder to flush any remaining data
		if err := encoder.Close(); err != nil {
			return err
		}

		// Write closing quote and comma
		if _, err := writer.Write([]byte(fmt.Sprintf("\"%s", commaStr))); err != nil {
			return err
		}
	} else {
		// For other types, use JSON encoding
		data, err := json.Marshal(value)
		if err != nil {
			return err
		}
		jsonStr := string(data) + commaStr
		if _, err := writer.Write([]byte(jsonStr)); err != nil {
			return err
		}
	}

	if _, err := writer.Write([]byte("\n")); err != nil {
		return err
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
	fmt.Println("  --savescans       Save directory scans to cache for faster subsequent patches")
	fmt.Println("  --rescan          Force rescan of cached versions (use with --savescans)")
	fmt.Println("  --scandata        Custom directory for scan cache (default: .data)")
	fmt.Println("  --jobs            Number of parallel workers (0=auto-detect CPU cores, 1=single-threaded, default: 0)")
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
	fmt.Println("\n  # Versions on different network locations")
	fmt.Println("  patch-gen --from-dir \\\\\\\\server1\\\\app\\\\v1 --to-dir \\\\\\\\server2\\\\app\\\\v2 --output .")
}
