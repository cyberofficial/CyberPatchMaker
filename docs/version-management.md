# Version Management Guide

Complete guide to managing versions in CyberPatchMaker.

## Overview

CyberPatchMaker uses a **version registry system** that tracks all registered software versions and their locations. Versions can be stored anywhere on your system - different drives, network paths, or even cloud storage mounts.

**Key Concepts:**
- **Version Directory**: A folder containing a complete snapshot of your software at a specific version
- **Version Number**: A unique identifier (e.g., 1.0.0, 2.1.3) following semantic versioning
- **Key File**: A designated file (usually main executable) that uniquely identifies the version
- **Version Registry**: Internal tracking of all registered versions and their locations

---

## Version Directory Structure

### Basic Structure

Each version is a complete directory containing all files for that version:

```
versions/
├── 1.0.0/                    # Version 1.0.0
│   ├── program.exe           # Key file (main executable)
│   ├── data/
│   │   ├── config.json
│   │   └── assets/
│   │       └── image.png
│   └── libs/
│       └── library.dll
├── 1.0.1/                    # Version 1.0.1
│   ├── program.exe
│   ├── data/
│   │   ├── config.json
│   │   └── assets/
│   │       ├── image.png
│   │       └── new_image.png  # New file
│   └── libs/
│       └── library.dll
└── 1.0.2/                    # Version 1.0.2
    ├── program.exe
    ├── data/
    │   └── config.json
    │       # assets/ folder removed
    └── libs/
        └── library.dll
```

---

## Key File System

### What is a Key File?

A **key file** is a designated file that uniquely identifies a version. It's typically the main program executable.

**Purpose:**
- Uniquely identifies each version
- Prevents applying wrong patches
- Detects corrupted installations

**How it works:**
1. System calculates SHA-256 hash of key file
2. Hash is stored in version manifest
3. Hash is embedded in patch files
4. Patch applier verifies hash before applying

---

### Key File Detection

The generator automatically detects key files using this priority order:

1. `program.exe` (highest priority)
2. `game.exe`
3. `app.exe`
4. `main.exe`
5. First `.exe` file found (Windows)
6. First file without extension (Linux/macOS)

**Example:**
```
Version directory contains:
├── myapp.exe
├── game.exe
├── launcher.exe
└── data/

Key file selected: game.exe (priority 2)
```

---

---

## Version Naming

### Semantic Versioning (Recommended)

Follow semantic versioning format: **MAJOR.MINOR.PATCH**

```
1.0.0 → 1.0.1   # Patch release (bug fixes)
1.0.1 → 1.1.0   # Minor release (new features)
1.1.0 → 2.0.0   # Major release (breaking changes)
```

**Why use semantic versioning?**
- Clear upgrade path
- Easy to understand changes
- Standard convention
- Predictable version ordering

---

### Version Numbering Best Practices

**Good:**
```
1.0.0, 1.0.1, 1.0.2, 1.1.0, 2.0.0
```

**Avoid:**
```
v1.0.0          # Don't use 'v' prefix
1.0             # Always use three parts
1.0.0.1         # Don't use four parts
2023-01-15      # Don't use dates
alpha, beta     # Don't use text-only versions
```

---

### Special Cases

**Pre-release versions:**
```
1.0.0-alpha
1.0.0-beta
1.0.0-rc1
```

**Build metadata:**
```
1.0.0+20230115
1.0.0+build.123
```

---

## Organizing Versions

### Option 1: Flat Structure (Simple)

All versions in one directory:

```
versions/
├── 1.0.0/
├── 1.0.1/
├── 1.0.2/
├── 1.1.0/
└── 2.0.0/
```

**Pros:**
- Simple and clear
- Easy to navigate
- Works for most projects

**Cons:**
- Can get cluttered with many versions

---

### Option 2: Major Version Grouping

Group by major version:

```
versions/
├── v1/
│   ├── 1.0.0/
│   ├── 1.0.1/
│   ├── 1.0.2/
│   └── 1.1.0/
└── v2/
    ├── 2.0.0/
    ├── 2.0.1/
    └── 2.1.0/
```

**Pros:**
- Organized for large projects
- Clear major version separation

**Cons:**
- More complex structure
- Nested paths

---

### Option 3: Date-Based (Archive)

Organize by release date:

```
versions/
├── 2023/
│   ├── 01-January/
│   │   └── 1.0.0/
│   └── 02-February/
│       └── 1.0.1/
└── 2024/
    └── 01-January/
        └── 2.0.0/
```

