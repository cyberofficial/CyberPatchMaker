# Backup Directory Exclusion

CyberPatchMaker automatically excludes the backup directory from patch operations to prevent infinite recursion and wasted space.

## Overview

When applying patches, CyberPatchMaker creates a backup of files that will be modified or deleted. This backup is stored in a special directory called `backup.cyberpatcher` located in the target directory.

To prevent the backup directory from being included in patch generation and creating an infinite loop, CyberPatchMaker automatically excludes it from all scanning operations.

## Automatic Exclusions

The following paths are always excluded from patch generation:

1. **`backup.cyberpatcher/`** - The backup directory created during patch application
2. **`.cyberignore`** - The ignore patterns file itself
3. **Any patterns in `.cyberignore`** - User-specified exclusions

## Backup Directory Structure

The backup directory maintains the same structure as the original directory:

```
target_directory/
├── app.exe
├── data/
│   └── game.dat
└── backup.cyberpatcher/
    ├── app.exe          ← Backup of original file
    └── data/
        └── game.dat     ← Backup of original file
```

## Backup Exclusion in Code

### Scanner Level (internal/core/scanner/)

The scanner automatically skips `backup.cyberpatcher` during directory traversal:

```go
// From scanner.go
if strings.Contains(relPath, "backup.cyberpatcher") {
    return nil  // Skip this directory
}
```

### Patcher Level (internal/core/patcher/)

The applier creates selective backups that exclude newly added files:

```go
// From applier.go
// Only backup files that will be modified or deleted
// OpAdd and OpAddDir operations are NOT backed up
for _, op := range operations {
    if op.Type == utils.OpModify || op.Type == utils.OpDelete || op.Type == utils.OpDeleteDir {
        // Add to backup
    }
}
```

## Why This Matters

### Prevents Infinite Recursion

Without backup exclusion, patching would work like this:

1. Generate patch from v1.0 to v1.1
2. Apply patch to v1.0 → creates backup
3. Generate patch from v1.1 to v1.2 → backup files are included
4. Apply patch → backup grows larger
5. **Result**: Infinite growth of patch sizes and backup directories

### Saves Disk Space

By excluding the backup directory:
- Patches remain small (only contain actual changes)
- Backup directories don't get backed up repeatedly
- Disk usage stays predictable

### Improves Performance

- Fewer files to scan during patch generation
- Smaller patch files transfer faster
- Backup operations complete quicker

## Verification

The backup exclusion system works automatically. No configuration is required.

### To Verify It's Working

Generate a patch after applying a previous patch:

```bash
# Apply first patch
patch-apply --patch v1.0.0-to-v1.0.1.patch --current-dir ./app

# Generate second patch
patch-gen --from 1.0.1 --to 1.0.2 --versions-dir ./versions --output ./patches

# Check patch size - should be small, not include backup files
ls -lh patches/1.0.1-to-1.0.2.patch
```

If backup exclusion is working, the patch size will only reflect actual changes between versions, not the backup directory contents.

## Edge Cases

### Nested Backup Directories

If you manually create nested backup directories:

```
app/
├── backup.cyberpatcher/
│   └── backup.cyberpatcher/  ← Still excluded
```

The exclusion pattern matches at any level, so all `backup.cyberpatcher` directories are excluded.

### Case Sensitivity

The backup exclusion is case-sensitive:
- **Excluded**: `backup.cyberpatcher`
- **NOT excluded**: `Backup.cyberpatcher`, `BACKUP.CYBERPATCHER`

### Symbolic Links

If `backup.cyberpatcher` is a symbolic link, it is still excluded from scanning.

## Troubleshooting

### Backup Directory Appearing in Patches

If you see backup files in your patches:

1. **Check the directory name**: Ensure it's exactly `backup.cyberpatcher` (all lowercase, hyphen not underscore)
2. **Check for manual additions**: If you manually added files to the backup directory, they won't be in the original manifest and won't be included anyway
3. **Verify scanner behavior**: Use `--dry-run` to see what files would be included in the patch

### Large Patch Sizes

If patches are unexpectedly large:

1. **Check for large files**: Use compression to reduce patch size
2. **Verify backup exclusion**: Ensure `backup.cyberpatcher` isn't being scanned
3. **Check for duplicates**: Look for files that appear multiple times

## Related Documentation

- [Backup System](backup-system.md) - How backups work
- [Backup Lifecycle](backup-lifecycle.md) - When backups are created and cleaned up
- [Scanner Module](architecture.md#scanner) - How directory scanning works
- [.cyberignore Guide](cyberignore-guide.md) - User-defined exclusions
