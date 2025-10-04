# CLI Examples - CyberPatchMaker

This document provides comprehensive examples for every command-line argument available in the CyberPatchMaker generator and applier tools.

## Table of Contents

- [Generator Tool (`generator.exe`)](#generator-tool-generatorexe)
  - [Basic Usage](#basic-usage)
  - [--versions-dir](#--versions-dir)
  - [--new-version](#--new-version)
  - [--from and --to](#--from-and---to)
  - [--output](#--output)
  - [--compression](#--compression)
  - [--level](#--level)
  - [--verify](#--verify)
  - [--help](#--help-generator)
- [Applier Tool (`applier.exe`)](#applier-tool-applierexe)
  - [Basic Usage](#basic-usage-1)
  - [--patch](#--patch)
  - [--current-dir](#--current-dir)
  - [--dry-run](#--dry-run)
  - [--verify](#--verify-1)
  - [--backup](#--backup)
  - [--help](#--help-applier)
- [Common Workflows](#common-workflows)
- [Troubleshooting](#troubleshooting)

---

## Generator Tool (`generator.exe`)

The generator tool creates binary patch files by comparing two versions of your application.

### Basic Usage

**Minimum required arguments:**
```powershell
generator.exe --versions-dir <path> --from <version> --to <version> --output <path>
```

**Example:**
```powershell
generator.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\patches
```

**Output:**
```
Scanning source version 1.0.0...
Scanning target version 1.0.1...
Comparing versions...
Generating patch file...
Patch created: .\patches\1.0.0-to-1.0.1.patch
Size: 2.3 MB
Operations: Add=5, Modify=12, Delete=2
```

---

### `--versions-dir`

**Purpose:** Specifies the root directory containing all version folders.

**Type:** String (required)

**Directory Structure Expected:**
```
versions/
├── 1.0.0/
│   ├── program.exe
│   └── ...
├── 1.0.1/
│   ├── program.exe
│   └── ...
└── 1.0.2/
    ├── program.exe
    └── ...
```

#### Example 1: Local Versions Directory
```powershell
generator.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\patches
```

#### Example 2: Absolute Path
```powershell
generator.exe --versions-dir C:\MyApp\releases --from 2.5.0 --to 2.5.1 --output C:\patches
```

#### Example 3: Network Path
```powershell
generator.exe --versions-dir \\server\share\versions --from 3.0.0 --to 3.1.0 --output .\patches
```

#### Example 4: Different Drive
```powershell
generator.exe --versions-dir D:\builds\versions --from 1.0.0 --to 1.0.1 --output E:\patches
```

**What This Does:**
- Scans the versions directory for subdirectories matching version numbers
- Each subdirectory should contain a complete version of your application
- Calculates SHA-256 hashes for all files in both versions
- Identifies changes between versions

---

### `--new-version`

**Purpose:** Generate patches FROM all existing versions TO a new version (batch mode).

**Type:** String (optional, mutually exclusive with `--from`/`--to`)

**Use Case:** When you release a new version and want to create patches for all previous versions at once.

#### Example 1: Generate All Patches for New Release
```powershell
generator.exe --versions-dir .\versions --new-version 1.0.3 --output .\patches
```

**What This Does:**
If you have versions 1.0.0, 1.0.1, 1.0.2, and 1.0.3, this creates:
- `1.0.0-to-1.0.3.patch`
- `1.0.1-to-1.0.3.patch`
- `1.0.2-to-1.0.3.patch`

**Output:**
```
Found existing versions: 1.0.0, 1.0.1, 1.0.2
Target version: 1.0.3
Generating patch 1.0.0 → 1.0.3...
  Created: 1.0.0-to-1.0.3.patch (3.2 MB)
Generating patch 1.0.1 → 1.0.3...
  Created: 1.0.1-to-1.0.3.patch (2.1 MB)
Generating patch 1.0.2 → 1.0.3...
  Created: 1.0.2-to-1.0.3.patch (1.5 MB)
Successfully generated 3 patches
```

#### Example 2: New Major Version
```powershell
generator.exe --versions-dir C:\releases --new-version 2.0.0 --output C:\patches
```

**Use Case:**
- Releasing version 2.0.0
- Need to provide upgrade paths from all 1.x versions
- One command generates all necessary patches

---

### `--from` and `--to`

**Purpose:** Generate a single patch from one specific version to another.

**Type:** String (required if `--new-version` not used)

**Must be used together:** Both `--from` and `--to` are required for single patch generation.

#### Example 1: Simple Version Update
```powershell
generator.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\patches
```

**Output:**
```
Patch created: .\patches\1.0.0-to-1.0.1.patch
```

#### Example 2: Skip a Version
```powershell
generator.exe --versions-dir .\versions --from 1.0.0 --to 1.0.2 --output .\patches
```

**Use Case:**
- User on version 1.0.0 wants to jump directly to 1.0.2
- Creates a combined patch skipping 1.0.1

#### Example 3: Major Version Jump
```powershell
generator.exe --versions-dir .\versions --from 1.5.0 --to 2.0.0 --output .\patches
```

#### Example 4: Semantic Versioning
```powershell
generator.exe --versions-dir .\versions --from 2.1.3 --to 2.1.4 --output .\patches
```

**Version Number Rules:**
- Can be any string format (not just semantic versioning)
- Must exactly match directory names in `--versions-dir`
- Case-sensitive on Linux/macOS

---

### `--output`

**Purpose:** Specifies where to save the generated patch files.

**Type:** String (required)

**Behavior:**
- Creates directory if it doesn't exist
- Patch files are named automatically: `{from}-to-{to}.patch`

#### Example 1: Relative Path
```powershell
generator.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\patches
```

**Result:** Creates `.\patches\1.0.0-to-1.0.1.patch`

#### Example 2: Absolute Path
```powershell
generator.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output C:\MyApp\updates\patches
```

**Result:** Creates `C:\MyApp\updates\patches\1.0.0-to-1.0.1.patch`

#### Example 3: Network Share
```powershell
generator.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output \\fileserver\updates\patches
```

**Use Case:** Store patches on a central server for distribution

#### Example 4: Nested Directory
```powershell
generator.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\release\v1.0.1\patches
```

**Result:** Creates nested directories if they don't exist

#### Example 5: Different Drive
```powershell
generator.exe --versions-dir C:\builds\versions --from 1.0.0 --to 1.0.1 --output E:\distribution\patches
```

---

### `--compression`

**Purpose:** Choose compression algorithm for the patch file.

**Type:** String (optional, default: `zstd`)

**Valid Options:**
- `zstd` - High performance, excellent compression (default)
- `gzip` - Universal compatibility, good compression
- `none` - No compression (useful for debugging)

#### Example 1: Default (zstd)
```powershell
generator.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\patches
```

**Result:** Creates patch with zstd compression (~2.3 MB)

#### Example 2: Gzip Compression
```powershell
generator.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\patches --compression gzip
```

**Use Case:** Maximum compatibility - gzip is supported everywhere

**Result:** Creates patch with gzip compression (~2.5 MB)

**Output:**
```
Patch created: .\patches\1.0.0-to-1.0.1.patch
Compression: gzip
Size: 2.5 MB
```

#### Example 3: No Compression
```powershell
generator.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\patches --compression none
```

**Use Case:** 
- Debugging patch generation
- Network already compresses data
- Storage is not a concern

**Result:** Creates uncompressed patch (~7.8 MB)

**Output:**
```
Patch created: .\patches\1.0.0-to-1.0.1.patch
Compression: none
Size: 7.8 MB (uncompressed)
```

#### Example 4: Comparing All Compression Methods
```powershell
# Generate with zstd
generator.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\patches-zstd --compression zstd

# Generate with gzip
generator.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\patches-gzip --compression gzip

# Generate with no compression
generator.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\patches-none --compression none
```

**Results:**
```
zstd: 2.3 MB (fastest, smallest)
gzip: 2.5 MB (universal compatibility)
none: 7.8 MB (no compression overhead)
```

**Compression Recommendations:**
- **zstd**: Best for most use cases (fast + small)
- **gzip**: Use when compatibility is critical
- **none**: Use for debugging or when network/storage compresses

---

### `--level`

**Purpose:** Set compression level for finer control over speed vs size.

**Type:** Integer (optional, default: `3`)

**Valid Ranges:**
- **zstd:** 1-4 (1=fastest, 4=smallest)
- **gzip:** 1-9 (1=fastest, 9=smallest)
- **none:** Ignored

#### Example 1: Default Level
```powershell
generator.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\patches --compression zstd
```

**Result:** Uses level 3 (balanced speed and size)

#### Example 2: Fastest Compression (zstd)
```powershell
generator.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\patches --compression zstd --level 1
```

**Use Case:** Quick patch generation, size less important

**Output:**
```
Patch created: .\patches\1.0.0-to-1.0.1.patch
Compression: zstd (level 1)
Size: 2.5 MB
Time: 0.8 seconds
```

#### Example 3: Maximum Compression (zstd)
```powershell
generator.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\patches --compression zstd --level 4
```

**Use Case:** Minimize download size, time less important

**Output:**
```
Patch created: .\patches\1.0.0-to-1.0.1.patch
Compression: zstd (level 4)
Size: 2.1 MB
Time: 2.3 seconds
```

#### Example 4: Gzip with Custom Level
```powershell
generator.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\patches --compression gzip --level 6
```

**Output:**
```
Patch created: .\patches\1.0.0-to-1.0.1.patch
Compression: gzip (level 6)
Size: 2.4 MB
```

#### Example 5: Maximum Gzip Compression
```powershell
generator.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\patches --compression gzip --level 9
```

**Use Case:** Absolute minimum file size with gzip

**Performance Comparison (zstd):**
```
Level 1: 2.5 MB, 0.8 seconds (fastest)
Level 2: 2.4 MB, 1.2 seconds
Level 3: 2.3 MB, 1.8 seconds (default, balanced)
Level 4: 2.1 MB, 2.3 seconds (maximum compression)
```

**Performance Comparison (gzip):**
```
Level 1: 3.1 MB, 0.5 seconds (fastest)
Level 5: 2.6 MB, 1.0 seconds
Level 6: 2.5 MB, 1.5 seconds (default gzip)
Level 9: 2.4 MB, 3.2 seconds (maximum)
```

---

### `--verify`

**Purpose:** Verify patch integrity after generation.

**Type:** Boolean (optional, default: `true`)

**Behavior:**
- Validates patch file can be read correctly
- Checks all binary diffs are valid
- Confirms metadata is correct
- **Recommended:** Always leave enabled for production

#### Example 1: Default (Verification Enabled)
```powershell
generator.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\patches
```

**Output:**
```
Generating patch...
Patch created: .\patches\1.0.0-to-1.0.1.patch
Verifying patch integrity...
✓ Patch file is valid
✓ All operations verified
✓ Metadata is correct
Verification passed
```

#### Example 2: Explicitly Enable Verification
```powershell
generator.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\patches --verify=true
```

**Same as default behavior**

#### Example 3: Disable Verification (Not Recommended)
```powershell
generator.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\patches --verify=false
```

**Use Case:**
- Extremely time-sensitive batch generation
- Already verified in previous test run
- Debugging patch generation process

**Output:**
```
Generating patch...
Patch created: .\patches\1.0.0-to-1.0.1.patch
⚠ Verification skipped (not recommended for production)
```

#### Example 4: Verification with Multiple Patches
```powershell
generator.exe --versions-dir .\versions --new-version 1.0.3 --output .\patches --verify=true
```

**Output:**
```
Generating patch 1.0.0 → 1.0.3...
  Created: 1.0.0-to-1.0.3.patch
  Verifying... ✓ Valid
Generating patch 1.0.1 → 1.0.3...
  Created: 1.0.1-to-1.0.3.patch
  Verifying... ✓ Valid
Generating patch 1.0.2 → 1.0.3...
  Created: 1.0.2-to-1.0.3.patch
  Verifying... ✓ Valid
All patches verified successfully
```

**What Verification Checks:**
- Patch file format is valid
- Compression/decompression works
- Binary diffs are complete
- File operations are consistent
- Metadata matches expected values
- No corruption during write

---

### `--help` (Generator)

**Purpose:** Display usage information and available options.

**Type:** Boolean (optional)

#### Example 1: Show Help
```powershell
generator.exe --help
```

**Output:**
```
CyberPatchMaker - Generator Tool
Version: 0.1.0

Usage:
  generator.exe [options]

Required Arguments (Mode 1 - Single Patch):
  --versions-dir string    Directory containing version folders
  --from string           Source version number
  --to string             Target version number
  --output string         Output directory for patch file

Required Arguments (Mode 2 - Batch Generation):
  --versions-dir string    Directory containing version folders
  --new-version string    New version to generate patches for
  --output string         Output directory for patch files

Optional Arguments:
  --compression string    Compression algorithm: zstd, gzip, none (default: zstd)
  --level int            Compression level (1-4 for zstd, 1-9 for gzip) (default: 3)
  --verify               Verify patches after creation (default: true)
  --help                 Show this help message

Examples:
  # Generate single patch
  generator.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\patches

  # Generate patches for new version
  generator.exe --versions-dir .\versions --new-version 1.0.3 --output .\patches

  # Use gzip compression
  generator.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\patches --compression gzip

  # Maximum compression
  generator.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\patches --compression zstd --level 4
```

#### Example 2: Help Shortcut
```powershell
generator.exe -help
```

**Same output as `--help`**

---

## Applier Tool (`applier.exe`)

The applier tool applies binary patch files to update an application from one version to another.

### Basic Usage

**Minimum required arguments:**
```powershell
applier.exe --patch <patch-file> --current-dir <directory>
```

**Example:**
```powershell
applier.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir C:\MyApp
```

**Output:**
```
Loading patch file...
Verifying current version (1.0.0)...
✓ Version verified
Creating backup...
Applying patch operations...
  Modified: program.exe
  Added: data/newfile.json
  Modified: libs/core.dll
Verifying patched version (1.0.1)...
✓ Verification passed
Patch applied successfully!
```

---

### `--patch`

**Purpose:** Specifies the patch file to apply.

**Type:** String (required)

#### Example 1: Relative Path
```powershell
applier.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir C:\MyApp
```

#### Example 2: Absolute Path
```powershell
applier.exe --patch C:\updates\patches\1.0.0-to-1.0.1.patch --current-dir C:\MyApp
```

#### Example 3: Network Path
```powershell
applier.exe --patch \\server\updates\1.0.0-to-1.0.1.patch --current-dir C:\MyApp
```

**Use Case:** Download patches from central server

#### Example 4: Different Compression Types
```powershell
# Apply zstd-compressed patch
applier.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir C:\MyApp

# Apply gzip-compressed patch (same command, auto-detected)
applier.exe --patch .\patches-gzip\1.0.0-to-1.0.1.patch --current-dir C:\MyApp

# Apply uncompressed patch (same command, auto-detected)
applier.exe --patch .\patches-none\1.0.0-to-1.0.1.patch --current-dir C:\MyApp
```

**Note:** Compression type is auto-detected from patch file metadata.

---

### `--current-dir`

**Purpose:** Specifies the directory containing the current version to be updated.

**Type:** String (required)

**Important:** This directory will be modified in-place (with backup).

#### Example 1: Local Application Directory
```powershell
applier.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir C:\MyApp
```

**What This Does:**
- Verifies C:\MyApp contains version 1.0.0
- Creates backup
- Updates files in C:\MyApp to version 1.0.1

#### Example 2: Relative Path
```powershell
applier.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir .\application
```

#### Example 3: Different Drive
```powershell
applier.exe --patch C:\patches\1.0.0-to-1.0.1.patch --current-dir D:\MyApp
```

#### Example 4: Network Installation
```powershell
applier.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir \\server\apps\myapp
```

**Use Case:** Update applications on network shares

#### Example 5: Program Files
```powershell
applier.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir "C:\Program Files\MyApp"
```

**Note:** Requires administrator privileges for Program Files

---

### `--dry-run`

**Purpose:** Simulate patch application without making any changes.

**Type:** Boolean (optional, default: `false`)

**Use Case:** Preview what the patch will do before committing to changes.

#### Example 1: Dry Run Before Applying
```powershell
applier.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir C:\MyApp --dry-run
```

**Output:**
```
DRY RUN MODE - No changes will be made

Loading patch file...
Patch details:
  From: 1.0.0
  To: 1.0.1
  Operations: 15
  Size: 2.3 MB

Verifying current version...
✓ Current version is 1.0.0

Operations that would be performed:
  MODIFY: program.exe (2.1 MB → 2.3 MB)
  ADD:    data/newfeature.json (15 KB)
  MODIFY: libs/core.dll (512 KB → 518 KB)
  ADD:    plugins/newplugin.dll (230 KB)
  DELETE: libs/deprecated.dll

✓ Dry run completed successfully
✓ Patch is compatible with current version
✓ All operations are valid

To apply this patch for real, run without --dry-run:
  applier.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir C:\MyApp
```

#### Example 2: Dry Run with Verification
```powershell
applier.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir C:\MyApp --dry-run --verify
```

**Additional Output:**
```
Verifying file hashes...
✓ All current files verified
✓ Patch operations validated
```

#### Example 3: Check Compatibility
```powershell
# Test if patch can be applied to current installation
applier.exe --patch .\patches\1.0.1-to-1.0.2.patch --current-dir C:\MyApp --dry-run
```

**If current version is wrong:**
```
DRY RUN MODE - No changes will be made

Loading patch file...
Verifying current version...
✗ ERROR: Version mismatch
  Expected: 1.0.1
  Found: 1.0.0

This patch cannot be applied to the current version.
```

#### Example 4: Production Validation Workflow
```powershell
# Step 1: Dry run to validate
applier.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir C:\MyApp --dry-run

# Step 2: If successful, apply for real
applier.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir C:\MyApp
```

**Benefits of Dry Run:**
- Test compatibility before production
- Preview changes
- Verify no unexpected operations
- Validate disk space requirements
- Check file permissions

---

### `--verify` (Applier)

**Purpose:** Verify file hashes before and after patching.

**Type:** Boolean (optional, default: `true`)

**Behavior:**
- **Before:** Verifies current version matches expected hashes
- **After:** Verifies patched version matches target hashes

#### Example 1: Default (Verification Enabled)
```powershell
applier.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir C:\MyApp
```

**Output:**
```
Loading patch file...
Verifying current version...
  ✓ Calculating hashes (234 files)...
  ✓ program.exe: a1b2c3d4...
  ✓ libs/core.dll: e5f6g7h8...
  ✓ All files verified (234/234)
Applying patch...
Verifying patched version...
  ✓ Calculating hashes (236 files)...
  ✓ program.exe: i9j0k1l2...
  ✓ data/newfile.json: m3n4o5p6...
  ✓ All files verified (236/236)
✓ Patch applied successfully
```

#### Example 2: Explicitly Enable Verification
```powershell
applier.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir C:\MyApp --verify=true
```

**Same as default**

#### Example 3: Disable Verification (Not Recommended)
```powershell
applier.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir C:\MyApp --verify=false
```

**Use Case:**
- Trusted environment
- Already verified manually
- Speed is critical

**Output:**
```
Loading patch file...
⚠ Verification disabled
Applying patch...
✓ Patch applied
⚠ Post-verification skipped

WARNING: Verification was disabled. Use --verify=true for production.
```

#### Example 4: Verification Catches Corruption
```powershell
applier.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir C:\MyApp --verify
```

**If a file is corrupted:**
```
Loading patch file...
Verifying current version...
  ✓ program.exe: a1b2c3d4...
  ✗ libs/core.dll: Checksum mismatch!
    Expected: e5f6g7h8...
    Found:    z9y8x7w6...

✗ ERROR: Pre-verification failed
  File 'libs/core.dll' has been modified or corrupted.
  This patch cannot be applied safely.

Recommendation:
  1. Verify your installation is not corrupted
  2. Reinstall from clean source if needed
  3. Ensure no other processes modified files
```

#### Example 5: Verification Catches Wrong Version
```powershell
applier.exe --patch .\patches\1.0.1-to-1.0.2.patch --current-dir C:\MyApp --verify
```

**If current version is 1.0.0:**
```
Loading patch file...
Verifying current version...
  ✗ program.exe: Version mismatch!
    Expected: b2c3d4e5... (version 1.0.1)
    Found:    a1b2c3d4... (version 1.0.0)

✗ ERROR: This patch requires version 1.0.1
  Current version appears to be: 1.0.0

You need to apply the 1.0.0-to-1.0.1 patch first.
```

**What Verification Checks:**
- **Pre-Verification:**
  - All required files exist
  - All file hashes match expected values
  - Current version matches patch source version
  - Detects corruption or modifications

- **Post-Verification:**
  - All modified files have correct hashes
  - All added files exist with correct hashes
  - All deleted files are removed
  - Final version matches patch target version

---

### `--backup`

**Purpose:** Create backup of current files before patching.

**Type:** Boolean (optional, default: `true`)

**Behavior:**
- Creates timestamped backup directory
- Copies all files before modification
- Enables manual rollback if needed

#### Example 1: Default (Backup Enabled)
```powershell
applier.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir C:\MyApp
```

**Output:**
```
Loading patch file...
Creating backup...
  Backup location: C:\MyApp\.backup_20251004_143522
  Copying 234 files...
  ✓ Backup created
Applying patch...
✓ Patch applied successfully

Backup is stored at: C:\MyApp\.backup_20251004_143522
You can delete it manually if update is successful.
```

#### Example 2: Explicitly Enable Backup
```powershell
applier.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir C:\MyApp --backup=true
```

**Same as default**

#### Example 3: Disable Backup (Not Recommended)
```powershell
applier.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir C:\MyApp --backup=false
```

**Use Case:**
- Disk space is extremely limited
- Already have external backup
- Testing in disposable environment

**Output:**
```
Loading patch file...
⚠ Backup disabled
Applying patch...
✓ Patch applied

WARNING: No backup was created. Cannot rollback if issues occur.
```

#### Example 4: Backup with Large Application
```powershell
applier.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir C:\MyApp --backup
```

**Output for 5GB application:**
```
Loading patch file...
Creating backup...
  Backup location: C:\MyApp\.backup_20251004_143522
  Copying 5,234 files (5.2 GB)...
  Progress: [████████████████████] 100%
  ✓ Backup created (took 45 seconds)
Applying patch...
```

#### Example 5: Manual Rollback Using Backup
```powershell
# If patch causes problems, manually rollback:

# Stop application
Stop-Process -Name MyApp

# Remove patched files
Remove-Item C:\MyApp\* -Recurse -Force -Exclude .backup_*

# Restore from backup
Copy-Item C:\MyApp\.backup_20251004_143522\* C:\MyApp\ -Recurse

# Cleanup backup
Remove-Item C:\MyApp\.backup_20251004_143522 -Recurse -Force

# Restart application
Start-Process C:\MyApp\program.exe
```

#### Example 6: Automatic Rollback on Failure
```powershell
applier.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir C:\MyApp --backup --verify
```

**If post-verification fails:**
```
Loading patch file...
Creating backup...
  ✓ Backup created
Applying patch...
Verifying patched version...
  ✗ Verification failed: File checksum mismatch

✗ ERROR: Post-verification failed
  Automatically rolling back to backup...
  ✓ Rollback complete
  ✓ Original version restored

The patch was NOT applied due to verification failure.
Backup location: C:\MyApp\.backup_20251004_143522
```

**Backup Best Practices:**
- **Always enable for production** (default)
- Test rollback procedure before production deployment
- Ensure sufficient disk space (need 2x application size temporarily)
- Delete old backups after confirming successful update
- Consider external backups for critical systems

---

### `--help` (Applier)

**Purpose:** Display usage information and available options.

**Type:** Boolean (optional)

#### Example 1: Show Help
```powershell
applier.exe --help
```

**Output:**
```
CyberPatchMaker - Applier Tool
Version: 0.1.0

Usage:
  applier.exe [options]

Required Arguments:
  --patch string          Path to patch file
  --current-dir string    Directory containing current version

Optional Arguments:
  --dry-run              Simulate patch without making changes (default: false)
  --verify               Verify file hashes before and after patching (default: true)
  --backup               Create backup before patching (default: true)
  --help                 Show this help message

Examples:
  # Apply patch with full verification and backup (recommended)
  applier.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir C:\MyApp

  # Dry run to preview changes
  applier.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir C:\MyApp --dry-run

  # Apply without backup (not recommended)
  applier.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir C:\MyApp --backup=false

  # Quick apply (skip verification, not recommended for production)
  applier.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir C:\MyApp --verify=false
```

#### Example 2: Help Shortcut
```powershell
applier.exe -help
```

**Same output as `--help`**

---

## Common Workflows

### Workflow 1: Creating and Applying a Single Patch

**Step 1: Generate Patch**
```powershell
generator.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\patches
```

**Step 2: Test with Dry Run**
```powershell
applier.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir C:\MyApp --dry-run
```

**Step 3: Apply Patch**
```powershell
applier.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir C:\MyApp
```

---

### Workflow 2: Release New Version to All Users

**Scenario:** You've released version 1.0.3 and need patches for users on 1.0.0, 1.0.1, and 1.0.2.

**Step 1: Generate All Patches at Once**
```powershell
generator.exe --versions-dir .\versions --new-version 1.0.3 --output .\patches --compression zstd
```

**Step 2: Distribute Patches**
```powershell
# Copy to web server
Copy-Item .\patches\* \\webserver\downloads\updates\
```

**Step 3: Users Apply Appropriate Patch**
```powershell
# User on 1.0.0 downloads and applies:
applier.exe --patch .\downloads\1.0.0-to-1.0.3.patch --current-dir C:\MyApp

# User on 1.0.1 downloads and applies:
applier.exe --patch .\downloads\1.0.1-to-1.0.3.patch --current-dir C:\MyApp

# User on 1.0.2 downloads and applies:
applier.exe --patch .\downloads\1.0.2-to-1.0.3.patch --current-dir C:\MyApp
```

---

### Workflow 3: Multi-Hop Patching

**Scenario:** User on 1.0.0 wants to upgrade to 1.0.2, but only 1.0.0→1.0.1 and 1.0.1→1.0.2 patches exist.

**Step 1: Apply First Patch**
```powershell
applier.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir C:\MyApp --verify
```

**Step 2: Apply Second Patch**
```powershell
applier.exe --patch .\patches\1.0.1-to-1.0.2.patch --current-dir C:\MyApp --verify
```

**Final Result:** Application is now at version 1.0.2

---

### Workflow 4: Automated CI/CD Pipeline

**build-and-patch.ps1:**
```powershell
# Build new version
Write-Host "Building version 1.0.3..."
.\build-scripts\build.ps1 -Version 1.0.3 -Output .\versions\1.0.3

# Generate all patches
Write-Host "Generating patches..."
.\generator.exe --versions-dir .\versions --new-version 1.0.3 --output .\patches --compression zstd --verify

# Run automated tests on each patch
Write-Host "Testing patches..."
foreach ($patch in Get-ChildItem .\patches\*.patch) {
    Write-Host "Testing $($patch.Name)..."
    
    # Extract source version from filename (e.g., "1.0.0" from "1.0.0-to-1.0.3.patch")
    $patchName = $patch.BaseName
    $sourceVersion = ($patchName -split '-to-')[0]
    
    # Create test directory
    $testDir = ".\test-$sourceVersion"
    Copy-Item ".\versions\$sourceVersion" $testDir -Recurse
    
    # Apply patch
    .\applier.exe --patch $patch.FullName --current-dir $testDir --verify
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ Patch $($patch.Name) verified successfully"
    } else {
        Write-Host "✗ Patch $($patch.Name) FAILED verification"
        exit 1
    }
    
    # Cleanup
    Remove-Item $testDir -Recurse -Force
}

# Upload to distribution server
Write-Host "Uploading patches..."
Copy-Item .\patches\*.patch \\fileserver\updates\1.0.3\

Write-Host "✓ Build and patch generation complete"
```

---

### Workflow 5: Testing Before Production

**test-patch.ps1:**
```powershell
param(
    [string]$PatchFile,
    [string]$TestDir
)

Write-Host "Testing patch: $PatchFile"
Write-Host "Target directory: $TestDir"

# Step 1: Dry run
Write-Host "`nStep 1: Dry run..."
.\applier.exe --patch $PatchFile --current-dir $TestDir --dry-run
if ($LASTEXITCODE -ne 0) {
    Write-Host "✗ Dry run failed"
    exit 1
}
Write-Host "✓ Dry run passed"

# Step 2: Apply with verification
Write-Host "`nStep 2: Applying patch with verification..."
.\applier.exe --patch $PatchFile --current-dir $TestDir --verify --backup
if ($LASTEXITCODE -ne 0) {
    Write-Host "✗ Patch application failed"
    exit 1
}
Write-Host "✓ Patch applied successfully"

# Step 3: Run application tests
Write-Host "`nStep 3: Running application tests..."
& "$TestDir\program.exe" --self-test
if ($LASTEXITCODE -ne 0) {
    Write-Host "✗ Application tests failed"
    
    # Rollback
    Write-Host "Rolling back..."
    $backupDir = Get-ChildItem "$TestDir\.backup_*" | Sort-Object -Descending | Select-Object -First 1
    Remove-Item "$TestDir\*" -Recurse -Force -Exclude .backup_*
    Copy-Item "$($backupDir.FullName)\*" $TestDir -Recurse
    Write-Host "✓ Rollback complete"
    exit 1
}
Write-Host "✓ Application tests passed"

Write-Host "`n✓ All tests passed - patch is safe to deploy"
```

**Usage:**
```powershell
.\test-patch.ps1 -PatchFile .\patches\1.0.0-to-1.0.1.patch -TestDir C:\test-environment\app
```

---

### Workflow 6: Compression Comparison

**Compare compression methods to choose the best for your use case:**

```powershell
# Generate with all compression types
Write-Host "Generating patches with different compression..."

# Zstd
Measure-Command {
    .\generator.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\patches-zstd --compression zstd
} | Select-Object -ExpandProperty TotalSeconds | Tee-Object -Variable zstdTime

# Gzip
Measure-Command {
    .\generator.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\patches-gzip --compression gzip
} | Select-Object -ExpandProperty TotalSeconds | Tee-Object -Variable gzipTime

# None
Measure-Command {
    .\generator.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\patches-none --compression none
} | Select-Object -ExpandProperty TotalSeconds | Tee-Object -Variable noneTime

# Compare sizes
$zstdSize = (Get-Item .\patches-zstd\1.0.0-to-1.0.1.patch).Length
$gzipSize = (Get-Item .\patches-gzip\1.0.0-to-1.0.1.patch).Length
$noneSize = (Get-Item .\patches-none\1.0.0-to-1.0.1.patch).Length

Write-Host "`nCompression Comparison:"
Write-Host "  Zstd: $([Math]::Round($zstdSize/1MB, 2)) MB in $([Math]::Round($zstdTime, 2))s"
Write-Host "  Gzip: $([Math]::Round($gzipSize/1MB, 2)) MB in $([Math]::Round($gzipTime, 2))s"
Write-Host "  None: $([Math]::Round($noneSize/1MB, 2)) MB in $([Math]::Round($noneTime, 2))s"

Write-Host "`nSize Comparison (vs zstd):"
Write-Host "  Gzip: $([Math]::Round(($gzipSize/$zstdSize - 1) * 100, 1))% difference"
Write-Host "  None: +$([Math]::Round(($noneSize/$zstdSize - 1) * 100, 1))% larger"

Write-Host "`nRecommendation:"
if ($zstdSize -le $gzipSize -and $zstdTime -le $gzipTime) {
    Write-Host "  Use zstd (smallest and fastest)"
} elseif ($gzipSize -lt $zstdSize) {
    Write-Host "  Use gzip (better compression for this data)"
} else {
    Write-Host "  Use zstd for speed, gzip for compatibility"
}
```

---

## Troubleshooting

### Error: "versions-dir is required"

**Command:**
```powershell
generator.exe --from 1.0.0 --to 1.0.1 --output .\patches
```

**Error:**
```
Error: --versions-dir is required
```

**Solution:** Add `--versions-dir` argument
```powershell
generator.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\patches
```

---

### Error: "Version directory not found"

**Command:**
```powershell
generator.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\patches
```

**Error:**
```
Error: Version directory not found: .\versions\1.0.0
```

**Causes:**
- Directory doesn't exist
- Wrong path
- Typo in version number

**Solution:** Verify directory exists
```powershell
# Check what versions exist
Get-ChildItem .\versions

# Fix command with correct version number
generator.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\patches
```

---

### Error: "Patch file not found"

**Command:**
```powershell
applier.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir C:\MyApp
```

**Error:**
```
Error: Patch file not found: .\patches\1.0.0-to-1.0.1.patch
```

**Solution:** Check patch file path
```powershell
# Verify file exists
Test-Path .\patches\1.0.0-to-1.0.1.patch

# List available patches
Get-ChildItem .\patches\*.patch
```

---

### Error: "Version mismatch"

**Command:**
```powershell
applier.exe --patch .\patches\1.0.1-to-1.0.2.patch --current-dir C:\MyApp
```

**Error:**
```
Error: Version mismatch
  Expected: 1.0.1
  Found: 1.0.0
```

**Cause:** Trying to apply wrong patch

**Solution:** Apply correct patch for current version
```powershell
# Apply 1.0.0 to 1.0.1 first
applier.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir C:\MyApp

# Then apply 1.0.1 to 1.0.2
applier.exe --patch .\patches\1.0.1-to-1.0.2.patch --current-dir C:\MyApp
```

---

### Error: "Insufficient disk space"

**Command:**
```powershell
applier.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir C:\MyApp
```

**Error:**
```
Error: Insufficient disk space
  Required: 10.5 GB (5 GB app + 5.5 GB backup)
  Available: 3.2 GB
```

**Solution:** Free up disk space or disable backup
```powershell
# Option 1: Free up space and retry
# Option 2: Disable backup (not recommended)
applier.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir C:\MyApp --backup=false
```

---

### Error: "Permission denied"

**Command:**
```powershell
applier.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir "C:\Program Files\MyApp"
```

**Error:**
```
Error: Permission denied
  Cannot write to: C:\Program Files\MyApp\program.exe
```

**Solution:** Run as administrator
```powershell
# Open PowerShell as Administrator, then run:
applier.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir "C:\Program Files\MyApp"
```

---

## Related Documentation

- [README.md](../README.md) - Project overview and quick start
- [ADVANCED-TEST-SUMMARY.md](ADVANCED-TEST-SUMMARY.md) - Detailed test suite explanation
- [Testing Guide](testing-guide.md) - Running tests
- [Generator Guide](generator-guide.md) - Generator tool documentation
- [Applier Guide](applier-guide.md) - Applier tool documentation
- [Troubleshooting](troubleshooting.md) - Common issues and solutions

---

## Conclusion

This guide covered every CLI argument with comprehensive examples. Key takeaways:

**Generator:**
- Use `--new-version` for batch generation (multiple patches at once)
- Use `--from`/`--to` for single patches
- Choose compression based on needs: zstd (speed), gzip (compatibility), none (debugging)
- Always leave `--verify` enabled for production

**Applier:**
- Use `--dry-run` to preview changes before applying
- Always leave `--verify` enabled to catch corruption and version mismatches
- Always leave `--backup` enabled for safety (can rollback if issues occur)
- Check available disk space (need 2x application size for backup)

**Best Practices:**
1. Test patches with `--dry-run` before production
2. Keep `--verify` and `--backup` enabled (defaults)
3. Choose appropriate compression for your use case
4. Automate testing in CI/CD pipelines
5. Maintain good version control of your builds

For more information, see the related documentation links above.
