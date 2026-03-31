# Performance

Optimization techniques and performance characteristics of CyberPatchMaker.

## Overview

CyberPatchMaker is designed for performance across multiple dimensions:
- **Fast patch generation**: Minimize time to create patches
- **Small patch sizes**: Minimize bandwidth for distribution
- **Low memory usage**: Handle large applications without exhaustion
- **Quick patch application**: Minimize user downtime

## Performance Benchmarks

### Real-World Example: War Thunder

| Metric | Value |
|--------|-------|
| **Application Size** | 56 GB |
| **File Count** | 34,650 files |
| **Initial Scan Time** | ~15 minutes |
| **Cached Scan Load** | <1 second |
| **Typical Patch Size** | 5-50 MB |
| **Patch Generation** | 3-5 minutes |
| **Patch Application** | 2-3 minutes |

### Large File Performance

| File Size | Generation Time | Application Time | Memory Usage |
|-----------|-----------------|------------------|--------------|
| 1 GB | ~30 seconds | ~20 seconds | <500 MB |
| 5 GB | ~2 minutes | ~90 seconds | <500 MB |
| 20 GB | ~8 minutes | ~6 minutes | <500 MB |

## Optimization Techniques

### 1. Scan Caching

**Problem**: Directory scanning with SHA-256 hashing is CPU-intensive

**Solution**: Cache complete scan results to disk

```go
// First scan: 15+ minutes for large projects
// Cached load: <1 second
cache := cache.NewScanCache(".data")
cache.SaveScan(version)  // Save after scanning
cache.LoadScan(versionNumber, location)  // Instant load
```

**Benefits:**
- 900x faster for cached versions
- Key file hash validation ensures integrity
- Location-based hashing prevents wrong cache usage

**Trade-offs:**
- Disk space for cache files
- Cache invalidation needed when files change

### 2. Parallel Processing

**Problem**: Sequential file hashing doesn't utilize multi-core CPUs

**Solution**: Parallel checksum calculation with worker pool

```go
// The CLI layer handles auto-detection:
// --jobs 0  →  runtime.NumCPU() cores (set in cmd/generator/main.go)
// The version manager then calls:
scan.ScanDirectoryParallelWithProgress(workerCount, progressCallback)

// Or specify workers explicitly
scan.ScanDirectoryParallelWithProgress(8, progressCallback)  // Use 8 workers
```

**Benefits:**
- Near-linear speedup on multi-core systems
- 4-8x faster on typical 4-8 core CPUs

**Trade-offs:**
- Higher memory usage during parallel scan
- Diminishing returns beyond CPU count

### 3. Selective Backup

**Problem**: Full backup duplicates entire application

**Solution**: Only backup files that will be modified/deleted

```go
// OpAdd and OpAddDir: NOT backed up (new files)
// OpModify, OpDelete, OpDeleteDir: Backed up (changed/removed)
```

**Benefits:**
- 90%+ reduction in backup size for typical updates
- Faster backup creation
- Less disk I/O

**Trade-offs:**
- Slightly more complex logic
- Must track operation types

### 4. Memory-Optimized Large File Handling

**Problem**: Loading multi-GB files causes memory exhaustion

**Solution**: Chunked processing with 128MB chunks

```go
const (
    ChunkSize          = 128 * 1024 * 1024  // 128 MB
    LargeFileThreshold = 1024 * 1024 * 1024 // 1 GB
)

// Process in chunks instead of loading entire file
for chunk := 0; chunk < totalChunks; chunk++ {
    offset := chunk * ChunkSize
    processChunk(offset, ChunkSize)
}
```

**Benefits:**
- Constant memory usage regardless of file size
- Can handle 20GB+ files with <500MB RAM
- Progress reporting during processing

**Trade-offs:**
- ~5-10% slower due to chunk overhead
- More complex code

### 5. Compression

**Problem**: Uncompressed patches waste bandwidth

**Solution**: zstd compression with configurable levels

```go
// Compression levels: 1-4 (zstd)
// Level 1: Fastest, larger size
// Level 4: Smallest size, slower
CompressData(data, "zstd", 3)  // Balanced
```

**Benefits:**
- ~60% size reduction on average
- Faster transfer outweighs compression time
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

### 6. Full File Replacement for Large Files

**Problem**: Binary diff generation on large files is slow and memory-intensive

**Solution**: Use full file replacement for files >1GB

```go
if fileSize > LargeFileThreshold {
    // Use full file instead of binary diff
    operation.NewFile = readFile(newFilePath)
} else {
    // Use binary diff
    operation.BinaryDiff = generateDiff(old, new)
}
```

**Benefits:**
- Avoids bsdiff memory requirements
- Often similar or better than diff for large files
- Simpler code path

**Trade-offs:**
- Larger patch size for small changes in large files
- No inter-file deduplication

## Time Complexity

