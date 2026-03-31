# Code Structure

Detailed organization of CyberPatchMaker source code.

## Directory Layout

```
CyberPatchMaker/
├── cmd/                        # Command-line interface tools
│   ├── applier/               # Patch application CLI
│   │   └── main.go           # Entry point for patch-apply
│   └── generator/             # Patch generation CLI
│       └── main.go           # Entry point for patch-gen
│
├── internal/                   # Private application code
│   └── core/                  # Core business logic
│       ├── cache/            # Scan caching system
│       │   └── scan_cache.go # Cache management for version scans
│       ├── config/           # Configuration management
│       │   └── config.go     # Config load/save/validate
│       ├── differ/           # Binary diff generation
│       │   └── differ.go     # bsdiff/bspatch wrapper
│       ├── manifest/         # Manifest operations
│       │   └── manager.go    # Manifest create/load/compare
│       ├── patcher/          # Patch generation & application
│       │   ├── applier.go    # Patch application logic
│       │   ├── generator.go  # Patch generation logic
│       │   └── multipart.go  # Multi-part patch handling
│       ├── scanner/          # Directory scanning
│       │   ├── ignore.go     # .cyberignore pattern matching
│       │   ├── parallel.go   # Parallel scanning
│       │   └── scanner.go    # Recursive directory traversal
│       └── version/          # Version management
│           ├── manager.go    # Version registry
│           └── version.go    # Version constants
│
├── pkg/                       # Public utilities
│   └── utils/                # Shared utilities
│       ├── checksum.go       # SHA-256 hashing
│       ├── compress.go       # Compression (zstd/gzip)
│       ├── fileops.go        # File operations
│       ├── patch_io.go       # Patch save/load
│       └── types.go          # Core data structures
│
├── docs/                      # Documentation
├── build/                     # Build output (versioned executables)
├── bin/                       # Compiled binaries (patch-gen.exe, etc.)
├── dist/                      # Distribution packages
├── build.ps1                  # Build script (PowerShell)
├── advanced-test.ps1          # Integration test suite (PowerShell)
├── EXAMPLE_apply_patch.bat    # Example batch file for applying patches
└── go.mod / go.sum           # Go module definitions
```

## File-by-File Organization

### CLI Layer (`cmd/`)

| File | Lines | Purpose |
|------|-------|---------|
| `applier/main.go` | 1,123 | Patch application CLI - parses flags, loads patches, coordinates patcher |
| `generator/main.go` | 1,192 | Patch generation CLI - parses flags, coordinates version manager and patcher |

**Responsibilities:**
- Flag parsing and validation
- User input/output
- Progress reporting
- Error presentation
- Self-contained executable creation
- Embedded patch detection

### Core Layer (`internal/core/`)

#### Cache Module (`cache/`)

| File | Lines | Purpose |
|------|-------|---------|
| `scan_cache.go` | 258 | Cache directory scans for instant reload |

**Exported Types:**
- `ScanCache` - Cache manager
- `CachedScan` - Complete cached scan data
- `CachedScanInfo` - Summary information

**Exported Functions:**
- `NewScanCache(cacheDir string)` - Create cache manager
- `SaveScan(version)` - Save scan to cache
- `LoadScan(versionNumber, location)` - Load with validation
- `HasCachedScan(versionNumber, location)` - Check existence
- `DeleteScan(versionNumber, location)` - Remove entry
- `ClearCache()` - Remove all entries
- `ListCachedScans()` - Get all cached scans

#### Config Module (`config/`)

| File | Lines | Purpose |
|------|-------|---------|
| `config.go` | 203 | Configuration management |

**Exported Types:**
- `Manager` - Configuration manager

**Exported Functions:**
- `NewManager()` - Create config manager
- `Load(configPath)` - Load from file
- `Save()` - Save to file
- `GetConfig()` - Get current config
- `SetConfig(config)` - Set config
- `UpdateConfig(updates)` - Update specific fields
- `ValidateConfig()` - Validate settings
- `GetDefaultConfigPath()` - Platform-specific config path
- `GetDefaultManifestPath()` - Default manifest directory

#### Differ Module (`differ/`)

| File | Lines | Purpose |
|------|-------|---------|
| `differ.go` | 265 | Binary diff generation and application |

**Exported Types:**
- `Differ` - Diff operations handler

