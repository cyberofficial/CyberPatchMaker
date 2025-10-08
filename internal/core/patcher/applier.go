package patcher

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cyberofficial/cyberpatchmaker/internal/core/differ"
	"github.com/cyberofficial/cyberpatchmaker/pkg/utils"
)

// Applier handles patch application
type Applier struct {
	differ *differ.Differ
}

// NewApplier creates a new patch applier
func NewApplier() *Applier {
	return &Applier{
		differ: differ.NewDiffer(),
	}
}

// ApplyPatch applies a patch to a target directory
func (a *Applier) ApplyPatch(patch *utils.Patch, targetDir string, verifyBefore, verifyAfter bool, createBackup bool) error {
	fmt.Printf("Applying patch from %s to %s...\n", patch.FromVersion, patch.ToVersion)

	// Verify target directory exists
	if !utils.FileExists(targetDir) {
		return fmt.Errorf("target directory does not exist: %s", targetDir)
	}

	// Pre-patch verification
	if verifyBefore {
		fmt.Println("Verifying current version...")
		if err := a.verifyKeyFile(targetDir, patch.FromKeyFile); err != nil {
			return fmt.Errorf("key file verification failed: %w", err)
		}

		if err := a.verifyRequiredFiles(targetDir, patch.RequiredFiles); err != nil {
			return fmt.Errorf("required files verification failed: %w", err)
		}
		fmt.Println("Pre-patch verification successful")
	}

	// Create backup AFTER verification passes but BEFORE applying operations
	if createBackup {
		fmt.Println("\nCreating backup...")
		backupDir := filepath.Join(targetDir, "backup.cyberpatcher")
		if err := a.createMirrorBackup(targetDir, backupDir, patch.Operations); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
		fmt.Printf("Backup created at: %s\n", backupDir)
		fmt.Println("Note: Backup will be preserved after patching for manual rollback")
	}

	// Apply operations
	fmt.Printf("Applying %d operations...\n", len(patch.Operations))
	for i, op := range patch.Operations {
		if err := a.applyOperation(targetDir, op); err != nil {
			// Operation failed - automatically restore from backup if it was created
			if createBackup {
				fmt.Printf("\nOperation %d failed, automatically restoring from backup...\n", i)
				backupDir := filepath.Join(targetDir, "backup.cyberpatcher")
				if restoreErr := a.restoreMirrorBackup(backupDir, targetDir, patch.Operations[:i+1]); restoreErr != nil {
					fmt.Printf("Warning: Failed to restore backup: %v\n", restoreErr)
				} else {
					fmt.Println("Backup restored successfully")
				}
			}
			return fmt.Errorf("failed to apply operation %d: %w", i, err)
		}
	}

	// Post-patch verification
	if verifyAfter {
		fmt.Println("Verifying patched version...")
		if err := a.verifyKeyFile(targetDir, patch.ToKeyFile); err != nil {
			// Post-verification failed - automatically restore from backup if it was created
			if createBackup {
				fmt.Println("\nPost-verification failed, automatically restoring from backup...")
				backupDir := filepath.Join(targetDir, "backup.cyberpatcher")
				if restoreErr := a.restoreMirrorBackup(backupDir, targetDir, patch.Operations); restoreErr != nil {
					fmt.Printf("Warning: Failed to restore backup: %v\n", restoreErr)
				} else {
					fmt.Println("Backup restored successfully")
				}
			}
			return fmt.Errorf("post-patch key file verification failed: %w", err)
		}

		if err := a.verifyPatchedFiles(targetDir, patch.Operations); err != nil {
			// Post-verification failed - automatically restore from backup if it was created
			if createBackup {
				fmt.Println("\nPost-verification failed, automatically restoring from backup...")
				backupDir := filepath.Join(targetDir, "backup.cyberpatcher")
				if restoreErr := a.restoreMirrorBackup(backupDir, targetDir, patch.Operations); restoreErr != nil {
					fmt.Printf("Warning: Failed to restore backup: %v\n", restoreErr)
				} else {
					fmt.Println("Backup restored successfully")
				}
			}
			return fmt.Errorf("post-patch verification failed: %w", err)
		}
		fmt.Println("Post-patch verification successful")
	}

	// Keep backup after successful patching for manual rollback if needed
	if createBackup {
		backupDir := filepath.Join(targetDir, "backup.cyberpatcher")
		fmt.Printf("\nBackup preserved at: %s\n", backupDir)
		fmt.Println("To rollback: Copy files from backup folder to their original locations")
	}

	fmt.Println("Patch applied successfully")
	return nil
}

