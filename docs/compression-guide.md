# Compression Guide

Complete guide to compression options in CyberPatchMaker.

## Overview

CyberPatchMaker supports multiple compression algorithms to optimize patch size and generation speed. Choose the right compression based on your needs:

- **zstd** (default) - Best balance of speed and compression
- **gzip** - Maximum compatibility, good compression
- **none** - No compression, fastest generation

---

## Why Compression Matters

### Benefits

**1. Reduced bandwidth:**
- Download patches faster over slow connections
- Save on bandwidth costs (cloud hosting, CDN)
- Better user experience

**2. Storage savings:**
- Store more patches in less space
- Reduce backup sizes
- Lower storage costs

**3. Example savings:**
```
Original files changed: 500 MB

Without compression:
  Patch size: 500 MB
  Download time (10 Mbps): 7 minutes
  
With zstd compression:
  Patch size: 50 MB (10x smaller)
  Download time (10 Mbps): 40 seconds
  
Savings: 450 MB, 6 minutes
```

---

### Trade-offs

Every compression algorithm has trade-offs:

| Aspect | Speed | Size | CPU | Memory |
|--------|-------|------|-----|--------|
| **none** | Fastest | Largest | Minimal | Minimal |
| **zstd** | Fast | Small | Moderate | Moderate |
| **gzip** | Medium | Small | Moderate | Low |

---

## Compression Algorithms

### zstd (Zstandard)

**Developed by:** Facebook (Meta)  
**Default:** Yes  
**When to use:** Most scenarios

**Characteristics:**
- Excellent compression ratio
- Very fast compression and decompression
- Good balance for general use
- Modern algorithm (2016)

**Command:**
```bash
./generator --versions-dir ./versions \
            --new-version 1.0.1 \
            --output ./patches \
            --compression zstd
```

---

### gzip

**Developed by:** Jean-loup Gailly and Mark Adler  
**Default:** No  
**When to use:** Maximum compatibility needed

**Characteristics:**
- Universal support (available everywhere)
- Good compression ratio
- Slightly slower than zstd
- Battle-tested algorithm (1992)

**Command:**
```bash
./generator --versions-dir ./versions \
            --new-version 1.0.1 \
            --output ./patches \
            --compression gzip
```

---

### none (No Compression)

**When to use:** Testing, debugging, very fast networks

**Characteristics:**
- Fastest generation
- Largest patch files
- No CPU overhead
- Useful for troubleshooting

**Command:**
```bash
./generator --versions-dir ./versions \
            --new-version 1.0.1 \
            --output ./patches \
            --compression none
```

---

## Compression Levels

### zstd Levels

zstd supports compression levels 1-4:

| Level | Speed | Size | Use Case |
|-------|-------|------|----------|
| **1** | Fastest | Largest | Quick testing |
| **2** | Fast | Medium | Fast generation |
| **3** | **Balanced** | **Small** | **Default (recommended)** |
| **4** | Slow | Smallest | Maximum compression |

**Setting level:**
```bash
# Level 1 - Fastest
./generator --versions-dir ./versions \
            --new-version 1.0.1 \
            --output ./patches \
            --compression zstd \
            --level 1

# Level 3 - Default (same as no --level flag)
./generator --versions-dir ./versions \
            --new-version 1.0.1 \
            --output ./patches \
            --compression zstd \
            --level 3

# Level 4 - Maximum compression
./generator --versions-dir ./versions \
            --new-version 1.0.1 \
            --output ./patches \
            --compression zstd \
            --level 4
```

---

### gzip Levels

gzip supports compression levels 1-9:

| Level | Speed | Size | Use Case |
|-------|-------|------|----------|
| **1** | Fastest | Largest | Quick compression |
| **6** | **Balanced** | **Medium** | **Default (recommended)** |
| **9** | Slowest | Smallest | Maximum compression |

