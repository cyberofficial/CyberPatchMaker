# Compression Guide

## Available Algorithms

| Algorithm | Levels | Default | Best For |
|-----------|--------|---------|----------|
| **zstd** | 1-4 | 3 | Default choice — best speed/size balance |
| **gzip** | 1-3 | 3 | Compatibility with systems lacking zstd |
| **none** | N/A | N/A | Testing/debugging only |

## Choosing a Level

### zstd
- Level 1: Fastest, largest. Use for rapid iteration or fast networks.
- Level 2: Default speed. Slightly faster than level 3 with slightly larger output.
- Level 3 (default in CLI): Balanced. Recommended for production.
- Level 4: Smallest, slowest. Use when bandwidth is the primary constraint.

### gzip
- Level 1: Fastest (BestSpeed). Level 2: Default (DefaultCompression). Level 3: Maximum (BestCompression).

## Usage

```bash
patch-gen --compression zstd --level 3     # default, recommended
patch-gen --compression zstd --level 4     # smallest patches
patch-gen --compression gzip --level 2     # maximum compatibility
patch-gen --compression none               # testing only
```

## What Compresses Well

- **Good (4-10x)**: Text, source code, JSON, XML, uncompressed images (BMP, TIFF)
- **Moderate (2-4x)**: Executables, DLLs, databases
- **Poor (1-1.2x)**: Already-compressed files — JPG, PNG, MP4, ZIP, PDF

Including already-compressed assets in patches? Consider excluding them via `.cyberignore` since compression won't help.

## API

```go
// In-memory (small data)
utils.CompressData(data []byte, algorithm string, level int) ([]byte, error)
utils.DecompressData(data []byte, algorithm string) ([]byte, error)

// Streaming (large data — used internally for patch files)
utils.CompressDataStreaming(src io.Reader, dst io.Writer, algorithm string, level int) error
utils.DecompressDataStreaming(src io.Reader, dst io.Writer, algorithm string) error
```

Streaming functions operate on `io.Reader`/`io.Writer` and use constant memory regardless of input size.
