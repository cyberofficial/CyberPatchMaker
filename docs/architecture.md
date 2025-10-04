# System Architecture

CyberPatchMaker is designed as a modular, maintainable system with clear separation of concerns and robust error handling.

## High-Level Overview

```
┌─────────────────────────────────────────────────────────────┐
│                     User Interface Layer                     │
├─────────────────────────────────────────────────────────────┤
│  CLI Tools                                                    │
│  ├─ generator.exe/generator    (Patch Generation)            │
│  └─ applier.exe/applier        (Patch Application)           │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                    Core Business Logic                       │
├─────────────────────────────────────────────────────────────┤
│  internal/core/                                               │
│  ├─ scanner/      Directory scanning & hashing               │
│  ├─ manifest/     Manifest creation & comparison             │
│  ├─ version/      Version management & registry              │
│  ├─ config/       Configuration management                   │
│  ├─ differ/       Binary diff generation (bsdiff)            │
│  └─ patcher/      Patch generation & application             │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                    Shared Utilities Layer                    │
├─────────────────────────────────────────────────────────────┤
│  pkg/utils/                                                   │
│  ├─ types.go      Core data structures                       │
│  ├─ checksum.go   SHA-256 calculation                        │
│  ├─ fileops.go    File operations (copy, ensure dir)         │
│  └─ compress.go   Compression (zstd, gzip)                   │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                   External Dependencies                      │
├─────────────────────────────────────────────────────────────┤
│  • github.com/gabstv/go-bsdiff       Binary diffing          │
│  • github.com/klauspost/compress     zstd compression        │
│  • Standard library                   File I/O, crypto       │
└─────────────────────────────────────────────────────────────┘
```

## Component Architecture

### 1. CLI Tools (`cmd/`)

**Purpose**: User-facing command-line interfaces

#### Generator (`cmd/generator/main.go`)
- Parses command-line flags
- Validates input parameters
- Delegates to core business logic
- Displays progress and results

**Key Responsibilities:**
- Flag parsing and validation
- User input/output
- Error presentation
- Progress reporting

**Lines of Code**: ~300 lines

#### Applier (`cmd/applier/main.go`)
- Parses command-line flags
- Validates patch and target directory
- Delegates to patcher for application
- Handles backup restoration on failure

**Key Responsibilities:**
- Flag parsing and validation
- Patch file loading
- Dry-run mode coordination
- Backup restoration (fallback)
- User feedback

**Lines of Code**: ~300 lines

---

### 2. Core Business Logic (`internal/core/`)

#### Scanner (`internal/core/scanner/`)

**Purpose**: Recursively scan directory trees and calculate file hashes

**Key Components:**
- `scanner.go`: Directory traversal
- Recursive file discovery
- SHA-256 hash calculation
- Manifest generation

**Responsibilities:**
- Walk entire directory tree
- Hash every file with SHA-256
- Identify empty directories
- Build complete file manifests
- Handle symbolic links and special files

**Lines of Code**: ~200 lines

---

#### Manifest (`internal/core/manifest/`)

**Purpose**: Create, load, and compare version manifests

**Key Components:**
- `manifest.go`: Manifest operations
- JSON serialization/deserialization
- Manifest comparison logic
- Change detection

**Data Structure:**
```go
type Manifest struct {
    Version     string        // "1.0.0"
    KeyFile     KeyFileInfo   // Main program identifier
    Files       []FileEntry   // All files in tree
    Timestamp   time.Time     // Creation time
    TotalSize   int64         // Total bytes
    Checksum    string        // Overall version hash
}
```

**Responsibilities:**
- Create manifest from scanned directory
- Save manifest to disk (JSON)
- Load manifest from disk
- Compare two manifests
- Identify added/modified/deleted files

**Lines of Code**: ~250 lines

---

#### Version (`internal/core/version/`)

**Purpose**: Manage version registry and version metadata

**Key Components:**
- `manager.go`: Version operations
- Version registration
- Version lookup
- Key file identification

**Data Structure:**
```go
type VersionEntry struct {
    Version   string       // "1.0.0"
    Location  string       // Absolute path
    Manifest  *Manifest    // Cached manifest
    ScannedAt time.Time    // Last scan time
}
```

**Responsibilities:**
- Register new versions
- Auto-detect key files
- Load existing versions
- Validate version directories
- Manage version registry

**Lines of Code**: ~200 lines

---

#### Config (`internal/core/config/`)

