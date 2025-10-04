package utils

import (
	"time"
)

// Version represents a registered software version
type Version struct {
	Number       string      // Version number (e.g., "1.0.0")
	Location     string      // Absolute path to version directory
	KeyFile      KeyFileInfo // Key file for version identification
	Manifest     *Manifest   // Complete file manifest
	RegisteredAt time.Time   // When version was registered
	LastScanned  time.Time   // When manifest was last updated
}

// KeyFileInfo identifies the main executable for version verification
type KeyFileInfo struct {
	Path     string // Relative path from version root (e.g., "program.exe")
	Checksum string // SHA-256 hash of the key file
	Size     int64  // File size in bytes
}

// Manifest describes the complete contents of a version directory tree
type Manifest struct {
	Version     string      // Version number
	KeyFile     KeyFileInfo // Key file information
	Files       []FileEntry // ALL files in the entire directory tree
	Directories []string    // All directories (for empty dir handling)
	Timestamp   time.Time   // When manifest was created
	TotalSize   int64       // Total size of all files combined
	TotalFiles  int         // Total number of files
	Checksum    string      // Overall version checksum (SHA-256 of all file hashes)
}

// FileEntry represents a single file in the directory tree
type FileEntry struct {
	Path         string    // Relative file path from version root
	Size         int64     // File size in bytes
	Checksum     string    // SHA-256 hash
	ModTime      time.Time // Modification time
	IsExecutable bool      // Executable flag
}

// Patch represents a delta between two versions
type Patch struct {
	Header        PatchHeader       // Patch metadata
	FromVersion   string            // Source version number
	ToVersion     string            // Target version number
	FromKeyFile   KeyFileInfo       // Source key file verification
	ToKeyFile     KeyFileInfo       // Target key file verification
	RequiredFiles []FileRequirement // Files that MUST exist with exact hashes
	Operations    []PatchOperation  // List of changes to apply
}

// FileRequirement specifies a file that must exist with exact hash
type FileRequirement struct {
	Path       string // Relative file path
	Checksum   string // Required SHA-256 hash
	Size       int64  // Expected file size
	IsRequired bool   // If true, patch fails if file missing/wrong
}

// PatchOperation represents a single change operation
type PatchOperation struct {
	Type        OperationType // Add, Modify, Delete, AddDir, DeleteDir
	FilePath    string        // Relative file path
	BinaryDiff  []byte        // Binary diff data (for modify)
	NewFile     []byte        // Full file data (for add)
	OldChecksum string        // Expected checksum before patch
	NewChecksum string        // Expected checksum after patch
	Size        int64         // Operation size
}

// OperationType defines the type of patch operation
type OperationType int

const (
	OpAdd       OperationType = iota // Add new file
	OpModify                         // Modify existing file
	OpDelete                         // Delete file
	OpAddDir                         // Add directory
	OpDeleteDir                      // Delete directory
)

// PatchHeader contains patch-level information
type PatchHeader struct {
	FormatVersion int       // Patch format version
	CreatedAt     time.Time // Creation timestamp
	Compression   string    // Compression algorithm used
	PatchSize     int64     // Compressed patch size
	Checksum      string    // Patch file checksum
	Signature     []byte    // Digital signature (optional)
}

// PatchOptions configures patch generation
type PatchOptions struct {
	Compression       string // "zstd", "gzip", "none"
	CompressionLevel  int    // 1-9 or algorithm-specific
	VerifyAfter       bool   // Verify patch after creation
	GenerateSignature bool   // Create digital signature
	ParallelWorkers   int    // Number of parallel workers
	DiffThresholdKB   int    // Skip diff for files smaller than this
	SkipIdentical     bool   // Skip binary-identical files
}

// Config stores application configuration
type Config struct {
	VersionRegistry    map[string]*Version // Registered versions
	DefaultPatchOutput string              // Default output directory
	TempDirectory      string              // Temp file location
	WorkerThreads      int                 // Parallel workers
	EnableParallel     bool                // Use parallel processing
	DiffThresholdKB    int                 // Skip files smaller than this
	SkipIdentical      bool                // Skip binary-identical files
	PreservePerms      bool                // Preserve file permissions
	VerifySignatures   bool                // Verify patch signatures
	SigningKeyPath     string              // Path to signing key
}

// VersionRegistry tracks all registered versions
type VersionRegistry struct {
	Versions map[string]*Version // Key: version number
}
