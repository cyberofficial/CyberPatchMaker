package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/cyberofficial/cyberpatchmaker/internal/core/patcher"
	"github.com/cyberofficial/cyberpatchmaker/internal/core/version"
	"github.com/cyberofficial/cyberpatchmaker/pkg/utils"
)

const (
	MAGIC_BYTES = "CPMPATCH"
	HEADER_SIZE = 128
)

type EmbeddedPatchHeader struct {
	Magic       [8]byte
	Version     uint32
	StubSize    uint64
	DataOffset  uint64
	DataSize    uint64
	Compression [16]byte
	Checksum    [32]byte
	Reserved    [44]byte
}

func main() {
	// Define flags
	patchFile := flag.String("patch", "", "Path to patch file")
	currentDir := flag.String("current-dir", "", "Directory containing current version")
	keyFile := flag.String("key-file", "", "Custom key file path (if renamed or moved)")
	dryRun := flag.Bool("dry-run", false, "Simulate patch without making changes")
	verify := flag.Bool("verify", true, "Verify file hashes before and after patching")
	backup := flag.Bool("backup", true, "Create backup before patching")
	ignore1GB := flag.Bool("ignore1gb", false, "Bypass 1GB patch size limit (use with caution)")
	silent := flag.Bool("silent", false, "Silent mode: apply patch automatically without prompts (for automation)")
	versionFlag := flag.Bool("version", false, "Show version information")
	help := flag.Bool("help", false, "Show help message")

	flag.Parse()

	// Show version if requested
	if *versionFlag {
		fmt.Printf("CyberPatchMaker Patch Applier v%s\n", version.GetVersion())
		return
	}

	if *help {
		printHelp()
		return
	}

	// Check if patch data is embedded in this executable
	patch, targetDir, isEmbedded := checkEmbeddedPatch(*ignore1GB)

	if isEmbedded && patch != nil {
		if *silent {
			// Silent mode: apply patch automatically
			runSilentMode(patch, targetDir, *currentDir, *keyFile)
			return
		}
		// Interactive console mode for embedded patch
		fmt.Println("==============================================")
		fmt.Println("  CyberPatchMaker - Self-Contained Patch")
		fmt.Println("==============================================")
		runInteractiveMode(patch, targetDir, *ignore1GB)
		return
	}

	// Standard mode - require arguments
	if *patchFile == "" || *currentDir == "" {
		fmt.Println("Error: --patch and --current-dir are required")
		printHelp()
		os.Exit(1)
	}

	// Check if patch file exists
	if !utils.FileExists(*patchFile) {
		fmt.Printf("Error: patch file not found: %s\n", *patchFile)
		os.Exit(1)
	}

	// Check if current directory exists
	if !utils.FileExists(*currentDir) {
		fmt.Printf("Error: current directory not found: %s\n", *currentDir)
		os.Exit(1)
	}

	// Load patch
	patch, err := loadPatch(*patchFile)
	if err != nil {
		fmt.Printf("Error: failed to load patch: %v\n", err)
		os.Exit(1)
	}

	// Display patch information
	displayPatchInfo(patch)

	// Override key file path if custom one is provided
	if *keyFile != "" {
		fmt.Printf("\nUsing custom key file: %s\n", *keyFile)
		patch.FromKeyFile.Path = *keyFile
	}

	if *dryRun {
		fmt.Println("\n=== DRY RUN MODE ===")
		fmt.Println("No changes will be made")
		performDryRun(patch, *currentDir, *keyFile)
		return
	}

	applier := patcher.NewApplier()
	if err := applier.ApplyPatch(patch, *currentDir, *verify, *verify, *backup); err != nil {
		fmt.Printf("Error: patch application failed: %v\n", err)
		if *backup {
			fmt.Println("\nNote: If backup was created, automatic rollback may have been performed to restore original files.")
		}
		os.Exit(1)
	}

	fmt.Println("\n=== Patch Applied Successfully ===")
	fmt.Printf("Version updated from %s to %s\n", patch.FromVersion, patch.ToVersion)
}