// applyOperation applies a single patch operation
func (a *Applier) applyOperation(targetDir string, op utils.PatchOperation) error {
	targetPath := filepath.Join(targetDir, op.FilePath)

	switch op.Type {
	case utils.OpAdd:
		return a.applyAdd(targetPath, op)
	case utils.OpModify:
		return a.applyModify(targetPath, op)
	case utils.OpDelete:
		return a.applyDelete(targetPath, op)
	case utils.OpAddDir:
		return a.applyAddDir(targetPath)
	case utils.OpDeleteDir:
		return a.applyDeleteDir(targetPath)
	default:
		return fmt.Errorf("unknown operation type: %d", op.Type)
	}
}

// applyAdd adds a new file
func (a *Applier) applyAdd(targetPath string, op utils.PatchOperation) error {
	// Ensure directory exists
	if err := utils.EnsureDir(filepath.Dir(targetPath)); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write new file
	if err := os.WriteFile(targetPath, op.NewFile, 0644); err != nil {
		return fmt.Errorf("failed to write new file: %w", err)
	}

	// Verify checksum
	match, err := utils.VerifyFileChecksum(targetPath, op.NewChecksum)
	if err != nil {
		return fmt.Errorf("failed to verify checksum after add: %w", err)
	} else if !match {
		return fmt.Errorf("checksum verification failed after add")
	}

	fmt.Printf("  Added: %s\n", op.FilePath)
	return nil
}

// applyModify modifies an existing file
func (a *Applier) applyModify(targetPath string, op utils.PatchOperation) error {
	// Verify old checksum
	match, err := utils.VerifyFileChecksum(targetPath, op.OldChecksum)
	if err != nil {
		return fmt.Errorf("failed to verify old file checksum: %w", err)
	} else if !match {
		return fmt.Errorf("old file checksum mismatch")
	}

	var newData []byte

	if len(op.BinaryDiff) > 0 {
		// Apply binary diff
		var patchErr error
		newData, patchErr = a.differ.ApplyPatch(targetPath, op.BinaryDiff)
		if patchErr != nil {
			return fmt.Errorf("failed to apply binary diff: %w", patchErr)
		}
	} else if len(op.NewFile) > 0 {
		// Use full file replacement
		newData = op.NewFile
	} else {
		return fmt.Errorf("no diff or new file data provided")
	}

	// Write modified file
	if err := os.WriteFile(targetPath, newData, 0644); err != nil {
		return fmt.Errorf("failed to write modified file: %w", err)
	}

	// Verify new checksum
	match, err = utils.VerifyFileChecksum(targetPath, op.NewChecksum)
	if err != nil {
		return fmt.Errorf("failed to verify checksum after modify: %w", err)
	} else if !match {
		return fmt.Errorf("checksum verification failed after modify")
	}

	fmt.Printf("  Modified: %s\n", op.FilePath)
	return nil
}

// applyDelete deletes a file
func (a *Applier) applyDelete(targetPath string, op utils.PatchOperation) error {
	// Verify file exists and has correct checksum
	if utils.FileExists(targetPath) {
		match, err := utils.VerifyFileChecksum(targetPath, op.OldChecksum)
		if err != nil {
			return fmt.Errorf("failed to verify file checksum before delete: %w", err)
		} else if !match {
			return fmt.Errorf("file checksum mismatch before delete")
		}

		if err := os.Remove(targetPath); err != nil {
			return fmt.Errorf("failed to delete file: %w", err)
		}

		fmt.Printf("  Deleted: %s\n", op.FilePath)
	}

	return nil
}

// applyAddDir creates a new directory
func (a *Applier) applyAddDir(targetPath string) error {
	if err := utils.EnsureDir(targetPath); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	fmt.Printf("  Created directory: %s\n", targetPath)
	return nil
}

// applyDeleteDir removes a directory and all its contents
func (a *Applier) applyDeleteDir(targetPath string) error {
	if utils.FileExists(targetPath) {
		if err := os.RemoveAll(targetPath); err != nil {
			return fmt.Errorf("failed to delete directory: %w", err)
		}

		fmt.Printf("  Deleted directory: %s\n", targetPath)
	}

	return nil
}

