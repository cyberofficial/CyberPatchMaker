package scanner

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// IgnorePatterns holds patterns to exclude from scanning
type IgnorePatterns struct {
	patterns []string
}

// NewIgnorePatterns creates a new IgnorePatterns instance
func NewIgnorePatterns() *IgnorePatterns {
	return &IgnorePatterns{
		patterns: make([]string, 0),
	}
}

// LoadFromFile loads ignore patterns from a .cyberignore file
func (ip *IgnorePatterns) LoadFromFile(rootPath string) error {
	ignorePath := filepath.Join(rootPath, ".cyberignore")

	// Check if file exists
	if _, err := os.Stat(ignorePath); os.IsNotExist(err) {
		// No .cyberignore file, that's okay
		return nil
	}

	file, err := os.Open(ignorePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines
		if line == "" {
			continue
		}

		// Skip comments (lines starting with ::)
		if strings.HasPrefix(line, "::") {
			continue
		}

		// Normalize path separators to forward slashes
		line = strings.ReplaceAll(line, "\\", "/")

		// Add pattern
		ip.patterns = append(ip.patterns, line)
	}

	return scanner.Err()
}

// ShouldIgnore checks if a path should be ignored based on loaded patterns
func (ip *IgnorePatterns) ShouldIgnore(relPath string) bool {
	// Normalize the path to forward slashes
	normalizedPath := strings.ReplaceAll(relPath, "\\", "/")

	// Always ignore .cyberignore file itself
	if normalizedPath == ".cyberignore" {
		return true
	}

	for _, pattern := range ip.patterns {
		if ip.matchPattern(normalizedPath, pattern) {
			return true
		}
	}

	return false
} // matchPattern checks if a path matches a pattern
func (ip *IgnorePatterns) matchPattern(path, pattern string) bool {
	// Exact match
	if path == pattern {
		return true
	}

	// Check if pattern is a directory (ends with /)
	if strings.HasSuffix(pattern, "/") {
		// Match if path starts with this directory
		if strings.HasPrefix(path, pattern) {
			return true
		}
		// Also match the directory itself without trailing slash
		if path == strings.TrimSuffix(pattern, "/") {
			return true
		}
	}

	// Check if path starts with pattern (for directory patterns without trailing /)
	// This handles patterns like "folder" matching "folder/file.txt"
	if strings.HasPrefix(path, pattern+"/") {
		return true
	}

	// Handle wildcard patterns (*.ext)
	if strings.Contains(pattern, "*") {
		// Simple glob matching
		matched, err := filepath.Match(pattern, filepath.Base(path))
		if err == nil && matched {
			return true
		}

		// Also try matching against the full path for patterns with directories
		matched, err = filepath.Match(pattern, path)
		if err == nil && matched {
			return true
		}

		// Handle patterns like "*.txt" that should match "folder/file.txt"
		if strings.HasPrefix(pattern, "*.") {
			ext := strings.TrimPrefix(pattern, "*")
			if strings.HasSuffix(path, ext) {
				return true
			}
		}
	}

	return false
}

// HasPatterns returns true if any patterns are loaded
func (ip *IgnorePatterns) HasPatterns() bool {
	return len(ip.patterns) > 0
}

// GetPatterns returns all loaded patterns (for debugging/testing)
func (ip *IgnorePatterns) GetPatterns() []string {
	return ip.patterns
}
