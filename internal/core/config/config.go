package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/cyberofficial/cyberpatchmaker/pkg/utils"
)

// Manager handles application configuration
type Manager struct {
	config     *utils.Config
	configPath string
}

// NewManager creates a new configuration manager
func NewManager() *Manager {
	return &Manager{
		config: getDefaultConfig(),
	}
}

// Load loads configuration from a file
func (m *Manager) Load(configPath string) error {
	m.configPath = configPath

	// Check if config file exists
	if !utils.FileExists(configPath) {
		// Use default config
		return nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var config utils.Config
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	m.config = &config
	return nil
}

// Save saves the current configuration to a file
func (m *Manager) Save() error {
	if m.configPath == "" {
		return fmt.Errorf("config path not set")
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(m.configPath)
	if err := utils.EnsureDir(dir); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(m.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(m.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetConfig returns the current configuration
func (m *Manager) GetConfig() *utils.Config {
	return m.config
}

// SetConfig sets the configuration
func (m *Manager) SetConfig(config *utils.Config) {
	m.config = config
}

// UpdateConfig updates specific configuration fields
func (m *Manager) UpdateConfig(updates map[string]interface{}) error {
	for key, value := range updates {
		switch key {
		case "default_patch_output":
			if v, ok := value.(string); ok {
				m.config.DefaultPatchOutput = v
			}
		case "temp_directory":
			if v, ok := value.(string); ok {
				m.config.TempDirectory = v
			}
		case "worker_threads":
			if v, ok := value.(int); ok {
				m.config.WorkerThreads = v
			}
		case "enable_parallel":
			if v, ok := value.(bool); ok {
				m.config.EnableParallel = v
			}
		case "diff_threshold_kb":
			if v, ok := value.(int); ok {
				m.config.DiffThresholdKB = v
			}
		case "skip_identical":
			if v, ok := value.(bool); ok {
				m.config.SkipIdentical = v
			}
		case "preserve_perms":
			if v, ok := value.(bool); ok {
				m.config.PreservePerms = v
			}
		case "verify_signatures":
			if v, ok := value.(bool); ok {
				m.config.VerifySignatures = v
			}
		case "signing_key_path":
			if v, ok := value.(string); ok {
				m.config.SigningKeyPath = v
			}
		default:
			return fmt.Errorf("unknown config key: %s", key)
		}
	}

	return nil
}

// GetDefaultConfigPath returns the default configuration file path
func GetDefaultConfigPath() string {
	var configDir string

	switch runtime.GOOS {
	case "windows":
		configDir = os.Getenv("APPDATA")
		if configDir == "" {
			configDir = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming")
		}
		configDir = filepath.Join(configDir, "CyberPatchMaker")
	case "darwin":
		configDir = filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "CyberPatchMaker")
	default: // linux and others
		configDir = os.Getenv("XDG_CONFIG_HOME")
		if configDir == "" {
			configDir = filepath.Join(os.Getenv("HOME"), ".config")
		}
		configDir = filepath.Join(configDir, "cyberpatchmaker")
	}

	return filepath.Join(configDir, "config.json")
}

// GetDefaultManifestPath returns the default manifest directory path
func GetDefaultManifestPath() string {
	configPath := GetDefaultConfigPath()
	return filepath.Join(filepath.Dir(configPath), "manifests")
}

// getDefaultConfig returns the default configuration
func getDefaultConfig() *utils.Config {
	var tempDir string
	switch runtime.GOOS {
	case "windows":
		tempDir = filepath.Join(os.Getenv("TEMP"), "cyberpatchmaker")
	default:
		tempDir = filepath.Join("/tmp", "cyberpatchmaker")
	}

	return &utils.Config{
		DefaultPatchOutput: filepath.Join(os.Getenv("HOME"), "patches"),
		TempDirectory:      tempDir,
		WorkerThreads:      runtime.NumCPU(),
		EnableParallel:     true,
		DiffThresholdKB:    1,
		SkipIdentical:      true,
		PreservePerms:      true,
		VerifySignatures:   false,
		SigningKeyPath:     "",
	}
}

// ValidateConfig validates the configuration
func (m *Manager) ValidateConfig() error {
	if m.config.WorkerThreads < 1 {
		return fmt.Errorf("worker_threads must be at least 1")
	}

	if m.config.DiffThresholdKB < 0 {
		return fmt.Errorf("diff_threshold_kb must be non-negative")
	}

	// Validate paths exist or can be created
	paths := []string{
		m.config.DefaultPatchOutput,
		m.config.TempDirectory,
	}

	for _, path := range paths {
		if path != "" {
			if err := utils.EnsureDir(path); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", path, err)
			}
		}
	}

	return nil
}
