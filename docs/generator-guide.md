# Generator Tool Guide

Complete guide to using the CyberPatchMaker generator tool for creating patches.

## Overview

The generator tool creates efficient binary patches between software versions. It compares two complete directory trees and generates a patch file containing only the changes (modified files are stored as full replacements — there is no binary diffing).

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
Scanning version 1.0.3 at ./versions/1.0.3...
Using key file: program.exe
Version 1.0.3 registered: 156 files, 12 directories

Processing version 1.0.0...
Generating patch from 1.0.0 to 1.0.3...
Changes detected: 5 added, 12 modified, 3 deleted files, 2 added dirs, 1 deleted dirs
Patch saved to: patches/1.0.0-to-1.0.3.patch

Processing version 1.0.1...
Generating patch from 1.0.1 to 1.0.3...
Changes detected: 2 added, 8 modified, 1 deleted files, 0 added dirs, 1 deleted dirs
Patch saved to: patches/1.0.1-to-1.0.3.patch

Processing version 1.0.2...
Generating patch from 1.0.2 to 1.0.3...
Changes detected: 1 added, 4 modified, 0 deleted files, 0 added dirs, 0 deleted dirs
Patch saved to: patches/1.0.2-to-1.0.3.patch

Successfully generated 3 patches
```

---

### Generate Single Patch Between Specific Versions

To create one patch between two specific versions:

```bash
patch-gen --versions-dir ./versions --from 1.0.0 --to 1.0.3 --output ./patches
```

This registers both versions automatically during the generation process.

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
- Directory where patch files will be saved (optional)
- Defaults to the `default_patch_output` setting in the config file, then `patches`
- Directory is created if it doesn't exist
- Patches named automatically: `{from}-to-{to}.patch`

### Required Options (Single Patch Mode)

**`--versions-dir <path>`**
- Directory containing version folders
- Required to locate source and target version directories
- Example: `./versions`

**`--from <version>`**
- Source version number
- Must match a folder name in `--versions-dir`
- Example: `1.0.0`

**`--to <version>`**
- Target version number
- Must match a folder name in `--versions-dir`
- Example: `1.0.3`

**`--output <path>`**
- Full path to the output directory for patch files (optional)
- Defaults to the `default_patch_output` setting in the config file, then `patches`
- Directory is created if it doesn't exist
- Patches named automatically: `{from}-to-{to}.patch`

### Optional Options

**`--compression <type>`**
- Compression algorithm to use
- Options: `zstd` (default), `gzip`, `none`
- zstd provides best compression ratio and speed
- gzip is more universally compatible
- none is fastest but largest patches

**`--key-file <filename>`**
- Specify a custom key file to use for version identification
- Overrides auto-detection of key file (program.exe, game.exe, app.exe, main.exe)
- Example: `--key-file my_app.exe` or `--key-file custom_executable.exe`
- Useful when the main executable has a non-standard name
- Default: Auto-detects from standard key file names

**`--level <1-4>`**
- Compression level (applies to zstd and gzip)
- **zstd**: Levels 1-4 (1 = fastest/largest, 4 = slowest/smallest)
- **gzip**: Levels 1-3 (1 = fastest/largest, 3 = slowest/smallest)
- Default: 3 (balanced, recommended)
- Higher levels take longer but save bandwidth

**`--verify`**
- Accepted for compatibility, but currently has no effect
- Patch validation always runs during generation regardless of this flag and cannot be disabled
- **Note:** This flag does not add time or change behavior — validation of the in-memory patch (operation consistency, checksums, sizes) is always performed

**`--create-exe`**
- Create self-contained CLI executable
- Embeds patch data into a standalone `.exe` file
- Creates both `.patch` file and `.exe` file
- Uses CLI applier (console interface)
- Requires `patch-apply.exe` to be located in the same directory as `patch-gen.exe`
- For multi-part patches, the exe embeds part 01 only and is created only if part 01 is under 3.75GB (remaining parts are distributed alongside the exe)
- See [Self-Contained Executables Guide](self-contained-executables.md) for details
- Works with all generation modes (single, batch, custom paths)

**`--silent`**
- Embed silent mode flag into generated self-contained executables
- Only works with `--create-exe` flag (no effect without it)
- When embedded, executable automatically applies patch without user prompts
- Overrides interactive console menu - patch applies immediately on launch
- Creates log file `log_<epochtime>.txt` with execution details
- Perfect for distributing automated patches to end users
- Example: `patch-gen --create-exe --silent --from-dir v1 --to-dir v2 --output patches`
- Users run: `1.0.0-to-1.0.1.exe` (no flags needed, applies automatically)
- See [Self-Contained Executables Guide](self-contained-executables.md) for details

**`--crp`** (Create Reverse Patch)
- Automatically create reverse patch for downgrades
- Generates both forward patch (A→B) and reverse patch (B→A), reusing the same scan data (no rescan)
- Enables easy version rollback without manual patch creation
- Works with `--create-exe` to generate reverse executables too
- Example: Generates `1.0.0-to-1.0.1.patch` AND `1.0.1-to-1.0.0_rev.patch` (reverse files carry the `_rev` suffix)
- Compatible with all generation modes (single, batch, custom paths)
- See [Downgrade Guide](downgrade-guide.md) for usage details

**`--savescans`** (Enable Scan Caching)
- Save directory scan results to cache for faster subsequent patch generation
- Cache stored in `.data/` directory (or custom location with `--scandata`)
- First generation: Scans and caches (normal speed)
- Subsequent generations: Loads from cache (instant, no rescanning)
- **Performance**: Caching skips the scan on subsequent runs; see the War Thunder example below
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
- `0` = Use `worker_threads` from the config file (default; built-in config default is the CPU count) — recommended
- `1` = Single-threaded (useful for debugging)
- `2+` = Specific number of workers
- **Performance**: Faster on multi-core systems (especially for large projects); gains are limited by disk throughput
- **Example**: 4-worker run hashes 4 files simultaneously
- Works with all generation modes and compression options

**`--splitsize <size>`** (Multi-Part Split Size)
- Custom size for splitting large patches into multiple parts
- Format: Number followed by unit (M/MB for megabytes, G/GB for gigabytes)
- Examples: `2G`, `2GB`, `500M`, `500MB`
- Default: `4GB` (4 gigabytes)
- Minimum recommended: `100MB` (requires confirmation if below)
- Use cases:
  - Hosting platforms with file size limits
  - Email attachments or transfer services with restrictions
  - Keeping individual patch part files small for distribution
- Note: splitting limits part file sizes on disk; it does not reduce peak memory during generation or application
- See [Multi-Part Patches Guide](multipart-patches.md) for details

**`--bypasssplitlimit`**
- Bypass the 100MB minimum split size confirmation prompt
- Only meaningful when used with `--splitsize` below 100MB
- Without this flag: Generator prompts for confirmation if split size < 100MB
- With this flag: Generator proceeds without asking
- Warning: Very small split sizes create many parts (not recommended)
- Use case: Automated scripts where confirmation prompts would block execution

**`--version`**
- Display version information for the generator tool
- Prints version string and exits
- Example: `patch-gen --version`

**`--help`**
- Display usage information
- Shows all available options

### Custom Path Options

**`--from-dir <path>`**
- Full path to source version directory
- Used together with `--to-dir` (both are required for this mode)
- Use when source version is not in versions directory
- Example: `D:\builds\old-version`
- Takes priority over `--from`/`--to` (but not over `--new-version`); version numbers for file naming are taken from the directory base names

**`--to-dir <path>`**
- Full path to target version directory
- Used together with `--from-dir` (both are required for this mode)
- Use when target version is not in versions directory
- Example: `C:\projects\new-release`
- Takes priority over `--from`/`--to` (but not over `--new-version`); version numbers for file naming are taken from the directory base names

**Custom Path Example:**
```bash
# Generate patch from arbitrary locations
patch-gen --from-dir D:\old\1.0.0 --to-dir C:\new\1.0.3 --output ./patches
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

