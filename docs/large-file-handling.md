# Large File Handling

## Overview

The generator uses **full file replacement** for all files — every modified or added file is read into memory via `os.ReadFile()` and stored as `PatchOperation.NewFile`. There is no binary diff generation (bsdiff) in the current generator flow.

During patch **application**, files larger than 1GB are written in 128MB chunks to avoid memory pressure from the write buffer.

## Constants

```go
ChunkSize          = 128 * 1024 * 1024  // 128 MB per chunk
LargeFileThreshold = 1024 * 1024 * 1024 // 1 GB threshold for chunked writing
DefaultMaxPartSize = 4 * 1024 * 1024 * 1024 // 4 GB max part size
```

## Patch Generation

All files (any size) are read fully into memory with `os.ReadFile()` and stored directly in the patch operation's `NewFile` field. Every added/modified file's data is held simultaneously in the in-memory `Patch` struct, so the generator requires sufficient RAM for the combined size of all changed files (a 20GB file needs 20GB+ RAM by itself).

## Patch Application

During application, files larger than `LargeFileThreshold` (1GB) are written in 128MB chunks instead of a single `os.WriteFile` call. However, the patch is fully decoded into memory first (all operations, including their full file data, are held by the parsed `Patch` struct), so peak memory is at minimum the size of the largest file in the patch. Chunked writes only limit write-buffer overhead.

Progress indicators show real-time status during chunked writes.

## Memory Considerations

- **Generation**: All added/modified files are held in memory at once, so RAM on the order of the total changed-file size is required. A 20GB file needs 20GB+ RAM by itself.
- **Application**: The patch is fully decoded into memory — every operation's `NewFile` data is held by the parsed `Patch` struct — so peak memory is on the order of the total size of the files contained in the patch. Chunked writes only limit write-buffer overhead.
- **Multi-part patches**: Patches >4GB are automatically split into parts to manage individual file sizes.

## Performance

Chunked writing adds a small amount of overhead compared to a single write call. It avoids allocating one additional full-size write buffer, but it does not reduce the peak memory needed to hold the file data itself (the operation's `NewFile` bytes are already in memory).
