package manifest

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/cyberofficial/cyberpatchmaker/pkg/utils"
)

// Manager handles manifest operations
type Manager struct{}

// NewManager creates a new manifest manager
func NewManager() *Manager {
	return &Manager{}
}

// CreateManifest creates a manifest from scanned files
func (m *Manager) CreateManifest(version string, keyFile utils.KeyFileInfo, files []utils.FileEntry, directories []string) (*utils.Manifest, error) {
	if len(files) == 0 {
		return nil, fmt.Errorf("no files provided for manifest")
	}

	// Calculate total size
	var totalSize int64
	for _, file := range files {
		totalSize += file.Size
	}

	// Calculate overall checksum
	overallChecksum, err := calculateOverallChecksum(files)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate overall checksum: %w", err)
	}

	manifest := &utils.Manifest{
		Version:     version,
		KeyFile:     keyFile,
		Files:       files,
		Directories: directories,
		Timestamp:   time.Now(),
		TotalSize:   totalSize,
		TotalFiles:  len(files),
		Checksum:    overallChecksum,
	}

	return manifest, nil
}

// SaveManifest saves a manifest to a JSON file
func (m *Manager) SaveManifest(manifest *utils.Manifest, filePath string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := utils.EnsureDir(dir); err != nil {
		return fmt.Errorf("failed to create manifest directory: %w", err)
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write manifest file: %w", err)
	}

	return nil
}

// LoadManifest loads a manifest from a JSON file
func (m *Manager) LoadManifest(filePath string) (*utils.Manifest, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest file: %w", err)
	}

	var manifest utils.Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to unmarshal manifest: %w", err)
	}

	return &manifest, nil
}

// CompareManifests compares two manifests and returns the differences
func (m *Manager) CompareManifests(source, target *utils.Manifest) (added, modified, deleted []utils.FileEntry) {
	// Create maps for efficient lookup
	sourceFiles := make(map[string]utils.FileEntry)
	targetFiles := make(map[string]utils.FileEntry)

	for _, file := range source.Files {
		sourceFiles[file.Path] = file
	}

	for _, file := range target.Files {
		targetFiles[file.Path] = file
	}

	// Find added and modified files
	for path, targetFile := range targetFiles {
		sourceFile, exists := sourceFiles[path]
		if !exists {
			// File was added
			added = append(added, targetFile)
		} else if sourceFile.Checksum != targetFile.Checksum {
			// File was modified
			modified = append(modified, targetFile)
		}
	}

	// Find deleted files
	for path, sourceFile := range sourceFiles {
		if _, exists := targetFiles[path]; !exists {
			// File was deleted
			deleted = append(deleted, sourceFile)
		}
	}

	return added, modified, deleted
}

// VerifyManifest verifies all files in a manifest match their checksums
func (m *Manager) VerifyManifest(manifest *utils.Manifest, basePath string) ([]string, error) {
	var mismatches []string

	for _, file := range manifest.Files {
		fullPath := filepath.Join(basePath, file.Path)

		// Check if file exists
		if !utils.FileExists(fullPath) {
			mismatches = append(mismatches, fmt.Sprintf("%s: file not found", file.Path))
			continue
		}

		// Verify checksum
		match, err := utils.VerifyFileChecksum(fullPath, file.Checksum)
		if err != nil {
			mismatches = append(mismatches, fmt.Sprintf("%s: failed to calculate checksum: %v", file.Path, err))
		} else if !match {
			currentChecksum, calcErr := utils.CalculateFileChecksum(fullPath)
			if calcErr != nil {
				mismatches = append(mismatches, fmt.Sprintf("%s: failed to calculate checksum: %v", file.Path, calcErr))
			} else {
				mismatches = append(mismatches, fmt.Sprintf("%s: checksum mismatch (expected %s, got %s)",
					file.Path, file.Checksum[:16], currentChecksum[:16]))
			}
		}
	}

	return mismatches, nil
}

// GetManifestStats returns statistics about a manifest
func (m *Manager) GetManifestStats(manifest *utils.Manifest) map[string]interface{} {
	stats := make(map[string]interface{})

	stats["version"] = manifest.Version
	stats["total_files"] = manifest.TotalFiles
	stats["total_size"] = manifest.TotalSize
	stats["total_directories"] = len(manifest.Directories)
	stats["timestamp"] = manifest.Timestamp
	stats["checksum"] = manifest.Checksum

	// Calculate executable count
	executableCount := 0
	for _, file := range manifest.Files {
		if file.IsExecutable {
			executableCount++
		}
	}
	stats["executable_files"] = executableCount

	return stats
}

// calculateOverallChecksum calculates a checksum of all file checksums
func calculateOverallChecksum(files []utils.FileEntry) (string, error) {
	// Sort files by path for consistent ordering
	sortedFiles := make([]utils.FileEntry, len(files))
	copy(sortedFiles, files)
	sort.Slice(sortedFiles, func(i, j int) bool {
		return sortedFiles[i].Path < sortedFiles[j].Path
	})

	// Concatenate all checksums
	hasher := sha256.New()
	for _, file := range sortedFiles {
		if _, err := hasher.Write([]byte(file.Checksum)); err != nil {
			return "", fmt.Errorf("failed to hash checksum: %w", err)
		}
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}
