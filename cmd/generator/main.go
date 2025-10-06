package main

import (
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
	compression := flag.String("compression", "zstd", "Compression algorithm (zstd, gzip, none)")
	level := flag.Int("level", 3, "Compression level (1-4 for zstd, 1-9 for gzip)")
	verify := flag.Bool("verify", true, "Verify patches after creation")
	help := flag.Bool("help", false, "Show help message")

	flag.Parse()

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
		generateAllPatches(versionMgr, *versionsDir, *newVersion, outputDir, *compression, *level, *verify)
	} else if *fromDir != "" && *toDir != "" {
		// Generate single patch using custom directory paths
		generateSinglePatchCustomPaths(versionMgr, *fromDir, *toDir, outputDir, *compression, *level, *verify)
	} else if *from != "" && *to != "" && *versionsDir != "" {
		// Generate single patch using versions-dir
		generateSinglePatch(versionMgr, *versionsDir, *from, *to, outputDir, *compression, *level, *verify)
	} else {
		fmt.Println("Error: insufficient arguments")
		printHelp()
		os.Exit(1)
	}
}

func generateAllPatches(versionMgr *version.Manager, versionsDir, newVersion, outputDir, compression string, level int, verify bool) {
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

	// Find key file (assume it's named program.exe, game.exe, or app.exe)
	keyFiles := []string{"program.exe", "game.exe", "app.exe", "main.exe"}
	var keyFile string
	for _, kf := range keyFiles {
		if utils.FileExists(filepath.Join(newVersionPath, kf)) {
			keyFile = kf
			break
		}
	}
	if keyFile == "" {
		fmt.Println("Error: could not find key file (program.exe, game.exe, app.exe, or main.exe)")
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

		fromVer, err := versionMgr.RegisterVersion(fromVersion, fromPath, keyFile)
		if err != nil {
			fmt.Printf("Warning: failed to register version %s: %v\n", fromVersion, err)
			continue
		}

		// Generate patch
		patchFile := filepath.Join(outputDir, fmt.Sprintf("%s-to-%s.patch", fromVersion, newVersion))
		if err := generatePatch(fromVer, toVer, patchFile, compression, level, verify); err != nil {
			fmt.Printf("Error: failed to generate patch from %s: %v\n", fromVersion, err)
			continue
		}

		patchCount++
	}

	fmt.Printf("\nSuccessfully generated %d patches\n", patchCount)
}

func generateSinglePatch(versionMgr *version.Manager, versionsDir, from, to, outputDir, compression string, level int, verify bool) {
	fmt.Printf("Generating patch from %s to %s\n", from, to)

	// Find key file
	fromPath := filepath.Join(versionsDir, from)
	keyFiles := []string{"program.exe", "game.exe", "app.exe", "main.exe"}
	var keyFile string
	for _, kf := range keyFiles {
		if utils.FileExists(filepath.Join(fromPath, kf)) {
			keyFile = kf
			break
		}
	}
	if keyFile == "" {
		fmt.Println("Error: could not find key file")
		os.Exit(1)
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

	// Generate patch
	patchFile := filepath.Join(outputDir, fmt.Sprintf("%s-to-%s.patch", from, to))
	if err := generatePatch(fromVer, toVer, patchFile, compression, level, verify); err != nil {
		fmt.Printf("Error: failed to generate patch: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Patch generated successfully")
}

// generateSinglePatchCustomPaths generates a patch using custom directory paths
// This allows versions to be on different drives or network locations
func generateSinglePatchCustomPaths(versionMgr *version.Manager, fromPath, toPath, outputDir, compression string, level int, verify bool) {
	// Extract version numbers from directory names
	fromVersion := extractVersionFromPath(fromPath)
	toVersion := extractVersionFromPath(toPath)

	fmt.Printf("Generating patch from %s (%s) to %s (%s)...\n", fromVersion, fromPath, toVersion, toPath)

	// Find key file in source directory
	keyFiles := []string{"program.exe", "game.exe", "app.exe", "main.exe"}
	var keyFile string
	for _, kf := range keyFiles {
		if utils.FileExists(filepath.Join(fromPath, kf)) {
			keyFile = kf
			break
		}
	}
	if keyFile == "" {
		fmt.Println("Error: could not find key file in source directory")
		os.Exit(1)
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

	// Generate patch
	patchFile := filepath.Join(outputDir, fmt.Sprintf("%s-to-%s.patch", fromVersion, toVersion))
	if err := generatePatch(fromVer, toVer, patchFile, compression, level, verify); err != nil {
		fmt.Printf("Error: failed to generate patch: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Patch generated successfully: %s\n", patchFile)
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
		DiffThresholdKB:  1,
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

func printHelp() {
	fmt.Println("CyberPatchMaker - Patch Generator")
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
	fmt.Println("  --compression     Compression algorithm: zstd, gzip, none (default: zstd)")
	fmt.Println("  --level           Compression level (default: 3)")
	fmt.Println("  --verify          Verify patches after creation (default: true)")
	fmt.Println("  --help            Show this help message")
	fmt.Println("\nExamples:")
	fmt.Println("  # Versions on different drives")
	fmt.Println("  patch-gen --from-dir C:\\releases\\1.0.0 --to-dir D:\\builds\\1.0.1 --output patches")
	fmt.Println("\n  # Versions on different network locations")
	fmt.Println("  patch-gen --from-dir \\\\server1\\app\\v1 --to-dir \\\\server2\\app\\v2 --output .")
}
