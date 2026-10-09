# Backup System

## Overview

When applying a patch, CyberPatchMaker creates a **selective mirror-structure backup** inside the target directory at `backup.cyberpatcher/`. Only files being modified or deleted are backed up — new files and directories being added are not backed up (they don't exist yet). The backup mirrors exact directory structure so manual rollback is drag-and-drop intuitive.

## Timing

Backups are created **after pre-verification passes but before any operations are applied**. This ensures:

- Backup captures a **verified clean state** (never backs up corrupted files)
- If pre-verification fails, **no backup is created** (nothing to safely restore)
- Backup exists before modifications, enabling rollback if operations fail

## What Gets Backed Up

| Operation | Backed Up? |
|-----------|-----------|
| Modified files (`OpModify`) | Yes |
| Deleted files (`OpDelete`) | Yes |
| Deleted directories (`OpDeleteDir`) | Yes (full `CopyDir` with all contents) |
| Added files (`OpAdd`) | No |
| Added directories (`OpAddDir`) | No |

## Behavior Summary

| Scenario | Result |
|----------|--------|
| **Pre-verification fails** | No backup created, no changes made |
| **Patch succeeds** | Backup preserved for manual rollback |
| **Operation fails mid-patch** | Automatic rollback from backup restores original state; backup kept for investigation |
| **Post-patch verification fails** | Automatic rollback from backup restores original state; backup kept |

## Automatic Exclusion

The scanner automatically skips `backup.cyberpatcher` during directory traversal (checked by relative path prefix). This prevents infinite recursion: without exclusion, patching v1.0→v1.1 would include the backup folder from the previous patch in the next scan cycle.

The `.cyberignore` file itself is also auto-excluded. Only the **root-level** `backup.cyberpatcher` is excluded — nested directories with the same name are not auto-excluded (add them to `.cyberignore` if needed).

## Selective Strategy (Why Not Full Backup?)

Backing up only the files touched by the patch keeps the backup a small fraction of the full installation size:

- **Disk space**: Only changed files are copied, not the entire application
- **Speed**: Copying a handful of files is much faster than duplicating everything
- **Rollback simplicity**: Mirror structure preserves exact original paths

## Implementation

Backup logic lives in `internal/core/patcher/applier.go`:

```go
// After pre-verification:
if createBackup {
    backupDir := filepath.Join(targetDir, "backup.cyberpatcher")
    a.createMirrorBackup(targetDir, backupDir, patch.Operations)
}
// On success: backup preserved
// On failure: a.restoreMirrorBackup() restores, cleans up added files/dirs
```

**`createMirrorBackup`**: Removes existing backup, iterates operations, copies OpModify/OpDelete files and OpDeleteDir directories with mirror structure. Skips OpAdd/OpAddDir.

**`restoreMirrorBackup`**: Restores backed-up files/dirs to original locations, then removes files/dirs added during the failed patch.

## Manual Rollback

```
# Mirror structure makes this straightforward:
Copy-Item .\backup.cyberpatcher\* . -Recurse -Force

# Then remove files added by the patch (not in backup)
# Then delete backup.cyberpatcher when confirmed
```

## CLI Usage

```bash
# Backup enabled by default
patch-apply --patch patch.patch --current-dir ./myapp --verify

# Disable backup (not recommended)
patch-apply --patch patch.patch --current-dir ./myapp --verify --backup=false
```