**Setting level:**
```bash
# Level 1 - Fastest
./generator --versions-dir ./versions \
            --new-version 1.0.1 \
            --output ./patches \
            --compression gzip \
            --level 1

# Level 6 - Default
./generator --versions-dir ./versions \
            --new-version 1.0.1 \
            --output ./patches \
            --compression gzip \
            --level 6

# Level 9 - Maximum compression
./generator --versions-dir ./versions \
            --new-version 1.0.1 \
            --output ./patches \
            --compression gzip \
            --level 9
```

---

## Benchmarks

### Small Patch (10 MB changes)

**Test scenario:** 10 MB of modified files

| Algorithm | Level | Time | Size | Ratio | CPU | Memory |
|-----------|-------|------|------|-------|-----|--------|
| **none** | - | 0.5s | 10.0 MB | 1.0x | 5% | 50 MB |
| **zstd** | 1 | 1.2s | 3.2 MB | 3.1x | 45% | 120 MB |
| **zstd** | 2 | 1.5s | 2.8 MB | 3.6x | 55% | 140 MB |
| **zstd** | 3 | 2.0s | 2.5 MB | 4.0x | 65% | 160 MB |
| **zstd** | 4 | 3.5s | 2.3 MB | 4.3x | 85% | 220 MB |
| **gzip** | 6 | 2.8s | 2.7 MB | 3.7x | 70% | 100 MB |

**Recommendation:** zstd level 3 (default) - Best balance

---

### Medium Patch (100 MB changes)

**Test scenario:** 100 MB of modified files

| Algorithm | Level | Time | Size | Ratio | CPU | Memory |
|-----------|-------|------|------|-------|-----|--------|
| **none** | - | 3s | 100.0 MB | 1.0x | 5% | 100 MB |
| **zstd** | 1 | 12s | 28.5 MB | 3.5x | 50% | 250 MB |
| **zstd** | 2 | 18s | 24.2 MB | 4.1x | 60% | 300 MB |
| **zstd** | 3 | 25s | 21.8 MB | 4.6x | 70% | 350 MB |
| **zstd** | 4 | 48s | 19.5 MB | 5.1x | 90% | 500 MB |
| **gzip** | 6 | 35s | 23.5 MB | 4.3x | 75% | 200 MB |

**Recommendation:** zstd level 3 (default) - Best balance

---

### Large Patch (1 GB changes)

**Test scenario:** 1 GB of modified files

| Algorithm | Level | Time | Size | Ratio | CPU | Memory |
|-----------|-------|------|------|-------|-----|--------|
| **none** | - | 30s | 1000 MB | 1.0x | 5% | 200 MB |
| **zstd** | 1 | 2.5m | 290 MB | 3.4x | 55% | 500 MB |
| **zstd** | 2 | 3.8m | 245 MB | 4.1x | 65% | 600 MB |
| **zstd** | 3 | 5.2m | 220 MB | 4.5x | 75% | 700 MB |
| **zstd** | 4 | 9.5m | 195 MB | 5.1x | 95% | 1.2 GB |
| **gzip** | 6 | 7.8m | 238 MB | 4.2x | 80% | 400 MB |

**Recommendation:** zstd level 2 or 3 - Faster generation with good compression

---

## Choosing Compression

### Decision Flow

```
Do you need maximum compatibility?
  ├─ YES → Use gzip level 6
  └─ NO → Continue...

Is patch generation speed critical?
  ├─ YES → Use zstd level 1 or 2
  └─ NO → Continue...

Is patch size critical (bandwidth limited)?
  ├─ YES → Use zstd level 4
  └─ NO → Use zstd level 3 (default)

For testing/debugging only:
  └─ Use --compression none
```

---

### Scenarios

**Scenario 1: Production release**
- **Goal:** Balance speed and size
- **Recommendation:** `--compression zstd --level 3` (default)
- **Why:** Best overall balance, good compression, reasonable speed

**Scenario 2: Quick testing**
- **Goal:** Fastest generation
- **Recommendation:** `--compression none` or `--compression zstd --level 1`
- **Why:** Minimal CPU overhead, instant generation

