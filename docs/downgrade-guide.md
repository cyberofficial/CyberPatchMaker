# Downgrade Guide

Complete guide to generating and applying downgrade patches with CyberPatchMaker.

## Overview

CyberPatchMaker fully supports **bidirectional patching** - you can generate patches to upgrade OR downgrade between any two versions. This is useful when:

- Rolling back problematic updates
- Testing with older versions
- Providing users a way to revert changes
- Supporting version-specific features

---

## How Downgrade Patches Work

**Key Insight:** A downgrade patch is just a regular patch with reversed source/target versions.

**Upgrade Patch:**
```bash
generator --from 1.0.0 --to 1.0.1  # 1.0.0 → 1.0.1
```

**Downgrade Patch:**
```bash
generator --from 1.0.1 --to 1.0.0  # 1.0.1 → 1.0.0
```

The generator creates a patch that transforms version 1.0.1 back to version 1.0.0, including:
- Restoring deleted files
- Removing added files
- Reverting modified files
- Recreating deleted directories
- Removing new directories

---

## Generating Downgrade Patches

### Single Downgrade Patch

**Scenario:** Generate patch to downgrade from 1.0.3 to 1.0.2

```bash
generator --versions-dir ./versions \
          --from 1.0.3 \
          --to 1.0.2 \
          --output ./patches/downgrade
```

**Result:** `patches/downgrade/1.0.3-to-1.0.2.patch`

---

### Generate All Downgrade Patches

**Scenario:** For version 1.0.3, create downgrade patches to all previous versions

```bash
# Downgrade to 1.0.2
generator --from 1.0.3 --to 1.0.2 \
          --versions-dir ./versions \
          --output ./patches/downgrade

# Downgrade to 1.0.1
generator --from 1.0.3 --to 1.0.1 \
          --versions-dir ./versions \
          --output ./patches/downgrade

# Downgrade to 1.0.0
generator --from 1.0.3 --to 1.0.0 \
          --versions-dir ./versions \
          --output ./patches/downgrade
```

**Result:**
```
patches/downgrade/
├── 1.0.3-to-1.0.2.patch
├── 1.0.3-to-1.0.1.patch
└── 1.0.3-to-1.0.0.patch
```

---

### Batch Script for Downgrade Patches

**PowerShell Script:**
```powershell
# generate-downgrades.ps1
param(
    [string]$FromVersion = "1.0.3",
    [string]$VersionsDir = "./versions",
    [string]$OutputDir = "./patches/downgrade"
)

Write-Host "Generating downgrade patches from $FromVersion"

# Get all versions lower than FromVersion
$allVersions = Get-ChildItem -Path $VersionsDir -Directory | 
               Where-Object { $_.Name -lt $FromVersion } |
               Sort-Object Name

foreach ($targetVersion in $allVersions) {
    $to = $targetVersion.Name
    Write-Host "`nGenerating: $FromVersion → $to"
    
    & .\generator.exe `
        --versions-dir $VersionsDir `
        --from $FromVersion `
        --to $to `
        --output $OutputDir `
        --verify
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ Downgrade patch created: $FromVersion-to-$to.patch"
    } else {
        Write-Error "✗ Failed to create downgrade patch"
    }
}
```

**Usage:**
```powershell
.\generate-downgrades.ps1 -FromVersion 1.0.3
```

---

### Bash Script for Downgrade Patches

```bash
#!/bin/bash
# generate-downgrades.sh

FROM_VERSION=${1:-"1.0.3"}
VERSIONS_DIR="./versions"
OUTPUT_DIR="./patches/downgrade"

echo "Generating downgrade patches from $FROM_VERSION"

# Get all versions lower than FROM_VERSION
for target_dir in "$VERSIONS_DIR"/*; do
    if [ -d "$target_dir" ]; then
        TO_VERSION=$(basename "$target_dir")
        
        # Only process if target is lower version
        if [[ "$TO_VERSION" < "$FROM_VERSION" ]]; then
            echo ""
            echo "Generating: $FROM_VERSION → $TO_VERSION"
            
            ./generator \
                --versions-dir "$VERSIONS_DIR" \
                --from "$FROM_VERSION" \
                --to "$TO_VERSION" \
                --output "$OUTPUT_DIR" \
                --verify
            
            if [ $? -eq 0 ]; then
                echo "✓ Downgrade patch created: $FROM_VERSION-to-$TO_VERSION.patch"
            else
                echo "✗ Failed to create downgrade patch"
            fi
        fi
    fi
done
```

**Usage:**
```bash
chmod +x generate-downgrades.sh
./generate-downgrades.sh 1.0.3
```

---

## Applying Downgrade Patches

### Basic Downgrade