// verifyKeyFile verifies the key file matches expected hash
func (a *Applier) verifyKeyFile(targetDir string, keyFile utils.KeyFileInfo) error {
	keyFilePath := filepath.Join(targetDir, keyFile.Path)

	if !utils.FileExists(keyFilePath) {
		return fmt.Errorf("key file not found: %s", keyFile.Path)
	}

	match, err := utils.VerifyFileChecksum(keyFilePath, keyFile.Checksum)
	if err != nil {
		return fmt.Errorf("failed to verify key file checksum: %w", err)
	} else if !match {
		currentChecksum, _ := utils.CalculateFileChecksum(keyFilePath)
		return fmt.Errorf("key file checksum mismatch: expected %s, got %s",
			keyFile.Checksum[:16], currentChecksum[:16])
	}

	return nil
}

// verifyRequiredFiles verifies all required files exist with correct checksums
func (a *Applier) verifyRequiredFiles(targetDir string, required []utils.FileRequirement) error {
	mismatches := make([]string, 0)

	for _, req := range required {
		if !req.IsRequired {
			continue
		}

		filePath := filepath.Join(targetDir, req.Path)

		if !utils.FileExists(filePath) {
			mismatches = append(mismatches, fmt.Sprintf("%s: file not found", req.Path))
			continue
		}

		match, err := utils.VerifyFileChecksum(filePath, req.Checksum)
		if err != nil {
			mismatches = append(mismatches, fmt.Sprintf("%s: failed to verify checksum: %v", req.Path, err))
			continue
		} else if !match {
			currentChecksum, _ := utils.CalculateFileChecksum(filePath)
			mismatches = append(mismatches, fmt.Sprintf("%s: checksum mismatch (expected %s, got %s)",
				req.Path, req.Checksum[:16], currentChecksum[:16]))
		}
	}

	if len(mismatches) > 0 {
		return fmt.Errorf("found %d mismatches:\n%v", len(mismatches), mismatches)
	}

	return nil
}

// verifyPatchedFiles verifies all modified files have correct checksums
func (a *Applier) verifyPatchedFiles(targetDir string, operations []utils.PatchOperation) error {
	mismatches := make([]string, 0)

	for _, op := range operations {
		if op.Type == utils.OpDelete || op.Type == utils.OpDeleteDir || op.Type == utils.OpAddDir {
			continue
		}

		filePath := filepath.Join(targetDir, op.FilePath)

		if !utils.FileExists(filePath) {
			mismatches = append(mismatches, fmt.Sprintf("%s: file not found after patching", op.FilePath))
			continue
		}

		expectedChecksum := op.NewChecksum
		match, err := utils.VerifyFileChecksum(filePath, expectedChecksum)
		if err != nil {
			mismatches = append(mismatches, fmt.Sprintf("%s: failed to verify checksum: %v", op.FilePath, err))
		} else if !match {
			currentChecksum, _ := utils.CalculateFileChecksum(filePath)
			mismatches = append(mismatches, fmt.Sprintf("%s: checksum mismatch (expected %s, got %s)",
				op.FilePath, expectedChecksum[:16], currentChecksum[:16]))
		}
	}

	if len(mismatches) > 0 {
		return fmt.Errorf("found %d mismatches:\n%v", len(mismatches), mismatches)
	}

	return nil
}