func loadPatch(filename string) (*utils.Patch, error) {
	// Read patch file
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read patch file: %w", err)
	}

	return parsePatchData(data)
}

// parsePatchData parses patch data, automatically detecting and decompressing if needed
func parsePatchData(data []byte) (*utils.Patch, error) {
	// Try to detect compression and decompress
	// First try as JSON directly
	var patch utils.Patch
	if err := json.Unmarshal(data, &patch); err != nil {
		// Try decompressing with zstd
		decompressed, err := utils.DecompressData(data, "zstd")
		if err != nil {
			// Try gzip
			decompressed, err = utils.DecompressData(data, "gzip")
			if err != nil {
				return nil, fmt.Errorf("failed to decompress or parse patch: %w", err)
			}
		}
		data = decompressed

		// Try parsing again
		if err := json.Unmarshal(data, &patch); err != nil {
			return nil, fmt.Errorf("failed to parse patch JSON: %w", err)
		}
	}

	return &patch, nil
}

func displayPatchInfo(patch *utils.Patch) {
	fmt.Println("\n=== Patch Information ===")
	fmt.Printf("From Version:     %s\n", patch.FromVersion)
	fmt.Printf("To Version:       %s\n", patch.ToVersion)
	fmt.Printf("Key File:         %s\n", patch.FromKeyFile.Path)
	fmt.Printf("Required Hash:    %s\n", patch.FromKeyFile.Checksum)
	fmt.Printf("Patch Size:       %d bytes\n", patch.Header.PatchSize)
	fmt.Printf("Compression:      %s\n", patch.Header.Compression)
	fmt.Printf("Created:          %s\n", patch.Header.CreatedAt.Format("2006-01-02 15:04:05"))

	// Count operations
	addCount := 0
	modifyCount := 0
	deleteCount := 0
	addDirCount := 0
	deleteDirCount := 0

	for _, op := range patch.Operations {
		switch op.Type {
		case utils.OpAdd:
			addCount++
		case utils.OpModify:
			modifyCount++
		case utils.OpDelete:
			deleteCount++
		case utils.OpAddDir:
			addDirCount++
		case utils.OpDeleteDir:
			deleteDirCount++
		}
	}

	fmt.Printf("Files Added:      %d\n", addCount)
	fmt.Printf("Files Modified:   %d\n", modifyCount)
	fmt.Printf("Files Deleted:    %d\n", deleteCount)
	fmt.Printf("Dirs Added:       %d\n", addDirCount)
	fmt.Printf("Dirs Deleted:     %d\n", deleteDirCount)
	fmt.Printf("Required Files:   %d (must match exact hashes)\n", len(patch.RequiredFiles))
}

// resolveKeyFilePath resolves the actual key file path, using custom path if provided
func resolveKeyFilePath(patch *utils.Patch, currentDir string, customKeyFile string) string {
	if customKeyFile != "" {
		// Use custom key file path (can be absolute or relative)
		if strings.Contains(customKeyFile, string(os.PathSeparator)) || strings.Contains(customKeyFile, "/") {
			// If it contains path separators, treat as-is
			return customKeyFile
		}
		// Otherwise, treat as relative to currentDir
		return currentDir + string(os.PathSeparator) + customKeyFile
	}
	// Use default key file from patch
	return currentDir + string(os.PathSeparator) + patch.FromKeyFile.Path
}