| Operation | Complexity | Notes |
|-----------|------------|-------|
| Directory Scan | O(n) | n = number of files |
| File Hashing | O(m) | m = total file size |
| Manifest Comparison | O(n) | n = number of files |
| Binary Diff Generation | O(m) | m = file size (bsdiff) |
| Patch Generation | O(k) | k = size of changed files |
| Patch Application | O(k) | k = number of operations |

## Space Complexity

| Component | Complexity | Notes |
|-----------|------------|-------|
| Manifest | O(n) | n = number of files |
| Patch | O(k) | k = size of changes |
| Backup | O(m) | m = size of modified/deleted files |
| Memory | O(1) | Streaming I/O for large files |

## Compression Performance

### Algorithm Comparison

| Algorithm | Ratio | Speed | Use Case |
|-----------|-------|-------|----------|
| **zstd** | 60-70% | Fast | Default, best balance |
| **gzip** | 55-65% | Medium | Maximum compatibility |
| **none** | 100% | N/A | Debugging, very small patches |

### zstd Level Performance

| Level | Ratio | Time | Recommendation |
|-------|-------|------|----------------|
| 1 | 50-55% | Fastest | Fast iteration |
| 2 | 55-60% | Fast | Development |
| 3 | 60-65% | Medium | **Default** |
| 4 | 65-70% | Slow | Production builds |

## Memory Management

### Peak Memory Usage

| Operation | Small Project | Medium Project | Large Project |
|-----------|---------------|----------------|---------------|
| Scan | 50 MB | 200 MB | 500 MB |
| Generate | 100 MB | 500 MB | 1 GB |
| Apply | 100 MB | 500 MB | 1 GB |

**Note**: Memory usage is bounded by:
- Chunk size (128MB) for large files
- Worker count for parallel operations
- Patch size for loading

### Memory Optimization Strategies

1. **Streaming**: Process data in chunks, never load full file
2. **Reuse buffers**: Reuse compression buffers
3. **Release early**: Free data immediately after use
4. **Limit workers**: Cap parallel operations based on available memory

## I/O Performance

### Disk Access Patterns

| Operation | Pattern | Optimization |
|-----------|---------|---------------|
| Scanning | Sequential read | OS read-ahead helps |
| Hashing | Random read | Limited by disk seek time |
| Compression | Sequential write | Large buffer writes |
| Backup | Copy + Verify | Parallel when possible |

### SSD vs HDD Impact

| Operation | SSD | HDD | Ratio |
|-----------|-----|-----|-------|
| Scan | 2 min | 8 min | 4x |
| Generate | 1 min | 3 min | 3x |
| Apply | 1 min | 2 min | 2x |

**Recommendation**: Use SSD for temp directory and version storage.

## Performance Tuning

### For Patch Creators

1. **Enable scan caching**: Use `--savescans` for repeated builds
2. **Use parallel workers**: `--jobs 0` for auto-detect (uses `runtime.NumCPU()`)
3. **Choose compression level**: `--level 3` for balance
4. **SSD for versions**: Store versions on fast storage
5. **Exclude unnecessary files**: Use `.cyberignore`

### For Patch Users

1. **Verify before applying**: `--verify` ensures correctness
2. **Keep backup enabled**: `--backup` for safe rollback
3. **SSD for target**: Faster application on SSD
4. **Close other apps**: Reduce disk contention

## Performance Monitoring

### Progress Indicators

CyberPatchMaker provides real-time progress for:

- **Scanning**: Files processed, percentage, ETA
- **Hashing**: Current file, files remaining
- **Diffing**: File being diffed, progress
- **Compression**: Percentage complete
- **Applying**: Operations completed, remaining

### Performance Profiling

To profile performance:

```bash
# Build with profiling
go build -o patch-gen ./cmd/generator

# Run with CPU profiling
go tool pprof -http=:8080 ./patch-gen [options]

# Check memory usage
/runtime/metrics
```

## Bottlenecks

### Common Bottlenecks

1. **Disk I/O**: Primary bottleneck for large projects
2. **Hash calculation**: CPU-intensive, parallelizable
3. **Compression**: CPU-intensive, tunable
4. **Network**: For distributed patching

### Mitigation Strategies

| Bottleneck | Mitigation |
|------------|------------|
| Disk I/O | SSD, reduce file count, exclude files |
| Hashing | Parallel workers, scan caching |
| Compression | Lower level, better algorithm |
| Network | Smaller patches, compression, CDN |

## Future Optimizations

### Planned

1. **Deduplication**: Cross-patch deduplication
2. **Delta compression**: Similar file detection
3. **Async I/O**: Overlap computation and I/O
4. **Memory mapping**: Zero-copy file access

### Experimental

1. **GPU acceleration**: CUDA/OpenCL hashing
2. **Distributed generation**: Cloud-based patch generation
3. **Machine learning**: Predict optimal chunk sizes

## Related Documentation

- [Compression Guide](compression-guide.md) - Compression options
- [Large File Handling](large-file-handling.md) - Memory optimization
- [Scan Caching](scan-caching.md) - Instant reload
- [Architecture](architecture.md#performance-characteristics) - System performance
