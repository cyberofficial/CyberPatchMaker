package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cyberofficial/cyberpatchmaker/pkg/utils"
)

// Scanner handles recursive directory scanning
type Scanner struct {
	rootPath       string
	ignorePatterns *IgnorePatterns
}

// NewScanner creates a new scanner for the given root path
func NewScanner(rootPath string) *Scanner {
	ignorePatterns := NewIgnorePatterns()
	// Try to load .cyberignore file (silently fail if not present)
	ignorePatterns.LoadFromFile(rootPath)

	return &Scanner{
		rootPath:       rootPath,
		ignorePatterns: ignorePatterns,
	}
}

// ScanDirectory recursively scans a directory tree and returns all files
func (s *Scanner) ScanDirectory() ([]utils.FileEntry, []string, error) {
	var files []utils.FileEntry
	var directories []string

	err := filepath.Walk(s.rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing path %s: %w", path, err)
		}

		// Get absolute path
		absPath, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("failed to get absolute path for %s: %w", path, err)
		}

		// Get relative path from root
		relPath, err := filepath.Rel(s.rootPath, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		// Skip the root directory itself
		if relPath == "." {
			return nil
		}

		// Convert to forward slashes for consistency
		relPath = filepath.ToSlash(relPath)

		// Skip backup.cyberpatcher directory and all its contents
		if relPath == "backup.cyberpatcher" || strings.HasPrefix(relPath, "backup.cyberpatcher/") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if path should be ignored based on .cyberignore patterns
		if s.ignorePatterns.ShouldIgnoreWithAbsPath(relPath, absPath) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if info.IsDir() {
			// Track directories for empty directory handling
			directories = append(directories, relPath)
		} else {
			// Calculate checksum for file
			checksum, err := utils.CalculateFileChecksum(path)
			if err != nil {
				return fmt.Errorf("failed to calculate checksum for %s: %w", path, err)
			}

			// Create file entry
			entry := utils.FileEntry{
				Path:         relPath,
				Size:         info.Size(),
				Checksum:     checksum,
				ModTime:      info.ModTime(),
				IsExecutable: utils.IsExecutable(path),
			}

			files = append(files, entry)
		}

		return nil
	})

	if err != nil {
		return nil, nil, fmt.Errorf("failed to scan directory: %w", err)
	}

	return files, directories, nil
}

// ScanDirectoryWithProgress scans directory with progress callback
func (s *Scanner) ScanDirectoryWithProgress(progressCallback func(current, total int, currentFile string)) ([]utils.FileEntry, []string, error) {
	// First pass: count total files
	totalFiles := 0
	filepath.Walk(s.rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		relPath, _ := filepath.Rel(s.rootPath, path)
		relPath = filepath.ToSlash(relPath)
		// Skip backup.cyberpatcher directory
		if relPath == "backup.cyberpatcher" || strings.HasPrefix(relPath, "backup.cyberpatcher/") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		// Check .cyberignore patterns
		absPath, _ := filepath.Abs(path)
		if s.ignorePatterns.ShouldIgnoreWithAbsPath(relPath, absPath) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if !info.IsDir() {
			totalFiles++
		}
		return nil
	})

	var files []utils.FileEntry
	var directories []string
	currentFile := 0

	err := filepath.Walk(s.rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing path %s: %w", path, err)
		}

		relPath, err := filepath.Rel(s.rootPath, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		if relPath == "." {
			return nil
		}

		relPath = filepath.ToSlash(relPath)

		// Skip backup.cyberpatcher directory and all its contents
		if relPath == "backup.cyberpatcher" || strings.HasPrefix(relPath, "backup.cyberpatcher/") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if path should be ignored based on .cyberignore patterns
		if s.ignorePatterns.ShouldIgnoreWithAbsPath(relPath, path) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if info.IsDir() {
			directories = append(directories, relPath)
		} else {
			currentFile++
			if progressCallback != nil {
				progressCallback(currentFile, totalFiles, relPath)
			}

			checksum, err := utils.CalculateFileChecksum(path)
			if err != nil {
				return fmt.Errorf("failed to calculate checksum for %s: %w", path, err)
			}

			entry := utils.FileEntry{
				Path:         relPath,
				Size:         info.Size(),
				Checksum:     checksum,
				ModTime:      info.ModTime(),
				IsExecutable: utils.IsExecutable(path),
			}

			files = append(files, entry)
		}

		return nil
	})

	if err != nil {
		return nil, nil, fmt.Errorf("failed to scan directory: %w", err)
	}

	return files, directories, nil
}

// FindFile searches for a file by relative path
func (s *Scanner) FindFile(relPath string) (utils.FileEntry, error) {
	fullPath := filepath.Join(s.rootPath, filepath.FromSlash(relPath))

	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return utils.FileEntry{}, fmt.Errorf("file not found: %s", relPath)
		}
		return utils.FileEntry{}, fmt.Errorf("failed to stat file: %w", err)
	}

	if info.IsDir() {
		return utils.FileEntry{}, fmt.Errorf("path is a directory, not a file: %s", relPath)
	}

	checksum, err := utils.CalculateFileChecksum(fullPath)
	if err != nil {
		return utils.FileEntry{}, fmt.Errorf("failed to calculate checksum: %w", err)
	}

	entry := utils.FileEntry{
		Path:         relPath,
		Size:         info.Size(),
		Checksum:     checksum,
		ModTime:      info.ModTime(),
		IsExecutable: utils.IsExecutable(fullPath),
	}

	return entry, nil
}

// ValidatePath checks if a path exists and is accessible
func (s *Scanner) ValidatePath() error {
	info, err := os.Stat(s.rootPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("path does not exist: %s", s.rootPath)
		}
		return fmt.Errorf("failed to access path: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", s.rootPath)
	}

	return nil
}

// GetAbsolutePath returns the absolute path for a relative path
func (s *Scanner) GetAbsolutePath(relPath string) string {
	return filepath.Join(s.rootPath, filepath.FromSlash(relPath))
}

// NormalizePath normalizes a path (converts backslashes to forward slashes)
func NormalizePath(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}