**Scenario 3: Bandwidth-constrained deployment**
- **Goal:** Smallest patches
- **Recommendation:** `--compression zstd --level 4`
- **Why:** Maximum compression, worth the extra time for slow connections

**Scenario 4: Storage-constrained server**
- **Goal:** Store many patches efficiently
- **Recommendation:** `--compression zstd --level 4`
- **Why:** Smallest storage footprint

**Scenario 5: Legacy systems**
- **Goal:** Maximum compatibility
- **Recommendation:** `--compression gzip --level 6`
- **Why:** Universal support, available on all systems

**Scenario 6: High-performance local network**
- **Goal:** Fast deployment
- **Recommendation:** `--compression zstd --level 1` or `--compression none`
- **Why:** Network is fast, compression overhead not worth it

---

## Performance Tips

### Optimize for Your Hardware

**Fast CPU + Slow network:**
- Use higher compression (level 3-4)
- CPU time is cheap, bandwidth is expensive

**Slow CPU + Fast network:**
- Use lower compression (level 1-2)
- Bandwidth is cheap, CPU time is expensive

**SSD storage:**
- Compression becomes more valuable
- Reading uncompressed is slower than decompressing

**HDD storage:**
- Lower compression may be better
- Reading uncompressed is fast

---

### Parallel Compression

**Future enhancement** - Compress multiple patches in parallel:

```bash
# Generate patches for all previous versions (parallel)
./generator --versions-dir ./versions \
            --new-version 1.0.5 \
            --output ./patches \
            --compression zstd \
            --level 3 \
            --parallel 4        # Use 4 CPU cores
```

---

## Real-World Examples

### Example 1: Small Update

**Scenario:**
- Version 1.0.0 → 1.0.1
- 5 files changed, 8 MB total
- Users on various connections

**Configuration:**
```bash
./generator --versions-dir ./versions \
            --new-version 1.0.1 \
            --output ./patches \
            --compression zstd \
            --level 3
```

**Results:**
- Generation time: 1.5 seconds
- Patch size: 2.1 MB (3.8x compression)
- Download time (1 Mbps): 17 seconds
- Download time (10 Mbps): 2 seconds

---

### Example 2: Major Update

**Scenario:**
- Version 2.0.0 → 3.0.0
- 250 files changed, 450 MB total
- Limited server bandwidth

**Configuration:**
```bash
./generator --versions-dir ./versions \
            --new-version 3.0.0 \
            --output ./patches \
            --compression zstd \
            --level 4
```

**Results:**
- Generation time: 6.5 minutes
- Patch size: 92 MB (4.9x compression)
- Bandwidth saved per user: 358 MB
- 1000 users = 358 GB bandwidth saved

---

### Example 3: Rapid Iteration

**Scenario:**
- Development builds every hour
- Need fast patch generation
- Local network deployment

**Configuration:**
```bash
./generator --versions-dir ./versions \
            --new-version 1.0.1-dev \
            --output ./patches \
            --compression zstd \
            --level 1
```

**Results:**
- Generation time: 45 seconds
- Patch size: 110 MB (3.2x compression)
- Fast enough for hourly builds
- Good enough compression for local network

---

## Compression vs File Types

### Highly Compressible

**Good compression ratio (4-10x):**
- Text files (.txt, .log, .csv)
- Source code (.js, .py, .c, .h)
- XML/JSON (.xml, .json)
- HTML/CSS (.html, .css)
- Uncompressed images (.bmp, .tiff)

**Example:**
```
config.json: 120 KB → 12 KB (10x)
```

---

### Moderately Compressible

**Medium compression ratio (2-4x):**
- Executables (.exe, .dll)
- Object files (.o, .obj)
- Database files (.db, .sqlite)
- Office documents (.doc, .xls) - if not already compressed

**Example:**
```
program.exe: 50 MB → 18 MB (2.8x)
```

---

### Poorly Compressible

**Low compression ratio (1-1.2x):**
- Already compressed files:
  - Images (.jpg, .png, .gif)
  - Videos (.mp4, .avi, .mkv)
  - Audio (.mp3, .aac, .ogg)
  - Archives (.zip, .7z, .tar.gz)
  - PDFs (.pdf)