**Exported Functions:**
- `NewDiffer()` - Create differ instance
- `GenerateDiff(oldPath, newPath)` - Generate binary diff
- `GenerateDiffFromData(oldData, newData)` - Diff from byte arrays
- `GenerateDiffStreaming(oldPath, newPath, output)` - Stream diff generation
- `IsLargeFile(path)` - Check if file > 1GB, returns (isLarge, size, error)
- `GenerateDiffChunked(oldPath, newPath, chunkSize, callback)` - Chunked diff for large files
- `CopyFileChunked(srcPath, dstPath, chunkSize, callback)` - Copy large file in chunks
- `ApplyPatch(oldPath, patchData)` - Apply patch to data
- `ApplyPatchToData(oldData, patchData)` - Apply patch to bytes
- `ApplyPatchToFile(oldPath, outputPath, patchData)` - Apply patch to file
- `CompareSizes(oldSize, newSize)` - Calculate size difference
- `EstimatePatchSize(oldSize, newSize, similarity)` - Estimate patch size
- `ValidatePatch(patchData)` - Verify bsdiff format

#### Manifest Module (`manifest/`)

| File | Lines | Purpose |
|------|-------|---------|
| `manager.go` | 230 | Manifest operations |

**Exported Types:**
- `Manager` - Manifest manager

**Exported Functions:**
- `NewManager()` - Create manifest manager
- `CreateManifest(version, keyFile, files, directories)` - Create manifest
- `SaveManifest(manifest, filePath)` - Save to JSON
- `LoadManifest(filePath)` - Load from JSON
- `CompareManifests(source, target)` - Find differences
- `VerifyManifest(manifest, basePath)` - Verify all files
- `GetManifestStats(manifest)` - Get statistics

#### Patcher Module (`patcher/`)

| File | Lines | Purpose |
|------|-------|---------|
| `applier.go` | 633 | Patch application |
| `generator.go` | 314 | Patch generation |
| `multipart.go` | 479 | Multi-part patch handling |

**Applier Exported Types:**
- `Applier` - Patch applier

**Applier Exported Functions:**
- `NewApplier()` - Create applier
- `ApplyPatch(patch, targetDir, verifyBefore, verifyAfter, createBackup)` - Main apply function
- `ApplyPatchWithPath(patch, targetDir, patchFilePath, ...)` - Apply with file location for streaming

**Generator Exported Types:**
- `Generator` - Patch generator

**Generator Exported Functions:**
- `NewGenerator()` - Create generator
- `GeneratePatch(fromVersion, toVersion, options)` - Generate patch
- `CalculatePatchSize(patch)` - Calculate patch size
- `ValidatePatch(patch)` - Validate before saving
- `SplitPatchIntoParts(patch, maxPartSize)` - Split into parts
- `SaveMultiPartPatch(parts, basePath, compression, chunkSize)` - Save parts

**Multi-Part Exported Functions:**
- `LoadMultiPartPatch(part1Path)` - Load all parts

#### Scanner Module (`scanner/`)

| File | Lines | Purpose |
|------|-------|---------|
| `scanner.go` | 266 | Directory traversal |
| `ignore.go` | 207 | Pattern matching |
| `parallel.go` | 252 | Parallel scanning |

**Scanner Exported Types:**
- `Scanner` - Directory scanner
- `IgnorePatterns` - Pattern matcher

**Scanner Exported Functions:**
- `NewScanner(rootPath)` - Create scanner
- `ScanDirectory()` - Scan recursively
- `ScanDirectoryWithProgress(callback)` - Scan with progress
- `FindFile(relPath)` - Find specific file
- `ValidatePath()` - Check path validity
- `GetAbsolutePath(relPath)` - Get absolute path
- `NormalizePath(path)` - Normalize path separators

**IgnorePatterns Exported Functions:**
- `NewIgnorePatterns()` - Create pattern matcher
- `LoadFromFile(rootPath)` - Load .cyberignore
- `ShouldIgnore(relPath)` - Check if should ignore
- `ShouldIgnoreWithAbsPath(relPath, absPath)` - Check with absolute path support
- `HasPatterns()` - Check if patterns loaded
- `GetPatterns()` - Get all patterns

**Scanner Exported Methods (parallel):**
- `ScanDirectoryParallel(workers)` - Parallel scan
- `ScanDirectoryParallelWithProgress(workers, callback)` - Parallel with progress

#### Version Module (`version/`)

| File | Lines | Purpose |
|------|-------|---------|
| `manager.go` | 480 | Version registry |
| `version.go` | 87 | Version constants |

**Version Manager Exported Types:**
- `Manager` - Version manager
- `Registry` - Version storage

**Version Manager Exported Functions:**
- `NewManager()` - Create manager
- `SetWorkerThreads(threads)` - Set parallel workers
- `EnableScanCache(cacheDir, forceRescan)` - Enable caching
- `GetScanCache()` - Get cache instance
- `RegisterVersion(versionNumber, location, keyFilePath)` - Register version
- `UnregisterVersion(versionNumber)` - Remove version
- `GetVersion(versionNumber)` - Get version
- `ListVersions()` - List all versions
- `RescanVersion(versionNumber)` - Re-scan version
- `VerifyVersion(versionNumber)` - Verify files
- `SaveRegistry(filePath)` - Save to disk
- `LoadRegistry(filePath)` - Load from disk
- `GetRegistry()` - Get registry

