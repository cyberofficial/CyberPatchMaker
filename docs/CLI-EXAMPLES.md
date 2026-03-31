# CLI Examples - CyberPatchMaker

This document provides comprehensive examples for every command-line argument available in the CyberPatchMaker generator and applier tools.

## Table of Contents

- [Generator Tool (`patch-gen.exe`)](#generator-tool-generatorexe)
  - [Basic Usage](#basic-usage)
  - [--versions-dir](#--versions-dir)
  - [--new-version](#--new-version)
  - [--from and --to](#--from-and---to)
  - [--from-dir and --to-dir](#--from-dir-and---to-dir)
  - [--output](#--output)
  - [--key-file (Generator)](#--key-file-generator)
  - [--compression](#--compression)
  - [--level](#--level)
  - [--verify (Generator)](#--verify-generator)
  - [--create-exe](#--create-exe)
  - [--silent (Generator)](#--silent-generator)
  - [--crp](#--crp)
  - [--savescans](#--savescans)
  - [--scandata](#--scandata)
  - [--rescan](#--rescan)
  - [--jobs](#--jobs)
  - [--splitsize](#--splitsize)
  - [--bypasssplitlimit](#--bypasssplitlimit)
  - [--version (Generator)](#--version-generator)
  - [--help (Generator)](#--help-generator)
- [Applier Tool (`patch-apply.exe`)](#applier-tool-applierexe)
  - [Basic Usage](#basic-usage-1)
  - [--patch](#--patch)
  - [--current-dir](#--current-dir)
  - [--key-file (Applier)](#--key-file-applier)
  - [--dry-run](#--dry-run)
  - [--verify (Applier)](#--verify-applier)
  - [--backup](#--backup)
  - [--ignore1gb](#--ignore1gb)
  - [--silent (Applier)](#--silent-applier)
  - [--version (Applier)](#--version-applier)
  - [--help (Applier)](#--help-applier)
- [Common Workflows](#common-workflows)
- [Troubleshooting](#troubleshooting)

---

## Generator Tool (`patch-gen.exe`)

The generator tool creates binary patch files by comparing two versions of your application.

### Basic Usage

**Minimum required arguments:**
```powershell
patch-gen.exe --versions-dir <path> --from <version> --to <version> --output <path>
```

**Example:**
```powershell
patch-gen.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\patches
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
patch-gen.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\patches
```

#### Example 2: Absolute Path
```powershell
patch-gen.exe --versions-dir C:\MyApp\releases --from 2.5.0 --to 2.5.1 --output C:\patches
```

#### Example 3: Network Path
```powershell
patch-gen.exe --versions-dir \\server\share\versions --from 3.0.0 --to 3.1.0 --output .\patches
```

#### Example 4: Different Drive
```powershell
patch-gen.exe --versions-dir D:\builds\versions --from 1.0.0 --to 1.0.1 --output E:\patches
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
patch-gen.exe --versions-dir .\versions --new-version 1.0.3 --output .\patches
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
patch-gen.exe --versions-dir C:\releases --new-version 2.0.0 --output C:\patches
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
patch-gen.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\patches
```

**Output:**
```
Patch created: .\patches\1.0.0-to-1.0.1.patch
```

#### Example 2: Skip a Version
```powershell
patch-gen.exe --versions-dir .\versions --from 1.0.0 --to 1.0.2 --output .\patches
```

**Use Case:**
- User on version 1.0.0 wants to jump directly to 1.0.2
- Creates a combined patch skipping 1.0.1

#### Example 3: Major Version Jump
```powershell
patch-gen.exe --versions-dir .\versions --from 1.5.0 --to 2.0.0 --output .\patches
```

#### Example 4: Semantic Versioning
```powershell
patch-gen.exe --versions-dir .\versions --from 2.1.3 --to 2.1.4 --output .\patches
```

**Version Number Rules:**
- Can be any string format (not just semantic versioning)
- Must exactly match directory names in `--versions-dir`
- Case-sensitive on Linux/macOS

---

### `--from-dir` and `--to-dir`

**Purpose:** Specify full paths to source and target version directories directly, bypassing `--versions-dir`/`--from`/`--to`.

**Type:** String (optional, overrides `--versions-dir`/`--from`/`--to`)

**Behavior:**
- Use when versions are on different drives, network locations, or not in a shared parent directory
- `--from-dir` sets the full path to the source version directory
- `--to-dir` sets the full path to the target version directory
- Both must be specified together

#### Example 1: Versions on Different Drives
```powershell
patch-gen.exe --from-dir "C:\releases\1.0.0" --to-dir "D:\builds\1.0.1" --output .\patches
```

#### Example 2: Network Locations
```powershell
patch-gen.exe --from-dir "\\server1\app\v1" --to-dir "\\server2\app\v2" --output .\patches
```

#### Example 3: Combined with Other Flags
```powershell
patch-gen.exe --from-dir C:\v1 --to-dir C:\v2 --output .\patches --create-exe --compression zstd --level 4
```

---

### `--output`

**Purpose:** Specifies where to save the generated patch files.

**Type:** String (required)

**Behavior:**
- Creates directory if it doesn't exist
- Patch files are named automatically: `{from}-to-{to}.patch`

#### Example 1: Relative Path
```powershell
patch-gen.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\patches
```

**Result:** Creates `.\patches\1.0.0-to-1.0.1.patch`

#### Example 2: Absolute Path
```powershell
patch-gen.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output C:\MyApp\updates\patches
```

**Result:** Creates `C:\MyApp\updates\patches\1.0.0-to-1.0.1.patch`

#### Example 3: Network Share
```powershell
patch-gen.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output \\fileserver\updates\patches
```

**Use Case:** Store patches on a central server for distribution

#### Example 4: Nested Directory
```powershell
patch-gen.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\release\v1.0.1\patches
```

**Result:** Creates nested directories if they don't exist

#### Example 5: Different Drive
```powershell
patch-gen.exe --versions-dir C:\builds\versions --from 1.0.0 --to 1.0.1 --output E:\distribution\patches
```

---

### `--key-file` (Generator)

**Purpose:** Specify which file to use as the key file for version identification.

**Type:** String (optional)

**Behavior:**
- By default, the generator auto-detects the key file (e.g., the main executable)
- Use this flag to explicitly specify which file should be used as the key file
- The key file's checksum is used to verify the correct version when applying patches

#### Example 1: Specify Custom Key File
```powershell
patch-gen.exe --from-dir C:\v1 --to-dir C:\v2 --output .\patches --key-file app.exe
```

#### Example 2: Game with Custom Launcher
```powershell
patch-gen.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\patches --key-file game_launcher.exe
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
patch-gen.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\patches
```

**Result:** Creates patch with zstd compression (~2.3 MB)

#### Example 2: Gzip Compression
```powershell
patch-gen.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\patches --compression gzip
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
patch-gen.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\patches --compression none
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
patch-gen.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\patches-zstd --compression zstd

# Generate with gzip
patch-gen.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\patches-gzip --compression gzip

# Generate with no compression
patch-gen.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\patches-none --compression none
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
- **gzip:** 1-3 (1=fastest, 3=smallest)
- **none:** Ignored

#### Example 1: Default Level
```powershell
patch-gen.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\patches --compression zstd
```

**Result:** Uses level 3 (balanced speed and size)

#### Example 2: Fastest Compression (zstd)
```powershell
patch-gen.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\patches --compression zstd --level 1
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
patch-gen.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\patches --compression zstd --level 4
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
patch-gen.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\patches --compression gzip --level 6
```

**Output:**
```
Patch created: .\patches\1.0.0-to-1.0.1.patch
Compression: gzip (level 6)
Size: 2.4 MB
```

#### Example 5: Maximum Gzip Compression
```powershell
patch-gen.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\patches --compression gzip --level 9
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
patch-gen.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\patches
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
patch-gen.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\patches --verify=true
```

**Same as default behavior**

#### Example 3: Disable Verification (Not Recommended)
```powershell
patch-gen.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\patches --verify=false
```

**Use Case:**
- Extremely time-sensitive batch generation
- Already verified in previous test run
- Debugging patch generation process

**Output:**
```
Generating patch...
Patch created: .\patches\1.0.0-to-1.0.1.patch
WARNING: Verification skipped (not recommended for production)
```

#### Example 4: Verification with Multiple Patches
```powershell
patch-gen.exe --versions-dir .\versions --new-version 1.0.3 --output .\patches --verify=true
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

### `--create-exe`

**Purpose:** Create self-contained CLI executable with embedded patch data.

**Type:** Boolean (optional, default: `false`)

**Behavior:**
- Creates both `.patch` file and `.exe` file
- Embeds patch data into CLI applier executable
- Results in standalone executable that users can run directly
- Uses console interface
- See [Self-Contained Executables Guide](self-contained-executables.md) for details

#### Example 1: Single Patch with Executable
```powershell
patch-gen.exe --from-dir "C:\releases\1.0.0" --to-dir "C:\releases\1.0.1" --output .\patches --create-exe
```

**Output:**
```
Generating patch from 1.0.0 to 1.0.1...
Patch saved to: .\patches\1.0.0-to-1.0.1.patch
✓ Created executable: .\patches\1.0.0-to-1.0.1.exe
```

**Result:**
```
patches/
├── 1.0.0-to-1.0.1.patch     (2.3 MB - standard patch file)
└── 1.0.0-to-1.0.1.exe       (52.8 MB - self-contained executable)
```

#### Example 2: Batch Mode with Executables
```powershell
patch-gen.exe --versions-dir .\versions --new-version 1.0.3 --output .\patches --create-exe
```

**Output:**
```
Generating patches for new version 1.0.3
Processing version 1.0.0...
  Patch saved: 1.0.0-to-1.0.3.patch
  ✓ Created executable: 1.0.0-to-1.0.3.exe
Processing version 1.0.1...
  Patch saved: 1.0.1-to-1.0.3.patch
  ✓ Created executable: 1.0.1-to-1.0.3.exe
Processing version 1.0.2...
  Patch saved: 1.0.2-to-1.0.3.patch
  ✓ Created executable: 1.0.2-to-1.0.3.exe
Successfully generated 3 patches and executables
```

#### Example 3: With Verification
```powershell
patch-gen.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\patches --create-exe --verify
```

**Verifies patch before creating executable**

#### Example 4: Custom Compression with Executable
```powershell
patch-gen.exe --from-dir "D:\v1" --to-dir "D:\v2" --output .\dist --compression zstd --level 4 --create-exe
```

**Creates highly compressed self-contained executable**

**User Experience:**
When users run the created `.exe` file:
```
==============================================
  CyberPatchMaker - Self-Contained Patch
==============================================

=== Patch Information ===
From Version:     1.0.0
To Version:       1.0.1
Key File:         program.exe
Files Added:      5
Files Modified:   12
Files Deleted:    2

Target directory [C:\Program Files\MyApp]:

==============================================
Options:
  1. Dry Run (simulate without changes)
  2. Apply Patch
  3. Toggle 1GB Bypass Mode (currently: Disabled)
  4. Change Target Directory
  5. Specify Custom Key File
     (Currently: program.exe - default)
  6. Exit
==============================================
Select option [1-6]:
```

**Benefits:**
- Users only need one file
- No separate tools required
- Interactive console interface
- Can't select wrong patch file
- Includes dry-run option
- 1GB bypass toggle available

**Considerations:**
- Larger file size (~50 MB base + patch data)
- Higher bandwidth for distribution
- Requires `patch-apply.exe` in same directory as generator

---

### `--silent` (Generator)

**Purpose:** Enable silent mode in the generated executable so it auto-applies without prompts.

**Type:** Boolean (optional, default: `false`)

**Behavior:**
- Only applies when used with `--create-exe`
- Embeds a silent flag into the self-contained executable header
- When end users run the executable, it applies the patch automatically without showing a menu
- No user interaction required -- perfect for automation and non-technical users

#### Example 1: Create Silent Self-Contained Executable
```powershell
patch-gen.exe --from-dir C:\v1 --to-dir C:\v2 --output .\patches --create-exe --silent
```

**Output:**
```
Generating patch from 1.0.0 to 1.0.1...
Patch saved to: .\patches\1.0.0-to-1.0.1.patch
✓ Created executable (silent mode): .\patches\1.0.0-to-1.0.1.exe
```

#### Example 2: Batch with Silent Mode
```powershell
patch-gen.exe --versions-dir .\versions --new-version 1.0.3 --output .\patches --create-exe --silent
```

**When end user runs the executable:**
```
========================================
CyberPatchMaker Silent Mode Log
Started: 2026-03-30 12:00:00
========================================

Patch Information:
  From Version: 1.0.0
  To Version:   1.0.1
  ...

Patch applied successfully: 1.0.0 → 1.0.1

========================================
Status: SUCCESS
========================================
```

---

### `--crp`

**Purpose:** Create a reverse patch (for downgrades) alongside the forward patch.

**Type:** Boolean (optional, default: `false`)

**Behavior:**
- Generates a second patch that allows downgrading from the target version back to the source version
- The reverse patch filename follows the pattern `<to>-to-<from>.patch`
- Useful for allowing users to roll back if an update causes issues

#### Example 1: Forward and Reverse Patches
```powershell
patch-gen.exe --from-dir C:\v1.0.0 --to-dir C:\v1.0.1 --output .\patches --crp
```

**Output:**
```
patches/
├── 1.0.0-to-1.0.1.patch     (forward patch)
└── 1.0.1-to-1.0.0.patch     (reverse patch)
```

#### Example 2: Reverse Patches with Self-Contained Executables
```powershell
patch-gen.exe --from-dir C:\v1.0.0 --to-dir C:\v1.0.1 --output .\patches --crp --create-exe
```

**Output:**
```
patches/
├── 1.0.0-to-1.0.1.patch
├── 1.0.0-to-1.0.1.exe       (upgrade executable)
├── 1.0.1-to-1.0.0.patch
└── 1.0.1-to-1.0.0.exe       (downgrade executable)
```

---

### `--savescans`

**Purpose:** Save directory scans to cache for faster subsequent patch generation.

**Type:** Boolean (optional, default: `false`)

**Behavior:**
- Scans directory contents and saves metadata to a cache directory
- On subsequent runs with the same versions, loads from cache instead of rescanning
- Dramatically speeds up repeated patch generation for large projects
- Default cache directory is `.data/` (can be changed with `--scandata`)

#### Example 1: Enable Scan Caching
```powershell
patch-gen.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\patches --savescans
```

**Output:**
```
✓ Scan caching enabled (cache dir: .data)
Scanning source version 1.0.0...
Scanning target version 1.0.1...
...
```

#### Example 2: Subsequent Run Uses Cache
```powershell
patch-gen.exe --versions-dir .\versions --from 1.0.1 --to 1.0.2 --output .\patches --savescans
```

**Note:** Version 1.0.1 is loaded from cache (no rescan needed), only 1.0.2 is scanned fresh.

---

### `--scandata`

**Purpose:** Specify a custom directory for the scan cache.

**Type:** String (optional, default: `.data`)

**Behavior:**
- Only applies when used with `--savescans`
- Overrides the default `.data/` cache directory
- Useful for shared cache locations or CI/CD environments

#### Example 1: Custom Cache Directory
```powershell
patch-gen.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\patches --savescans --scandata ./shared-cache
```

---

### `--rescan`

**Purpose:** Force rescan of cached versions, ignoring existing cache data.

**Type:** Boolean (optional, default: `false`)

**Behavior:**
- Only applies when used with `--savescans`
- Forces a fresh scan even if cache exists for the version
- Useful when files have changed on disk since the last scan

#### Example 1: Force Rescan
```powershell
patch-gen.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\patches --savescans --rescan
```

---

### `--jobs`

**Purpose:** Set the number of parallel worker threads for patch generation.

**Type:** Integer (optional, default: `0`)

**Behavior:**
- `0` = Auto-detect (uses number of CPU cores)
- `1` = Single-threaded (no parallelism)
- `2+` = Use that many worker threads
- Improves performance for large projects with many files

#### Example 1: Auto-Detect Workers
```powershell
patch-gen.exe --from-dir C:\v1 --to-dir C:\v2 --output .\patches --jobs 0
```

**Output:**
```
✓ Using 8 worker threads for parallel operations
...
```

#### Example 2: Single-Threaded
```powershell
patch-gen.exe --from-dir C:\v1 --to-dir C:\v2 --output .\patches --jobs 1
```

#### Example 3: Custom Worker Count
```powershell
patch-gen.exe --from-dir C:\v1 --to-dir C:\v2 --output .\patches --jobs 16
```

---

### `--splitsize`

**Purpose:** Set a custom split size for multi-part patch files.

**Type:** String (optional, default: `4GB`)

**Behavior:**
- Accepts sizes like `2G`, `2GB`, `500M`, `500MB`
- Patches larger than this size are split into multiple parts
- Default split size is 4GB
- Minimum recommended size is 100MB (see `--bypasssplitlimit`)

#### Example 1: Split at 2GB
```powershell
patch-gen.exe --from-dir C:\v1 --to-dir C:\v2 --output .\patches --splitsize 2G
```

#### Example 2: Split at 500MB
```powershell
patch-gen.exe --from-dir C:\v1 --to-dir C:\v2 --output .\patches --splitsize 500MB
```

---

### `--bypasssplitlimit`

**Purpose:** Bypass the 100MB minimum split size safety check.

**Type:** Boolean (optional, default: `false`)

**Behavior:**
- When `--splitsize` is set below 100MB, the generator prompts for confirmation
- Use this flag to skip the confirmation prompt
- Warning: Very small split sizes may create many small parts, which is not recommended

#### Example 1: Small Split Size with Bypass
```powershell
patch-gen.exe --from-dir C:\v1 --to-dir C:\v2 --output .\patches --splitsize 50M --bypasssplitlimit
```

---

### `--version` (Generator)

**Purpose:** Display the generator version information.

**Type:** Boolean (optional)

#### Example 1: Show Version
```powershell
patch-gen.exe --version
```

**Output:**
```
CyberPatchMaker Patch Generator v1.0.16
```

---

### `--help` (Generator)

**Purpose:** Display usage information and available options.

**Type:** Boolean (optional)

#### Example 1: Show Help
```powershell
patch-gen.exe --help
```

**Output:**
```
CyberPatchMaker - Patch Generator v1.0.16

Usage:
  Generate patches from all versions to new version:
    patch-gen --versions-dir <dir> --new-version <version>

  Generate single patch (versions in same directory):
    patch-gen --versions-dir <dir> --from <version> --to <version>

  Generate single patch (custom paths, different drives/locations):
    patch-gen --from-dir <path> --to-dir <path>

Options:
  --versions-dir    Directory containing version folders
  --new-version     New version number to generate patches for
  --from            Source version number (with --versions-dir)
  --to              Target version number (with --versions-dir)
  --from-dir        Full path to source version directory
  --to-dir          Full path to target version directory
  --output          Output directory for patches (default: patches)
  --key-file        Specific key file to use (e.g., app_name.exe)
  --compression     Compression algorithm: zstd, gzip, none (default: zstd)
  --level           Compression level (default: 3)
  --verify          Verify patches after creation (default: true)
  --create-exe      Create self-contained CLI executable
  --silent          Enable silent mode in generated executable (auto-apply without prompts)
  --crp             Create reverse patch (for downgrades)
  --savescans       Save directory scans to cache for faster subsequent patches
  --rescan          Force rescan of cached versions (use with --savescans)
  --scandata        Custom directory for scan cache (default: .data)
  --jobs            Number of parallel workers (0=auto-detect CPU cores, 1=single-threaded, default: 0)
  --splitsize       Custom multi-part split size (e.g., '2G', '2GB', '500M', '500MB', default: 4GB)
  --bypasssplitlimit Bypass 100MB minimum split size confirmation
  --version         Show version information
  --help            Show this help message

Examples:
  # Versions on different drives
  patch-gen --from-dir C:\releases\1.0.0 --to-dir D:\builds\1.0.1 --output patches

  # Create self-contained executable
  patch-gen --from-dir C:\\v1 --to-dir C:\\v2 --output patches --create-exe

  # Create forward and reverse patches with executables
  patch-gen --from-dir C:\\v1.0.0 --to-dir C:\\v1.0.1 --output patches --crp --create-exe

  # Use scan caching for faster subsequent patches
  patch-gen --versions-dir C:\\versions --from 1.0.0 --to 1.0.1 --output patches --savescans
  patch-gen --versions-dir C:\\versions --from 1.0.1 --to 1.0.2 --output patches --savescans

  # Use parallel workers for faster processing (large projects)
  patch-gen --from-dir C:\\v1 --to-dir C:\\v2 --output patches --jobs 0
  patch-gen --from-dir C:\\v1 --to-dir C:\\v2 --output patches --jobs 8

  # Force rescan of cached versions
  patch-gen --versions-dir C:\\versions --from 1.0.0 --to 1.0.1 --output patches --savescans --rescan

  # Custom split size for multi-part patches
  patch-gen --from-dir C:\\v1 --to-dir C:\\v2 --output patches --splitsize 2G
  patch-gen --from-dir C:\\v1 --to-dir C:\\v2 --output patches --splitsize 500MB

  # Small split size (below 100MB) with bypass
  patch-gen --from-dir C:\\v1 --to-dir C:\\v2 --output patches --splitsize 50M --bypasssplitlimit

  # Versions on different network locations
  patch-gen --from-dir \\server1\app\v1 --to-dir \\server2\app\v2 --output .
```

#### Example 2: Help Shortcut
```powershell
patch-gen.exe -help
```

**Same output as `--help`**

---

## Applier Tool (`patch-apply.exe`)

The applier tool applies binary patch files to update an application from one version to another.

### Basic Usage

**Minimum required arguments:**
```powershell
patch-apply.exe --patch <patch-file> --current-dir <directory>
```

**Example:**
```powershell
patch-apply.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir C:\MyApp
```

**Output:**
```
Loading patch file...
Verifying current version (1.0.0)...
✓ Version verified
Creating selective backup...
  Backing up: program.exe
  Backing up: libs\core.dll
  ✓ Selective backup created
Applying patch operations...
  Modified: program.exe
  Added: data/newfile.json (not backed up - didn't exist)
  Modified: libs/core.dll
Verifying patched version (1.0.1)...
✓ Verification passed
Patch applied successfully!
Backup preserved in: C:\MyApp\backup.cyberpatcher
```

---

### `--patch`

**Purpose:** Specifies the patch file to apply.

**Type:** String (required)

#### Example 1: Relative Path
```powershell
patch-apply.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir C:\MyApp
```

#### Example 2: Absolute Path
```powershell
patch-apply.exe --patch C:\updates\patches\1.0.0-to-1.0.1.patch --current-dir C:\MyApp
```

#### Example 3: Network Path
```powershell
patch-apply.exe --patch \\server\updates\1.0.0-to-1.0.1.patch --current-dir C:\MyApp
```

**Use Case:** Download patches from central server

#### Example 4: Different Compression Types
```powershell
# Apply zstd-compressed patch
patch-apply.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir C:\MyApp

# Apply gzip-compressed patch (same command, auto-detected)
patch-apply.exe --patch .\patches-gzip\1.0.0-to-1.0.1.patch --current-dir C:\MyApp

# Apply uncompressed patch (same command, auto-detected)
patch-apply.exe --patch .\patches-none\1.0.0-to-1.0.1.patch --current-dir C:\MyApp
```

**Note:** Compression type is auto-detected from patch file metadata.

---

### `--current-dir`

**Purpose:** Specifies the directory containing the current version to be updated.

**Type:** String (required)

**Important:** This directory will be modified in-place (with selective backup of changed files preserved in `backup.cyberpatcher` subfolder).

#### Example 1: Local Application Directory
```powershell
patch-apply.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir C:\MyApp
```

**What This Does:**
- Verifies C:\MyApp contains version 1.0.0
- Creates selective backup of changed files in C:\MyApp\backup.cyberpatcher
- Updates files in C:\MyApp to version 1.0.1

#### Example 2: Relative Path
```powershell
patch-apply.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir .\application
```

#### Example 3: Different Drive
```powershell
patch-apply.exe --patch C:\patches\1.0.0-to-1.0.1.patch --current-dir D:\MyApp
```

#### Example 4: Network Installation
```powershell
patch-apply.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir \\server\apps\myapp
```

**Use Case:** Update applications on network shares

#### Example 5: Program Files
```powershell
patch-apply.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir "C:\Program Files\MyApp"
```

**Note:** Requires administrator privileges for Program Files

---

### `--key-file` (Applier)

**Purpose:** Specify a custom key file path if the original key file was renamed or moved.

**Type:** String (optional)

**Behavior:**
- Overrides the key file path stored in the patch
- Can be a relative path (resolved against `--current-dir`) or an absolute path
- Useful when the key file has been renamed or is in a non-standard location

#### Example 1: Renamed Key File
```powershell
patch-apply.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir C:\MyApp --key-file app.exe
```

**Output:**
```
Using custom key file: app.exe
...
```

#### Example 2: Absolute Path
```powershell
patch-apply.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir C:\MyApp --key-file C:\MyApp\renamed_program.exe
```

---

### `--dry-run`

**Purpose:** Simulate patch application without making any changes.

**Type:** Boolean (optional, default: `false`)

**Use Case:** Preview what the patch will do before committing to changes.

#### Example 1: Dry Run Before Applying
```powershell
patch-apply.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir C:\MyApp --dry-run
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
  patch-apply.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir C:\MyApp
```

#### Example 2: Dry Run with Verification
```powershell
patch-apply.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir C:\MyApp --dry-run --verify
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
patch-apply.exe --patch .\patches\1.0.1-to-1.0.2.patch --current-dir C:\MyApp --dry-run
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
patch-apply.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir C:\MyApp --dry-run

# Step 2: If successful, apply for real
patch-apply.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir C:\MyApp
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
patch-apply.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir C:\MyApp
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
patch-apply.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir C:\MyApp --verify=true
```

**Same as default**

#### Example 3: Disable Verification (Not Recommended)
```powershell
patch-apply.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir C:\MyApp --verify=false
```

**Use Case:**
- Trusted environment
- Already verified manually
- Speed is critical

**Output:**
```
Loading patch file...
WARNING: Verification disabled
Applying patch...
✓ Patch applied
WARNING: Post-verification skipped

WARNING: Verification was disabled. Use --verify=true for production.
```

#### Example 4: Verification Catches Corruption
```powershell
patch-apply.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir C:\MyApp --verify
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
patch-apply.exe --patch .\patches\1.0.1-to-1.0.2.patch --current-dir C:\MyApp --verify
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

### `--ignore1gb`

**Purpose:** Bypass 1GB patch size limit for self-contained executables.

**Type:** Boolean (optional, default: `false`)

**Behavior:**
- Only applies to self-contained executables (created with `--create-exe`)
- Allows loading patch data larger than 1GB into memory
- Standard mode limits patch size to 1GB to prevent memory exhaustion
- Use with caution on systems with limited RAM

#### Example 1: Running Self-Contained Executable with 1GB Bypass
```powershell
# When patch is larger than 1GB
.\1.0.0-to-1.0.1.exe --ignore1gb
```

**Output:**
```
==============================================
  CyberPatchMaker - Self-Contained Patch
==============================================

=== Patch Information ===
From Version:     1.0.0
To Version:       1.0.1
Patch Size:       1.8 GB
Files Modified:   2,456

Target directory [C:\MyApp]:

==============================================
Options:
  1. Dry Run (simulate without changes)
  2. Apply Patch
  3. Toggle 1GB Bypass Mode (currently: Enabled)
  4. Change Target Directory
  5. Specify Custom Key File
     (Currently: program.exe - default)
  6. Exit
==============================================
Select option [1-6]:
```

#### Example 2: Without Bypass (Default)
```powershell
.\large-patch-1.0.0-to-1.0.1.exe
```

**Output (if patch > 1GB):**
```
Warning: Patch size (1.8 GB) exceeds 1GB limit
Use --ignore1gb flag if you want to proceed anyway
Press any key to exit...
```

#### Example 3: Toggle in Interactive Menu
Users can also toggle the 1GB bypass from within the interactive console:
```
Select option [1-6]: 3

1GB Bypass Mode: Enabled
Warning: Large patches may consume significant memory!
```

**When to Use:**
- Large game updates with many assets
- Systems with 8GB+ RAM
- Deploying to known hardware configurations

**When NOT to Use:**
- Low-memory systems (< 4GB RAM)
- Unknown target hardware
- When patch can be split into smaller updates

**Considerations:**
- Large patches require significant RAM to decompress and process
- On 32-bit systems, memory limitations are more severe
- Consider splitting very large patches into incremental updates instead

---

### `--backup`

**Purpose:** Create selective backup of files being changed before patching.

**Type:** Boolean (optional, default: `true`)

**Behavior:**
- Creates selective backup of only modified/deleted files (NOT added files)
- Backup stored INSIDE target directory at `backup.cyberpatcher`
- Uses mirror directory structure preserving exact paths
- Backup is PRESERVED after success (manual cleanup required)
- Enables easy manual rollback using drag-and-drop

#### Example 1: Default (Backup Enabled)
```powershell
patch-apply.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir C:\MyApp
```

**Output:**
```
Loading patch file...
Creating selective backup...
  Backing up: program.exe
  Backing up: libs\core.dll
  Backing up: data\oldconfig.json
  Backup created in: C:\MyApp\backup.cyberpatcher
Applying patch...
  Modified: program.exe
  Modified: libs\core.dll
  Added: data\newfeature.json (NOT backed up - didn't exist)
  Deleted: data\oldconfig.json
✓ Patch applied successfully

Backup preserved in: C:\MyApp\backup.cyberpatcher
Files can be manually restored if needed. Delete backup.cyberpatcher when no longer needed.
```

#### Example 2: Explicitly Enable Backup
```powershell
patch-apply.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir C:\MyApp --backup=true
```

**Same as default**

#### Example 3: Disable Backup (Not Recommended)
```powershell
patch-apply.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir C:\MyApp --backup=false
```

**Use Case:**
- Disk space is extremely limited (selective backup uses minimal space)
- Already have external backup
- Testing in disposable environment

**Output:**
```
Loading patch file...
WARNING: Backup disabled
Applying patch...
✓ Patch applied

WARNING: No backup was created. Cannot rollback if issues occur.
```

#### Example 4: Backup with Large Application
```powershell
patch-apply.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir C:\MyApp --backup
```

**Output for 5GB application (only 15 files changed):**
```
Loading patch file...
Creating selective backup...
  Backing up: program.exe (2.1 MB)
  Backing up: libs\core.dll (512 KB)
  Backing up: libs\deprecated.dll (230 KB)
  ... (12 more files)
  Backup created in: C:\MyApp\backup.cyberpatcher (2.8 MB total)
  ✓ Selective backup created (took 2 seconds)
Applying patch...
```

**Note:** Selective backup only backs up changed files (2.8 MB), not the entire 5GB application. This saves significant disk space and time.

#### Example 5: Manual Rollback Using Backup
```powershell
# If patch causes problems, manually rollback using mirror structure:

# Stop application
Stop-Process -Name MyApp

# Restore files from backup (mirror structure makes this easy)
# Each file in backup.cyberpatcher has the exact path it should go to
Copy-Item C:\MyApp\backup.cyberpatcher\program.exe C:\MyApp\program.exe -Force
Copy-Item C:\MyApp\backup.cyberpatcher\libs\core.dll C:\MyApp\libs\core.dll -Force
Copy-Item C:\MyApp\backup.cyberpatcher\data\oldconfig.json C:\MyApp\data\oldconfig.json -Force

# Or restore all backed up files at once
Copy-Item C:\MyApp\backup.cyberpatcher\* C:\MyApp\ -Recurse -Force

# Delete new files that were added by patch (these weren't backed up)
Remove-Item C:\MyApp\data\newfeature.json -Force

# Cleanup backup when rollback is confirmed working
Remove-Item C:\MyApp\backup.cyberpatcher -Recurse -Force

# Restart application
Start-Process C:\MyApp\program.exe
```

**Note:** The mirror structure in `backup.cyberpatcher` preserves exact paths, making manual rollback intuitive - just copy files back to their original locations.

#### Example 6: Automatic Rollback on Failure
```powershell
patch-apply.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir C:\MyApp --backup --verify
```

**If post-verification fails:**
```
Loading patch file...
Creating selective backup...
  Backing up: program.exe
  Backing up: libs\core.dll
  ✓ Selective backup created
Applying patch...
Verifying patched version...
  ✗ Verification failed: File checksum mismatch

✗ ERROR: Post-verification failed
  Automatically rolling back from backup...
  Restored: program.exe
  Restored: libs\core.dll
  ✓ Rollback complete
  ✓ Original version restored

The patch was NOT applied due to verification failure.
Backup preserved in: C:\MyApp\backup.cyberpatcher for investigation.
```

**Backup Best Practices:**
- **Always enable for production** (default)
- Test rollback procedure before production deployment
- Ensure sufficient disk space (need space for changed files only - much less than full application)
- Delete `backup.cyberpatcher` folder after confirming successful update
- Consider external backups for critical systems
- Selective backup minimizes disk space requirements

---

### `--silent` (Applier)

**Purpose:** Run the applier in silent mode, applying the patch automatically without prompts.

**Type:** Boolean (optional, default: `false`)

**Behavior:**
- Applies patch immediately without showing menus or prompts
- Uses default settings: verify=true, backup=true
- Creates a log file named `log_<timestamp>.txt` in the current directory
- Returns exit code 0 on success, 1 on failure
- Ideal for automation, CI/CD pipelines, and scripting

#### Example 1: Silent Mode with Self-Contained Executable
```powershell
1.0.0-to-1.0.1.exe --silent
```

**Output:**
```
========================================
CyberPatchMaker Silent Mode Log
Started: 2026-03-30 12:00:00
========================================

Patch Information:
  From Version: 1.0.0
  To Version:   1.0.1
  ...

Patch applied successfully: 1.0.0 → 1.0.1

========================================
Status: SUCCESS
Completed: 2026-03-30 12:00:05
========================================

Log saved to: log_1743331200.txt
```

#### Example 2: Silent Mode with Custom Target Directory
```powershell
1.0.0-to-1.0.1.exe --silent --current-dir C:\MyApp
```

#### Example 3: Silent Mode with Custom Key File
```powershell
1.0.0-to-1.0.1.exe --silent --current-dir C:\MyApp --key-file renamed.exe
```

---

### `--version` (Applier)

**Purpose:** Display the applier version information.

**Type:** Boolean (optional)

#### Example 1: Show Version
```powershell
patch-apply.exe --version
```

**Output:**
```
CyberPatchMaker Patch Applier v1.0.16
```

---

### `--help` (Applier)

**Purpose:** Display usage information and available options.

**Type:** Boolean (optional)

#### Example 1: Show Help
```powershell
patch-apply.exe --help
```

**Output:**
```
CyberPatchMaker - Patch Applier v1.0.16

Usage:
  patch-apply --patch <file> --current-dir <directory>

Options:
  --patch         Path to patch file (required)
  --current-dir   Directory containing current version (required)
  --key-file      Custom key file path (if renamed or moved)
  --dry-run       Simulate patch without making changes
  --verify        Verify file hashes before and after patching (default: true)
  --backup        Create backup before patching (default: true)
  --ignore1gb     Bypass 1GB patch size limit (use with caution)
  --silent        Silent mode: apply patch automatically without prompts
  --version       Show version information
  --help          Show this help message

Self-Contained Executable Mode:
  When run as a self-contained executable, an interactive console
  interface will guide you through the patch application process.
  Use --silent flag for automated patching without user interaction.

Examples:
  # Apply patch
  patch-apply --patch 1.0.0-to-1.0.3.patch --current-dir C:\MyApp

  # Dry run (simulate only)
  patch-apply --patch 1.0.0-to-1.0.3.patch --current-dir C:\MyApp --dry-run

  # Run self-contained executable with 1GB bypass
  1.0.0-to-1.0.1.exe --ignore1gb

  # Run self-contained executable in silent mode (automation)
  1.2.4-to-1.2.5.exe --silent

  # Silent mode with custom target directory
  1.2.4-to-1.2.5.exe --silent --current-dir C:\MyApp
```

#### Example 2: Help Shortcut
```powershell
patch-apply.exe -help
```

**Same output as `--help`**

---

## Common Workflows

### Workflow 1: Creating and Applying a Single Patch

**Step 1: Generate Patch**
```powershell
patch-gen.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\patches
```

**Step 2: Test with Dry Run**
```powershell
patch-apply.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir C:\MyApp --dry-run
```

**Step 3: Apply Patch**
```powershell
patch-apply.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir C:\MyApp
```

---

### Workflow 2: Release New Version to All Users

**Scenario:** You've released version 1.0.3 and need patches for users on 1.0.0, 1.0.1, and 1.0.2.

**Step 1: Generate All Patches at Once**
```powershell
patch-gen.exe --versions-dir .\versions --new-version 1.0.3 --output .\patches --compression zstd
```

**Step 2: Distribute Patches**
```powershell
# Copy to web server
Copy-Item .\patches\* \\webserver\downloads\updates\
```

**Step 3: Users Apply Appropriate Patch**
```powershell
# User on 1.0.0 downloads and applies:
patch-apply.exe --patch .\downloads\1.0.0-to-1.0.3.patch --current-dir C:\MyApp

# User on 1.0.1 downloads and applies:
patch-apply.exe --patch .\downloads\1.0.1-to-1.0.3.patch --current-dir C:\MyApp

# User on 1.0.2 downloads and applies:
patch-apply.exe --patch .\downloads\1.0.2-to-1.0.3.patch --current-dir C:\MyApp
```

---

### Workflow 3: Multi-Hop Patching

**Scenario:** User on 1.0.0 wants to upgrade to 1.0.2, but only 1.0.0→1.0.1 and 1.0.1→1.0.2 patches exist.

**Step 1: Apply First Patch**
```powershell
patch-apply.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir C:\MyApp --verify
```

**Step 2: Apply Second Patch**
```powershell
patch-apply.exe --patch .\patches\1.0.1-to-1.0.2.patch --current-dir C:\MyApp --verify
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
.\patch-gen.exe --versions-dir .\versions --new-version 1.0.3 --output .\patches --compression zstd --verify

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
    .\patch-apply.exe --patch $patch.FullName --current-dir $testDir --verify
    
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
.\patch-apply.exe --patch $PatchFile --current-dir $TestDir --dry-run
if ($LASTEXITCODE -ne 0) {
    Write-Host "✗ Dry run failed"
    exit 1
}
Write-Host "✓ Dry run passed"

# Step 2: Apply with verification
Write-Host "`nStep 2: Applying patch with verification..."
.\patch-apply.exe --patch $PatchFile --current-dir $TestDir --verify --backup
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
    
    # Rollback using mirror structure backup
    Write-Host "Rolling back from backup.cyberpatcher..."
    Copy-Item "$TestDir\backup.cyberpatcher\*" $TestDir -Recurse -Force
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
    .\patch-gen.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\patches-zstd --compression zstd
} | Select-Object -ExpandProperty TotalSeconds | Tee-Object -Variable zstdTime

# Gzip
Measure-Command {
    .\patch-gen.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\patches-gzip --compression gzip
} | Select-Object -ExpandProperty TotalSeconds | Tee-Object -Variable gzipTime

# None
Measure-Command {
    .\patch-gen.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\patches-none --compression none
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
patch-gen.exe --from 1.0.0 --to 1.0.1 --output .\patches
```

**Error:**
```
Error: --versions-dir is required
```

**Solution:** Add `--versions-dir` argument
```powershell
patch-gen.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\patches
```

---

### Error: "Version directory not found"

**Command:**
```powershell
patch-gen.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\patches
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
patch-gen.exe --versions-dir .\versions --from 1.0.0 --to 1.0.1 --output .\patches
```

---

### Error: "Patch file not found"

**Command:**
```powershell
patch-apply.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir C:\MyApp
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
patch-apply.exe --patch .\patches\1.0.1-to-1.0.2.patch --current-dir C:\MyApp
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
patch-apply.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir C:\MyApp

# Then apply 1.0.1 to 1.0.2
patch-apply.exe --patch .\patches\1.0.1-to-1.0.2.patch --current-dir C:\MyApp
```

---

### Error: "Insufficient disk space"

**Command:**
```powershell
patch-apply.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir C:\MyApp
```

**Error:**
```
Error: Insufficient disk space
  Required: 50 MB (5 GB app + 50 MB selective backup for changed files)
  Available: 30 MB
```

**Solution:** Free up minimal disk space (selective backup is much smaller than old system) or disable backup
```powershell
# Option 1: Free up space and retry
# Option 2: Disable backup (not recommended)
patch-apply.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir C:\MyApp --backup=false
```

---

### Error: "Permission denied"

**Command:**
```powershell
patch-apply.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir "C:\Program Files\MyApp"
```

**Error:**
```
Error: Permission denied
  Cannot write to: C:\Program Files\MyApp\program.exe
```

**Solution:** Run as administrator
```powershell
# Open PowerShell as Administrator, then run:
patch-apply.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir "C:\Program Files\MyApp"
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
- Check available disk space (need minimal space for selective backup - only changed files, not entire application)

**Best Practices:**
1. Test patches with `--dry-run` before production
2. Keep `--verify` and `--backup` enabled (defaults)
3. Choose appropriate compression for your use case
4. Automate testing in CI/CD pipelines
5. Maintain good version control of your builds

For more information, see the related documentation links above.
