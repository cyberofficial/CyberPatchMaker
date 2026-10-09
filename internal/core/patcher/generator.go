package patcher

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/cyberofficial/cyberpatchmaker/internal/core/manifest"
	"github.com/cyberofficial/cyberpatchmaker/pkg/utils"
)

// Generator handles patch generation
type Generator struct {
	manifestManager *manifest.Manager
}

// NewGenerator creates a new patch generator
func NewGenerator() *Generator {
	return &Generator{
		manifestManager: manifest.NewManager(),
	}
}

// GeneratePatch generates a patch between two versions
func (g *Generator) GeneratePatch(fromVersion, toVersion *utils.Version, options *utils.PatchOptions) (*utils.Patch, error) {
	fmt.Printf("Generating patch from %s to %s...\n", fromVersion.Number, toVersion.Number)

	// Compare manifests
	added, modified, deleted := g.manifestManager.CompareManifests(fromVersion.Manifest, toVersion.Manifest)

	// Compare directories
	addedDirs, deletedDirs := g.compareDirectories(fromVersion.Manifest.Directories, toVersion.Manifest.Directories)

	fmt.Printf("Changes detected: %d added, %d modified, %d deleted files, %d added dirs, %d deleted dirs\n",
		len(added), len(modified), len(deleted), len(addedDirs), len(deletedDirs))

	// Create patch
	patch := &utils.Patch{
		FromVersion:   fromVersion.Number,
		ToVersion:     toVersion.Number,
		FromKeyFile:   fromVersion.KeyFile,
		ToKeyFile:     toVersion.KeyFile,
		RequiredFiles: make([]utils.FileRequirement, 0),
		Operations:    make([]utils.PatchOperation, 0),
	}

	// Add required files (all files from source version)
	for _, file := range fromVersion.Manifest.Files {
		patch.RequiredFiles = append(patch.RequiredFiles, utils.FileRequirement{
			Path:       file.Path,
			Checksum:   file.Checksum,
			Size:       file.Size,
			IsRequired: true,
		})
	}

	// Process added directories first (before adding files to them)
	for _, dir := range addedDirs {
		patch.Operations = append(patch.Operations, utils.PatchOperation{
			Type:     utils.OpAddDir,
			FilePath: dir,
			Size:     0,
		})
		fmt.Printf("  Add directory: %s\n", dir)
	}

	// Process deleted files
	for _, file := range deleted {
		patch.Operations = append(patch.Operations, utils.PatchOperation{
			Type:        utils.OpDelete,
			FilePath:    file.Path,
			OldChecksum: file.Checksum,
			Size:        0,
		})
	}

	// Process deleted directories last (after files are deleted, so dirs are empty)
	// Sort deleted directories by depth (deepest first) to ensure child dirs deleted before parents
	sort.Slice(deletedDirs, func(i, j int) bool {
		// Count path separators - more separators = deeper path
		iDepth := strings.Count(deletedDirs[i], string(filepath.Separator))
		jDepth := strings.Count(deletedDirs[j], string(filepath.Separator))
		return iDepth > jDepth // Deeper paths first
	})

	for _, dir := range deletedDirs {
		patch.Operations = append(patch.Operations, utils.PatchOperation{
			Type:     utils.OpDeleteDir,
			FilePath: dir,
			Size:     0,
		})
		fmt.Printf("  Delete directory: %s\n", dir)
	}

	// Process added files
	totalAdded := len(added)
	if totalAdded > 0 {
		fmt.Printf("Processing %d added files...\n", totalAdded)
	}
	for _, file := range added {
		fullPath := filepath.Join(toVersion.Location, file.Path)

		// Read file directly (no streaming for large files)
		fileData, err := os.ReadFile(fullPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read new file %s: %w", file.Path, err)
		}

		// The manifest checksums from the scanner are trusted; no need to re-verify

		patch.Operations = append(patch.Operations, utils.PatchOperation{
			Type:        utils.OpAdd,
			FilePath:    file.Path,
			NewFile:     fileData,
			NewChecksum: file.Checksum,
			Size:        file.Size,
		})

		fmt.Printf("  Add: %s (%d bytes)\n", file.Path, file.Size)
	}

	// Process modified files
	totalModified := len(modified)
	if totalModified > 0 {
		fmt.Printf("Processing %d modified files...\n", totalModified)
	}
	deltaOps, fullOps, savedBytes := 0, 0, int64(0)
	for _, file := range modified {
		// Find corresponding source file
		var sourceFile *utils.FileEntry
		for i := range fromVersion.Manifest.Files {
			if fromVersion.Manifest.Files[i].Path == file.Path {
				sourceFile = &fromVersion.Manifest.Files[i]
				break
			}
		}

		if sourceFile == nil {
			return nil, fmt.Errorf("source file not found: %s", file.Path)
		}

		// Skip identical files if option is enabled
		if options.SkipIdentical && sourceFile.Checksum == file.Checksum {
			continue
		}

		// Get file path
		newPath := filepath.Join(toVersion.Location, file.Path)

		// Read the new file data directly
		newFileData, err := os.ReadFile(newPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read new file %s: %w", file.Path, err)
		}

		op := utils.PatchOperation{
			Type:        utils.OpModify,
			FilePath:    file.Path,
			OldChecksum: sourceFile.Checksum,
			NewChecksum: file.Checksum,
			Size:        file.Size,
		}

		// Try to store the change as a compact delta against the old file;
		// fall back to full file replacement whenever the delta is not
		// strictly smaller or cannot be verified.
		if g.tryDelta(&op, fromVersion.Location, sourceFile, newFileData, options.CompressionLevel) {
			deltaOps++
			savedBytes += int64(len(newFileData)) - int64(len(op.BinaryDiff))
			fmt.Printf("  Modify (delta): %s (%d -> %d bytes, %.1f%% of original)\n",
				file.Path, len(newFileData), len(op.BinaryDiff),
				float64(len(op.BinaryDiff))/float64(len(newFileData))*100)
		} else {
			op.NewFile = newFileData
			fullOps++
			fmt.Printf("  Modify (full replacement): %s (%d bytes)\n", file.Path, file.Size)
		}

		patch.Operations = append(patch.Operations, op)
	}
	if deltaOps > 0 {
		fmt.Printf("Delta compression: %d file(s) stored as deltas, %d as full replacements (delta savings: %d bytes)\n",
			deltaOps, fullOps, savedBytes)
	}

	// Create patch header
	patch.Header = utils.PatchHeader{
		FormatVersion: 2, // 2 = modified files may be stored as zstd dictionary deltas (BinaryDiff)
		CreatedAt:     time.Now(),
		Compression:   options.Compression,
		PatchSize:     g.CalculatePatchSize(patch),
		Checksum:      "", // Will be calculated when saving
	}

	fmt.Printf("Patch generation complete: %d operations\n", len(patch.Operations))

	return patch, nil
}

