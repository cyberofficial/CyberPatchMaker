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

All files (any size) are read fully into memory with `os.ReadFile()` and stored directly in the patch operation's `NewFile` field. This means the generator requires sufficient RAM for the largest single file being processed.

## Patch Application

During application, files larger than `LargeFileThreshold` (1GB) are written in 128MB chunks instead of a single `os.WriteFile` call. However, the patch file is fully loaded and deserialized into memory first (via `os.ReadFile` and JSON decoding), so peak memory is at minimum the size of the largest file in the patch. Chunked writes only limit write-buffer overhead.

Progress indicators show real-time status during chunked writes.

## Memory Considerations

- **Generation**: Requires enough RAM for the largest file being read. A 20GB file needs 20GB+ RAM.
- **Application**: The patch file must be fully loaded and deserialized into memory, so peak memory is at minimum the size of the largest file in the patch. Chunked writes only limit write-buffer overhead.
- **Multi-part patches**: Patches >4GB are automatically split into parts to manage individual file sizes.

## Performance

Chunked writing during application adds ~5-10% overhead compared to a single write, but prevents memory exhaustion on systems with limited RAM.