**Test first (dry-run):**
```bash
applier --patch ./patches/downgrade/1.0.3-to-1.0.2.patch \
        --current-dir ./app \
        --dry-run
```

**Apply downgrade:**
```bash
applier --patch ./patches/downgrade/1.0.3-to-1.0.2.patch \
        --current-dir ./app \
        --verify
```

---

### Downgrade with Backup

**Always recommended for production:**

```bash
# Downgrade safely with verification and selective backup
applier --patch ./patches/downgrade/1.0.3-to-1.0.2.patch \
        --current-dir C:\Production\MyApp \
        --verify
```

This will:
1. Verify current version is 1.0.3
2. **Create selective backup** of files being changed (modified/deleted from 1.0.3) in `C:\Production\MyApp\backup.cyberpatcher`
3. Apply downgrade operations (transform 1.0.3 → 1.0.2)
4. Verify result is 1.0.2
5. **Preserve backup** for manual rollback to 1.0.3 if needed

**Backup System Benefit:** Downgrade backup only contains changed files (minimal disk space), preserved with mirror structure for easy restoration.

---

## Organizing Patches

### Recommended Directory Structure

```
patches/
├── upgrade/                    # Forward patches
│   ├── 1.0.0-to-1.0.1.patch
│   ├── 1.0.0-to-1.0.2.patch
│   ├── 1.0.0-to-1.0.3.patch
│   ├── 1.0.1-to-1.0.2.patch
│   ├── 1.0.1-to-1.0.3.patch
│   └── 1.0.2-to-1.0.3.patch
└── downgrade/                  # Reverse patches
    ├── 1.0.3-to-1.0.2.patch
    ├── 1.0.3-to-1.0.1.patch
    ├── 1.0.3-to-1.0.0.patch
    ├── 1.0.2-to-1.0.1.patch
    ├── 1.0.2-to-1.0.0.patch
    └── 1.0.1-to-1.0.0.patch
```

---

## Best Practices

### 1. Always Generate Both Directions

When releasing a new version, generate both upgrade and downgrade patches:

```bash
# Upgrades
generator --versions-dir ./versions --new-version 1.0.3 --output ./patches/upgrade

# Downgrades (from 1.0.3 to all previous)
generator --from 1.0.3 --to 1.0.2 --output ./patches/downgrade
generator --from 1.0.3 --to 1.0.1 --output ./patches/downgrade
generator --from 1.0.3 --to 1.0.0 --output ./patches/downgrade
```

---

### 2. Test Downgrade Patches

**Test the complete cycle:**

```bash
# Start with 1.0.0
# Apply upgrade to 1.0.1
applier --patch upgrade/1.0.0-to-1.0.1.patch --current-dir ./test --verify

# Apply downgrade back to 1.0.0
applier --patch downgrade/1.0.1-to-1.0.0.patch --current-dir ./test --verify

# Verify result matches original 1.0.0
```

---

### 3. Document Downgrade Path

**Create a downgrade guide for users:**

```markdown
# Downgrade Instructions for MyApp

If you encounter issues with version 1.0.3, you can safely downgrade:

## From 1.0.3 to 1.0.2
1. Download: https://myapp.com/patches/1.0.3-to-1.0.2.patch
2. Run: applier --patch 1.0.3-to-1.0.2.patch --current-dir C:\MyApp --verify
3. Restart application

## From 1.0.3 to 1.0.1
1. Download: https://myapp.com/patches/1.0.3-to-1.0.1.patch
2. Run: applier --patch 1.0.3-to-1.0.1.patch --current-dir C:\MyApp --verify
3. Restart application
```

---

### 4. Version Verification

**Downgrade patches include strict verification:**

- **Pre-verification**: Confirms you're on version 1.0.3 before downgrading
- **Post-verification**: Confirms you're on version 1.0.2 after downgrading
- **Automatic rollback**: If verification fails, restores version 1.0.3

---

### 5. Storage Considerations

**Downgrade patches are the same size as upgrade patches:**

```
Version 1.0.0: 5GB
Version 1.0.1: 5GB (with changes)

Upgrade patch (1.0.0 → 1.0.1): 10MB
Downgrade patch (1.0.1 → 1.0.0): 10MB (same size!)
```

**Storage planning:**
- For N versions, you may want N² patches (every version to every other version)
- Or just N patches (each version to previous)
- Balance availability vs storage cost

---

## Common Scenarios

### Scenario 1: Rollback Problematic Update

**Problem:** Version 1.0.3 has a critical bug

```bash
# Users on 1.0.3 can downgrade to stable 1.0.2
applier --patch https://myapp.com/patches/emergency/1.0.3-to-1.0.2.patch \
        --current-dir ./app \
        --verify
```

---

### Scenario 2: Feature Testing

