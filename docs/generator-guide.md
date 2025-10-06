# Generator Tool Guide

Complete guide to using the CyberPatchMaker generator tool for creating delta patches.

## Overview

The generator tool creates efficient binary patches between software versions. It compares two complete directory trees and generates a small patch file containing only the changes.

> **💡 GUI Alternative Available**
> 
> For a visual interface with additional features like self-contained executables, see the [GUI Usage Guide](gui-usage.md).
> 
> The GUI includes:
> - User-friendly interface for all options
> - Real-time validation and progress monitoring
> - **Self-contained executable creation** (generates standalone `.exe` files with embedded patches)
> - Batch mode for generating multiple patches at once
> 
> This guide focuses on the command-line tool.

## Basic Usage

### Generate Patches from All Versions to New Version

**The most common use case** - generate patches from all existing versions to your new release:

```bash
generator --versions-dir ./versions --new-version 1.0.3 --output ./patches
```

This will:
1. Scan the new version directory (versions/1.0.3)
2. Auto-detect the key file (program.exe, game.exe, app.exe, or main.exe)
3. Register the new version
4. Find all existing versions in the versions directory
5. Generate a patch from EACH existing version to 1.0.3
6. Save patches as `{from}-to-{to}.patch`

**Example Output:**
```
Generating patches for new version 1.0.3
Scanning version 1.0.3...
Using key file: program.exe
Version 1.0.3 registered: 156 files, 12 directories

Processing version 1.0.0...
Generating patch from 1.0.0 to 1.0.3...
  5 files added
  12 files modified
  3 files deleted
  2 directories added
Patch saved to: patches/1.0.0-to-1.0.3.patch (2.1 MB)

Processing version 1.0.1...
Generating patch from 1.0.1 to 1.0.3...
  2 files added
  8 files modified
  1 file deleted
Patch saved to: patches/1.0.1-to-1.0.3.patch (1.5 MB)

Processing version 1.0.2...
Generating patch from 1.0.2 to 1.0.3...
  1 file added
  4 files modified
Patch saved to: patches/1.0.2-to-1.0.3.patch (0.8 MB)

Successfully generated 3 patches
```

---

### Generate Single Patch Between Specific Versions

To create one patch between two specific versions:

```bash
generator --from 1.0.0 --to 1.0.3 --output ./patches/custom-patch.patch
```

This requires both versions to already be registered in the system.

---

## Command-Line Options

### Required Options (Batch Mode)

**`--versions-dir <path>`**
- Directory containing version folders
- Each subfolder should be a version (e.g., 1.0.0/, 1.0.1/, 1.0.2/)
- Required when using `--new-version`

**`--new-version <version>`**
- Version number of the new release
- Must match a folder name in `--versions-dir`
- Example: `1.0.3`
- Required when using `--versions-dir`

**`--output <path>`**
- Directory where patch files will be saved
- Directory is created if it doesn't exist
- Patches named automatically: `{from}-to-{to}.patch`

### Required Options (Single Patch Mode)

**`--from <version>`**
- Source version number
- Must be a registered version
- Example: `1.0.0`

**`--to <version>`**
- Target version number
- Must be a registered version
- Example: `1.0.3`

**`--output <path>`**
- Full path to output patch file
- Example: `./patches/1.0.0-to-1.0.3.patch`

### Optional Options

**`--compression <type>`**
- Compression algorithm to use
- Options: `zstd` (default), `gzip`, `none`
- zstd provides best compression ratio and speed
- gzip is more universally compatible
- none is fastest but largest patches

**`--level <1-4>`**
- Compression level (zstd only)
- 1 = Fastest, largest files
- 3 = Default, balanced (recommended)
- 4 = Slowest, smallest files
- Higher levels take longer but save bandwidth

**`--verify`**
- Verify patches after creation
- Simulates patch application to ensure it works
- Recommended for production patches
- Adds time to generation process

**`--help`**
- Display usage information
- Shows all available options

### Custom Path Options

**`--from-dir <path>`**
- Full path to source version directory
- Overrides `--versions-dir` and `--from`
- Use when source version is not in versions directory
- Example: `D:\builds\old-version`
- Cannot be used with `--versions-dir`

