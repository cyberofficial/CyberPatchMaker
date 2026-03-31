# Multi-Part Patch System

## Overview

CyberPatchMaker automatically splits large patches into multiple parts when the patch size exceeds 4GB. This prevents memory exhaustion issues and makes large patches more manageable.

## How It Works

### Automatic Splitting

When generating a patch, the system:

1. **Calculates total patch size** before saving
2. **Sorts operations by size** with key file operations prioritized
3. **Splits into 4GB parts** if total size > 4GB
4. **Generates multi-part metadata** for verification
5. **Saves parts as separate files** (.01.patch, .02.patch, etc.)

### File Naming Convention

```
patch-name.01.patch  ← Part 1 (contains key file + metadata + small files)
patch-name.02.patch  ← Part 2
patch-name.03.patch  ← Part 3 (if needed)
...
```

### Part 1 Special Role

Part 1 (.01.patch) contains:
- **Key file verification data** (always in part 1)
- **Multi-part metadata** (total parts, part hashes)
- **Small files** (fill up to 4GB limit)
- **Part verification hashes** (SHA-256 of all parts)

## Size Distribution Example

Given these files:
```
File 1: 100MB
File 2: 5.5GB
File 3 (keyfile): 450MB
File 4: 2.2GB
File 5: 50MB
File 6: 9.8GB
File 7: 1.6GB
File 8: 300MB
File 9: 7.4GB
File 10: 120MB
```

**Sorting Order:**
```
File 3 (keyfile): 450MB  ← Always first
File 5: 50MB
File 1: 100MB
File 10: 120MB
File 8: 300MB
File 7: 1.6GB
File 4: 2.2GB
File 2: 5.5GB
File 9: 7.4GB
File 6: 9.8GB
```

**Part Distribution:**

**Part 1 (.01.patch):** Files 3,5,1,10,8,7 = ~2.6GB
- Key file (File 3): 450MB
- Small files (5,1,10,8): 570MB
- Medium file (7): 1.6GB
- Multi-part metadata + hashes

**Part 2 (.02.patch):** File 4 = 2.2GB

**Part 3 (.03.patch):** File 2 = 5.5GB (exceeds 4GB, gets own part)

**Part 4 (.04.patch):** File 9 = 7.4GB (exceeds 4GB, gets own part)

**Part 5 (.05.patch):** File 6 = 9.8GB (exceeds 4GB, gets own part)

## Generating Multi-Part Patches

### CLI Generator

```bash
# Automatically splits if needed (no extra flags required)
patch-gen.exe --from-dir "C:\Version1" --to-dir "C:\Version2" --output patches

# Output:
# Patch size (15.2 GB) exceeds 4GB limit, splitting into multiple parts...
# Split patch into 5 parts
#   Part 1: 6 operations, ~2.6GB
#   Part 2: 1 operation, ~2.2GB
#   Part 3: 1 operation, ~5.5GB
#   Part 4: 1 operation, ~7.4GB
#   Part 5: 1 operation, ~9.8GB
# ✓ Multi-part patch saved: 5 parts
```

### Custom Split Sizes

The CLI automatically detects and splits large patches. Use `--splitsize` to customize the part size:

```bash
# Auto-split with default 4GB limit
patch-gen.exe --from-dir "C:\v1" --to-dir "C:\v2" --output patches

# Custom split size (2GB parts)
patch-gen.exe --from-dir "C:\v1" --to-dir "C:\v2" --output patches --splitsize 2G

# Small parts for limited upload space
patch-gen.exe --from-dir "C:\v1" --to-dir "C:\v2" --output patches --splitsize 500MB --bypasssplitlimit
```

Progress indicators show:
- Total patch size
- Number of parts
- Part generation progress

## Applying Multi-Part Patches

### Automatic Detection

The applier automatically detects multi-part patches:

```bash
# Just specify part 1 - all parts loaded automatically
patch-apply.exe --patch "patches\Version1-to-Version2.01.patch" --current-dir "C:\MyApp"

# Output:
# Detected multi-part patch: 5 parts
# Loading part 2: Version1-to-Version2.02.patch
#   ✓ Part 2 hash verified
# Loading part 3: Version1-to-Version2.03.patch
#   ✓ Part 3 hash verified
# Loading part 4: Version1-to-Version2.04.patch
#   ✓ Part 4 hash verified
# Loading part 5: Version1-to-Version2.05.patch
#   ✓ Part 5 hash verified
# ✓ Loaded 15 total operations from 5 parts
```

### User-Friendly Behavior

- **Specify any part:** If you run `.02.patch`, applier redirects to `.01.patch`
- **Hash verification:** All parts verified before applying
- **Missing parts detected:** Error if any part is missing or corrupted
- **Progress feedback:** Clear status for each part loaded

## Technical Details

### Part Metadata Structure

