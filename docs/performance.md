# Performance

Optimization techniques and performance characteristics of CyberPatchMaker.

## Overview

CyberPatchMaker is designed for performance across multiple dimensions:
- **Fast patch generation**: Scan caching and parallel hashing minimize time to create patches
- **Small patch sizes**: Compression reduces bandwidth for distribution
- **Quick patch application**: Only changed files are backed up, written, and verified

Note: generation and application hold patch data in memory (see below), so very large changed files require commensurate RAM.

## Real-World Example

The following figures are a real-world observation from scanning a large game installation (War Thunder), not a benchmark suite:

| Metric | Value |
|--------|-------|
| **Application Size** | 56 GB |
| **File Count** | 34,650 files |
| **Initial Scan Time** | ~15 minutes |
| **Cached Scan Load** | <1 second |

## Optimization Techniques

### 1. Scan Caching

**Problem**: Directory scanning with SHA-256 hashing is CPU- and I/O-intensive

**Solution**: Cache complete scan results to disk

```go
// First scan: full directory walk + hashing
// Cached load: instant (validated against the key file on disk)
cache := cache.NewScanCache(".data")
cache.SaveScan(version)  // Save after scanning
cache.LoadScan(versionNumber, location)  // Instant load
```

**Benefits:**
- Cached versions skip the scan entirely on subsequent runs
- Key file hash validation ensures integrity
- Location-based hashing prevents wrong cache usage

**Trade-offs:**
- Disk space for cache files
- Cache invalidation needed when files change

### 2. Parallel Processing

**Problem**: Sequential file hashing doesn't utilize multi-core CPUs

**Solution**: Parallel checksum calculation with a worker pool

```go
// The CLI layer handles auto-detection:
// --jobs 0  →  worker_threads from config file (built-in default: runtime.NumCPU())
// The version manager then calls:
scan.ScanDirectoryParallelWithProgress(workerCount, progressCallback)

// Or specify workers explicitly
scan.ScanDirectoryParallelWithProgress(8, progressCallback)  // Use 8 workers
```

**Benefits:**
- Spreads SHA-256 hashing across all cores; largest gains when disk throughput is not the bottleneck

**Trade-offs:**
- Diminishing returns beyond CPU count or when I/O-bound

### 3. Selective Backup

**Problem**: Full backup duplicates entire application

**Solution**: Only backup files that will be modified/deleted

```go
// OpAdd and OpAddDir: NOT backed up (new files)
// OpModify, OpDelete, OpDeleteDir: Backed up (changed/removed)
```

**Benefits:**
- Backup size is a small fraction of the installation for typical updates
- Faster backup creation
- Less disk I/O

**Trade-offs:**
- Slightly more complex logic
- Must track operation types

### 4. Large File Handling

**Problem**: Loading multi-GB files causes memory exhaustion

**Current behavior**: The generator reads all files entirely into memory via `os.ReadFile` (see `large-file-handling.md` for details). Chunked writes during application (128MB chunks, `ChunkSize` constant) limit write-buffer overhead but do not reduce overall memory because the patch data is fully loaded during deserialization.

### 5. Compression

**Problem**: Uncompressed patches waste bandwidth

**Solution**: zstd compression with configurable levels

```go
// Compression levels: 1-4 (zstd)
// Level 1: Fastest, larger size
// Level 4: Smallest size, slower
CompressData(data, "zstd", 3)  // Balanced (CLI default)
```

**Benefits:**
- Substantial size reduction on compressible content
- Multiple algorithm options (zstd, gzip, none)

### Streaming Compression

For large data that should not be buffered entirely in memory, `CompressDataStreaming` and `DecompressDataStreaming` operate on `io.Reader`/`io.Writer` interfaces:

```go
// Streaming compression - constant memory regardless of input size
CompressDataStreaming(src, dst, "zstd", 3)

// Streaming decompression
DecompressDataStreaming(src, dst, "zstd")
```

**Performance characteristics:**
- Memory usage is constant (bounded by internal encoder buffers), not proportional to data size
- Used internally by `SavePatch` for patch file output with optional compression
- Algorithm and level options are identical to the in-memory `CompressData`/`DecompressData` functions

**Trade-offs:**
- CPU time for compression/decompression
- Higher levels have diminishing returns

### 6. Full File Replacement

**Problem**: Binary diff generation is slow and memory-intensive for any file size

**Solution**: Use full file replacement for all modified files regardless of size

```go
// All modified files: store entire new file content
operation.NewFile = readFile(newFilePath)
```

**Benefits:**
- Avoids bsdiff memory requirements entirely
- Simpler code path, faster generation
- Predictable memory usage proportional to changed file sizes

**Trade-offs:**
- Larger patch size for small changes in large files
- No inter-file deduplication

## Time Complexity

| Operation | Complexity | Notes |
|-----------|------------|-------|
| Directory Scan | O(n) | n = number of files |
| File Hashing | O(m) | m = total file size |
| Manifest Comparison | O(n) | n = number of files |
| Patch Generation | O(k) | k = size of changed files |
| Patch Application | O(k) | k = total bytes backed up, written, and re-verified |

## Space Complexity

| Component | Complexity | Notes |
|-----------|------------|-------|
| Manifest | O(n) | n = number of files |
| Patch | O(k) | k = size of changes |
| Backup | O(m) | m = size of modified/deleted files |
| Memory | O(k) | k = total size of changed files (all held in memory simultaneously during generation; the whole decoded patch is held during application) |

## Performance Tuning

### For Patch Creators

1. **Enable scan caching**: Use `--savescans` for repeated builds
2. **Use parallel workers**: `--jobs 0` for auto-detect (uses `runtime.NumCPU()` unless overridden in the config file)
3. **Choose compression level**: `--level 3` for balance
4. **SSD for versions**: Store versions on fast storage
5. **Exclude unnecessary files**: Use `.cyberignore`

### For Patch Users

1. **Verify before applying**: `--verify` ensures correctness
2. **Keep backup enabled**: `--backup` for safe rollback
3. **SSD for target**: Faster application on SSD
4. **Close other apps**: Reduce disk contention

## Progress Indicators

CyberPatchMaker provides real-time progress for:

- **Scanning**: Files processed, percentage, elapsed time, ETA (parallel scans also show worker count)
- **Manifest comparison**: Percentage of files analyzed
- **Generation**: Per-file add/modify lines as operations are built
- **Large file writes** (>1GB): Write progress percentage during application

## Bottlenecks

### Common Bottlenecks

1. **Disk I/O**: Primary bottleneck for large projects (scanning, backup, writes)
2. **Hash calculation**: CPU-intensive, parallelizable
3. **Compression**: CPU-intensive, tunable via level

### Mitigation Strategies

| Bottleneck | Mitigation |
|------------|------------|
| Disk I/O | SSD, reduce file count, exclude files |
| Hashing | Parallel workers, scan caching |
| Compression | Lower level, better algorithm |

## Related Documentation

- [Compression Guide](compression-guide.md) - Compression options
- [Large File Handling](large-file-handling.md) - Memory optimization
- [Scan Caching](scan-caching.md) - Instant reload
- [Architecture](architecture.md) - System design and code organization
