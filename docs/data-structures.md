# Data Structures

Complete reference for all data structures used in CyberPatchMaker.

## Core Types (`pkg/utils/types.go`)

### Version

Represents a registered software version that can be used for patch generation.

```go
type Version struct {
    Number       string      // Version number (e.g., "1.0.0")
    Location     string      // Absolute path to version directory
    KeyFile      KeyFileInfo // Key file for version identification
    Manifest     *Manifest   // Complete file manifest
    RegisteredAt time.Time   // When version was registered
    LastScanned  time.Time   // When manifest was last updated
}
```

**Usage:**
- Created by version manager during registration
- Stored in version registry
- Used as input for patch generation

---

### KeyFileInfo

Identifies the main executable for version verification.

```go
type KeyFileInfo struct {
    Path     string // Relative path from version root (e.g., "program.exe")
    Checksum string // SHA-256 hash of the key file
    Size     int64  // File size in bytes
}
```

**Purpose:**
- Serves as unique identifier for a version
- Verified before and after patching
- Used to detect wrong version application

**Detection Priority:**
1. `program.exe`
2. `game.exe`
3. `app.exe`
4. `main.exe`

---

### Manifest

Describes the complete contents of a version directory tree.

```go
type Manifest struct {
    Version     string       // Version number
    KeyFile     KeyFileInfo  // Key file information
    Files       []FileEntry  // ALL files in the entire directory tree
    Directories []string     // All directories (for empty dir handling)
    Timestamp   time.Time    // When manifest was created
    TotalSize   int64        // Total size of all files combined
    TotalFiles  int          // Total number of files
    Checksum    string       // Overall version checksum
}
```

**Checksum Calculation:**
- Concatenates all file checksums (sorted by path)
- Calculates SHA-256 of the concatenated string
- Provides unique fingerprint for entire version

---

### FileEntry

Represents a single file in the directory tree.

```go
type FileEntry struct {
    Path         string    // Relative file path from version root
    Size         int64     // File size in bytes
    Checksum     string    // SHA-256 hash
    ModTime      time.Time // Modification time
    IsExecutable bool      // Executable flag (platform-specific)
}
```

**Path Format:**
- Always uses forward slashes
- Relative to version root
- Example: `data/textures/player.png`

---

### Patch

Represents a delta between two versions.

```go
type Patch struct {
    Header        PatchHeader        // Patch metadata
    FromVersion   string             // Source version number
    ToVersion     string             // Target version number
    FromKeyFile   KeyFileInfo        // Source key file verification
    ToKeyFile     KeyFileInfo        // Target key file verification
    RequiredFiles []FileRequirement  // Files that MUST exist with exact hashes
    Operations    []PatchOperation   // List of changes to apply
    SimpleMode    bool               // Simplified UI for end users
    MultiPart     *MultiPartInfo     // Multi-part metadata (nil if single-part)
}
```

**Storage Format:**
- JSON serialization
- Optionally compressed (zstd, gzip)
- Can be embedded in self-contained executables

---

### PatchOperation

Represents a single change operation in a patch.

```go
type PatchOperation struct {
    Type        OperationType // Add, Modify, Delete, AddDir, DeleteDir
    FilePath    string        // Relative file path
    BinaryDiff  []byte        // Binary diff data (for modify) - small files
    NewFile     []byte        // Full file data (for add/modify) - all data
    OldChecksum string        // Expected checksum before patch
    NewChecksum string        // Expected checksum after patch
    Size        int64         // Operation size in bytes
}
```

**Operation Types:**
- `OpAdd` (0): Add new file
- `OpModify` (1): Modify existing file
- `OpDelete` (2): Delete file
- `OpAddDir` (3): Create directory
- `OpDeleteDir` (4): Delete directory

**Data Storage Strategy:**
- All modified files: Use `NewFile` with full replacement
- `BinaryDiff` is never populated by the current generator flow and is reserved for future use

---

### PatchHeader

Contains patch-level metadata.

```go
type PatchHeader struct {
    FormatVersion int       // Patch format version (currently 1)
    CreatedAt     time.Time // Creation timestamp
    Compression   string    // Compression algorithm: "zstd", "gzip", "none"
    PatchSize     int64     // Compressed patch size in bytes
    Checksum      string    // SHA-256 of patch data
    Signature     []byte    // Digital signature (optional, future feature)
}
```

---

### MultiPartInfo

Contains metadata for multi-part patches.