func performDryRun(patch *utils.Patch, currentDir string, customKeyFile string) {
	fmt.Println("\nSimulating patch application...")

	// Verify key file
	if customKeyFile != "" {
		fmt.Printf("\nVerifying custom key file: %s\n", customKeyFile)
	} else {
		fmt.Printf("\nVerifying key file: %s\n", patch.FromKeyFile.Path)
	}
	keyFilePath := resolveKeyFilePath(patch, currentDir, customKeyFile)
	if !utils.FileExists(keyFilePath) {
		fmt.Printf("✗ Key file not found: %s\n", keyFilePath)
		return
	}

	checksum, err := utils.CalculateFileChecksum(keyFilePath)
	if err != nil {
		fmt.Printf("✗ Failed to calculate key file checksum: %v\n", err)
		return
	}

	if checksum != patch.FromKeyFile.Checksum {
		fmt.Printf("✗ Key file hash mismatch\n")
		fmt.Printf("  Expected: %s\n", patch.FromKeyFile.Checksum[:16]+"...")
		fmt.Printf("  Got:      %s\n", checksum[:16]+"...")
		return
	}
	fmt.Println("✓ Key file verified")

	// Verify required files
	fmt.Printf("\nVerifying %d required files...\n", len(patch.RequiredFiles))
	mismatches := 0
	for i, req := range patch.RequiredFiles {
		if i < 5 || mismatches > 0 { // Show first 5 or any mismatches
			filePath := currentDir + string(os.PathSeparator) + req.Path
			if !utils.FileExists(filePath) {
				fmt.Printf("✗ Required file missing: %s\n", req.Path)
				mismatches++
				continue
			}

			checksum, err := utils.CalculateFileChecksum(filePath)
			if err != nil {
				fmt.Printf("✗ Failed to verify: %s\n", req.Path)
				mismatches++
				continue
			}

			if checksum != req.Checksum {
				fmt.Printf("✗ Hash mismatch: %s\n", req.Path)
				mismatches++
			}
		}
	}

	if mismatches > 0 {
		fmt.Printf("\n✗ %d file(s) have mismatches - patch cannot be applied\n", mismatches)
		return
	}

	fmt.Println("✓ All required files verified")

	// Show operations that would be performed
	fmt.Println("\nOperations that would be performed:")
	for _, op := range patch.Operations {
		switch op.Type {
		case utils.OpAdd:
			fmt.Printf("  ADD: %s\n", op.FilePath)
		case utils.OpModify:
			fmt.Printf("  MODIFY: %s\n", op.FilePath)
		case utils.OpDelete:
			fmt.Printf("  DELETE: %s\n", op.FilePath)
		case utils.OpAddDir:
			fmt.Printf("  ADD DIR: %s\n", op.FilePath)
		case utils.OpDeleteDir:
			fmt.Printf("  DELETE DIR: %s\n", op.FilePath)
		}
	}

	fmt.Println("\n✓ Dry run completed - patch can be applied safely")
}