```json
{
  "MultiPart": {
    "IsMultiPart": true,
    "PartNumber": 1,
    "TotalParts": 5,
    "MaxPartSize": 4294967296,
    "PartHashes": [
      {
        "PartNumber": 1,
        "Checksum": "abc123...",
        "Size": 2684354560
      },
      {
        "PartNumber": 2,
        "Checksum": "def456...",
        "Size": 2200000000
      }
      // ... more parts
    ]
  }
}
```

### Verification Process

1. **Load Part 1** first (contains metadata)
2. **Check PartHashes** array for total parts
3. **Load each remaining part** sequentially
4. **Verify SHA-256 hash** before accepting part
5. **Merge operations** from all parts
6. **Apply combined patch** as single unit

### Memory Efficiency

Multi-part patches solve the memory exhaustion problem:

- **Single 18GB patch:** Tries to allocate 64GB+ for JSON parsing → OOM crash
- **Multi-part (5x 3-4GB):** Loads parts sequentially, max ~4GB per part → Success

## Chunk Sidecar System (Very Large Patches)

For patches where individual parts exceed 3.75GB (after compression overhead), CyberPatchMaker uses a **chunk sidecar system** to further break down parts into smaller chunk files.

### When Chunking Occurs

- Default part size limit: 4GB
- When a compressed part approaches this limit (~3.75GB), the system:
  1. Chunks the part data into smaller files (~500MB-1GB each)
  2. Creates a `.chunks.json` sidecar file with chunk metadata
  3. Reconstructs the part on-the-fly when loading

### Chunk File Structure

```
patch-name.01.patch                    ← Part 1 (small, contains metadata)
patch-name.02.patch                    ← Part 2 (regular, not chunked)
patch-name.part3.chunks.json           ← Sidecar for chunked Part 3
patch-name.part3.1.patch               ← Chunk 1 of Part 3
patch-name.part3.2.patch               ← Chunk 2 of Part 3
patch-name.part3.3.patch               ← Chunk 3 of Part 3
```

### Sidecar JSON Format

```json
{
  "part_number": 3,
  "chunks": [
    {
      "FileName": "patch-name.part3.1.patch",
      "PartNumber": 3,
      "ChunkNumber": 1,
      "Size": 536870912,
      "Checksum": "abc123def456..."
    },
    {
      "FileName": "patch-name.part3.2.patch",
      "PartNumber": 3,
      "ChunkNumber": 2,
      "Size": 536870912,
      "Checksum": "789ghi012jkl..."
    }
  ]
}
```

### Chunk Reconstruction Process

When applying a patch with chunked parts:

1. **Detect sidecar:** If `<part>.chunks.json` exists, part is chunked
2. **Load sidecar:** Read chunk metadata (file names, offsets, checksums)
3. **Read chunks:** Load each chunk file and verify SHA-256 checksum
4. **Reassemble:** Combine chunks in order into temporary file
5. **Verify reconstructed hash:** Ensure full part matches expected hash
6. **Load part:** Parse reconstructed part as normal patch data

### Benefits of Chunking

- **Handles any patch size:** Even 50GB+ patches can be processed
- **Memory-safe:** Never loads more than one chunk at a time
- **Parallel download ready:** Chunks can be downloaded independently
- **Checksum verification:** Each chunk verified individually

### Self-Contained Executables with Chunks

When creating self-contained executables with very large patches:

- **Chunk sidecar embedded** in the executable
- **Chunk files distributed separately** alongside the .exe
- **Executable auto-detects** chunks and reassembles during application

Example distribution:
```
MyApp-Patch-1.0.3.exe                  ← ~50MB (contains applier + patch metadata)
MyApp-Patch-1.0.3.exe.chunks.json     ← Chunk manifest
MyApp-Patch-1.0.3.exe.part1.1.patch    ← First chunk file
MyApp-Patch-1.0.3.exe.part1.2.patch    ← Second chunk file
```

### Binary Sidecar Embedding Format

When creating a self-contained executable with `--create-exe`, the generator embeds chunk sidecar JSON files directly into the executable binary. The layout on disk is:

```
+---------------------------+
| Applier EXE (stub)       |  ← patch-apply.exe binary
+---------------------------+
| Patch Data (part 01)      |  ← Compressed .patch content
+---------------------------+
| Sidecar Blob (optional)   |  ← Embedded chunk sidecar data
+---------------------------+
| 128-byte Header           |  ← Metadata trailer
+---------------------------+
```

**Sidecar Blob Format** (binary, little-endian):

| Field | Type | Description |
|-------|------|-------------|
| Count | uint32 | Number of embedded sidecar files |
| For each sidecar: | | |
| &nbsp;&nbsp;Name Length | uint16 | Length of sidecar filename in bytes |
| &nbsp;&nbsp;Name | []byte | Sidecar filename (e.g., `mypass.part2.chunks.json`) |
| &nbsp;&nbsp;Data Length | uint64 | Length of sidecar data in bytes |
| &nbsp;&nbsp;Data | []byte | Raw sidecar JSON content |

**128-byte Header Layout:**

