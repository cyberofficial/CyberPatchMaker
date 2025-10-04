package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/cyberofficial/cyberpatchmaker/internal/core/patcher"
	"github.com/cyberofficial/cyberpatchmaker/pkg/utils"
)

func main() {
	// Define flags
	patchFile := flag.String("patch", "", "Path to patch file")
	currentDir := flag.String("current-dir", "", "Directory containing current version")
	dryRun := flag.Bool("dry-run", false, "Simulate patch without making changes")
	verify := flag.Bool("verify", true, "Verify file hashes before and after patching")
	backup := flag.Bool("backup", true, "Create backup before patching")
	help := flag.Bool("help", false, "Show help message")

	flag.Parse()

	if *help {
		printHelp()
		return
	}

	// Validate arguments
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

	if *dryRun {
		fmt.Println("\n=== DRY RUN MODE ===")
		fmt.Println("No changes will be made")
		performDryRun(patch, *currentDir)
		return
	}

	// Create backup if requested
	if *backup {
		fmt.Println("\nCreating backup...")
		backupDir := *currentDir + ".backup"
		if err := createBackup(*currentDir, backupDir); err != nil {
			fmt.Printf("Error: failed to create backup: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Backup created at: %s\n", backupDir)
	}

	// Apply patch
	fmt.Println("\nApplying patch...")
	applier := patcher.NewApplier()
	if err := applier.ApplyPatch(patch, *currentDir, *verify, *verify); err != nil {
		fmt.Printf("Error: patch application failed: %v\n", err)
		if *backup {
			fmt.Println("Restoring from backup...")
			if restoreErr := restoreBackup(*currentDir+".backup", *currentDir); restoreErr != nil {
				fmt.Printf("Error: failed to restore backup: %v\n", restoreErr)
			} else {
				fmt.Println("Backup restored successfully")
			}
		}
		os.Exit(1)
	}

	fmt.Println("\n=== Patch Applied Successfully ===")
	fmt.Printf("Version updated from %s to %s\n", patch.FromVersion, patch.ToVersion)

	// Clean up backup if successful
	if *backup {
		fmt.Println("Removing backup...")
		if err := os.RemoveAll(*currentDir + ".backup"); err != nil {
			fmt.Printf("Warning: failed to remove backup: %v\n", err)
		}
	}
}

func loadPatch(filename string) (*utils.Patch, error) {
	// Read patch file
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read patch file: %w", err)
	}

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
	fmt.Printf("Required Hash:    %s\n", patch.FromKeyFile.Checksum[:16]+"...")
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

func performDryRun(patch *utils.Patch, currentDir string) {
	fmt.Println("\nSimulating patch application...")

	// Verify key file
	fmt.Printf("\nVerifying key file: %s\n", patch.FromKeyFile.Path)
	keyFilePath := currentDir + string(os.PathSeparator) + patch.FromKeyFile.Path
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
	for i, op := range patch.Operations {
		if i >= 10 {
			fmt.Printf("... and %d more operations\n", len(patch.Operations)-10)
			break
		}

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

func createBackup(srcDir, backupDir string) error {
	// Remove existing backup if it exists
	if utils.FileExists(backupDir) {
		if err := os.RemoveAll(backupDir); err != nil {
			return fmt.Errorf("failed to remove existing backup: %w", err)
		}
	}

	// Create backup directory
	if err := utils.EnsureDir(backupDir); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Copy all files
	return copyDir(srcDir, backupDir)
}

func restoreBackup(backupDir, targetDir string) error {
	// Remove current directory
	if err := os.RemoveAll(targetDir); err != nil {
		return fmt.Errorf("failed to remove current directory: %w", err)
	}

	// Restore from backup
	return copyDir(backupDir, targetDir)
}

func copyDir(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := src + string(os.PathSeparator) + entry.Name()
		dstPath := dst + string(os.PathSeparator) + entry.Name()

		if entry.IsDir() {
			if err := utils.EnsureDir(dstPath); err != nil {
				return err
			}
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := utils.CopyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func printHelp() {
	fmt.Println("CyberPatchMaker - Patch Applier")
	fmt.Println("\nUsage:")
	fmt.Println("  patch-apply --patch <file> --current-dir <directory>")
	fmt.Println("\nOptions:")
	fmt.Println("  --patch         Path to patch file (required)")
	fmt.Println("  --current-dir   Directory containing current version (required)")
	fmt.Println("  --dry-run       Simulate patch without making changes")
	fmt.Println("  --verify        Verify file hashes before and after patching (default: true)")
	fmt.Println("  --backup        Create backup before patching (default: true)")
	fmt.Println("  --help          Show this help message")
	fmt.Println("\nExamples:")
	fmt.Println("  # Apply patch")
	fmt.Println("  patch-apply --patch 1.0.0-to-1.0.3.patch --current-dir C:\\MyApp")
	fmt.Println("\n  # Dry run (simulate only)")
	fmt.Println("  patch-apply --patch 1.0.0-to-1.0.3.patch --current-dir C:\\MyApp --dry-run")
}