// checkEmbeddedPatch checks if this executable contains an embedded patch
func checkEmbeddedPatch(ignore1GB bool) (*utils.Patch, string, bool) {
	// Get path to this executable
	exePath, err := os.Executable()
	if err != nil {
		return nil, "", false
	}

	// Open executable for reading
	file, err := os.Open(exePath)
	if err != nil {
		return nil, "", false
	}
	defer file.Close()

	// Get file size
	stat, err := file.Stat()
	if err != nil {
		return nil, "", false
	}
	fileSize := stat.Size()

	// Check if file is large enough for header
	if fileSize < HEADER_SIZE {
		return nil, "", false
	}

	// Read header from end of file
	headerOffset := fileSize - HEADER_SIZE
	if _, err := file.Seek(headerOffset, io.SeekStart); err != nil {
		return nil, "", false
	}

	headerBytes := make([]byte, HEADER_SIZE)
	if _, err := io.ReadFull(file, headerBytes); err != nil {
		return nil, "", false
	}

	// Parse header
	var header EmbeddedPatchHeader
	buf := bytes.NewReader(headerBytes)
	if err := binary.Read(buf, binary.LittleEndian, &header); err != nil {
		return nil, "", false
	}

	// Validate magic bytes
	magic := string(bytes.TrimRight(header.Magic[:], "\x00"))
	if magic != MAGIC_BYTES {
		return nil, "", false
	}

	// Validate version
	if header.Version != 1 {
		return nil, "", false
	}

	// Validate structure
	if header.DataOffset != header.StubSize {
		return nil, "", false
	}

	expectedSize := header.StubSize + header.DataSize + HEADER_SIZE
	if expectedSize != uint64(fileSize) {
		return nil, "", false
	}

	// Check size limit (1GB)
	const maxPatchSize = 1 << 30 // 1 GB
	if !ignore1GB && header.DataSize > maxPatchSize {
		fmt.Printf("Warning: Patch size (%d bytes) exceeds 1GB limit\n", header.DataSize)
		fmt.Println("Use --ignore1gb flag if you want to proceed anyway")
		return nil, "", false
	}

	// Read patch data
	if _, err := file.Seek(int64(header.DataOffset), io.SeekStart); err != nil {
		return nil, "", false
	}

	patchData := make([]byte, header.DataSize)
	if _, err := io.ReadFull(file, patchData); err != nil {
		return nil, "", false
	}

	// Verify checksum
	actualChecksum := sha256.Sum256(patchData)
	if !bytes.Equal(actualChecksum[:], header.Checksum[:]) {
		return nil, "", false
	}

	// The embedded patch data is the raw .patch file content
	// We need to parse it the same way loadPatch() does
	patch, err := parsePatchData(patchData)
	if err != nil {
		return nil, "", false
	}

	// Get current directory as default target
	targetDir, _ := os.Getwd()

	return patch, targetDir, true
}

