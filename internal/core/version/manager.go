package version

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/cyberofficial/cyberpatchmaker/internal/core/cache"
	"github.com/cyberofficial/cyberpatchmaker/internal/core/manifest"
	"github.com/cyberofficial/cyberpatchmaker/internal/core/scanner"
	"github.com/cyberofficial/cyberpatchmaker/pkg/utils"
)

// Manager handles version registration and management
type Manager struct {
	registry        *Registry
	manifestManager *manifest.Manager
	scanCache       *cache.ScanCache
	useScanCache    bool
	forceRescan     bool
	mu              sync.RWMutex
}

// Registry stores registered versions
type Registry struct {
	Versions map[string]*utils.Version
	mu       sync.RWMutex
}

// NewManager creates a new version manager
func NewManager() *Manager {
	return &Manager{
		registry: &Registry{
			Versions: make(map[string]*utils.Version),
		},
		manifestManager: manifest.NewManager(),
		scanCache:       nil,
		useScanCache:    false,
		forceRescan:     false,
	}
}

// EnableScanCache enables scan caching with the specified cache directory
func (m *Manager) EnableScanCache(cacheDir string, forceRescan bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.scanCache = cache.NewScanCache(cacheDir)
	m.useScanCache = true
	m.forceRescan = forceRescan
}

// GetScanCache returns the scan cache instance
func (m *Manager) GetScanCache() *cache.ScanCache {
	return m.scanCache
}

// RegisterVersion registers a new version with key file
func (m *Manager) RegisterVersion(versionNumber, location, keyFilePath string) (*utils.Version, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if version already exists
	if _, exists := m.registry.Versions[versionNumber]; exists {
		return nil, fmt.Errorf("version %s is already registered", versionNumber)
	}

	// Validate location
	if !utils.FileExists(location) {
		return nil, fmt.Errorf("version location does not exist: %s", location)
	}

	// Try to load from cache if enabled and not forcing rescan
	if m.useScanCache && !m.forceRescan && m.scanCache != nil {
		if m.scanCache.HasCachedScan(versionNumber, location) {
			fmt.Printf("Loading cached scan for version %s...\n", versionNumber)
			cachedVersion, err := m.scanCache.LoadScan(versionNumber, location)
			if err == nil {
				// Verify key file still matches
				keyFilePath := cachedVersion.KeyFile.Path
				fullKeyPath := filepath.Join(location, keyFilePath)
				if utils.FileExists(fullKeyPath) {
					match, _ := utils.VerifyFileChecksum(fullKeyPath, cachedVersion.KeyFile.Checksum)
					if match {
						// Cache is valid, use it
						m.registry.Versions[versionNumber] = cachedVersion
						fmt.Printf("✓ Loaded from cache: %d files, %d directories\n",
							len(cachedVersion.Manifest.Files), len(cachedVersion.Manifest.Directories))
						fmt.Printf("Version %s registered: %d files, %d directories\n",
							versionNumber, len(cachedVersion.Manifest.Files), len(cachedVersion.Manifest.Directories))
						return cachedVersion, nil
					}
				}
				fmt.Printf("Cache invalid (key file changed), rescanning...\n")
			} else {
				fmt.Printf("Failed to load cache: %v, rescanning...\n", err)
			}
		}
	}

	// Scan the directory
	scan := scanner.NewScanner(location)
	if err := scan.ValidatePath(); err != nil {
		return nil, fmt.Errorf("invalid version path: %w", err)
	}

	fmt.Printf("Scanning version %s at %s...\n", versionNumber, location)

	// Track scanning start time for ETA calculation
	startTime := time.Now()

	// Use progress callback to show scan progress with percentage, ETA and elapsed time
	files, directories, err := scan.ScanDirectoryWithProgress(func(current, total int, currentFile string) {
		elapsed := time.Since(startTime).Seconds()
		elapsedStr := formatDuration(elapsed)
		percentage := 0
		if total > 0 {
			percentage = (current * 100) / total
		}
		if current > 0 && elapsed > 0 {
			rate := float64(current) / elapsed
			remaining := float64(total-current) / rate
			eta := formatDuration(remaining)
			fmt.Printf("\rScanning: %d/%d files (%d%%) | Elapsed: %s | ETA: %s                    ", current, total, percentage, elapsedStr, eta)
		} else {
			fmt.Printf("\rScanning: %d/%d files (%d%%) | Elapsed: %s                    ", current, total, percentage, elapsedStr)
		}
	})
	if err != nil {
		fmt.Println() // New line after progress
		return nil, fmt.Errorf("failed to scan version directory: %w", err)
	}
	fmt.Println() // New line after progress completes

	// Find and verify key file
	keyFileEntry, err := scan.FindFile(keyFilePath)
	if err != nil {
		return nil, fmt.Errorf("key file not found: %w", err)
	}

	keyFileInfo := utils.KeyFileInfo{
		Path:     keyFilePath,
		Checksum: keyFileEntry.Checksum,
		Size:     keyFileEntry.Size,
	}

	// Create manifest
	manifestData, err := m.manifestManager.CreateManifest(versionNumber, keyFileInfo, files, directories)
	if err != nil {
		return nil, fmt.Errorf("failed to create manifest: %w", err)
	}

	// Create version
	version := &utils.Version{
		Number:       versionNumber,
		Location:     location,
		KeyFile:      keyFileInfo,
		Manifest:     manifestData,
		RegisteredAt: time.Now(),
		LastScanned:  time.Now(),
	}

	m.registry.Versions[versionNumber] = version

	fmt.Printf("Version %s registered: %d files, %d directories\n",
		versionNumber, len(files), len(directories))

	// Save to cache if enabled
	if m.useScanCache && m.scanCache != nil {
		if err := m.scanCache.SaveScan(version); err != nil {
			fmt.Printf("Warning: Failed to save scan to cache: %v\n", err)
		} else {
			fmt.Printf("✓ Scan cached for future use\n")
		}
	}

	return version, nil
}

