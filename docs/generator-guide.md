# Generator Tool Guide

Complete guide to using the CyberPatchMaker generator tool for creating delta patches.

## Overview

The generator tool creates efficient binary patches between software versions. It compares two complete directory trees and generates a small patch file containing only the changes.

> **NOTE: GUI Alternative Available**
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
patch-gen --versions-dir ./versions --new-version 1.0.3 --output ./patches
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
patch-gen --from 1.0.0 --to 1.0.3 --output ./patches/custom-patch.patch
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

**`--create-exe`**
- Create self-contained CLI executable
- Embeds patch data into a standalone `.exe` file
- Creates both `.patch` file and `.exe` file
- Uses CLI applier (console interface) instead of GUI
- See [Self-Contained Executables Guide](self-contained-executables.md) for details
- Works with all generation modes (single, batch, custom paths)

**`--crp`** (Create Reverse Patch)
- Automatically create reverse patch for downgrades
- Generates both forward patch (A→B) and reverse patch (B→A)
- Enables easy version rollback without manual patch creation
- Works with `--create-exe` to generate reverse executables too
- Example: Generates `1.0.0-to-1.0.1.patch` AND `1.0.1-to-1.0.0.patch`
- Compatible with all generation modes (single, batch, custom paths)
- See [Downgrade Guide](downgrade-guide.md) for usage details

**`--savescans`** (Enable Scan Caching)
- Save directory scan results to cache for faster subsequent patch generation
- Cache stored in `.data/` directory (or custom location with `--scandata`)
- First generation: Scans and caches (normal speed)
- Subsequent generations: Loads from cache (instant, no rescanning)
- **Performance**: Small projects (5-10ms saved), Large projects (15+ minutes → <1 second)
- **Example**: War Thunder (34,650 files) - 15 minute scan → instant cache load
- Cache validates key file hash to prevent using stale data
- Works with all generation modes (batch, single, custom paths)

**`--scandata <directory>`**
- Specify custom directory for scan cache storage
- Default: `.data` (in current working directory)
- Useful for shared cache locations or specific storage needs
- Only meaningful when used with `--savescans`
- Cache files named: `scan_<version>_<hash>.json`

**`--rescan`**
- Force fresh directory scan, ignoring cached data
- Updates cache with latest file data
- Useful when files changed but need to rebuild cache
- Only meaningful when used with `--savescans`
- Ensures cache is up-to-date after file modifications

**`--jobs <n>`** (Parallel Processing)
- Number of parallel workers to use for file hashing and processing
- `0` = Auto-detect CPU cores (default, recommended)
- `1` = Single-threaded (useful for debugging)
- `2+` = Specific number of workers
- **Performance**: Significantly faster on multi-core systems (especially for large projects)
- **Example**: 4-core system can process 4 files simultaneously
- Scales based on available CPU cores and I/O bandwidth
- Works with all generation modes and compression options

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
patch-gen --from-dir D:\old\1.0.0 --to-dir C:\new\1.0.3 --output ./patch.patch
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
patch-gen --versions-dir ./versions --new-version 1.0.3 --output ./patches --verify
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
patch-gen --from 1.0.0 --to 1.0.2 --output ./patches/legacy.patch --compression gzip --verify
```

---

### Example 3: Maximum Compression

Generate patches with maximum compression for slow internet users:

```bash
patch-gen --versions-dir ./versions --new-version 1.0.3 --output ./patches --compression zstd --level 4
```

Note: Level 4 takes longer but creates smallest patches.

---

### Example 4: Downgrade Patches (Rollback)

**Generate downgrade patch to roll back from 1.0.3 to 1.0.2:**

```bash
patch-gen --from 1.0.3 --to 1.0.2 --versions-dir ./versions --output ./patches/downgrade --verify
```

**Generate all downgrade paths from current version:**

```bash
# From 1.0.3 to each previous version
patch-gen --from 1.0.3 --to 1.0.2 --versions-dir ./versions --output ./patches/downgrade
patch-gen --from 1.0.3 --to 1.0.1 --versions-dir ./versions --output ./patches/downgrade
patch-gen --from 1.0.3 --to 1.0.0 --versions-dir ./versions --output ./patches/downgrade
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
patch-gen --versions-dir ./versions --new-version 1.0.3 --output ./patches --compression none
```

---

### Example 6: Self-Contained Executables (CLI)

Create standalone executables with embedded patches for easy distribution:

```bash
# Single patch with self-contained executable
patch-gen --from-dir "C:\releases\1.0.0" --to-dir "C:\releases\1.0.1" --output ./patches --create-exe

# Batch mode with executables for all versions
patch-gen --versions-dir ./versions --new-version 1.0.3 --output ./patches --create-exe --verify
```

**Result:**
```
patches/
├── 1.0.0-to-1.0.3.patch     ← Standard patch file
├── 1.0.0-to-1.0.3.exe       ← Self-contained CLI executable
├── 1.0.1-to-1.0.3.patch
├── 1.0.1-to-1.0.3.exe       ← Self-contained CLI executable
├── 1.0.2-to-1.0.3.patch
└── 1.0.2-to-1.0.3.exe       ← Self-contained CLI executable
```

**User Experience:**
- Users download the `.exe` matching their version
- Double-click to run - shows interactive console menu
- Choose "Dry Run" to simulate, or "Apply Patch" to update
- Can toggle 1GB bypass mode if needed
- No need to download separate patch files or tools

See [Self-Contained Executables Guide](self-contained-executables.md) for complete documentation.

---

### Example 7: Create Reverse Patches for Easy Downgrades

Automatically generate reverse patches to enable version rollback:

```bash
# Single patch with reverse patch
patch-gen --from-dir "C:\releases\1.0.0" --to-dir "C:\releases\1.0.1" --output ./patches --crp --create-exe