// restoreMirrorBackup restores files from a selective backup created by createMirrorBackup
// This restores only the files that were backed up, putting them back in their original locations
// It also cleans up any files/directories that were added during the failed patch application
func (a *Applier) restoreMirrorBackup(backupDir, targetDir string, operations []utils.PatchOperation) error {
	if !utils.FileExists(backupDir) {
		return fmt.Errorf("backup directory does not exist: %s", backupDir)
	}

	restoredCount := 0
	cleanedCount := 0

	// First, restore files that were backed up (modified/deleted files)
	for _, op := range operations {
		if op.Type == utils.OpModify || op.Type == utils.OpDelete {
			// Restore individual files
			backupPath := filepath.Join(backupDir, op.FilePath)
			targetPath := filepath.Join(targetDir, op.FilePath)

			if !utils.FileExists(backupPath) {
				continue // File wasn't backed up, skip
			}

			// Ensure target directory exists
			if err := utils.EnsureDir(filepath.Dir(targetPath)); err != nil {
				return fmt.Errorf("failed to create target directory for %s: %w", op.FilePath, err)
			}

			// Copy file back from backup
			if err := utils.CopyFile(backupPath, targetPath); err != nil {
				return fmt.Errorf("failed to restore file %s: %w", op.FilePath, err)
			}

			restoredCount++

		} else if op.Type == utils.OpDeleteDir {
			// Restore entire directory that was deleted
			backupPath := filepath.Join(backupDir, op.FilePath)
			targetPath := filepath.Join(targetDir, op.FilePath)

			if !utils.FileExists(backupPath) {
				continue // Directory wasn't backed up, skip
			}

			// Copy entire directory back from backup
			if err := utils.CopyDir(backupPath, targetPath); err != nil {
				return fmt.Errorf("failed to restore directory %s: %w", op.FilePath, err)
			}

			// Count files in restored directory
			fileCount, err := utils.CountFilesInDir(targetPath)
			if err != nil {
				return fmt.Errorf("failed to count files in restored directory %s: %w", op.FilePath, err)
			}

			restoredCount += fileCount
		}
	}

	// Second, clean up any files/directories that were added during the failed patch
	for _, op := range operations {
		if op.Type == utils.OpAdd {
			// Remove newly added files
			targetPath := filepath.Join(targetDir, op.FilePath)
			if utils.FileExists(targetPath) {
				if err := os.Remove(targetPath); err != nil {
					return fmt.Errorf("failed to remove added file %s during rollback: %w", op.FilePath, err)
				}
				cleanedCount++
			}
		} else if op.Type == utils.OpAddDir {
			// Remove newly added directories
			targetPath := filepath.Join(targetDir, op.FilePath)
			if utils.FileExists(targetPath) {
				if err := os.RemoveAll(targetPath); err != nil {
					return fmt.Errorf("failed to remove added directory %s during rollback: %w", op.FilePath, err)
				}
				cleanedCount++
			}
		}
	}

	if restoredCount > 0 {
		fmt.Printf("Restored %d files from backup\n", restoredCount)
	}
	if cleanedCount > 0 {
		fmt.Printf("Cleaned up %d added files/directories\n", cleanedCount)
	}

	return nil
}

// createMirrorBackup creates a selective backup of only files/directories that will be modified or deleted
// The backup mirrors the directory structure for easy manual rollback
func (a *Applier) createMirrorBackup(targetDir, backupDir string, operations []utils.PatchOperation) error {
	// Remove existing backup if it exists
	if utils.FileExists(backupDir) {
		if err := os.RemoveAll(backupDir); err != nil {
			return fmt.Errorf("failed to remove existing backup: %w", err)
		}
	}

	// Create root backup directory
	if err := utils.EnsureDir(backupDir); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Backup only files and directories that will be modified or deleted
	backedUpFileCount := 0
	backedUpDirCount := 0

	for _, op := range operations {
		if op.Type == utils.OpModify || op.Type == utils.OpDelete {
			// Backup individual files that will be modified or deleted
			srcPath := filepath.Join(targetDir, op.FilePath)
			dstPath := filepath.Join(backupDir, op.FilePath)

			// Skip if source file doesn't exist (shouldn't happen, but be safe)
			if !utils.FileExists(srcPath) {
				continue
			}

			// Create parent directories in backup (mirror structure)
			if err := utils.EnsureDir(filepath.Dir(dstPath)); err != nil {
				return fmt.Errorf("failed to create backup subdirectory for %s: %w", op.FilePath, err)
			}

			// Copy the file to backup location
			if err := utils.CopyFile(srcPath, dstPath); err != nil {
				return fmt.Errorf("failed to backup file %s: %w", op.FilePath, err)
			}

			backedUpFileCount++

		} else if op.Type == utils.OpDeleteDir {
			// Backup entire directory that will be deleted (with all contents)
			srcPath := filepath.Join(targetDir, op.FilePath)
			dstPath := filepath.Join(backupDir, op.FilePath)

			// Skip if source directory doesn't exist
			if !utils.FileExists(srcPath) {
				continue
			}

			// Copy entire directory tree to backup
			if err := utils.CopyDir(srcPath, dstPath); err != nil {
				return fmt.Errorf("failed to backup directory %s: %w", op.FilePath, err)
			}

			// Count files in backed up directory
			fileCount, err := utils.CountFilesInDir(dstPath)
			if err != nil {
				return fmt.Errorf("failed to count files in backed up directory %s: %w", op.FilePath, err)
			}

			backedUpFileCount += fileCount
			backedUpDirCount++
		}
	}

	if backedUpDirCount > 0 {
		fmt.Printf("Backed up %d files and %d directories\n", backedUpFileCount, backedUpDirCount)
	} else {
		fmt.Printf("Backed up %d files\n", backedUpFileCount)
	}

	return nil
}