**Version Exported Constants:**
- `Major`, `Minor`, `Patch`, `PreRelease` - Version components

**Version Exported Functions:**
- `GetVersion()` - Get full version string
- `GetShortVersion()` - Get version without pre-release

### Utilities Layer (`pkg/utils/`)

| File | Lines | Purpose |
|------|-------|---------|
| `types.go` | 164 | Core data structures |
| `checksum.go` | 46 | SHA-256 hashing |
| `fileops.go` | 142 | File operations |
| `compress.go` | 238 | Compression/decompression |
| `patch_io.go` | 338 | Patch serialization |

**types.go Exported Types:**
- `Version` - Version information
- `KeyFileInfo` - Key file identification
- `Manifest` - Version manifest
- `FileEntry` - File metadata
- `Patch` - Complete patch
- `MultiPartInfo` - Multi-part metadata
- `PartHash` - Part hash information
- `PartChunk` - Part chunk information
- `FileRequirement` - Required file
- `PatchOperation` - Single operation
- `OperationType` - Operation type enum
- `PatchHeader` - Patch metadata
- `PatchOptions` - Generation options
- `Config` - Application config
- `VersionRegistry` - Version storage

**types.go Constants:**
- `OpAdd`, `OpModify`, `OpDelete`, `OpAddDir`, `OpDeleteDir` - Operation types
- `ChunkSize` - 128MB chunk size
- `LargeFileThreshold` - 1GB threshold
- `DefaultMaxPartSize` - 4GB part size

**checksum.go Exported Functions:**
- `CalculateFileChecksum(path)` - Hash file
- `CalculateDataChecksum(data)` - Hash bytes
- `CalculateStringChecksum(text)` - Hash string
- `VerifyFileChecksum(path, expected)` - Verify file hash

**fileops.go Exported Functions:**
- `CopyFile(src, dst)` - Copy file
- `EnsureDir(path)` - Create directory
- `RemoveDir(path)` - Remove directory
- `FileExists(path)` - Check existence
- `GetFileSize(path)` - Get size
- `IsExecutable(path)` - Check if executable
- `CopyDir(src, dst)` - Copy directory
- `CountFilesInDir(path)` - Count files

**compress.go Exported Functions:**
- `CompressData(data, algorithm, level)` - Compress data
- `DecompressData(data, algorithm)` - Decompress data
- `CompressDataStreaming(src, dst, algorithm, level)` - Stream compress
- `DecompressDataStreaming(src, dst, algorithm)` - Stream decompress

**patch_io.go Exported Functions:**
- `SavePatch(patch, filename, compression)` - Save patch
- `LoadPatch(filename)` - Load patch

## Module Dependencies

```
┌─────────────────────────────────────────────────────────────┐
│                         cmd/                                 │
│  ┌──────────────┐         ┌──────────────┐                 │
│  │   applier/   │         │  generator/  │                 │
│  └──────┬───────┘         └──────┬───────┘                 │
└─────────┼────────────────────────┼──────────────────────────┘
          │                        │
          ▼                        ▼
┌─────────────────────────────────────────────────────────────┐
│                      internal/core/                          │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐   │
│  │ patcher/ │──│ differ/  │  │manifest/ │──│ scanner/ │   │
│  └──────────┘  └──────────┘  └────┬─────┘  └────┬─────┘   │
│       │                             │             │          │
│       └──────────┬──────────────────┴─────────────┘          │
│                  ▼                                          │
│           ┌──────────┐  ┌──────────┐  ┌──────────┐        │
│           │ version/ │──│  cache/  │  │ config/  │        │
│           └──────────┘  └──────────┘  └──────────┘        │
└───────────────────────────┬────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                       pkg/utils/                             │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  │
│  │ types.go │  │checksum.go│ │fileops.go│ │compress.go│  │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘  │
│  ┌────────────────────┐                                   │
│  │   patch_io.go      │                                   │
│  └────────────────────┘                                   │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                    External Dependencies                       │
│  github.com/gabstv/go-bsdiff    (Binary diffing)             │
│  github.com/klauspost/compress  (zstd compression)           │
└─────────────────────────────────────────────────────────────┘
```

## Adding New Features

When adding new features to CyberPatchMaker, follow this decision tree:

1. **Is it user-facing?** → Add to `cmd/` layer
2. **Is it business logic?** → Add to appropriate `internal/core/` module
3. **Is it a reusable utility?** → Add to `pkg/utils/`
4. **Does it modify core types?** → Update `pkg/utils/types.go`

## Related Documentation

- [Architecture](architecture.md) - System design overview
- [Development Setup](development-setup.md) - Setting up development environment
- [Data Structures](data-structures.md) - Core type definitions
