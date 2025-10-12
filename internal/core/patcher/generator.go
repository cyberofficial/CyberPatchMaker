package patcher

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/cyberofficial/cyberpatchmaker/internal/core/differ"
	"github.com/cyberofficial/cyberpatchmaker/internal/core/manifest"
	"github.com/cyberofficial/cyberpatchmaker/pkg/utils"
)

// Generator handles patch generation
type Generator struct {
	differ          *differ.Differ
	manifestManager *manifest.Manager
}

// NewGenerator creates a new patch generator
func NewGenerator() *Generator {
	return &Generator{
		differ:          differ.NewDiffer(),
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

		// Check if file is large (>1GB)
		isLarge, fileSize, err := g.differ.IsLargeFile(fullPath)
		if err != nil {
			return nil, fmt.Errorf("failed to check file size for %s: %w", file.Path, err)
		}

		var fileData []byte
		if isLarge {
			// For large files, use chunked copying to avoid memory issues
			fmt.Printf("  Large file detected (%d MB), using chunked copy: %s\n", fileSize/(1024*1024), file.Path)
			// For added files, we still need the full data but we'll read it in chunks
			tmpFile, err := os.CreateTemp("", "cyberpatch-add-*.tmp")
			if err != nil {
				return nil, fmt.Errorf("failed to create temp file for %s: %w", file.Path, err)
			}
			defer os.Remove(tmpFile.Name())
			defer tmpFile.Close()

			// Copy file in chunks to temp location
			progressCallback := func(processed, total int64) {
				percent := float64(processed) / float64(total) * 100
				fmt.Printf("\r  Progress: %.1f%% (%d/%d MB)", percent, processed/(1024*1024), total/(1024*1024))
			}
			err = g.differ.CopyFileChunked(fullPath, tmpFile.Name(), utils.ChunkSize, progressCallback)
			if err != nil {
				return nil, fmt.Errorf("failed to copy large file %s: %w", file.Path, err)
			}
			fmt.Println() // New line after progress

			// Read the temp file data (now we have it all in one place)
			fileData, err = os.ReadFile(tmpFile.Name())
			if err != nil {
				return nil, fmt.Errorf("failed to read temp file for %s: %w", file.Path, err)
			}
		} else {
			// Read normal-sized file directly
			fileData, err = os.ReadFile(fullPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read new file %s: %w", file.Path, err)
			}
		}

		patch.Operations = append(patch.Operations, utils.PatchOperation{
			Type:        utils.OpAdd,
			FilePath:    file.Path,
			NewFile:     fileData,
			NewChecksum: file.Checksum,
			Size:        file.Size,
		})

		if isLarge {
			fmt.Printf("  Add (large): %s (%d MB)\n", file.Path, file.Size/(1024*1024))
		} else {
			fmt.Printf("  Add: %s (%d bytes)\n", file.Path, file.Size)
		}
	}

	// Process modified files
	totalModified := len(modified)
	if totalModified > 0 {
		fmt.Printf("Processing %d modified files (generating diffs)...\n", totalModified)
	}
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

		// Check if either old or new file is large
		oldPath := filepath.Join(fromVersion.Location, file.Path)
		newPath := filepath.Join(toVersion.Location, file.Path)

		isOldLarge, oldSize, err := g.differ.IsLargeFile(oldPath)
		if err != nil {
			return nil, fmt.Errorf("failed to check old file size for %s: %w", file.Path, err)
		}

		isNewLarge, newSize, err := g.differ.IsLargeFile(newPath)
		if err != nil {
			return nil, fmt.Errorf("failed to check new file size for %s: %w", file.Path, err)
		}

		var diffData []byte

		// Use chunked processing for large files
		if isOldLarge || isNewLarge {
			fmt.Printf("  Large file detected (old: %d MB, new: %d MB), using chunked diff: %s\n",
				oldSize/(1024*1024), newSize/(1024*1024), file.Path)

			progressCallback := func(processed, total int64) {
				percent := float64(processed) / float64(total) * 100
				fmt.Printf("\r  Progress: %.1f%% (%d/%d MB)", percent, processed/(1024*1024), total/(1024*1024))
			}

			diffData, err = g.differ.GenerateDiffChunked(oldPath, newPath, utils.ChunkSize, progressCallback)
			if err != nil {
				return nil, fmt.Errorf("failed to generate chunked diff for %s: %w", file.Path, err)
			}
			fmt.Println() // New line after progress
		} else {
			// Use standard bsdiff for normal-sized files
			diffData, err = g.differ.GenerateDiff(oldPath, newPath)
			if err != nil {
				return nil, fmt.Errorf("failed to generate diff for %s: %w", file.Path, err)
			}
		}

		patch.Operations = append(patch.Operations, utils.PatchOperation{
			Type:        utils.OpModify,
			FilePath:    file.Path,
			BinaryDiff:  diffData,
			OldChecksum: sourceFile.Checksum,
			NewChecksum: file.Checksum,
			Size:        int64(len(diffData)),
		})

		if isOldLarge || isNewLarge {
			fmt.Printf("  Modify (chunked diff): %s (diff: %d MB, orig: %d MB, new: %d MB)\n",
				file.Path, len(diffData)/(1024*1024), sourceFile.Size/(1024*1024), file.Size/(1024*1024))
		} else {
			fmt.Printf("  Modify (diff): %s (diff: %d bytes, orig: %d bytes, new: %d bytes)\n",
				file.Path, len(diffData), sourceFile.Size, file.Size)
		}
	}

	// Create patch header
	patch.Header = utils.PatchHeader{
		FormatVersion: 1,
		CreatedAt:     time.Now(),
		Compression:   options.Compression,
		PatchSize:     g.calculatePatchSize(patch),
		Checksum:      "", // Will be calculated when saving
	}

	fmt.Printf("Patch generation complete: %d operations\n", len(patch.Operations))

	return patch, nil
}

// calculatePatchSize calculates the total size of patch operations
func (g *Generator) calculatePatchSize(patch *utils.Patch) int64 {
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
			if len(op.NewFile) == 0 {
				return fmt.Errorf("operation %d (add): new file data is empty", i)
			}
			if op.NewChecksum == "" {
				return fmt.Errorf("operation %d (add): new checksum is empty", i)
			}
		case utils.OpModify:
			if len(op.BinaryDiff) == 0 && len(op.NewFile) == 0 {
				return fmt.Errorf("operation %d (modify): both diff and new file are empty", i)
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