// tryDelta attempts to store a modification as a compact zstd delta instead
// of a full file replacement: the old file content is used as the encoding
// dictionary for the new file content (the `zstd --patch-from` mechanism).
// On success it fills op.BinaryDiff and sets op.Size to the delta size;
// otherwise the caller keeps the full-replacement behavior. Every stored
// delta is verified to reconstruct the new file exactly before it is
// accepted, and the base content is checked against the manifest checksum so
// the patch only carries deltas that will verify at apply time.
func (g *Generator) tryDelta(op *utils.PatchOperation, fromDir string, sourceFile *utils.FileEntry, newFileData []byte, level int) bool {
	// The base (dictionary) and the new content must be within the delta
	// size limits; anything larger keeps the full-replacement behavior.
	// Both ends are checked against manifest/actual sizes before any
	// extra file reads or encoding work.
	if sourceFile.Size <= 0 || sourceFile.Size > utils.MaxDeltaBaseSize ||
		len(newFileData) == 0 || int64(len(newFileData)) > utils.MaxDeltaTargetSize {
		return false
	}

	oldFileData, err := os.ReadFile(filepath.Join(fromDir, op.FilePath))
	if err != nil {
		fmt.Printf("  Warning: could not read base file for delta (%s), using full replacement: %v\n", op.FilePath, err)
		return false
	}

	// The delta must be built against the exact content the patch will
	// verify at apply time (op.OldChecksum); otherwise fall back.
	if utils.CalculateDataChecksum(oldFileData) != sourceFile.Checksum {
		fmt.Printf("  Warning: base file %s does not match its manifest checksum, using full replacement\n", op.FilePath)
		return false
	}

	delta, err := utils.EncodeDelta(oldFileData, newFileData, level)
	if err != nil {
		fmt.Printf("  Warning: delta encoding failed for %s, using full replacement: %v\n", op.FilePath, err)
		return false
	}

	// Only keep the delta when it is smaller than the full new file.
	if !utils.ShouldUseDelta(int64(len(delta)), int64(len(newFileData))) {
		return false
	}

	// Verify the delta reconstructs the new file exactly before storing it.
	reconstructed, err := utils.DecodeDelta(oldFileData, delta)
	if err != nil || !bytes.Equal(reconstructed, newFileData) {
		fmt.Printf("  Warning: delta verification failed for %s, using full replacement\n", op.FilePath)
		return false
	}

	op.BinaryDiff = delta
	op.Size = int64(len(delta)) // size of the payload this operation carries
	return true
}

