# System Architecture

## High-Level Design

```
User Input (CLI)
    |
patch-gen (cmd/generator/)          patch-apply (cmd/applier/)
    |                                     |
Version Manager (version/)            Patcher.Applier (patcher/)
    |  register, scan, cache               |  verify, backup, apply, rollback
Scanner (scanner/)                    +--> Differ (differ/ -- bsdiff wrapper)
    |  walk tree, hash files
Manifest Manager (manifest/)
    |  create, compare, detect changes
Differ (differ/)
    |  bsdiff wrapper (available but NOT used by generator)
    v
patcher.Generator -> streaming JSON + compression -> .patch file
```

## Component Responsibilities

### CLI Tools (`cmd/`)
- `generator/main.go`: flag parsing, version registration, patch generation, self-contained EXE creation
- `applier/main.go`: flag parsing, patch loading, embedded patch detection, interactive/silent/simple mode dispatch

### Core Logic (`internal/core/`)

**Version (`version/`)**: Manages version registry. `RegisterVersion()` scans directories, creates manifests, integrates scan cache. Supports parallel scanning via `SetWorkerThreads()`. Key file auto-detection (program.exe > game.exe > app.exe > main.exe) is handled by the CLI layer in `cmd/generator/main.go` before calling `RegisterVersion()`.

**Patcher (`patcher/`)**: `generator.go` — compares manifests, reads all added/modified files into memory as full replacements (no bsdiff), builds `Patch` struct. `applier.go` — pre-verification, selective backup, operation application, post-verification, automatic rollback on failure. `multipart.go` — splits large patches into parts, chunk sidecar system.

**Scanner (`scanner/`)**: Recursive directory traversal, SHA-256 hashing, `.cyberignore` pattern matching, backup folder exclusion. Supports parallel checksum computation via worker pool.

**Manifest (`manifest/`)**: Creates JSON manifests from scanned directories, loads/saves/compares manifests. Overall checksum via concatenation of sorted file checksums.

**Cache (`cache/`)**: Scan result caching to `.data/` JSON files. Key file hash validation prevents stale cache usage. Location-based hashing for isolation.

**Config (`config/`)**: Application configuration load/save with platform-specific paths.

**Differ (`differ/`)**: bsdiff/bspatch wrapper. Available but the generator currently uses full file replacement (`os.ReadFile`) for all files, not binary diffs.

### Utilities (`pkg/utils/`)

- `types.go`: All shared data structures (Version, Patch, PatchOperation, Manifest, FileEntry, Config, etc.)
- `checksum.go`: SHA-256 file/data/string hashing and verification
- `fileops.go`: CopyFile, EnsureDir, RemoveDir, CopyDir, FileExists, IsExecutable
- `compress.go`: zstd/gzip compression/decompression (in-memory and streaming)
- `patch_io.go`: SavePatch/LoadPatch with streaming JSON encoding and auto-compression detection

## Data Flow

### Patch Generation
```
1. Scan source directory -> build manifest with file hashes
2. Scan target directory -> build manifest with file hashes
3. Compare manifests -> identify added/modified/deleted files and directories
4. For each modified/added file: read full content via os.ReadFile
5. Package everything into Patch struct with operations
6. Stream JSON encode -> compress -> write .patch file
```

### Patch Application
```
1. Load patch file (auto-detect compression)
2. Pre-verify: key file hash matches + all required files match
3. Create selective backup to backup.cyberpatcher/
4. Apply operations in order (add dirs first, delete files, delete dirs deepest-first, add files, modify files last)
5. Post-verify: modified files match target hashes
6. On failure: automatic rollback from backup
7. On success: preserve backup for manual rollback
```

## Self-Contained Executable Format

`patch-gen --create-exe` writes: `[patch-apply.exe] [patch data] [sidecar blob] [128-byte header]`

128-byte header at end of file (little-endian):
- Bytes 0-7: Magic `CPMPATCH`
- Bytes 8-11: Version uint32 (currently 1)
- Bytes 12-19: StubSize uint64
- Bytes 20-27: DataOffset uint64 (== StubSize)
- Bytes 28-35: DataSize uint64
- Bytes 36-51: Compression type string
- Bytes 52-83: SHA-256 checksum of patch data
- Byte 84: Flags (bit 0 = silent mode embedded)
- Bytes 85-127: Reserved

Applier detects by reading last 128 bytes of its own file, validating magic, version, bounds, and checksum.

## External Dependencies

- `github.com/klauspost/compress` — zstd compression