**Pros:**
- Historical context
- Easy to find by date

**Cons:**
- Harder to find specific version
- Not version-number ordered

---

### Option 4: Multiple Locations

Versions on different storage:

```
C:\releases\             # Production releases
├── 1.0.0/
├── 1.0.1/
└── 1.0.2/

D:\dev-builds\           # Development builds
├── 1.1.0-alpha/
└── 1.1.0-beta/

\\server\archive\        # Archived versions
├── 0.9.0/
└── 0.9.5/
```

**Pros:**
- Separation of concerns
- Different storage types
- Network/cloud storage support

**Cons:**
- More complex management
- Multiple locations to track

---

## Registering Versions

### Automatic Registration

When generating patches, versions are automatically registered:

```bash
# This automatically registers all versions in ./versions
./generator --versions-dir ./versions \
            --new-version 1.0.2 \
            --output ./patches
```

---



## Version Detection Process

### How Generator Detects Versions

1. **Scan versions directory**: Find all subdirectories
2. **Validate version names**: Must be valid version numbers
3. **Find key file**: Use priority order (program.exe → game.exe → app.exe → main.exe)
4. **Calculate hash**: SHA-256 of key file
5. **Scan files**: Recursively scan all files and directories
6. **Create manifest**: Store file list with hashes
7. **Register version**: Add to internal registry

---

### Version Validation

**Required for valid version:**
- ✓ Directory exists
- ✓ Contains at least one file
- ✓ Has detectable key file
- ✓ Name is valid version number

**Example validation:**
```bash
# Valid
./versions/1.0.0/           # ✓ Valid version number
├── program.exe             # ✓ Has key file
└── data/
    └── config.json         # ✓ Has files

# Invalid
./versions/latest/          # ✗ "latest" not a version number
./versions/1.0.0/           # ✗ Directory empty
./versions/v1.0.0/          # ✗ "v" prefix not allowed
```

---

## Managing Many Versions

### Best Practices for 100+ Versions

**1. Archive old versions:**
```bash
# Move old versions to archive
mkdir versions/archive
mv versions/0.* versions/archive/
```

**2. Use symbolic links:**
```bash
# Link to versions on different drives
ln -s /mnt/storage/versions/1.0.0 versions/1.0.0
```

**3. Generate patches incrementally:**
```bash
# Generate patches for recent versions only
./generator --versions-dir ./versions \
            --new-version 2.1.0 \
            --from 2.0.0 \
            --output ./patches

./generator --versions-dir ./versions \
            --new-version 2.1.0 \
            --from 2.0.5 \
            --output ./patches
```

**4. Clean up old patches:**
```bash
# Keep only recent patches
rm patches/0.*-to-*.patch
rm patches/1.0.*-to-*.patch
```

---

### Performance Considerations

**Scanning time:**
- 10 versions: < 1 second
- 100 versions: < 10 seconds
- 1000 versions: < 2 minutes

**Factors affecting performance:**
- Number of files per version
- Storage speed (SSD vs HDD)
- Network latency (for network paths)

**Optimization tips:**
- Use local storage for active versions
- Archive old versions to slower storage
- Use SSD for version directories
- Limit number of active versions

---

## Version Cleanup

### Safe Cleanup Process

1. **Identify versions to remove:**
   ```bash
   # List all versions
   ls -la versions/
   
   # Check which patches exist
   ls -la patches/
   ```

2. **Verify not needed:**
   - Are patches still being generated from this version?
   - Do users still need to upgrade from this version?
   - Is this version used for testing?

3. **Archive before deleting:**
   ```bash
   # Create archive
   mkdir versions/archive
   
   # Move old versions
   mv versions/0.9.* versions/archive/
   
   # Compress archive (optional)
   tar -czf versions-archive-2023.tar.gz versions/archive/
   ```

4. **Delete safely:**
   ```bash
   # Delete archived versions
   rm -rf versions/archive/
   ```

---

### Cleanup Automation

**Example cleanup script (Bash):**
```bash
#!/bin/bash

# Keep versions from last 6 months
CUTOFF_DATE=$(date -d '6 months ago' +%s)

# Find old versions
for VERSION_DIR in versions/*/; do
    VERSION_TIME=$(stat -c %Y "$VERSION_DIR")
    
    if [ "$VERSION_TIME" -lt "$CUTOFF_DATE" ]; then
        echo "Archiving old version: $VERSION_DIR"
        mv "$VERSION_DIR" versions/archive/
    fi
done

# Compress archive
tar -czf versions-archive-$(date +%Y%m%d).tar.gz versions/archive/
rm -rf versions/archive/
```