If none of these files exist, generation fails with an error for the new/target version and in single-patch modes. In batch mode (`--new-version`), source versions without a detectable key file are skipped with a warning.

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
patch-gen --versions-dir ./versions --from 1.0.0 --to 1.0.2 --output ./patches --compression gzip --verify
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
├── 1.0.0-to-1.0.3.patch       ← Forward patch (upgrade)
├── 1.0.0-to-1.0.3.exe         ← Forward executable
├── 1.0.3-to-1.0.0_rev.patch   ← Reverse patch (downgrade)
├── 1.0.3-to-1.0.0_rev.exe     ← Reverse executable
├── 1.0.1-to-1.0.3.patch
├── 1.0.1-to-1.0.3.exe
├── 1.0.3-to-1.0.1_rev.patch
├── 1.0.3-to-1.0.1_rev.exe
├── 1.0.2-to-1.0.3.patch
├── 1.0.2-to-1.0.3.exe
├── 1.0.3-to-1.0.2_rev.patch
└── 1.0.3-to-1.0.2_rev.exe
```

**Benefits:**
- Users can easily rollback if issues occur
- No need to manually create downgrade patches
- Both upgrade and downgrade executables ready to distribute
- Automatic version safety net for production deployments

See [Downgrade Guide](downgrade-guide.md) for complete documentation.

---

### Example 8: Simple Mode for End Users

**Available since v1.0.9**: The `SimpleMode` field in the `Patch` struct and `runSimpleMode()` function exist in the codebase. When enabled, Simple Mode provides a fully automated patching experience (automatic dry-run validation followed by patch application with no user interaction).

**Current status:** No generator code path currently sets `SimpleMode = true`. The field exists in the patch structure and the applier respects it, but it is reserved for future use. Silent Mode (`--silent` flag) is the currently available automation option for self-contained executables.

**Note:** Simple Mode is different from the `--silent` flag (Simple Mode = automated two-step flow with verbose output; Silent Mode = minimal-output automation for scripting).

---

### Example 9: Embedded Silent Mode for Automated Patches

**Available since v1.0.15**: Create self-contained executables that automatically apply patches without user interaction.

The `--silent` flag embeds silent mode into generated executables, making them perfect for automated deployments where you want users to simply run the file and have it patch automatically.

**Create Silent Mode Executable:**

```bash
# Single patch with embedded silent mode
patch-gen --from-dir "C:\releases\1.0.0" --to-dir "C:\releases\1.0.1" --output ./patches --create-exe --silent

