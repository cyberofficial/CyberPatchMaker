package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/cyberofficial/cyberpatchmaker/pkg/utils"
)

// ScanDirectoryParallel scans directory with parallel checksum computation
func (s *Scanner) ScanDirectoryParallel(workers int) ([]utils.FileEntry, []string, error) {
	// First pass: collect all file paths and directories
	var filePaths []string
	var directories []string

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

		if info.IsDir() {
			directories = append(directories, relPath)
		} else {
			filePaths = append(filePaths, path)
		}

		return nil
	})

	if err != nil {
		return nil, nil, fmt.Errorf("failed to scan directory: %w", err)
	}

	// Second pass: compute checksums in parallel
	files := make([]utils.FileEntry, len(filePaths))
	var wg sync.WaitGroup
	var mu sync.Mutex
	errChan := make(chan error, workers)
	jobs := make(chan int, len(filePaths))

	// Start workers
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range jobs {
				path := filePaths[idx]
				info, err := os.Stat(path)
				if err != nil {
					errChan <- fmt.Errorf("failed to stat %s: %w", path, err)
					return
				}

				checksum, err := utils.CalculateFileChecksum(path)
				if err != nil {
					errChan <- fmt.Errorf("failed to calculate checksum for %s: %w", path, err)
					return
				}

				relPath, _ := filepath.Rel(s.rootPath, path)
				relPath = filepath.ToSlash(relPath)

				entry := utils.FileEntry{
					Path:         relPath,
					Size:         info.Size(),
					Checksum:     checksum,
					ModTime:      info.ModTime(),
					IsExecutable: utils.IsExecutable(path),
				}

				mu.Lock()
				files[idx] = entry
				mu.Unlock()
			}
		}()
	}

	// Send jobs
	for idx := range filePaths {
		jobs <- idx
	}
	close(jobs)

	// Wait for completion
	wg.Wait()
	close(errChan)

	// Check for errors
	if len(errChan) > 0 {
		return nil, nil, <-errChan
	}

	return files, directories, nil
}

// ScanDirectoryParallelWithProgress scans directory with parallel checksum computation and progress callback
func (s *Scanner) ScanDirectoryParallelWithProgress(workers int, progressCallback func(current, total int, currentFile string)) ([]utils.FileEntry, []string, error) {
	// First pass: collect all file paths and directories
	var filePaths []string
	var directories []string

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

		if info.IsDir() {
			directories = append(directories, relPath)
		} else {
			filePaths = append(filePaths, path)
		}

		return nil
	})

	if err != nil {
		return nil, nil, fmt.Errorf("failed to scan directory: %w", err)
	}

	totalFiles := len(filePaths)

	// Second pass: compute checksums in parallel
	files := make([]utils.FileEntry, len(filePaths))
	var wg sync.WaitGroup
	var mu sync.Mutex
	var progressMu sync.Mutex
	errChan := make(chan error, workers)
	jobs := make(chan int, len(filePaths))
	currentFile := 0

	// Start workers
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range jobs {
				path := filePaths[idx]
				info, err := os.Stat(path)
				if err != nil {
					errChan <- fmt.Errorf("failed to stat %s: %w", path, err)
					return
				}

				checksum, err := utils.CalculateFileChecksum(path)
				if err != nil {
					errChan <- fmt.Errorf("failed to calculate checksum for %s: %w", path, err)
					return
				}

				relPath, _ := filepath.Rel(s.rootPath, path)
				relPath = filepath.ToSlash(relPath)

				entry := utils.FileEntry{
					Path:         relPath,
					Size:         info.Size(),
					Checksum:     checksum,
					ModTime:      info.ModTime(),
					IsExecutable: utils.IsExecutable(path),
				}

				mu.Lock()
				files[idx] = entry
				mu.Unlock()

				// Update progress
				if progressCallback != nil {
					progressMu.Lock()
					currentFile++
					current := currentFile
					progressMu.Unlock()
					progressCallback(current, totalFiles, relPath)
				}
			}
		}()
	}

	// Send jobs
	for idx := range filePaths {
		jobs <- idx
	}
	close(jobs)

	// Wait for completion
	wg.Wait()
	close(errChan)

	// Check for errors
	if len(errChan) > 0 {
		return nil, nil, <-errChan
	}

	return files, directories, nil
}
