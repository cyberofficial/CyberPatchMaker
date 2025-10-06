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
| `--versions-dir <path>` | Yes (batch) | Directory containing version folders |
| `--new-version <version>` | Yes (batch) | New version number to generate patches for |
| `--from <version>` | Yes (single) | Source version number |
| `--to <version>` | Yes (single) | Target version number |
| `--output <path>` | Yes | Output directory for patches |
| `--compression <type>` | No | Compression: `zstd` (default), `gzip`, `none` |
| `--level <n>` | No | Compression level: 1-4 (default: 3) |
| `--verify` | No | Verify patches after creation |
| `--help` | No | Display help information |

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Invalid arguments |
| 3 | Version not found |
| 4 | Key file detection failed |
| 5 | Manifest generation failed |
| 6 | Patch generation failed |
| 7 | Verification failed |

### Examples

**Batch Mode** (generate all patches to new version):
```bash
patch-gen --versions-dir ./versions --new-version 1.0.3 --output ./patches
```

**Single Patch Mode**:
```bash
patch-gen --from 1.0.1 --to 1.0.3 --versions-dir ./versions --output ./patches
```

**With Compression**:
```bash
patch-gen --versions-dir ./versions --new-version 1.0.3 --output ./patches --compression zstd --level 4
```

**With Verification**:
```bash
patch-gen --versions-dir ./versions --new-version 1.0.3 --output ./patches --verify
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
| `--verify` | No | Enable full verification (recommended!) |
| `--dry-run` | No | Preview changes without applying |
| `--create-backup` | No | Create selective backup before patching (default: `true`) |
| `--no-backup` | No | Disable backup creation (NOT recommended for production!) |
| `--help` | No | Display help information |

#### Backup Flag Details

**`--create-backup` (default: `true`)**
- **Strategy**: Selective backup of only modified/deleted files (NOT new files)
- **Location**: `backup.cyberpatcher` folder created inside `--current-dir`
- **Structure**: Mirror directory hierarchy preserving exact original paths
- **Preservation**: Kept after successful patching for manual rollback capability
- **Benefits**:
  - Minimal disk space (e.g., 2.8MB vs 5.2GB = 99.5% reduction)
  - Fast backup creation (e.g., 2s vs 45s = 95% faster)
  - Intuitive rollback (mirror structure = drag-and-drop restore)
  - Transparent about changes (shows exactly what was backed up)

**`--no-backup`**
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

See [Backup Lifecycle](backup-lifecycle.md) for complete backup system documentation.

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Invalid arguments |
| 3 | Patch file not found |
| 4 | Current directory not found |
| 5 | Pre-verification failed |
| 6 | Backup creation failed |
| 7 | Operation failed |
| 8 | Post-verification failed |
| 9 | Restoration failed |

### Examples

**Safe Application** (with verification):
```bash
patch-apply --patch ./patches/1.0.0-to-1.0.3.patch --current-dir ./myapp --verify
```

**Dry-Run** (preview only):
```bash
patch-apply --patch ./patches/1.0.0-to-1.0.3.patch --current-dir ./myapp --dry-run
```

**Quick Application** (no verification - RISKY!):
```bash
patch-apply --patch ./patches/1.0.0-to-1.0.3.patch --current-dir ./myapp
```

**Without Backup** (for testing only - NOT recommended!):
```bash
patch-apply --patch ./patches/1.0.0-to-1.0.3.patch --current-dir ./myapp --no-backup
```

**Explicit Backup** (default behavior):
```bash
patch-apply --patch ./patches/1.0.0-to-1.0.3.patch --current-dir ./myapp --create-backup --verify
```

**Custom Key File** (if the key file was renamed):
```bash
# If program.exe was renamed to app.exe
patch-apply --patch ./patches/1.0.0-to-1.0.3.patch \
            --current-dir ./myapp \
            --key-file app.exe

# Or with absolute path
patch-apply --patch ./patches/1.0.0-to-1.0.3.patch \
            --current-dir ./myapp \
            --key-file C:\MyApp\renamed_program.exe
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
        --current-dir ./test-app
```

---

## Environment Variables

Currently, no environment variables are used. All configuration is via command-line flags.

---

## Configuration Files

Currently, no configuration files are used. All configuration is via command-line flags.

---

## Output Format

### Generator Output

```
Scanning versions directory: ./versions
Found versions: 1.0.0, 1.0.1, 1.0.2
New version: 1.0.3

Generating patch 1.0.0 -> 1.0.3...
  Loading manifests...
  Comparing versions...
  Generating binary diffs...
  Creating patch file...
  Compressing (zstd level 3)...
  ✓ Success: patches/1.0.0-to-1.0.3.patch (2.1 MB)

Generating patch 1.0.1 -> 1.0.3...
  ...
  ✓ Success: patches/1.0.1-to-1.0.3.patch (1.8 MB)

Generating patch 1.0.2 -> 1.0.3...
  ...
  ✓ Success: patches/1.0.2-to-1.0.3.patch (1.2 MB)

All patches generated successfully!
Total: 3 patches, 5.1 MB
```

### Applier Output

```
Loading patch: patches/1.0.0-to-1.0.3.patch

=== Patch Information ===
From Version: 1.0.0
To Version:   1.0.3
Key File:     program.exe
Created:      2025-10-04 10:30:00
Patch Size:   2.1 MB
Compression:  zstd

Operations:
  5 files to add
  12 files to modify
  3 files to delete

Applying patch from 1.0.0 to 1.0.3...
Verifying current version...
✓ Pre-patch verification successful

Creating selective backup...
  Backing up: program.exe
  Backing up: data/config.json
  Backing up: libs/oldfeature.dll
  ... (9 more files - only changed/deleted files)
✓ Selective backup created

Applying 20 operations...
  Modified: program.exe
  Modified: data/config.json
  Added: libs/newfeature.dll (NOT backed up - didn't exist before)
  Deleted: libs/oldfeature.dll
  ... (16 more operations)

✓ Post-patch verification successful

Backup preserved in: ./myapp/backup.cyberpatcher

=== Patch Applied Successfully ===
Version updated from 1.0.0 to 1.0.3
Time elapsed: 8.2 seconds (selective backup saved time!)
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
./generator --versions-dir ./versions --new-version 1.0.3 --output ./patches
./applier --patch ./patches/1.0.0-to-1.0.3.patch --current-dir ./myapp --verify
```

**Paths:** Use forward slashes `/`

**Permissions:** May need to make executables:
```bash
chmod +x generator applier
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

if [ $? -eq 0 ]; then
    echo "Success! Patches generated in $OUTPUT_DIR"
else
    echo "Patch generation failed with exit code $?"
    exit $?
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