# Batch mode with silent executables for all versions
patch-gen --versions-dir ./versions --new-version 1.0.3 --output ./patches --create-exe --silent --verify
```

**Result:**
```
patches/
├── 1.0.0-to-1.0.3.patch     ← Standard patch file
├── 1.0.0-to-1.0.3.exe       ← Silent mode embedded executable
├── 1.0.1-to-1.0.3.patch
├── 1.0.1-to-1.0.3.exe       ← Silent mode embedded executable
├── 1.0.2-to-1.0.3.patch
└── 1.0.2-to-1.0.3.exe       ← Silent mode embedded executable
```

**User Experience:**
- User downloads `1.0.0-to-1.0.3.exe`
- User double-clicks the executable
- Patch applies automatically (no prompts, no menus)
- Creates log file `log_<epochtime>.txt` (in the working directory) with patch info, status, and any errors
- Returns exit code 0 on success, 1 on failure

**Difference from Applier --silent Flag:**
- **Generator --silent**: Embeds silent mode INTO the executable itself
  - User runs: `1.0.0-to-1.0.1.exe` (applies automatically)
  - No command-line flags needed by end user
  - Silent behavior is built into the file
  
- **Applier --silent**: Command-line flag for interactive executables
  - User runs: `1.0.0-to-1.0.1.exe --silent` (requires flag)
  - Works with any self-contained executable
  - Silent behavior activated by user

**Combined with --crp for Bidirectional Silent Patches:**

```bash
# Create both upgrade and downgrade silent executables
patch-gen --from-dir "C:\releases\1.0.0" --to-dir "C:\releases\1.0.1" --output ./patches --create-exe --silent --crp
```

**Result:**
```
patches/
├── 1.0.0-to-1.0.1.patch       ← Forward patch
├── 1.0.0-to-1.0.1.exe         ← Forward silent executable (auto-upgrade)
├── 1.0.1-to-1.0.0_rev.patch   ← Reverse patch
└── 1.0.1-to-1.0.0_rev.exe     ← Reverse silent executable (auto-downgrade)
```

**When to Use Embedded Silent Mode:**
- Distributing to end users for one-click automated updates
- Enterprise deployments with minimal user interaction
- Kiosk or unattended systems that need automatic patching
- Game patches where users just want updates to "work"
- Any scenario where you want simplest possible user experience

**When NOT to Use Embedded Silent Mode:**
- Users need to review patch details before applying
- Testing environments where control is needed
- Situations where backup/dry-run decisions should be user-controlled
- When users need to specify custom target directories

**Log File Details:**
When a silent mode executable runs, it creates a timestamped log file:
- Filename: `log_<unix_timestamp>.txt`
- Location: created in the current working directory (the executable's folder when launched by double-click)
- Contents: Patch info, target directory, status, and error messages (per-operation applier progress goes to the console only)
- Enables audit trails for automated deployments

**Security Note:**
Silent mode executables apply patches immediately without confirmation. Ensure:
- Patches are from trusted sources
- Users understand what the executable does
- Files are distributed through secure channels
- Digital signatures verify authenticity

See [Self-Contained Executables Guide](self-contained-executables.md) for complete documentation on silent mode behavior and log file formats.

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

Because modified files are stored as full replacements, a patch contains roughly the full (compressed) size of every changed file.

---

## Performance Tips

### For Large Applications (5GB+)

1. **Use Default Settings**: zstd level 3 is well-optimized
2. **Sufficient RAM**: All added/modified files are held in memory during generation — ensure RAM on the order of the total changed-file size
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
- Skips the full directory scan on every subsequent generation
- **Large projects** benefit most — e.g. the real-world War Thunder scan (~34,650 files, ~15 minutes) loads from cache instantly
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

1. **Validation is automatic**: Patches are always validated during generation (the `--verify` flag has no additional effect)
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
