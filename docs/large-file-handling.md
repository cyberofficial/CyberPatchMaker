# Large File Handling

CyberPatchMaker now includes automatic memory-optimized handling for large files (>1GB). This feature prevents memory exhaustion when working with massive game codebases or enterprise applications.

## Overview

When generating or applying patches, CyberPatchMaker automatically detects large files and switches to **chunked processing mode**. This ensures that even files measuring 20GB+ can be processed without exhausting system memory.

## Key Features

### Automatic Detection
- Files larger than **1GB** are automatically processed using chunked operations
- No configuration required - the system handles this transparently
- Progress indicators show real-time status for large file operations

### Memory-Efficient Operations

#### Patch Generation
- **Large file additions**: Files are copied in 128MB chunks
- **Large file modifications**: Binary diffs are computed chunk-by-chunk
- **Progress tracking**: Shows percentage and MB processed/total

#### Patch Application
- **Chunked writing**: Large results are written in 128MB chunks
- **Memory limits**: Never loads both file versions simultaneously
- **Safe operations**: Checksums verified after each operation

## Technical Details

### Constants
```go
ChunkSize = 128 * 1024 * 1024  // 128 MB per chunk
LargeFileThreshold = 1024 * 1024 * 1024  // 1 GB threshold
```

### Processing Strategy

#### For Added Files (>1GB)
1. Detect file size exceeds threshold
2. Copy source file in 128MB chunks to temporary location
3. Read complete file data from temp location
4. Include in patch operation
5. Show progress during copy

#### For Modified Files (>1GB)
1. Detect either old or new file exceeds threshold
2. Open both files for streaming access
3. Compare files chunk-by-chunk (128MB chunks)
4. Generate binary diff for each changed chunk using bsdiff
5. Write chunks incrementally to temp file
6. Release memory after each chunk
7. Return complete diff data

#### For Applying Patches
1. Check if target file exceeds threshold
2. Apply binary diff to generate result
3. Write result in 128MB chunks if large
4. Verify checksums
5. Show progress during write

### Performance Benefits

**Before (without chunked processing):**
- 23.4GB file + 23.4GB file = 46.8GB RAM required
- System with 32GB RAM: Memory exhaustion, page file usage, system slowdown

**After (with chunked processing):**
- Peak memory usage: ~256MB (2x 128MB chunks)
- Stable operation on 32GB RAM systems
- No page file spillover

## Example Output

### Generating Patch with Large File
```
Processing 1 added files...
  Large file detected (23456 MB), using chunked copy: assets/game.pak
  Progress: 100.0% (23456/23456 MB)
  Add (large): assets/game.pak (23456 MB)

Processing 1 modified files (generating diffs)...
  Large file detected (old: 12000 MB, new: 13500 MB), using chunked diff: data/world.bin
  Progress: 100.0% (13500/13500 MB)
  Modify (chunked diff): data/world.bin (diff: 1200 MB, orig: 12000 MB, new: 13500 MB)
```

### Applying Patch with Large File
```
Applying 2 operations...
  Large file add detected (23456 MB), writing in chunks: assets/game.pak
  Write progress: 100.0% (23456/23456 MB)
  Added (large): assets/game.pak (23456 MB)

  Large file modify detected (12000 MB), applying patch in chunks: data/world.bin
  Writing large result (13500 MB) in chunks...
  Write progress: 100.0% (13500/13500 MB)
  Modified (large): data/world.bin (13500 MB)
```

## Best Practices

### For Developers
1. **System Requirements**: Recommend at least 8GB RAM for typical operations
2. **Large Projects**: 16GB+ RAM recommended for game projects with 20GB+ files
3. **Progress Monitoring**: Console output shows real-time progress for large operations
4. **Disk Space**: Ensure adequate temp space (2x largest file size recommended)

### For System Administrators
1. **Temp Directory**: Ensure temp partition has sufficient space
2. **I/O Performance**: SSD recommended for temp directory location
3. **Process Priority**: Consider running with normal priority to avoid system impact

## Limitations

### Current Implementation
- Chunked processing adds minor overhead (~5-10% slower)
- Still requires enough RAM for patch metadata
- Temp files created during processing (automatically cleaned up)

### Future Improvements
- Reduce chunk size for very memory-constrained systems
- Add configuration option for custom chunk sizes
- Implement even more aggressive memory optimization

## Troubleshooting

### High Memory Usage
If you still experience high memory usage:
1. Check that files are actually >1GB (threshold check)
2. Verify sufficient temp space available
3. Close other memory-intensive applications
4. Consider processing files individually instead of batch mode

### Slow Performance
Chunked processing is slightly slower but prevents crashes:
1. Expected: ~5-10% slower than non-chunked
2. Check disk I/O performance (temp directory)
3. Consider upgrading to SSD if using HDD
4. Ensure antivirus isn't scanning temp files

## Version History

- **v1.0.6**: Initial implementation of large file handling
  - Automatic detection of files >1GB
  - Chunked processing (128MB chunks)
  - Progress indicators for large operations
  - Streaming compression support

## Related Documentation

- [Hash Verification](hash-verification.md) - How checksums work with large files
- [Compression Guide](compression-guide.md) - Compression with chunked data
- [Troubleshooting](troubleshooting.md) - Common issues and solutions