**`--to-dir <path>`**
- Full path to target version directory
- Overrides `--versions-dir` and `--to`
- Use when target version is not in versions directory
- Example: `C:\projects\new-release`
- Cannot be used with `--versions-dir`

**Custom Path Example:**
```bash
# Generate patch from arbitrary locations
generator --from-dir D:\old\1.0.0 --to-dir C:\new\1.0.3 --output ./patch.patch
```

**When to Use Custom Paths:**
- Versions stored on different drives
- Network shares or external storage
- Build output directories
- Testing without copying files

---

## Version Directory Structure

Your versions directory should look like this:

```
versions/
├── 1.0.0/                  # Version folder
│   ├── program.exe         # Key file (main executable)
│   ├── data/
│   │   ├── config.json
│   │   └── assets/
│   │       └── textures/
│   └── libs/
│       └── core.dll
├── 1.0.1/                  # Another version
│   ├── program.exe
│   ├── data/
│   └── libs/
└── 1.0.2/                  # Yet another version
    ├── program.exe
    ├── data/
    └── libs/
```

**Key Points:**
- Each version must have its own folder
- Folder name should be the version number
- Each version must contain a key file (program.exe, game.exe, app.exe, or main.exe)
- Complete directory tree is scanned and hashed

---

## Key File Auto-Detection

The generator automatically detects your key file by looking for:
1. `program.exe` (most common)
2. `game.exe` (for games)
3. `app.exe` (for applications)
4. `main.exe` (alternative)

**Priority:** Checked in the order above, first one found is used.

**Key File Purpose:**
- Uniquely identifies the version
- Prevents applying patches to wrong versions
- Verified before patch application

If none of these files exist, generation will fail with an error.

---

## Compression Comparison

### zstd (Default - Recommended)

**Pros:**
- Best compression ratio (smallest patches)
- Very fast compression and decompression
- Modern algorithm optimized for binary data
- Industry standard (used by Facebook, kernel.org)

**Cons:**
- Requires zstd library (included)

**Use When:**
- Default choice for production
- Bandwidth is a concern
- Fast patching is important

**Example:** 5MB of changes → 1.2MB patch file

---

### gzip

**Pros:**
- Universally compatible
- Widely supported
- Decent compression ratio

**Cons:**
- Slower than zstd
- Larger patches than zstd
- Older algorithm

**Use When:**
- Compatibility is critical
- Target systems may not have modern libraries

**Example:** 5MB of changes → 1.8MB patch file

---

### none

**Pros:**
- Fastest generation
- No compression overhead
- Useful for testing

**Cons:**
- Largest patch files
- Wastes bandwidth
- Not recommended for production

**Use When:**
- Local testing only
- Network speed not a concern
- Debugging patch generation

**Example:** 5MB of changes → 5MB patch file

---

## Examples

### Example 1: New Release to Production

You have versions 1.0.0, 1.0.1, 1.0.2 and just built 1.0.3:

```bash
# Copy new version to versions directory
mkdir versions\1.0.3
xcopy /E C:\builds\v1.0.3\* versions\1.0.3\

# Generate all patches
generator --versions-dir ./versions --new-version 1.0.3 --output ./patches --verify
```

Result:
- `patches/1.0.0-to-1.0.3.patch` (users on 1.0.0)
- `patches/1.0.1-to-1.0.3.patch` (users on 1.0.1)
- `patches/1.0.2-to-1.0.3.patch` (users on 1.0.2)

Upload these patches to your update server.

---

### Example 2: Custom Patch with Gzip

Generate a specific patch with gzip compression:

```bash
generator --from 1.0.0 --to 1.0.2 --output ./patches/legacy.patch --compression gzip --verify
```

---

### Example 3: Maximum Compression

Generate patches with maximum compression for slow internet users:

```bash
generator --versions-dir ./versions --new-version 1.0.3 --output ./patches --compression zstd --level 4
```

Note: Level 4 takes longer but creates smallest patches.

---

### Example 4: Downgrade Patches (Rollback)

**Generate downgrade patch to roll back from 1.0.3 to 1.0.2:**

```bash
generator --from 1.0.3 --to 1.0.2 --versions-dir ./versions --output ./patches/downgrade --verify
```

**Generate all downgrade paths from current version:**