**Purpose**: Manage application configuration and settings

**Key Components:**
- `config.go`: Configuration management
- Default settings
- Config persistence

**Data Structure:**
```go
type Config struct {
    DefaultPatchOutput string
    TempDirectory      string
    WorkerThreads      int
    EnableParallel     bool
    // ... compression settings, etc.
}
```

**Lines of Code**: ~150 lines

---

#### Differ (`internal/core/differ/`)

**Purpose**: Generate binary diffs between file versions

**Key Components:**
- `differ.go`: Binary diff wrapper
- bsdiff algorithm integration
- Diff optimization

**Responsibilities:**
- Generate binary diff using bsdiff
- Optimize diff size
- Handle large files efficiently
- Skip binary-identical files

**External Dependency**: `github.com/gabstv/go-bsdiff`

**Lines of Code**: ~100 lines

---

#### Patcher (`internal/core/patcher/`)

**Purpose**: Generate and apply patches with full verification

**Key Components:**
- `applier.go`: Complete patch lifecycle
- Pre-verification
- Backup management
- Operation application
- Post-verification
- Rollback handling

**Key Methods:**
- `GeneratePatch(from, to *Manifest) (*Patch, error)`
- `ApplyPatch(patch, targetDir string, verify, backup bool) error`
- `verifyKeyFile(dir string, keyFile KeyFileInfo) error`
- `verifyRequiredFiles(dir string, files []FileRequirement) error`
- `verifyPatchedFiles(dir string, ops []PatchOperation) error`
- `createBackup(srcDir, backupDir string) error`
- `copyDir(src, dst string) error`

**Responsibilities:**
- Generate patch from two manifests
- Apply patch operations
- Verify current version (pre-verification)
- Create backup after verification
- Apply add/modify/delete operations
- Verify patched version (post-verification)
- Cleanup backup on success
- Restore backup on failure

**Lines of Code**: ~400 lines (most complex component)

---

### 3. Shared Utilities (`pkg/utils/`)

**Purpose**: Reusable utility functions used across the system

#### types.go
**Defines core data structures:**
- `Patch`: Complete patch structure
- `PatchOperation`: Single operation (add/modify/delete)
- `FileEntry`: File metadata
- `KeyFileInfo`: Key file identification
- `Manifest`: Version manifest
- Enums: `OperationType`, `CompressionType`

**Lines of Code**: ~200 lines

---

#### checksum.go
**Hash calculation utilities:**
- `CalculateFileChecksum(path string) (string, error)`
- `CalculateDirectoryChecksum(dir string) (string, error)`
- SHA-256 implementation
- Efficient file streaming

**Lines of Code**: ~100 lines

---

#### fileops.go
**File operation utilities:**
- `CopyFile(src, dst string) error`
- `EnsureDir(path string) error`
- `FileExists(path string) bool`
- `RemoveAll(path string) error`
- Cross-platform path handling

**Lines of Code**: ~150 lines

---

#### compress.go
**Compression/decompression:**
- `Compress(data []byte, method string, level int) ([]byte, error)`
- `Decompress(data []byte, method string) ([]byte, error)`
- zstd support
- gzip support

**External Dependencies**: 
- `github.com/klauspost/compress/zstd`
- Standard library `compress/gzip`

**Lines of Code**: ~150 lines

---

## Data Flow

### Patch Generation Flow

```
User Input (CLI)
    ↓
Generator Main (cmd/generator/main.go)
    ↓
Version Manager (internal/core/version/manager.go)
    ├─ Load/Register Versions
    └─ Identify Key Files
    ↓
Scanner (internal/core/scanner/scanner.go)
    ├─ Walk Directory Tree
    ├─ Calculate File Hashes
    └─ Build Manifest
    ↓
Manifest Comparator (internal/core/manifest/manifest.go)
    ├─ Compare Manifests
    └─ Identify Changes
    ↓
Differ (internal/core/differ/differ.go)
    ├─ Generate Binary Diffs
    └─ Optimize Sizes
    ↓
Patcher (internal/core/patcher/applier.go)
    ├─ Create Patch Structure
    ├─ Package Operations
    └─ Add Metadata
    ↓
Compression (pkg/utils/compress.go)
    ├─ Compress Patch Data
    └─ Write to File
    ↓
Output: Patch File (.patch)
```

---

### Patch Application Flow

