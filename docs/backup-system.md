# Backup System Documentation

## Overview

CyberPatchMaker includes an intelligent backup system that creates a mirror-structure backup of only the files being modified or deleted during patch application. This makes manual rollback simple and intuitive.

## Backup Behavior

### What Gets Backed Up

- **Modified Files**: Files that will be changed by the patch
- **Deleted Files**: Files that will be removed by the patch
- **Deleted Directories**: Directories that will be removed, backed up with full contents using CopyDir
- **Not Backed Up**: New files or directories being added (since they don't exist yet)

### Backup Location

Backups are created inside the target directory as:
```
target_directory/backup.cyberpatcher/
```

### Directory Structure

The backup mirrors the exact directory structure of the original files:

**Example:**

Original structure:
```
C:\MyApp\
├── program.exe
├── folder1\
│   └── somefile.dll
└── folder5\
    └── folder4\
        └── someotherfile.map
```

If `program.exe`, `folder1\somefile.dll`, and `folder5\folder4\someotherfile.map` are being patched, the backup will be:

```
C:\MyApp\backup.cyberpatcher\
├── program.exe
├── folder1\
│   └── somefile.dll
└── folder5\
    └── folder4\
        └── someotherfile.map
```

## Manual Rollback

To rollback a patch manually:

1. Copy files from `backup.cyberpatcher` to their corresponding locations
2. The directory structure in the backup matches the original, so just copy:
   - `backup.cyberpatcher\program.exe` → `program.exe`
   - `backup.cyberpatcher\folder1\somefile.dll` → `folder1\somefile.dll`
   - `backup.cyberpatcher\folder5\folder4\someotherfile.map` → `folder5\folder4\someotherfile.map`

3. Delete the `backup.cyberpatcher` folder when done

## Automatic Behavior

- **Before Patching**: Backup is created automatically if enabled (default: enabled)
- **After Successful Patching**: Backup is **preserved** for manual rollback
- **After Failed Patching**: Automatic rollback restores files from backup, then the backup remains for manual investigation

## CLI Usage

```bash
# Apply patch with backup (default)
.\patch-apply.exe --patch patch.patch --current-dir C:\MyApp --verify

# Apply patch without backup (NOT recommended)
.\patch-apply.exe --patch patch.patch --current-dir C:\MyApp --verify --backup=false
```

## Benefits

- **Selective**: Only backs up files that will change (saves disk space)
- **Mirror Structure**: Easy to understand and manually copy back
- **Preserved**: Backup stays after successful patching for safety
- **Quality of Life**: Simple drag-and-drop rollback if needed
- **Transparent**: Shows exactly which files were modified/deleted

## Notes

- The backup folder is named `backup.cyberpatcher` to avoid conflicts
- Each patch application overwrites the previous backup
- Backup is created AFTER pre-patch verification succeeds
- If verification fails, no backup is created (nothing is changed)