```go
type MultiPartInfo struct {
    IsMultiPart bool       // True if this is a multi-part patch
    PartNumber  int        // Current part number (1-indexed)
    TotalParts  int        // Total number of parts
    PartHashes  []PartHash // Hashes of all parts for verification (only in part 1)
    MaxPartSize int64      // Maximum size per part (default 4GB)
}
```

**Usage:**
- Only present when patch is split into multiple parts
- Part 1 contains the complete `PartHashes` array
- Other parts have `nil` for `PartHashes`

---

### PartHash

Stores hash information for a patch part.

```go
type PartHash struct {
    PartNumber int    // Part number (1-indexed)
    Checksum   string // SHA-256 hash of the part file
    Size       int64  // Part file size in bytes
}
```

---

### PartChunk

Describes a chunk of a larger part file when further splitting is needed.

```go
type PartChunk struct {
    PartNumber  int    // Parent part number (1-indexed)
    ChunkNumber int    // Chunk order within the part (1-indexed)
    FileName    string // Filename of the chunk (relative to patch directory)
    Checksum    string // SHA-256 checksum of this chunk
    Size        int64  // Size in bytes of this chunk
}
```

**Usage:**
- When a single part exceeds ~3.75GB
- Chunks are stored separately with `.chunks.json` sidecar
- Reassembled during patch loading

---

### FileRequirement

Specifies a file that must exist with exact hash.

```go
type FileRequirement struct {
    Path       string // Relative file path
    Checksum   string // Required SHA-256 hash
    Size       int64  // Expected file size
    IsRequired bool   // If true, patch fails if file missing/wrong
}
```

**Purpose:**
- Pre-verification before patch application
- Ensures source version matches expected state
- Prevents applying patches to wrong versions

---

### PatchOptions

Configures patch generation behavior.

```go
type PatchOptions struct {
    Compression       string // "zstd", "gzip", "none"
    CompressionLevel  int    // 1-4 for zstd, 1-3 for gzip
    GenerateSignature bool   // Create digital signature (future)
    ParallelWorkers   int    // Number of parallel workers
    SkipIdentical     bool   // Skip binary-identical files
}
```

---

### Config

Stores application configuration.

```go
type Config struct {
    VersionRegistry    map[string]*Version // Registered versions
    DefaultPatchOutput string              // Default output directory
    TempDirectory      string              // Temp file location
    WorkerThreads      int                 // Parallel workers
    EnableParallel     bool                // Use parallel processing
    SkipIdentical      bool                // Skip binary-identical files
    PreservePerms      bool                // Preserve file permissions
    VerifySignatures   bool                // Verify patch signatures
    SigningKeyPath     string              // Path to signing key
}
```

**Default Values:**
- `WorkerThreads`: `runtime.NumCPU()`
- `EnableParallel`: `true`
- `SkipIdentical`: `true`
- `PreservePerms`: `true`
- `VerifySignatures`: `false`

---

### VersionRegistry

Tracks all registered versions.

```go
type VersionRegistry struct {
    Versions map[string]*Version // Key: version number
}
```

---

### OperationType

Defines the type of patch operation.

```go
type OperationType int

const (
    OpAdd       OperationType = iota // 0: Add new file
    OpModify                         // 1: Modify existing file
    OpDelete                         // 2: Delete file
    OpAddDir                         // 3: Add directory
    OpDeleteDir                      // 4: Delete directory
)
```

---

## Constants

### Memory Optimization

```go
const (
    ChunkSize          = 128 * 1024 * 1024 // 128 MB per chunk
    LargeFileThreshold = 1024 * 1024 * 1024 // 1 GB threshold
    DefaultMaxPartSize = 4 * 1024 * 1024 * 1024 // 4 GB per part
)
```

**Usage:**
- `ChunkSize`: Size of chunks when processing large files
- `LargeFileThreshold`: Files larger than this use chunked processing
- `DefaultMaxPartSize`: Patches larger than this are split

---

## Cached Scan Types (`internal/core/cache/`)

### CachedScan

Represents a complete cached scan result.

```go
type CachedScan struct {
    Version     string       // Version number
    Location    string       // Directory that was scanned
    KeyFile     utils.KeyFileInfo
    Manifest    *utils.Manifest
    CachedAt    time.Time
    LocationHash string      // Hash of location path for validation
}
```

### CachedScanInfo

Summary information about a cached scan.

```go
type CachedScanInfo struct {
    Version    string
    Location   string
    CachedAt   time.Time
    TotalFiles int
    TotalSize  int64
}
```

---

## Related Documentation

- [Architecture](architecture.md) - System design and code organization