```
User Input (CLI)
    ↓
Applier Main (cmd/applier/main.go)
    ├─ Parse Flags
    ├─ Load Patch File
    └─ Display Info
    ↓
Patcher.ApplyPatch (internal/core/patcher/applier.go)
    ↓
1. PRE-VERIFICATION
    ├─ Verify Target Directory Exists
    ├─ Verify Key File (checksum)
    └─ Verify All Required Files (checksums)
    ↓
2. BACKUP CREATION (if enabled)
    ├─ Create Backup Directory
    └─ Recursively Copy All Files
    ↓
3. APPLY OPERATIONS
    ├─ For Each Operation:
    │   ├─ Add New Files
    │   ├─ Modify Files (apply binary diffs)
    │   └─ Delete Files/Directories
    └─ Report Progress
    ↓
4. POST-VERIFICATION
    ├─ Verify All Modified Files (checksums)
    └─ Verify Key File Matches Target
    ↓
5. CLEANUP
    ├─ If Success: Remove Backup
    └─ If Failure: Restore from Backup (in main.go)
    ↓
Output: Success/Error Message
```

---

## Design Patterns

### 1. Separation of Concerns
- **CLI Layer**: User interaction only
- **Core Layer**: Business logic only
- **Utils Layer**: Reusable utilities only

### 2. Dependency Injection
- Components receive dependencies via constructors
- Enables testing and modularity

### 3. Error Propagation
- Errors bubble up with context
- `fmt.Errorf` with `%w` for error wrapping
- Clear error messages at CLI layer

### 4. Single Responsibility
- Each component has one clear purpose
- Functions are small and focused
- No god objects or classes

### 5. Encapsulation
- Backup logic owned by patcher
- Configuration owned by config manager
- Version metadata owned by version manager

---

## Security Considerations

### Hash Verification
- SHA-256 used for all file hashing
- Pre-verification prevents wrong patch application
- Post-verification ensures correct patching

### Atomic Operations
- Backups created before modifications
- Operations applied to verified clean state
- Automatic rollback on failure

### Path Safety
- `filepath.Join` for cross-platform paths
- No string concatenation with path separators
- Path validation before operations

---

## Performance Characteristics

### Time Complexity
- **Directory Scan**: O(n) where n = number of files
- **Hash Calculation**: O(m) where m = total file size
- **Manifest Comparison**: O(n) where n = number of files
- **Patch Generation**: O(k) where k = size of changed files
- **Patch Application**: O(k) where k = number of operations

### Space Complexity
- **Manifest**: O(n) where n = number of files
- **Patch**: O(k) where k = size of changes
- **Backup**: O(m) where m = installation size
- **Memory**: O(1) - streaming I/O for large files

### Optimizations
- Streaming for large files (no full load)
- Skip binary-identical files
- Parallel processing option (planned)
- Incremental checksums (planned)

---

## Testing Architecture

### Test Suite Location
- `test.ps1` (Windows PowerShell)
- `test.sh` (Linux/macOS Bash)

### Test Data
- `testdata/versions/` - Sample versions
- `testdata/test-output/` - Test execution workspace

### Test Coverage
1. Build validation
2. Environment setup
3. Test data verification
4. Patch generation (zstd)
5. Dry-run mode
6. Patch application
7. File verification
8. Error handling (rejection)
9. Compression alternatives (gzip)
10. End-to-end validation

See [Testing Guide](testing-guide.md) for details.

---

## Future Architecture

### Planned Phase 4: GUI Application
```
┌─────────────────────────────────────────────────────────────┐
│                       GUI Layer (Fyne)                       │
├─────────────────────────────────────────────────────────────┤
│  cmd/patch-gui/                                               │
│  ├─ mainwindow.go      Main application window               │
│  ├─ versionpanel.go    Version management UI                 │
│  ├─ patchgenpanel.go   Patch generation UI                   │
│  ├─ patchapplypanel.go Patch application UI                  │
│  └─ progress.go        Progress dialogs                      │
└─────────────────────────────────────────────────────────────┘
                              ↓
        (Reuses all existing core business logic)
```

### Planned Enhancements
- Parallel file processing
- Resume interrupted operations
- Network path support optimization
- Delta-of-delta compression
- Patch chaining (1.0.0→1.0.1→1.0.2)

---

## Related Documentation

- [Code Structure](code-structure.md) - Detailed file organization
- [Data Structures](data-structures.md) - Core types explained
- [Backup Lifecycle](backup-lifecycle.md) - Backup timing details
- [Performance](performance.md) - Optimization techniques