# Batch mode with reverse patches for all versions
patch-gen --versions-dir ./versions --new-version 1.0.3 --output ./patches --crp --create-exe
```

**Result:**
```
patches/
├── 1.0.0-to-1.0.1.patch     ← Forward patch (upgrade)
├── 1.0.0-to-1.0.1.exe       ← Forward executable
├── 1.0.1-to-1.0.0.patch     ← Reverse patch (downgrade)
├── 1.0.1-to-1.0.0.exe       ← Reverse executable
├── 1.0.2-to-1.0.3.patch
├── 1.0.2-to-1.0.3.exe
├── 1.0.3-to-1.0.2.patch     ← Reverse patch
└── 1.0.3-to-1.0.2.exe       ← Reverse executable
```

**Benefits:**
- Users can easily rollback if issues occur
- No need to manually create downgrade patches
- Both upgrade and downgrade executables ready to distribute
- Automatic version safety net for production deployments

See [Downgrade Guide](downgrade-guide.md) for complete documentation.

---

### Example 8: Simple Mode for End Users

**NEW in v1.0.9**: Enable simplified interface for end users when creating self-contained executables.

When you distribute patches to clients who will give the executables to their users, you can enable **Simple Mode** to provide a streamlined, user-friendly experience. This hides advanced options and shows only what end users need.

**Using GUI Generator:**

1. Open the Patch Generator GUI
2. Configure your patch settings (from/to versions, compression, etc.)
3. Check **"Enable Simple Mode for End Users"** checkbox
4. Click "Generate Patch" or "Generate + Create EXE"

**Using CLI Generator:**

Currently, Simple Mode can only be enabled via the GUI. The CLI generator does not have a `--simple-mode` flag. The `SimpleMode` field in the patch structure is set by the GUI when the checkbox is checked.

**Note:** Simple Mode is different from the `--silent` flag (which is for fully automatic patching with no user interaction).

**What Users See (GUI exe):**
- Simple message: "You are about to patch from [version] to [version]"
- Create backup checkbox (checked by default)
- Dry Run button (to test without changes)
- Apply Patch button
- Advanced options are hidden/disabled

**What Users See (CLI exe):**
- Clean console interface showing patch info
- Simple menu with only 3 options:
  1. Dry Run (test without making changes)
  2. Apply Patch
  3. Exit
- Backup option available before applying
- No confusing technical details

**Benefits:**
- **Simplified UX**: End users see only what they need
- **Reduced Support**: Fewer questions about advanced options
- **Professional**: Cleaner interface for client distributions
- **Safety**: Critical options (verify, auto-detect) forced on
- **Flexibility**: Backup and dry run still available

**Example Workflow:**

```bash
# Software vendor creates patches (enable Simple Mode via GUI checkbox)
patch-gen --versions-dir ./releases \
          --new-version 2.0.0 \
          --output ./dist \
          --create-exe \
          --verify

# Distribute the .exe files to clients
# Clients give them to end users
# End users double-click and see simple interface
```

**When to Use Simple Mode:**
- Distributing to non-technical end users
- Client deployments where support is limited
- Enterprise environments with IT policies
- Any scenario where simplified UX is desired

**When NOT to Use Simple Mode:**
- Internal development/testing
- Technical users who need full control
- Debugging or troubleshooting patches
- Advanced deployment scenarios

**Remember:** Simple Mode (SimpleMode field) = Simplified UI | Silent Mode (--silent flag) = Automation

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

### For Large Projects (1000+ Files) - Use Scan Cache

**Enable scan caching** for massive time savings on projects with many files:

```bash
# First generation: Enable caching (scans and saves to cache)
patch-gen --versions-dir ./versions \
          --new-version 1.0.3 \
          --output ./patches \
          --savescans

# Subsequent generations: Load from cache (instant)
patch-gen --versions-dir ./versions \
          --new-version 1.0.4 \
          --output ./patches \
          --savescans
```

**Performance Benefits:**
- **Small projects** (< 1,000 files): 5-10ms saved (minimal benefit)
- **Medium projects** (1,000-10,000 files): 1-5 seconds saved
- **Large projects** (10,000+ files): 15+ minutes → <1 second (massive improvement)
- **Example**: War Thunder (34,650 files) - 15 minute scan → instant cache load

**How it Works:**
- Cache stored in `.data/` directory as JSON files
- Cache file format: `scan_<version>_<hash>.json`
- Validates key file hash before using cache (prevents stale data)
- Works with all generation modes (batch, single, custom paths)

**Custom Cache Location:**
```bash
# Shared cache for multiple developers
patch-gen --versions-dir ./versions \
          --new-version 1.0.3 \
          --output ./patches \
          --savescans \
          --scandata /shared/cache
```

**Force Rescan (Update Cache):**
```bash
# Files changed, need to update cache
patch-gen --versions-dir ./versions \
          --new-version 1.0.3 \
          --output ./patches \
          --savescans \
          --rescan
```

### For Many Versions

If you have 10+ versions:
- Generation scales linearly
- Each patch is independent
- No extra memory usage
- Time = (number of versions) × (time per patch)
- **Use scan cache** to speed up each patch dramatically

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