// UnregisterVersion removes a version from the registry
func (m *Manager) UnregisterVersion(versionNumber string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.registry.Versions[versionNumber]; !exists {
		return fmt.Errorf("version %s is not registered", versionNumber)
	}

	delete(m.registry.Versions, versionNumber)
	return nil
}

// GetVersion retrieves a registered version
func (m *Manager) GetVersion(versionNumber string) (*utils.Version, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	version, exists := m.registry.Versions[versionNumber]
	if !exists {
		return nil, fmt.Errorf("version %s is not registered", versionNumber)
	}

	return version, nil
}

// ListVersions returns all registered versions
func (m *Manager) ListVersions() []*utils.Version {
	m.mu.RLock()
	defer m.mu.RUnlock()

	versions := make([]*utils.Version, 0, len(m.registry.Versions))
	for _, version := range m.registry.Versions {
		versions = append(versions, version)
	}

	return versions
}

// RescanVersion rescans a version's directory and updates its manifest
func (m *Manager) RescanVersion(versionNumber string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	version, exists := m.registry.Versions[versionNumber]
	if !exists {
		return fmt.Errorf("version %s is not registered", versionNumber)
	}

	fmt.Printf("Rescanning version %s...\n", versionNumber)

	// Scan the directory
	scan := scanner.NewScanner(version.Location)

	// Track scanning start time for ETA calculation
	startTime := time.Now()

	// Use progress callback to show scan progress with percentage, ETA and elapsed time
	files, directories, err := scan.ScanDirectoryWithProgress(func(current, total int, currentFile string) {
		elapsed := time.Since(startTime).Seconds()
		elapsedStr := formatDuration(elapsed)
		percentage := 0
		if total > 0 {
			percentage = (current * 100) / total
		}
		if current > 0 && elapsed > 0 {
			rate := float64(current) / elapsed
			remaining := float64(total-current) / rate
			eta := formatDuration(remaining)
			fmt.Printf("\rScanning: %d/%d files (%d%%) | Elapsed: %s | ETA: %s", current, total, percentage, elapsedStr, eta)
		} else {
			fmt.Printf("\rScanning: %d/%d files (%d%%) | Elapsed: %s", current, total, percentage, elapsedStr)
		}
	})
	if err != nil {
		fmt.Println() // New line after progress
		return fmt.Errorf("failed to scan version directory: %w", err)
	}
	fmt.Println() // New line after progress completes

	// Verify key file still matches
	keyFileEntry, err := scan.FindFile(version.KeyFile.Path)
	if err != nil {
		return fmt.Errorf("key file not found during rescan: %w", err)
	}

	if keyFileEntry.Checksum != version.KeyFile.Checksum {
		return fmt.Errorf("key file has been modified (expected %s, got %s)",
			version.KeyFile.Checksum[:16], keyFileEntry.Checksum[:16])
	}

	// Update manifest
	manifestData, err := m.manifestManager.CreateManifest(versionNumber, version.KeyFile, files, directories)
	if err != nil {
		return fmt.Errorf("failed to create manifest: %w", err)
	}

	version.Manifest = manifestData
	version.LastScanned = time.Now()

	fmt.Printf("Version %s rescanned: %d files, %d directories\n",
		versionNumber, len(files), len(directories))

	return nil
}