```bash
# From 1.0.3 to each previous version
generator --from 1.0.3 --to 1.0.2 --versions-dir ./versions --output ./patches/downgrade
generator --from 1.0.3 --to 1.0.1 --versions-dir ./versions --output ./patches/downgrade
generator --from 1.0.3 --to 1.0.0 --versions-dir ./versions --output ./patches/downgrade
```

**Result:**
```
patches/downgrade/
├── 1.0.3-to-1.0.2.patch
├── 1.0.3-to-1.0.1.patch
└── 1.0.3-to-1.0.0.patch
```

**Key Points:**
- Downgrade patches work exactly like upgrade patches
- Simply swap the `--from` and `--to` parameters
- The generator creates a patch that reverses all changes
- Users can safely rollback to previous versions
- See [Downgrade Guide](downgrade-guide.md) for complete documentation

---

### Example 5: Fast Generation for Testing

Generate patches quickly without compression for local testing:

```bash
generator --versions-dir ./versions --new-version 1.0.3 --output ./patches --compression none
```

---

## Understanding Patch Size

### Factors Affecting Patch Size

1. **Amount of Changes**
   - More changed files = larger patch
   - Larger changed files = larger patch

2. **Type of Changes**
   - Text files compress very well
   - Binary files (images, videos) compress poorly
   - Executables vary based on changes

3. **Compression Settings**
   - zstd level 4: smallest, slowest
   - zstd level 1: larger, fastest
   - gzip: medium size, medium speed
   - none: no reduction

### Typical Patch Sizes

For a 5GB application:
- **Few bug fixes** (10MB changed): ~2-5MB patch
- **Feature update** (50MB changed): ~10-20MB patch
- **Major overhaul** (500MB changed): ~100-200MB patch

---

## Performance Tips

### For Large Applications (5GB+)

1. **Use Default Settings**: zstd level 3 is well-optimized
2. **Sufficient RAM**: Generation uses ~500MB max
3. **Fast Storage**: SSD recommended for large version folders

### For Many Versions

If you have 10+ versions:
- Generation scales linearly
- Each patch is independent
- No extra memory usage
- Time = (number of versions) × (time per patch)

---

## Troubleshooting

### "Key file not found"

**Problem:** None of the expected key files exist

**Solution:**
- Ensure your version has: program.exe, game.exe, app.exe, or main.exe
- If using different name, rename it to one of the above
- File must be in root of version directory

---

### "Version directory not found"

**Problem:** Specified version doesn't exist

**Solution:**
- Check folder name matches version number exactly
- Verify `--versions-dir` path is correct
- Use absolute paths if having issues

---

### "Failed to create manifest"

**Problem:** Cannot scan version directory

**Solution:**
- Check directory is readable
- Verify no permission issues
- Ensure directory is not empty
- Check disk is not full

---

### "Patch size is larger than expected"

**Problem:** Patch file is bigger than anticipated

**Solution:**
- Check what files actually changed (use file comparison tool)
- Binary files (images, videos) don't compress well
- Consider what changed - large files = large patches
- Verify compression is enabled (not `--compression none`)

---

## Best Practices

### Version Management

1. **Keep All Versions**: Never delete old version folders
2. **Consistent Structure**: Use same directory layout across versions
3. **Clean Builds**: Generate from clean, tested builds
4. **Version Numbers**: Use semantic versioning (major.minor.patch)

### Patch Generation

1. **Always Verify**: Use `--verify` flag for production patches
2. **Default Compression**: Stick with zstd unless you have a reason not to
3. **Test First**: Generate and test patches before distributing
4. **Document Changes**: Keep changelog for each version

### Distribution

1. **Multiple Paths**: Generate patches from all recent versions
2. **Legacy Support**: Keep patches for last 3-4 versions
3. **Server Storage**: Upload patches to reliable servers
4. **Checksums**: Provide SHA-256 checksums for patch files

---

## Related Documentation

- [Applier Tool Guide](applier-guide.md) - Applying patches
- [How It Works](how-it-works.md) - Understanding the patch system
- [Compression Guide](compression-guide.md) - Detailed compression info
- [Version Management](version-management.md) - Managing versions
- [Backup System](backup-system.md) - Understanding backup behavior during patching
