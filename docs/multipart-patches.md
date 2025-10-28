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

### CLI Generator

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

## Compatibility

### Backward Compatibility

- **Single-part patches** continue to work exactly as before
- **No version detection needed** - format is self-describing
- **Legacy appliers** will fail gracefully on multi-part patches (missing metadata error)

### Forward Compatibility

- **MaxPartSize is configurable** in metadata (future: adjustable limit)
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

1. **4GB fixed limit** - not user-configurable (planned for future)
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

1. **Configurable part size** - let users choose limit
2. **Parallel part loading** - speed up multi-part application
3. **Delta compression** - cross-part deduplication
4. **Smart part sizing** - better balance for many small files
5. **Integrity validation** - verify parts before generation completes
6. **Resume support** - restart failed multi-part downloads
7. **Part compression** - optimize individual part sizes further

## FAQ

**Q: Why 4GB limit?**
A: Balances between manageable part sizes and avoiding too many parts. Prevents 32-bit integer overflow issues.

**Q: Can I combine parts manually?**
A: No - the format requires proper metadata merging. Use the applier to load all parts.

**Q: Do I need all parts to view patch info?**
A: Yes - Part 1 contains metadata but operations span all parts.

**Q: Can I create executables from multi-part patches?**
A: Not currently supported - self-contained executables work with single-part only.

**Q: What if I lose Part 2 but have others?**
A: Application will fail - all parts required for complete patch.

**Q: Can old appliers use multi-part patches?**
A: No - they'll error on missing MultiPart metadata. Upgrade to latest version.

**Q: Is compression per-part or across all parts?**
A: Per-part - each .XX.patch file is independently compressed.

**Q: How much disk space for multi-part patches?**
A: Same as single-part (slightly more due to metadata duplication).
