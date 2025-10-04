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
		backupDir := targetDir + ".backup"
		if err := a.createBackup(targetDir, backupDir); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
		fmt.Printf("Backup created at: %s\n", backupDir)
	}

	// Apply operations
	fmt.Printf("Applying %d operations...\n", len(patch.Operations))
	for i, op := range patch.Operations {
		if err := a.applyOperation(targetDir, op); err != nil {
			return fmt.Errorf("failed to apply operation %d: %w", i, err)
		}
	}

	// Post-patch verification
	if verifyAfter {
		fmt.Println("Verifying patched version...")
		if err := a.verifyKeyFile(targetDir, patch.ToKeyFile); err != nil {
			return fmt.Errorf("post-patch key file verification failed: %w", err)
		}

		if err := a.verifyPatchedFiles(targetDir, patch.Operations); err != nil {
			return fmt.Errorf("post-patch verification failed: %w", err)
		}
		fmt.Println("Post-patch verification successful")
	}

	// Clean up backup if successful
	if createBackup {
		fmt.Println("Removing backup...")
		backupDir := targetDir + ".backup"
		if err := os.RemoveAll(backupDir); err != nil {
			fmt.Printf("Warning: failed to remove backup: %v\n", err)
		}
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

// createBackup creates a backup of the target directory
func (a *Applier) createBackup(srcDir, backupDir string) error {
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
	return a.copyDir(srcDir, backupDir)
}

// copyDir recursively copies a directory
func (a *Applier) copyDir(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := utils.EnsureDir(dstPath); err != nil {
				return err
			}
			if err := a.copyDir(srcPath, dstPath); err != nil {
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