// runSilentMode applies the patch automatically without user interaction (for automation)
func runSilentMode(patch *utils.Patch, defaultTargetDir string, customTargetDir string, customKeyFile string) {
	// Use custom target directory if provided, otherwise use default (current directory)
	targetDir := defaultTargetDir
	if customTargetDir != "" {
		targetDir = customTargetDir
	}

	// Create log file with current epoch timestamp
	logFileName := fmt.Sprintf("log_%d.txt", time.Now().Unix())
	logFile, err := os.Create(logFileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to create log file: %v\n", err)
		logFile = nil
	}
	defer func() {
		if logFile != nil {
			logFile.Close()
		}
	}()

	// Helper function to write to both console and log
	logOutput := func(format string, args ...interface{}) {
		msg := fmt.Sprintf(format, args...)
		fmt.Print(msg)
		if logFile != nil {
			logFile.WriteString(msg)
		}
	}

	// Log header with timestamp
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logOutput("========================================\n")
	logOutput("CyberPatchMaker Silent Mode Log\n")
	logOutput("Started: %s\n", timestamp)
	logOutput("========================================\n\n")

	// Check if directory exists
	if !utils.FileExists(targetDir) {
		logOutput("Error: Target directory not found: %s\n", targetDir)
		logOutput("\n========================================\n")
		logOutput("Status: FAILED\n")
		logOutput("Completed: %s\n", time.Now().Format("2006-01-02 15:04:05"))
		logOutput("========================================\n")
		if logFile != nil {
			logOutput("\nLog saved to: %s\n", logFileName)
		}
		os.Exit(1)
	}

	// Override key file path if custom one is provided
	if customKeyFile != "" {
		patch.FromKeyFile.Path = customKeyFile
		logOutput("Using custom key file: %s\n", customKeyFile)
	}

	// Log patch details
	logOutput("Patch Information:\n")
	logOutput("  From Version: %s\n", patch.FromVersion)
	logOutput("  To Version:   %s\n", patch.ToVersion)
	logOutput("  Key File:     %s\n", patch.FromKeyFile.Path)
	logOutput("  Target Dir:   %s\n", targetDir)
	logOutput("  Compression:  %s\n", patch.Header.Compression)
	logOutput("\n")

	// Display simple startup message
	logOutput("Applying patch...\n\n")

	// Apply patch with default settings (verify=true, backup=true)
	applier := patcher.NewApplier()
	if err := applier.ApplyPatch(patch, targetDir, true, true, true); err != nil {
		logOutput("\nError: Patch application failed: %v\n", err)
		logOutput("\n========================================\n")
		logOutput("Status: FAILED\n")
		logOutput("Completed: %s\n", time.Now().Format("2006-01-02 15:04:05"))
		logOutput("========================================\n")
		if logFile != nil {
			logOutput("\nLog saved to: %s\n", logFileName)
		}
		os.Exit(1)
	}

	// Success - output minimal message
	logOutput("\n")
	logOutput("Patch applied successfully: %s → %s\n", patch.FromVersion, patch.ToVersion)
	logOutput("\n========================================\n")
	logOutput("Status: SUCCESS\n")
	logOutput("Completed: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	logOutput("========================================\n")
	if logFile != nil {
		logOutput("\nLog saved to: %s\n", logFileName)
	}
	os.Exit(0)
}

// runSimpleMode runs a simplified interface for end users when patch creator enabled simple mode
func runSimpleMode(patch *utils.Patch, defaultTargetDir string, reader *bufio.Reader) {
	fmt.Println()
	fmt.Println("==============================================")
	fmt.Println("          Simple Patch Application")
	fmt.Println("==============================================")
	fmt.Println()
	fmt.Printf("You are about to patch \"%s\" to \"%s\"\n", patch.FromVersion, patch.ToVersion)
	fmt.Println()

	// Ask for target directory
	fmt.Printf("Target directory [%s]: ", defaultTargetDir)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	targetDir := defaultTargetDir
	if input != "" {
		targetDir = input
	}

	// Check if directory exists
	if !utils.FileExists(targetDir) {
		fmt.Printf("Error: Directory not found: %s\n", targetDir)
		fmt.Println("\nPress Enter to exit...")
		reader.ReadString('\n')
		os.Exit(1)
	}

	// Ask about backup (default: yes)
	fmt.Print("\nCreate backup before patching? (Y/n): ")
	backupInput, _ := reader.ReadString('\n')
	backupInput = strings.TrimSpace(strings.ToLower(backupInput))
	createBackup := backupInput == "" || backupInput == "y" || backupInput == "yes"

	// Show options menu
	for {
		fmt.Println("\n==============================================")
		fmt.Println("Options:")
		fmt.Println("  1. Dry Run (test without making changes)")
		fmt.Println("  2. Apply Patch")
		fmt.Println("  3. Exit")
		fmt.Println("==============================================")
		fmt.Print("Select option [1-3]: ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		switch input {
		case "1":
			// Dry run
			fmt.Println("\n=== DRY RUN ===")
			fmt.Println("Testing patch application...")
			performDryRun(patch, targetDir, "")
			fmt.Println("\nPress Enter to continue...")
			reader.ReadString('\n')

		case "2":
			// Apply patch
			fmt.Println("\n=== APPLYING PATCH ===")
			fmt.Printf("Target: %s\n", targetDir)
			fmt.Printf("Backup: %s\n", formatBoolState(createBackup))
			fmt.Println()

			fmt.Print("Proceed with patching? (yes/no): ")
			confirm, _ := reader.ReadString('\n')
			confirm = strings.TrimSpace(strings.ToLower(confirm))

			if confirm == "yes" || confirm == "y" {
				fmt.Println("\nApplying patch...")
				applier := patcher.NewApplier()
				// Use default settings: verify before and after
				if err := applier.ApplyPatch(patch, targetDir, true, true, createBackup); err != nil {
					fmt.Printf("\nError: Patch application failed: %v\n", err)
					if createBackup {
						fmt.Println("\nNote: Automatic rollback may have been performed to restore original files.")
					}
					fmt.Println("\nPress Enter to exit...")
					reader.ReadString('\n')
					os.Exit(1)
				}

				fmt.Println("\n=== SUCCESS ===")
				fmt.Printf("Patch applied successfully!\n")
				fmt.Printf("Version updated from %s to %s\n", patch.FromVersion, patch.ToVersion)
				fmt.Println("\nPress Enter to exit...")
				reader.ReadString('\n')
				return
			} else {
				fmt.Println("Patch application cancelled")
			}

		case "3":
			// Exit
			fmt.Println("\nExiting...")
			return

		default:
			fmt.Println("Invalid option. Please select 1-3.")
		}
	}
}

// runInteractiveMode runs the interactive console interface for embedded patches
func runInteractiveMode(patch *utils.Patch, defaultTargetDir string, ignore1GB bool) {
	reader := bufio.NewReader(os.Stdin)
	customKeyFile := "" // Track custom key file path

	// Check if patch creator enabled simple mode for end users
	if patch.SimpleMode {
		runSimpleMode(patch, defaultTargetDir, reader)
		return
	}

	// Display patch information
	fmt.Println()
	displayPatchInfo(patch)

	// Ask for target directory
	fmt.Printf("\nTarget directory [%s]: ", defaultTargetDir)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	targetDir := defaultTargetDir
	if input != "" {
		targetDir = input
	}

	// Check if directory exists
	if !utils.FileExists(targetDir) {
		fmt.Printf("Error: Directory not found: %s\n", targetDir)
		fmt.Println("\nPress Enter to exit...")
		reader.ReadString('\n')
		os.Exit(1)
	}

	// Show menu
	for {
		fmt.Println("\n==============================================")
		fmt.Println("Options:")
		fmt.Println("  1. Dry Run (simulate without changes)")
		fmt.Println("  2. Apply Patch")
		fmt.Println("  3. Toggle 1GB Bypass Mode (currently: " + formatBoolState(ignore1GB) + ")")
		fmt.Println("  4. Change Target Directory")
		fmt.Println("  5. Specify Custom Key File")
		if customKeyFile != "" {
			fmt.Printf("     (Currently: %s)\n", customKeyFile)
		} else {
			fmt.Printf("     (Currently: %s - default)\n", patch.FromKeyFile.Path)
		}
		fmt.Println("  6. Exit")
		fmt.Println("==============================================")
		fmt.Print("Select option [1-6]: ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		switch input {
		case "1":
			// Dry run
			fmt.Println("\n=== DRY RUN MODE ===")
			fmt.Println("Simulating patch application...")
			performDryRun(patch, targetDir, customKeyFile)
			fmt.Println("\nPress Enter to continue...")
			reader.ReadString('\n')

		case "2":
			// Apply patch
			fmt.Println("\n=== APPLYING PATCH ===")
			fmt.Printf("Target: %s\n", targetDir)

			// Override key file path if custom one is provided
			if customKeyFile != "" {
				fmt.Printf("Using custom key file: %s\n", customKeyFile)
				patch.FromKeyFile.Path = customKeyFile
			}

			fmt.Print("Are you sure you want to apply this patch? (yes/no): ")
			confirm, _ := reader.ReadString('\n')
			confirm = strings.TrimSpace(strings.ToLower(confirm))

			if confirm == "yes" || confirm == "y" {
				fmt.Println("\nApplying patch...")
				applier := patcher.NewApplier()
				if err := applier.ApplyPatch(patch, targetDir, true, true, true); err != nil {
					fmt.Printf("\nError: Patch application failed: %v\n", err)
					fmt.Println("\nNote: Automatic rollback may have been performed to restore original files.")
					fmt.Println("\nPress Enter to exit...")
					reader.ReadString('\n')
					os.Exit(1)
				}

				fmt.Println("\n=== SUCCESS ===")
				fmt.Printf("Patch applied successfully!\n")
				fmt.Printf("Version updated from %s to %s\n", patch.FromVersion, patch.ToVersion)
				fmt.Println("\nPress Enter to exit...")
				reader.ReadString('\n')
				return
			} else {
				fmt.Println("Patch application cancelled")
			}

		case "3":
			// Toggle 1GB bypass
			ignore1GB = !ignore1GB
			fmt.Printf("\n1GB Bypass Mode: %s\n", formatBoolState(ignore1GB))
			if ignore1GB {
				fmt.Println("Warning: Large patches may consume significant memory!")
			}

		case "4":
			// Change target directory
			fmt.Print("\nEnter new target directory: ")
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)
			if input != "" {
				if !utils.FileExists(input) {
					fmt.Printf("Error: Directory not found: %s\n", input)
				} else {
					targetDir = input
					fmt.Printf("Target directory changed to: %s\n", targetDir)
				}
			}

		case "5":
			// Specify custom key file
			fmt.Print("\nEnter custom key file path (or press Enter to use default): ")
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)
			if input == "" {
				customKeyFile = ""
				fmt.Printf("Using default key file: %s\n", patch.FromKeyFile.Path)
			} else {
				// Validate the custom key file exists
				testPath := input
				if !strings.Contains(input, string(os.PathSeparator)) && !strings.Contains(input, "/") {
					// Relative path - check in target directory
					testPath = targetDir + string(os.PathSeparator) + input
				}
				if !utils.FileExists(testPath) {
					fmt.Printf("Warning: File not found: %s\n", testPath)
					fmt.Println("You can still proceed - verification will happen during patch application.")
				}
				customKeyFile = input
				fmt.Printf("Custom key file set to: %s\n", customKeyFile)
			}

		case "6":
			// Exit
			fmt.Println("\nExiting...")
			return

		default:
			fmt.Println("Invalid option. Please select 1-6.")
		}
	}
}

