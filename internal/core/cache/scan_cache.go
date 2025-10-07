package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cyberofficial/cyberpatchmaker/pkg/utils"
)

// ScanCache manages cached directory scans
type ScanCache struct {
	cacheDir string
}

// NewScanCache creates a new scan cache manager
func NewScanCache(cacheDir string) *ScanCache {
	// If cacheDir is empty, use default .data directory in current working directory
	if cacheDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			// Fallback to executable directory
			exe, exeErr := os.Executable()
			if exeErr == nil {
				cwd = filepath.Dir(exe)
			} else {
				cwd = "."
			}
		}
		cacheDir = filepath.Join(cwd, ".data")
	}

	return &ScanCache{
		cacheDir: cacheDir,
	}
}

// GetCacheDir returns the cache directory path
func (sc *ScanCache) GetCacheDir() string {
	return sc.cacheDir
}

// SaveScan saves a scan result to cache
func (sc *ScanCache) SaveScan(version *utils.Version) error {
	if version.Manifest == nil {
		return fmt.Errorf("version manifest is nil")
	}

	// Ensure cache directory exists
	if err := utils.EnsureDir(sc.cacheDir); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Generate cache filename from version number and location
	cacheFilename := sc.generateCacheFilename(version.Number, version.Location)
	cachePath := filepath.Join(sc.cacheDir, cacheFilename)

	// Create cache entry with all necessary information
	cacheEntry := CachedScan{
		Version:      version.Number,
		Location:     version.Location,
		KeyFile:      version.KeyFile,
		Manifest:     version.Manifest,
		CachedAt:     version.LastScanned,
		LocationHash: sc.hashLocation(version.Location),
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(cacheEntry, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache entry: %w", err)
	}

	// Write to file
	if err := os.WriteFile(cachePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// LoadScan loads a scan result from cache
func (sc *ScanCache) LoadScan(versionNumber, location string) (*utils.Version, error) {
	cacheFilename := sc.generateCacheFilename(versionNumber, location)
	cachePath := filepath.Join(sc.cacheDir, cacheFilename)

	// Check if cache file exists
	if !utils.FileExists(cachePath) {
		return nil, fmt.Errorf("no cached scan found for version %s at %s", versionNumber, location)
	}

	// Read cache file
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read cache file: %w", err)
	}

	// Unmarshal cache entry
	var cacheEntry CachedScan
	if err := json.Unmarshal(data, &cacheEntry); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cache entry: %w", err)
	}

	// Validate location hash matches
	if cacheEntry.LocationHash != sc.hashLocation(location) {
		return nil, fmt.Errorf("location hash mismatch - cache may be for different directory")
	}

	// Create version from cache
	version := &utils.Version{
		Number:       cacheEntry.Version,
		Location:     cacheEntry.Location,
		KeyFile:      cacheEntry.KeyFile,
		Manifest:     cacheEntry.Manifest,
		RegisteredAt: cacheEntry.CachedAt,
		LastScanned:  cacheEntry.CachedAt,
	}

	return version, nil
}

// HasCachedScan checks if a cached scan exists
func (sc *ScanCache) HasCachedScan(versionNumber, location string) bool {
	cacheFilename := sc.generateCacheFilename(versionNumber, location)
	cachePath := filepath.Join(sc.cacheDir, cacheFilename)
	return utils.FileExists(cachePath)
}

// DeleteScan removes a cached scan
func (sc *ScanCache) DeleteScan(versionNumber, location string) error {
	cacheFilename := sc.generateCacheFilename(versionNumber, location)
	cachePath := filepath.Join(sc.cacheDir, cacheFilename)

	if !utils.FileExists(cachePath) {
		return nil // Already deleted
	}

	if err := os.Remove(cachePath); err != nil {
		return fmt.Errorf("failed to delete cache file: %w", err)
	}

	return nil
}

// ClearCache removes all cached scans
func (sc *ScanCache) ClearCache() error {
	if !utils.FileExists(sc.cacheDir) {
		return nil // Nothing to clear
	}

	// Read all files in cache directory
	entries, err := os.ReadDir(sc.cacheDir)
	if err != nil {
		return fmt.Errorf("failed to read cache directory: %w", err)
	}

	// Delete all .json cache files
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			filePath := filepath.Join(sc.cacheDir, entry.Name())
			if err := os.Remove(filePath); err != nil {
				return fmt.Errorf("failed to delete cache file %s: %w", entry.Name(), err)
			}
		}
	}

	return nil
}

// ListCachedScans returns a list of all cached scans
func (sc *ScanCache) ListCachedScans() ([]CachedScanInfo, error) {
	if !utils.FileExists(sc.cacheDir) {
		return []CachedScanInfo{}, nil
	}

	entries, err := os.ReadDir(sc.cacheDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read cache directory: %w", err)
	}

	var cachedScans []CachedScanInfo
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		cachePath := filepath.Join(sc.cacheDir, entry.Name())
		data, err := os.ReadFile(cachePath)
		if err != nil {
			continue // Skip files we can't read
		}

		var cacheEntry CachedScan
		if err := json.Unmarshal(data, &cacheEntry); err != nil {
			continue // Skip invalid cache files
		}

		info := CachedScanInfo{
			Version:    cacheEntry.Version,
			Location:   cacheEntry.Location,
			CachedAt:   cacheEntry.CachedAt,
			TotalFiles: cacheEntry.Manifest.TotalFiles,
			TotalSize:  cacheEntry.Manifest.TotalSize,
		}
		cachedScans = append(cachedScans, info)
	}

	return cachedScans, nil
}

// generateCacheFilename creates a unique cache filename
func (sc *ScanCache) generateCacheFilename(versionNumber, location string) string {
	// Create a safe filename from version number and location hash
	safeVersion := strings.ReplaceAll(versionNumber, string(filepath.Separator), "_")
	safeVersion = strings.ReplaceAll(safeVersion, ":", "_")
	safeVersion = strings.ReplaceAll(safeVersion, " ", "_")

	locationHash := sc.hashLocation(location)[:16] // First 16 chars of hash

	return fmt.Sprintf("scan_%s_%s.json", safeVersion, locationHash)
}

// hashLocation creates a consistent hash of a location path
func (sc *ScanCache) hashLocation(location string) string {
	// Normalize path (convert to absolute, clean, and use consistent separators)
	absPath, err := filepath.Abs(location)
	if err != nil {
		absPath = location
	}
	absPath = filepath.Clean(absPath)
	absPath = filepath.ToSlash(absPath)
	absPath = strings.ToLower(absPath) // Case-insensitive

	// Calculate hash
	return utils.CalculateStringChecksum(absPath)
}

// CachedScan represents a cached scan result
type CachedScan struct {
	Version      string            `json:"version"`
	Location     string            `json:"location"`
	KeyFile      utils.KeyFileInfo `json:"key_file"`
	Manifest     *utils.Manifest   `json:"manifest"`
	CachedAt     time.Time         `json:"cached_at"`
	LocationHash string            `json:"location_hash"`
}

// CachedScanInfo provides summary information about a cached scan
type CachedScanInfo struct {
	Version    string    `json:"version"`
	Location   string    `json:"location"`
	CachedAt   time.Time `json:"cached_at"`
	TotalFiles int       `json:"total_files"`
	TotalSize  int64     `json:"total_size"`
}