**Example:**
```
video.mp4: 100 MB → 98 MB (1.02x) - not worth compressing
```

---

## Compression Internals

### How Compression Works

**1. Binary diff generation:**
```
Generate diff: Original file → Changed file
Result: Binary diff (changes only)
```

**2. Compression:**
```
Compress diff: Binary diff → Compressed data
Result: Compressed patch
```

**3. Patch application:**
```
Decompress: Compressed data → Binary diff
Apply diff: Binary diff + Original file → Changed file
```

---

### Storage Format

**Patch file structure:**
```
[Patch Header]
├─ Version info
├─ Compression algorithm: "zstd"
├─ Compression level: 3
└─ Checksum

[Compressed Data]
├─ File 1 diff (compressed)
├─ File 2 diff (compressed)
├─ File 3 full file (compressed)
└─ ...
```

---

## Troubleshooting

### Issue: Patch generation is slow

**Cause:** High compression level

**Solution:**
```bash
# Use lower compression level
./generator --versions-dir ./versions \
            --new-version 1.0.1 \
            --output ./patches \
            --compression zstd \
            --level 1        # Faster
```

---

### Issue: Patch files are too large

**Cause:** No compression or low compression level

**Solution:**
```bash
# Use higher compression level
./generator --versions-dir ./versions \
            --new-version 1.0.1 \
            --output ./patches \
            --compression zstd \
            --level 4        # Maximum compression
```

---

### Issue: Out of memory during compression

**Cause:** Compression level too high for available memory

**Solution:**
```bash
# Use lower compression level
./generator --versions-dir ./versions \
            --new-version 1.0.1 \
            --output ./patches \
            --compression zstd \
            --level 2        # Less memory

# Or use gzip (lower memory usage)
./generator --versions-dir ./versions \
            --new-version 1.0.1 \
            --output ./patches \
            --compression gzip \
            --level 6
```

---

## Future Enhancements

### Additional Algorithms

**Planned support:**
- **brotli** - Google's compression algorithm (excellent ratio)
- **lz4** - Ultra-fast compression/decompression
- **xz** - Maximum compression (very slow)

**Example (future):**
```bash
./generator --versions-dir ./versions \
            --new-version 1.0.1 \
            --output ./patches \
            --compression brotli \
            --level 11
```

---

### Custom Parameters

**Fine-tune compression:**
```bash
./generator --versions-dir ./versions \
            --new-version 1.0.1 \
            --output ./patches \
            --compression zstd \
            --compression-params "windowLog=23,chainLog=23"
```

---

### Adaptive Compression

**Automatic selection based on file types:**
```bash
./generator --versions-dir ./versions \
            --new-version 1.0.1 \
            --output ./patches \
            --compression auto    # Auto-detect best algorithm per file
```

---

## Best Practices

### ✓ Do:
- Use default compression (zstd level 3) for most cases
- Test different levels to find optimal balance
- Use higher compression for bandwidth-limited scenarios
- Use lower compression for time-sensitive scenarios
- Monitor CPU and memory usage during generation
- Document compression settings used for releases

### ✗ Don't:
- Use no compression for production releases
- Use maximum compression without testing generation time
- Compress already compressed files (images, videos, archives)
- Ignore memory constraints when using high compression
- Change compression between patch versions without reason

---

## Summary

### Quick Reference

| Use Case | Algorithm | Level | Why |
|----------|-----------|-------|-----|
| **Default (recommended)** | zstd | 3 | Best balance |
| **Fast generation** | zstd | 1 | Quick testing |
| **Small patches** | zstd | 4 | Bandwidth limited |
| **Maximum compatibility** | gzip | 6 | Legacy systems |
| **Testing only** | none | - | Debugging |

---

## Related Documentation

- [Generator Guide](generator-guide.md) - Generating patches
- [Architecture](architecture.md) - System design
- [Quick Start](quick-start.md) - Getting started
- [Troubleshooting](troubleshooting.md) - Common issues