| Offset | Size | Field |
|--------|------|-------|
| 0 | 8 bytes | Magic: `CPMPATCH` |
| 8 | 4 bytes | Version (uint32, currently 1) |
| 12 | 8 bytes | Stub size (uint64, size of applier EXE) |
| 20 | 8 bytes | Data offset (uint64, same as stub size) |
| 28 | 8 bytes | Data size (uint64, size of patch data) |
| 36 | 16 bytes | Compression type string (e.g., "zstd") |
| 52 | 32 bytes | SHA-256 checksum of patch data |
| 84 | 1 byte | Flags (bit 0: silent mode) |
| 85 | 43 bytes | Reserved |

At runtime, the applier reads the 128-byte header from the end of the executable, locates and extracts any embedded sidecar files, writes them to the same directory as the executable, and uses them to reconstruct chunked parts before applying the patch. Temporary files (extracted sidecars and part 01 data) are cleaned up after loading.

## Compatibility

### Backward Compatibility

- **Single-part patches** continue to work exactly as before
- **No version detection needed** - format is self-describing
- **Legacy appliers** will fail gracefully on multi-part patches (missing metadata error)

### Forward Compatibility

- **MaxPartSize is configurable** in metadata (adjustable via `--splitsize` flag)
- **Version field** allows format evolution
- **Reserved fields** for future enhancements

## Best Practices

### For Patch Creators

1. **Let the system auto-split** - don't try to manually split
2. **Keep all parts together** in same directory
3. **Distribute all parts** when sharing patches
4. **Test with part 1** - verify all parts present
5. **Include README** explaining multi-part patches to users

### For End Users

1. **Download all parts** before applying
2. **Keep parts in same folder** during application
3. **Run part 1** - system loads others automatically
4. **Don't rename parts** - breaks automatic detection
5. **Verify all parts present** - check for .01, .02, .03, etc.

## Troubleshooting

### "Part X missing" Error

**Cause:** Not all parts downloaded or incorrect directory

**Solution:** Ensure all .XX.patch files in same directory as .01.patch

### "Hash mismatch for part X"

**Cause:** Part file corrupted during download/transfer

**Solution:** Re-download the corrupted part file

### "Failed to load multi-part patch"

**Cause:** Part 1 doesn't contain metadata (corrupted or wrong file)

**Solution:** Re-download part 1 (.01.patch file)

### Out of Memory Despite Multi-Part

**Cause:** Individual operations still too large (>4GB single file)

**Solution:** Currently unsupported - file size exceeds system limits

## Performance Considerations

### Patch Generation

- **Sorting overhead:** Minimal (< 1 second for 100k files)
- **Part calculation:** Linear time O(n) operations
- **Hash generation:** Proportional to patch size
- **Disk I/O:** Main bottleneck (same as single-part)

### Patch Application

- **Sequential loading:** Each part loaded/verified in order
- **Memory usage:** Peak = largest single part (~4GB)
- **Verification time:** Proportional to number of parts
- **Overall time:** Similar to single-part (I/O bound)

## Limitations

### Current Limitations

1. **Configurable part size** - default 4GB, adjustable via `--splitsize` flag
2. **Sequential loading** - parts loaded one at a time (could be parallelized)
3. **No compression across parts** - each part compressed independently
4. **Metadata only in Part 1** - other parts don't have standalone info

### Known Edge Cases

1. **Single file > 4GB:** Gets own part (working as designed)
2. **Key file > 4GB:** Part 1 = just key file + metadata
3. **Many tiny files:** May create very unbalanced parts
4. **Rename detection:** Currently none (user must not rename)

## Future Enhancements

### Planned Features

1. **Parallel part loading** - speed up multi-part application
2. **Delta compression** - cross-part deduplication
3. **Smart part sizing** - better balance for many small files
4. **Integrity validation** - verify parts before generation completes
5. **Resume support** - restart failed multi-part downloads
6. **Part compression** - optimize individual part sizes further

## FAQ

**Q: Why 4GB default limit?**
A: Balances between manageable part sizes and avoiding too many parts. Prevents 32-bit integer overflow issues. The limit is configurable via `--splitsize` (e.g., `--splitsize 2G`).

**Q: Can I combine parts manually?**
A: No - the format requires proper metadata merging. Use the applier to load all parts.

**Q: Do I need all parts to view patch info?**
A: Yes - Part 1 contains metadata but operations span all parts.

**Q: Can I create executables from multi-part patches?**
A: Yes - use `--create-exe` with multi-part patches. The generator embeds part 01 data inside the executable. External parts (.02.patch, etc.) and chunk sidecar files must be distributed alongside the .exe. The embedded executable auto-detects and loads external parts at runtime.

**Q: What if I lose Part 2 but have others?**
A: Application will fail - all parts required for complete patch.

**Q: Can old appliers use multi-part patches?**
A: No - they'll error on missing MultiPart metadata. Upgrade to latest version.

**Q: Is compression per-part or across all parts?**
A: Per-part - each .XX.patch file is independently compressed.

**Q: How much disk space for multi-part patches?**
A: Same as single-part (slightly more due to metadata duplication).