// CalculatePatchSize calculates the total size of patch operations
func (g *Generator) CalculatePatchSize(patch *utils.Patch) int64 {
	var totalSize int64
	for _, op := range patch.Operations {
		totalSize += op.Size
		if op.NewFile != nil {
			totalSize += int64(len(op.NewFile))
		}
		if op.BinaryDiff != nil {
			totalSize += int64(len(op.BinaryDiff))
		}
	}
	return totalSize
}

// compareDirectories compares two directory lists and returns added and deleted directories
func (g *Generator) compareDirectories(sourceDirs, targetDirs []string) (added, deleted []string) {
	// Create maps for efficient lookup
	sourceMap := make(map[string]bool)
	targetMap := make(map[string]bool)

	for _, dir := range sourceDirs {
		sourceMap[dir] = true
	}

	for _, dir := range targetDirs {
		targetMap[dir] = true
	}

	// Find added directories
	for _, dir := range targetDirs {
		if !sourceMap[dir] {
			added = append(added, dir)
		}
	}

	// Find deleted directories
	for _, dir := range sourceDirs {
		if !targetMap[dir] {
			deleted = append(deleted, dir)
		}
	}

	return added, deleted
}

// ValidatePatch validates a patch before saving
func (g *Generator) ValidatePatch(patch *utils.Patch) error {
	if patch.FromVersion == "" {
		return fmt.Errorf("source version is empty")
	}
	if patch.ToVersion == "" {
		return fmt.Errorf("target version is empty")
	}
	if patch.FromKeyFile.Checksum == "" {
		return fmt.Errorf("source key file checksum is empty")
	}
	if patch.ToKeyFile.Checksum == "" {
		return fmt.Errorf("target key file checksum is empty")
	}
	if len(patch.Operations) == 0 {
		return fmt.Errorf("patch has no operations")
	}

	// Validate each operation
	for i, op := range patch.Operations {
		if op.FilePath == "" && op.Type != utils.OpAddDir && op.Type != utils.OpDeleteDir {
			return fmt.Errorf("operation %d has empty file path", i)
		}

		switch op.Type {
		case utils.OpAdd:
			if len(op.NewFile) != int(op.Size) {
				return fmt.Errorf("operation %d (add): file data size mismatch (expected %d bytes, got %d bytes)", i, op.Size, len(op.NewFile))
			}
			if op.NewChecksum == "" {
				return fmt.Errorf("operation %d (add): new checksum is empty", i)
			}
		case utils.OpModify:
			if len(op.BinaryDiff) == 0 && len(op.NewFile) == 0 {
				return fmt.Errorf("operation %d (modify): both diff and new file are empty", i)
			}
			if len(op.BinaryDiff) > 0 && len(op.NewFile) > 0 {
				return fmt.Errorf("operation %d (modify): both diff and new file are set", i)
			}
			if len(op.NewFile) > 0 && len(op.NewFile) != int(op.Size) {
				return fmt.Errorf("operation %d (modify): file data size mismatch (expected %d bytes, got %d bytes)", i, op.Size, len(op.NewFile))
			}
			if op.OldChecksum == "" {
				return fmt.Errorf("operation %d (modify): old checksum is empty", i)
			}
			if op.NewChecksum == "" {
				return fmt.Errorf("operation %d (modify): new checksum is empty", i)
			}
		case utils.OpDelete:
			if op.OldChecksum == "" {
				return fmt.Errorf("operation %d (delete): old checksum is empty", i)
			}
		case utils.OpAddDir, utils.OpDeleteDir:
			// Directory operations don't require checksums
			if op.FilePath == "" {
				return fmt.Errorf("operation %d (dir): directory path is empty", i)
			}
		default:
			return fmt.Errorf("operation %d has invalid type: %d", i, op.Type)
		}
	}

	return nil
}