// VerifyVersion verifies all files in a version match their checksums
func (m *Manager) VerifyVersion(versionNumber string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	version, exists := m.registry.Versions[versionNumber]
	if !exists {
		return nil, fmt.Errorf("version %s is not registered", versionNumber)
	}

	fmt.Printf("Verifying version %s...\n", versionNumber)

	mismatches, err := m.manifestManager.VerifyManifest(version.Manifest, version.Location)
	if err != nil {
		return nil, fmt.Errorf("failed to verify manifest: %w", err)
	}

	if len(mismatches) == 0 {
		fmt.Printf("Version %s verified successfully - all files match\n", versionNumber)
	} else {
		fmt.Printf("Version %s verification found %d mismatches\n", versionNumber, len(mismatches))
	}

	return mismatches, nil
}

// SaveRegistry saves the version registry to a JSON file
func (m *Manager) SaveRegistry(filePath string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := utils.EnsureDir(dir); err != nil {
		return fmt.Errorf("failed to create registry directory: %w", err)
	}

	// Save each version's manifest separately
	manifestDir := filepath.Join(dir, "manifests")
	if err := utils.EnsureDir(manifestDir); err != nil {
		return fmt.Errorf("failed to create manifest directory: %w", err)
	}

	for versionNum, version := range m.registry.Versions {
		manifestPath := filepath.Join(manifestDir, versionNum+".json")
		if err := m.manifestManager.SaveManifest(version.Manifest, manifestPath); err != nil {
			return fmt.Errorf("failed to save manifest for version %s: %w", versionNum, err)
		}
	}

	return nil
}

// LoadRegistry loads the version registry from a JSON file
func (m *Manager) LoadRegistry(filePath string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	manifestDir := filepath.Join(filepath.Dir(filePath), "manifests")

	// Load all manifest files
	entries, err := os.ReadDir(manifestDir)
	if err != nil {
		if os.IsNotExist(err) {
			// No manifests to load
			return nil
		}
		return fmt.Errorf("failed to read manifest directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		manifestPath := filepath.Join(manifestDir, entry.Name())
		manifestData, err := m.manifestManager.LoadManifest(manifestPath)
		if err != nil {
			fmt.Printf("Warning: failed to load manifest %s: %v\n", entry.Name(), err)
			continue
		}

		// Create version from manifest (location needs to be set separately or stored)
		version := &utils.Version{
			Number:       manifestData.Version,
			Location:     "", // Will need to be configured
			KeyFile:      manifestData.KeyFile,
			Manifest:     manifestData,
			RegisteredAt: manifestData.Timestamp,
			LastScanned:  manifestData.Timestamp,
		}

		m.registry.Versions[version.Number] = version
	}

	return nil
}

// GetRegistry returns the version registry
func (m *Manager) GetRegistry() *Registry {
	return m.registry
}

// formatDuration formats seconds into a human-readable duration string
func formatDuration(seconds float64) string {
	if seconds < 1 {
		return "<1s"
	} else if seconds < 60 {
		return fmt.Sprintf("%ds", int(seconds))
	} else if seconds < 3600 {
		minutes := int(seconds / 60)
		secs := int(seconds) % 60
		if secs > 0 {
			return fmt.Sprintf("%dm %ds", minutes, secs)
		}
		return fmt.Sprintf("%dm", minutes)
	} else {
		hours := int(seconds / 3600)
		minutes := int(seconds/60) % 60
		if minutes > 0 {
			return fmt.Sprintf("%dh %dm", hours, minutes)
		}
		return fmt.Sprintf("%dh", hours)
	}
}