// formatBoolState formats a boolean as "Enabled" or "Disabled"
func formatBoolState(enabled bool) string {
	if enabled {
		return "Enabled"
	}
	return "Disabled"
}

func printHelp() {
	fmt.Printf("CyberPatchMaker - Patch Applier v%s\n", version.GetVersion())
	fmt.Println("\nUsage:")
	fmt.Println("  patch-apply --patch <file> --current-dir <directory>")
	fmt.Println("\nOptions:")
	fmt.Println("  --patch         Path to patch file (required)")
	fmt.Println("  --current-dir   Directory containing current version (required)")
	fmt.Println("  --key-file      Custom key file path (if renamed or moved)")
	fmt.Println("  --dry-run       Simulate patch without making changes")
	fmt.Println("  --verify        Verify file hashes before and after patching (default: true)")
	fmt.Println("  --backup        Create backup before patching (default: true)")
	fmt.Println("  --ignore1gb     Bypass 1GB patch size limit (use with caution)")
	fmt.Println("  --silent        Silent mode: apply patch automatically without prompts")
	fmt.Println("  --version       Show version information")
	fmt.Println("  --help          Show this help message")
	fmt.Println("\nSelf-Contained Executable Mode:")
	fmt.Println("  When run as a self-contained executable, an interactive console")
	fmt.Println("  interface will guide you through the patch application process.")
	fmt.Println("  Use --silent flag for automated patching without user interaction.")
	fmt.Println("\nExamples:")
	fmt.Println("  # Apply patch")
	fmt.Println("  patch-apply --patch 1.0.0-to-1.0.3.patch --current-dir C:\\MyApp")
	fmt.Println("\n  # Dry run (simulate only)")
	fmt.Println("  patch-apply --patch 1.0.0-to-1.0.3.patch --current-dir C:\\MyApp --dry-run")
	fmt.Println("\n  # Run self-contained executable with 1GB bypass")
	fmt.Println("  1.0.0-to-1.0.1.exe --ignore1gb")
	fmt.Println("\n  # Run self-contained executable in silent mode (automation)")
	fmt.Println("  1.2.4-to-1.2.5.exe --silent")
	fmt.Println("\n  # Silent mode with custom target directory")
	fmt.Println("  1.2.4-to-1.2.5.exe --silent --current-dir C:\\MyApp")
}
