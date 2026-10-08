# CLI Reference

Quick reference for CyberPatchMaker command-line tools.

## Generator Tool

### Basic Syntax

```bash
patch-gen [options]
```

### Options

| Option | Required | Description |
|--------|----------|-------------|
| `--versions-dir <path>` | Mode 1 | Directory containing version folders |
| `--new-version <version>` | Mode 1 | New version number to generate patches for |
| `--from <version>` | Mode 2 | Source version number (with --versions-dir) |
| `--to <version>` | Mode 2 | Target version number (with --versions-dir) |
| `--from-dir <path>` | Mode 3 | Full path to source version directory |
| `--to-dir <path>` | Mode 3 | Full path to target version directory |
| `--output <path>` | No (default: patches) | Output directory for patches (falls back to the config file's `default_patch_output`, then `patches`) |
| `--key-file <name>` | No | Specific key file to use (e.g., app_name.exe) |
| `--compression <type>` | No | Compression: `zstd` (default), `gzip`, `none` |
| `--level <n>` | No | Compression level: zstd (1-4), gzip (1-3), default: 3 |
| `--verify` | No | Accepted for compatibility; currently has no effect (patch validation always runs) |
| `--create-exe` | No | Create self-contained CLI executable (requires patch-apply.exe next to patch-gen.exe) |
| `--silent` | No | Embed silent mode into generated executables (requires --create-exe) |
| `--crp` | No | Create reverse patch for downgrades (named `<to>-to-<from>_rev.patch`) |
| `--savescans` | No | Enable scan caching to `.data/` directory |
| `--scandata <dir>` | No | Custom cache directory (default: `.data`) |
| `--rescan` | No | Force rescan, ignoring cached data (requires --savescans) |
| `--jobs <n>` | No | Number of parallel workers (0 = config `worker_threads` which defaults to CPU count, 1 = single-threaded) |
| `--splitsize <size>` | No | Custom multi-part split size (e.g., '2G', '500M'). Default: 4GB |
| `--bypasssplitlimit` | No | Bypass 100MB minimum split size confirmation |
| `--version` | No | Show version information |
| `--help` | No | Display help information |

### Exit Codes

| Code | Meaning |
|------|------|
| 0 | Success |
| 1 | Error (all error conditions) |

Notes: invalid flag syntax exits with code 2 (Go flag package behavior); answering "no" to the sub-100MB `--splitsize` confirmation also exits 0.

### Examples

**Mode 1: Batch Mode** (generate all patches to new version):
```bash
patch-gen --versions-dir ./versions --new-version 1.0.3 --output ./patches
```

**Mode 2: Single Patch** (versions in same directory):
```bash
patch-gen --from 1.0.1 --to 1.0.3 --versions-dir ./versions --output ./patches
```

**Mode 3: Custom Paths** (different drives/locations):
```bash
patch-gen --from-dir C:\releases\1.0.0 --to-dir D:\builds\1.0.1 --output ./patches
```

**With Compression**:
```bash
patch-gen --versions-dir ./versions --new-version 1.0.3 --output ./patches --compression zstd --level 4
```

**With Verification**:
```bash
patch-gen --versions-dir ./versions --new-version 1.0.3 --output ./patches --verify
```

**With Self-Contained Executables (Silent Mode Embedded)**:
```bash
# Create executables with embedded silent mode (auto-apply)
patch-gen --versions-dir ./versions --new-version 1.0.3 --output ./patches --create-exe --silent --verify

# Result: Users just run 1.0.0-to-1.0.3.exe and patch applies automatically
```

---

## Applier Tool

### Basic Syntax

```bash
patch-apply [options]
```

### Options

| Option | Required | Description |
|--------|----------|-------------|
| `--patch <path>` | Yes | Path to patch file |
| `--current-dir <path>` | Yes | Directory containing current installation |
| `--key-file <path>` | No | Custom key file path (if renamed or moved) |
| `--dry-run` | No | Simulate patch without making changes |
| `--verify` | No | Verify file hashes before and after patching (default: true) |
| `--backup` | No | Create backup before patching (default: true) |
| `--ignore1gb` | No | Bypass 1GB patch size limit (use with caution) |
| `--silent` | No | Silent mode: apply patch automatically without prompts (for automation) |
| `--version` | No | Show version information |
| `--help` | No | Show this help message |

#### Backup Flag Details

**`--backup` (default: `true`)**
- **Strategy**: Selective backup of only modified/deleted files (NOT new files)
- **Location**: `backup.cyberpatcher` folder created inside `--current-dir`
- **Structure**: Mirror directory hierarchy preserving exact original paths
- **Preservation**: Kept after successful patching for manual rollback capability
- **Benefits**:
  - Minimal disk space (only changed files, not the whole installation)
  - Fast backup creation
  - Intuitive rollback (mirror structure = drag-and-drop restore)
  - Transparent about changes (shows exactly what was backed up)

**`--backup=false`**
- **Disables**: Backup creation entirely
- **Risk**: Cannot automatically rollback on failure
- **Use Case**: Testing environments, CI/CD pipelines with external backups
- **WARNING**: Not recommended for production systems!

**Rollback Procedure** (if backup exists):
```powershell
# Manual rollback from backup.cyberpatcher
Copy-Item C:\MyApp\backup.cyberpatcher\* C:\MyApp -Recurse -Force

# Delete any files that were added (not in backup)
# Then delete backup folder after confirming restoration
Remove-Item C:\MyApp\backup.cyberpatcher -Recurse -Force
```

See [Backup System](backup-system.md) for complete backup system documentation.

### Exit Codes

| Code | Meaning |
|------|------|
| 0 | Success |
| 1 | Error (all error conditions) |

### Examples

**Safe Application** (with verification):
```bash
patch-apply --patch ./patches/1.0.0-to-1.0.3.patch --current-dir ./myapp --verify
```

**Dry-Run** (preview only):
```bash
patch-apply --patch ./patches/1.0.0-to-1.0.3.patch --current-dir ./myapp --dry-run
```

**Without Backup** (for testing only - NOT recommended!):
```bash
patch-apply --patch ./patches/1.0.0-to-1.0.3.patch --current-dir ./myapp --backup=false
```

**Custom Key File** (if the key file was renamed):
```bash
# If program.exe was renamed to app.exe
patch-apply --patch ./patches/1.0.0-to-1.0.3.patch \
            --current-dir ./myapp \
            --key-file app.exe

# Key file in a subdirectory (resolved relative to current-dir)
patch-apply --patch ./patches/1.0.0-to-1.0.3.patch \
            --current-dir ./myapp \
            --key-file bin/program.exe
```

**Silent Mode** (for automation/CI-CD):
```bash
# Apply patch automatically without prompts
# Only works with self-contained executables
1.0.0-to-1.0.1.exe --silent

# Silent mode with custom target directory
1.0.0-to-1.0.1.exe --silent --current-dir C:\MyApp

# Silent mode with custom key file
1.0.0-to-1.0.1.exe --silent --current-dir C:\MyApp --key-file renamed.exe
```

---

## Common Workflows

### New Production Release

```bash
# 1. Generate all patches to new version
patch-gen --versions-dir ./versions \
          --new-version 1.0.3 \
          --output ./patches \
          --verify

# 2. Test with dry-run
patch-apply --patch ./patches/1.0.2-to-1.0.3.patch \
        --current-dir ./test-app \
        --dry-run

# 3. Apply to production
patch-apply --patch ./patches/1.0.2-to-1.0.3.patch \
        --current-dir C:\Production\MyApp \
        --verify
```

### Downgrade Patches (Rollback to Previous Version)

**Generate downgrade patch:**
```bash
# Generate patch to downgrade from 1.0.3 back to 1.0.2
patch-gen --from 1.0.3 \
          --to 1.0.2 \
          --versions-dir ./versions \
          --output ./patches/downgrade \
          --verify
```

**Apply downgrade patch:**
```bash
# Test rollback
patch-apply --patch ./patches/downgrade/1.0.3-to-1.0.2.patch \
        --current-dir ./test-app \
        --dry-run

# Apply rollback to production
patch-apply --patch ./patches/downgrade/1.0.3-to-1.0.2.patch \
        --current-dir C:\Production\MyApp \
        --verify
```

**Generate all downgrade paths from current version:**
```bash
# From 1.0.3 to all previous versions
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

> **Note:** For complete downgrade documentation, see [Downgrade Guide](downgrade-guide.md)

### Custom Patch with Maximum Compression

```bash
# Generate single patch with highest compression
patch-gen --from 1.0.1 \
          --to 1.0.3 \
          --versions-dir ./versions \
          --output ./patches \
          --compression zstd \
          --level 4 \
          --verify
```

### Quick Testing

```bash
# Generate without compression (fastest)
patch-gen --versions-dir ./versions \
          --new-version 1.0.3 \
          --output ./patches \
          --compression none

# Apply without verification (fastest)
patch-apply --patch ./patches/1.0.0-to-1.0.3.patch \
        --current-dir ./test-app \
        --verify=false
```

### Large Projects with Scan Cache

```bash
# First generation: Enable scan caching (scans and saves to cache)
patch-gen --versions-dir ./versions \
          --new-version 1.0.3 \
          --output ./patches \
          --savescans

# Subsequent generations: Load from cache (instant, no rescanning)
patch-gen --versions-dir ./versions \
          --new-version 1.0.4 \
          --output ./patches \
          --savescans

# Custom cache location
patch-gen --versions-dir ./versions \
          --new-version 1.0.3 \
          --output ./patches \
          --savescans \
          --scandata ./shared-cache

# Force rescan (update cache with latest file data)
patch-gen --versions-dir ./versions \
          --new-version 1.0.3 \
          --output ./patches \
          --savescans \
          --rescan
```

**Benefits:**
- **Large projects**: Massive improvement (full scan skipped entirely, e.g. a ~15 minute scan loads instantly from cache)
- **Example**: War Thunder (34,650 files) - 15 min scan → instant cache load
- Cache validates key file hash to prevent using stale data
- Works with both `--versions-dir` and custom paths (`--from-dir`/`--to-dir`)
- Cache files stored as JSON with complete file metadata

---

## Environment Variables

No tool-specific environment variables are defined. Standard OS variables (`HOME`/`USERPROFILE`, `APPDATA`, `XDG_CONFIG_HOME`, `TEMP`) are used only to locate the optional configuration file and default temp/output paths; all tool behavior is configured via command-line flags.

---

## Configuration Files

A `config.json` file is supported (loaded by the generator from the OS-specific config directory: `%APPDATA%\CyberPatchMaker\config.json` on Windows, `~/.config/cyberpatchmaker/config.json` on Linux, `~/Library/Application Support/CyberPatchMaker/config.json` on macOS). It can set `default_patch_output` and `worker_threads`, among other defaults.

---

## Output Format

### Generator Output

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
Changes detected: 2 added, 8 modified, 1 deleted files, 0 added dirs, 0 deleted dirs
Patch saved to: patches/1.0.1-to-1.0.3.patch

Successfully generated 2 patches
```

### Applier Output

```
=== Patch Information ===
From Version:     1.0.0
To Version:       1.0.3
Key File:         program.exe
Required Hash:    63573ff071ea5fa2...
Patch Size:       2202010 bytes
Compression:      zstd
Created:          2025-10-04 10:30:00
Files Added:      5
Files Modified:   12
Files Deleted:    3
Dirs Added:       2
Dirs Deleted:     1
Required Files:   41 (must match exact hashes)

Applying patch from 1.0.0 to 1.0.3...
Verifying current version...
Pre-patch verification successful

Creating backup...
Backed up 15 files
Backup created at: ./myapp/backup.cyberpatcher
Note: Backup will be preserved after patching for manual rollback

Applying 20 operations...
  Modified: program.exe
  Modified: data/config.json
  Added: libs/newfeature.dll
  Deleted: libs/oldfeature.dll
  ... (16 more operations)

Verifying patched version...
Post-patch verification successful
Patch applied successfully

Backup preserved at: ./myapp/backup.cyberpatcher
To rollback: Copy files from backup folder to their original locations

=== Patch Applied Successfully ===
Version updated from 1.0.0 to 1.0.3
```

---

## Platform-Specific Notes

### Windows

**PowerShell:**
```powershell
.\patch-gen.exe --versions-dir .\versions --new-version 1.0.3 --output .\patches
.\patch-apply.exe --patch .\patches\1.0.0-to-1.0.3.patch --current-dir .\myapp --verify
```

**Command Prompt (cmd):**
```batch
patch-gen.exe --versions-dir .\versions --new-version 1.0.3 --output .\patches
patch-apply.exe --patch .\patches\1.0.0-to-1.0.3.patch --current-dir .\myapp --verify
```

**Paths:** Use backslashes `\` or forward slashes `/` (both work)

---

### Linux/macOS

**Bash:**
```bash
./patch-gen --versions-dir ./versions --new-version 1.0.3 --output ./patches
./patch-apply --patch ./patches/1.0.0-to-1.0.3.patch --current-dir ./myapp --verify
```

**Paths:** Use forward slashes `/`

**Permissions:** May need to make executables:
```bash
chmod +x patch-gen patch-apply
```

---

## Automation Scripts

### PowerShell Script Example

```powershell
# generate-patches.ps1
param(
    [string]$NewVersion = "1.0.3",
    [string]$VersionsDir = "./versions",
    [string]$OutputDir = "./patches"
)

Write-Host "Generating patches for version $NewVersion..."

& .\patch-gen.exe `
    --versions-dir $VersionsDir `
    --new-version $NewVersion `
    --output $OutputDir `
    --verify

if ($LASTEXITCODE -eq 0) {
    Write-Host "Success! Patches generated in $OutputDir"
} else {
    Write-Error "Patch generation failed with exit code $LASTEXITCODE"
    exit $LASTEXITCODE
}
```

**Usage:**
```powershell
.\generate-patches.ps1 -NewVersion 1.0.4
```

---

### Bash Script Example

```bash
#!/bin/bash
# generate-patches.sh

NEW_VERSION=${1:-"1.0.3"}
VERSIONS_DIR="./versions"
OUTPUT_DIR="./patches"

echo "Generating patches for version $NEW_VERSION..."

./patch-gen \
    --versions-dir "$VERSIONS_DIR" \
    --new-version "$NEW_VERSION" \
    --output "$OUTPUT_DIR" \
    --verify
status=$?

if [ $status -eq 0 ]; then
    echo "Success! Patches generated in $OUTPUT_DIR"
else
    echo "Patch generation failed with exit code $status"
    exit $status
fi
```

**Usage:**
```bash
chmod +x generate-patches.sh
./generate-patches.sh 1.0.4
```

---

## Related Documentation

- [Generator Guide](generator-guide.md) - Detailed generator documentation
- [Applier Guide](applier-guide.md) - Detailed applier documentation
- [Quick Start](quick-start.md) - Getting started tutorial
- [Troubleshooting](troubleshooting.md) - Common issues