---

## Migration Strategies

### Moving Versions to New Location

**Scenario:** Moving versions from C: to D:

```bash
# 1. Copy versions to new location
xcopy C:\versions D:\versions /E /I

# 2. Update generator usage
./generator --versions-dir D:\versions \
            --new-version 1.0.3 \
            --output D:\patches

# 3. Verify patches work
./applier --patch D:\patches\1.0.0-to-1.0.3.patch \
          --current-dir ./test-app \
          --dry-run \
          --verify

# 4. Delete old versions (after verification)
rm -rf C:\versions
```

---

### Migrating to Network Storage

**Scenario:** Moving to network share

```powershell
# Windows - Map network drive
net use Z: \\server\software-versions

# Copy versions
xcopy C:\versions Z:\versions /E /I

# Update generator usage
.\generator.exe --versions-dir Z:\versions `
                --new-version 1.0.3 `
                --output Z:\patches
```

---

### Cloud Storage Migration

**Scenario:** Moving to cloud storage (Dropbox, OneDrive, etc.)

```bash
# 1. Move versions to cloud folder
mv versions ~/Dropbox/CyberPatchMaker/versions

# 2. Create symbolic link (optional)
ln -s ~/Dropbox/CyberPatchMaker/versions versions

# 3. Generate patches
./generator --versions-dir ~/Dropbox/CyberPatchMaker/versions \
            --new-version 1.0.3 \
            --output ~/Dropbox/CyberPatchMaker/patches
```

**Note:** Cloud storage may be slower due to sync time.

---

## Version Comparison

### Comparing Two Versions

**Manual comparison:**
```bash
# Linux/macOS - Use diff
diff -r versions/1.0.0 versions/1.0.1

# Show only differences
diff -rq versions/1.0.0 versions/1.0.1

# Windows - Use fc or PowerShell
fc /b versions\1.0.0\program.exe versions\1.0.1\program.exe
```

---

### Understanding Differences

**Output from generator shows:**
```
Comparing versions 1.0.0 → 1.0.1:
  Files added:    5
  Files modified: 12
  Files deleted:  3
  Total changes:  20
```

---

## Version History

### Tracking Version History

**Recommended: Version changelog:**
```
versions/
├── 1.0.0/
├── 1.0.1/
├── 1.0.2/
└── CHANGELOG.md      # Track changes
```

**CHANGELOG.md example:**
```markdown
# Changelog

## 1.0.2 - 2024-01-15
### Added
- New configuration options

### Fixed
- Bug in file processing

## 1.0.1 - 2024-01-10
### Fixed
- Critical security vulnerability

## 1.0.0 - 2024-01-01
### Initial Release
```

---

## Best Practices Summary

### ✓ Do:
- Use semantic versioning (MAJOR.MINOR.PATCH)
- Keep version directories organized
- Archive old versions regularly
- Document changes in CHANGELOG
- Verify versions after copying/moving
- Use consistent naming conventions
- Test patches after generating

### ✗ Don't:
- Use 'v' prefix in version numbers
- Mix version formats (1.0 vs 1.0.0)
- Delete versions without archiving
- Store versions on slow network drives (if possible)
- Modify version directories after registration
- Use spaces or special characters in version numbers

---

## Troubleshooting

### Issue: "Version not found"

**Cause:** Version directory doesn't exist or path is wrong

**Solution:**
```bash
# Check versions directory
ls -la versions/

# Verify version exists
ls -la versions/1.0.0/

# Use absolute path
./generator --versions-dir /full/path/to/versions \
            --new-version 1.0.1 \
            --output ./patches
```

---

### Issue: "Invalid version number"

**Cause:** Version name doesn't follow semantic versioning

**Solution:**
```bash
# Rename version directory
mv versions/v1.0.0 versions/1.0.0
mv versions/version-1.0 versions/1.0.0
mv versions/latest versions/1.0.0
```

---

### Issue: "Key file not found"

**Cause:** No recognized executable in version directory

**Solution:**
```bash
# Add a key file
cp myapp.exe versions/1.0.0/program.exe

# Or rename existing file
mv versions/1.0.0/myapp.exe versions/1.0.0/program.exe
```

---

## Related Documentation

- [Quick Start](quick-start.md) - Getting started
- [Generator Guide](generator-guide.md) - Generating patches
- [Architecture](architecture.md) - System design
- [Troubleshooting](troubleshooting.md) - Common issues