**Testing:** User wants to try new version but keep option to revert

```bash
# Upgrade to test version
applier --patch upgrade/1.0.2-to-1.0.3.patch --current-dir ./app --verify

# Test features...

# If not satisfied, downgrade
applier --patch downgrade/1.0.3-to-1.0.2.patch --current-dir ./app --verify
```

---

### Scenario 3: Incremental Downgrade

**Downgrade multiple versions:**

```bash
# From 1.0.3 down to 1.0.1 (two steps)
applier --patch downgrade/1.0.3-to-1.0.2.patch --current-dir ./app --verify
applier --patch downgrade/1.0.2-to-1.0.1.patch --current-dir ./app --verify

# Or direct downgrade (if patch exists)
applier --patch downgrade/1.0.3-to-1.0.1.patch --current-dir ./app --verify
```

---

### Scenario 4: Automated Rollback Script

**PowerShell:**
```powershell
# rollback.ps1
param(
    [string]$CurrentVersion = "1.0.3",
    [string]$TargetVersion = "1.0.2",
    [string]$AppDir = "C:\MyApp"
)

Write-Host "Rolling back from $CurrentVersion to $TargetVersion"

$patchFile = ".\patches\downgrade\$CurrentVersion-to-$TargetVersion.patch"

if (-not (Test-Path $patchFile)) {
    Write-Error "Downgrade patch not found: $patchFile"
    exit 1
}

# Dry run first
Write-Host "Testing rollback (dry-run)..."
.\applier.exe --patch $patchFile --current-dir $AppDir --dry-run

if ($LASTEXITCODE -ne 0) {
    Write-Error "Dry-run failed"
    exit 1
}

# Apply rollback
Write-Host "`nApplying rollback..."
.\applier.exe --patch $patchFile --current-dir $AppDir --verify

if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ Successfully rolled back to $TargetVersion"
} else {
    Write-Error "✗ Rollback failed"
    exit 1
}
```

**Usage:**
```powershell
.\rollback.ps1 -CurrentVersion 1.0.3 -TargetVersion 1.0.2
```

---

## Safety Considerations

### 1. Data Loss Risk

**Important:** Downgrading may cause data loss if:
- Newer version created data formats incompatible with older version
- Newer version added database schema changes
- Newer version stored data in new locations

**Recommendation:**
- Test downgrade on copy of application first
- Back up user data separately
- Document any data compatibility issues

---

### 2. Configuration Files

**Newer versions may have updated configuration:**

```bash
# 1.0.3 config.json
{
  "version": "1.0.3",
  "new_feature_enabled": true  ← New in 1.0.3
}

# After downgrade to 1.0.2, config.json reverts to:
{
  "version": "1.0.2"
}
```

---

### 3. Database Migrations

**If your application uses databases:**

- Downgrade patches revert FILES but not database schemas
- You may need separate database rollback scripts
- Test database compatibility after downgrade

---

### 4. User Expectations

**Inform users about downgrade implications:**

```markdown
## Before Downgrading

⚠️ Downgrading will:
- Revert application files to older version
- Remove features added in newer version
- May cause data compatibility issues
- Requires application restart

✓ Downgrading will NOT:
- Delete your user data (unless incompatible)
- Affect your license/activation
- Modify system settings
```

---

## Troubleshooting

### Error: "Version mismatch"

**Problem:** Trying to apply wrong downgrade patch

```bash
# Trying to apply 1.0.3→1.0.2 patch to version 1.0.1
Error: Version mismatch
  Expected: 1.0.3
  Found: 1.0.1
```

**Solution:** Use correct downgrade patch for current version

```bash
# Apply 1.0.1→1.0.0 instead
applier --patch downgrade/1.0.1-to-1.0.0.patch --current-dir ./app --verify
```

---

### Error: "Post-verification failed"

**Problem:** Downgrade completed but verification failed

**This indicates:**
- Patch file may be corrupted
- Target version directory changed since patch was generated
- File system issues

**Solution:**
```bash
# System will automatically rollback to previous version
# Check logs for specific file mismatches
# Re-download patch file and try again
```

---

### Patch File Not Found

**Problem:** Missing downgrade patch

**Solution:** Generate the needed downgrade patch

```bash
generator --versions-dir ./versions \
          --from 1.0.3 \
          --to 1.0.2 \
          --output ./patches/downgrade
```

---

## Related Documentation

- [Generator Guide](generator-guide.md) - Detailed patch generation
- [Applier Guide](applier-guide.md) - Detailed patch application
- [CLI Reference](cli-reference.md) - Complete command reference
- [Version Management](version-management.md) - Managing multiple versions
- [Backup Lifecycle](backup-lifecycle.md) - Backup and restoration
