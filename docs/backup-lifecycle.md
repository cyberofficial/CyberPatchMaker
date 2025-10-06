# Backup Lifecycle

Understanding how CyberPatchMaker manages backups is crucial for understanding the safety guarantees of the system.

## Overview

CyberPatchMaker creates **selective backups** of your installation **only when necessary** and at **exactly the right time** to ensure maximum data integrity and safety. Unlike traditional full-directory backups, the system intelligently backs up only the files that will be modified or deleted.

## Critical Timing: When Backups Are Created

Backups are created **AFTER pre-verification passes** but **BEFORE any operations are applied**.

This timing is critical because:
- ✅ Pre-verification ensures the backup captures a **verified clean state**
- ✅ Backup exists before any modifications, enabling restoration if operations fail
- ✅ Only files being changed are backed up (selective strategy)
- ❌ Never backs up corrupted or unverified state
- ❌ New files being added are NOT backed up (they don't exist yet)

## The Three Scenarios

### Scenario 1: Successful Patch Application ✅

**Flow:**
```
1. Load patch and display information
2. Pre-verification (verify all files match source version)
   ✓ Key file hash matches
   ✓ All required files match expected hashes
3. Create selective backup ← Captures VERIFIED files that will CHANGE
   ✓ Backs up files to be modified
   ✓ Backs up files to be deleted
   ✗ Does NOT back up new files being added
4. Apply operations (add/modify/delete files)
5. Post-verification (verify all files match target version)
   ✓ All modified files match expected hashes
6. Backup PRESERVED (kept for manual rollback if needed)
7. Success message
```

**Result:** Installation upgraded from 1.0.0 to 1.0.1, backup **preserved** in `target\backup.cyberpatcher\`

**Example Output:**
```
Applying patch from 1.0.0 to 1.0.1...
Verifying current version...
Pre-patch verification successful

Creating selective backup...
Backing up: program.exe
Backing up: data\config.json
Backup created in: C:\MyApp\backup.cyberpatcher

Applying 20 operations...
  Modified: program.exe
  Modified: data\config.json
  Added: libs\newfeature.dll

Post-patch verification successful

=== Patch Applied Successfully ===
Version updated from 1.0.0 to 1.0.1
Backup preserved in: C:\MyApp\backup.cyberpatcher
```

---

### Scenario 2: Pre-Verification Failure ❌ (Corrupted Installation)

**Flow:**
```
1. Load patch and display information
2. Pre-verification (verify all files match source version)
   ✗ Key file hash DOES NOT MATCH
   or
   ✗ Required file hash DOES NOT MATCH
3. STOP - NO BACKUP CREATED ← Why backup corrupted state?
4. Return error immediately
5. Exit with error code
```

**Result:** Installation remains in its current state (corrupted), no changes made, no backup created

**Example Output:**
```
Applying patch from 1.0.0 to 1.0.1...
Verifying current version...
Error: patch application failed: key file verification failed: 
  key file checksum mismatch: 
  expected 63573ff071ea5fa2...
  got      8f3c9d2e1a4b7c5e...

This patch requires version 1.0.0
Your installation may be corrupted or modified
```

**Why No Backup?**
- Pre-verification **failed** → Current state is **not** version 1.0.0
- Current state is either corrupted or a different version
- **We don't backup corrupted/wrong state** → Nothing to safely restore to
- **We don't attempt to "fix" corrupted installations** → User must resolve

**What Should User Do?**
1. Verify they have the correct version
2. Re-install clean version 1.0.0 if corrupted
3. Use correct patch for their version
4. Check for file modifications or corruption

---

### Scenario 3: Operation/Post-Verification Failure ❌ (Mid-Patch Failure)

**Flow:**
```
1. Load patch and display information
2. Pre-verification (verify all files match source version)
   ✓ Key file hash matches
   ✓ All required files match expected hashes
3. Create selective backup ← Captures VERIFIED files that will CHANGE
4. Apply operations (add/modify/delete files)
   ✗ OPERATION FAILS (permission error, disk full, corrupted diff, etc.)
   or
5. Post-verification (verify all files match target version)
   ✗ Modified file hash DOES NOT MATCH expected
6. Backup still exists in target\backup.cyberpatcher
7. Automatic rollback restores from backup
8. Exit with error code
```

**Result:** Installation automatically restored to original clean state from backup

**Example Output:**
```
Applying patch from 1.0.0 to 1.0.1...
Verifying current version...
Pre-patch verification successful

Creating selective backup...
Backing up: program.exe
Backing up: data\config.json
Backup created in: C:\MyApp\backup.cyberpatcher

Applying 20 operations...
  Modified: program.exe
  Error: failed to write file: permission denied

Rolling back from backup...
Restored: program.exe
Rollback complete

Error: patch application failed
Installation restored to original state
```

**Why Restoration Works:**
- Backup was created from **verified clean state** (after pre-verification)
- Restoring from backup **guarantees** return to clean version 1.0.0
- User can retry after fixing the issue (e.g., file permissions, disk space)
- Failed backup remains at `target\backup.cyberpatcher` for investigation

---

## Backup Storage

### Location
Backup is stored **inside the target directory** at: `<targetDir>\backup.cyberpatcher\`

Example:
- Target directory: `C:\MyApp\`
- Backup location: `C:\MyApp\backup.cyberpatcher\`

### Contents
The backup contains a **selective mirror-structure copy** of only the files being changed:
- **Modified files**: Files that will be changed by the patch
- **Deleted files**: Files that will be removed by the patch
- **NOT included**: New files being added (they don't exist yet)
- **Directory structure**: Mirrored exactly to preserve original paths
- **File permissions**: Preserved where supported

### Cleanup Behavior
- **On success**: Backup is **PRESERVED** for manual rollback if needed
- **On failure**: Backup **PRESERVED** for automatic rollback or investigation
- **User responsibility**: Delete `backup.cyberpatcher` folder when no longer needed

## Implementation Details

### Code Location
Backup management is implemented in `internal/core/patcher/applier.go`:

```go
// In ApplyPatch function (after pre-verification):
if createBackup {
    fmt.Println("\nCreating selective backup...")
    backupDir := filepath.Join(targetDir, "backup.cyberpatcher")
    if err := a.createSelectiveBackup(targetDir, backupDir, patch); err != nil {
        return fmt.Errorf("failed to create backup: %w", err)
    }
    fmt.Printf("Backup created in: %s\n", backupDir)
}

// ... apply operations ...

// On success: Backup is PRESERVED (NOT removed)
// On failure: Automatic rollback uses backup, then preserves it
```

### Backup Methods

**createSelectiveBackup(targetDir, backupDir string, patch *Patch)**:
1. Removes existing `backup.cyberpatcher` if present
2. Iterates through patch operations
3. For each OpModify or OpDelete operation:
   - Determines source file path in targetDir
   - Creates matching directory structure in backupDir
   - Copies file from targetDir to backupDir with preserved path
4. For OpAdd operations: **Skips** (new files don't exist yet)
5. Uses `filepath.Join` for cross-platform paths

**rollbackFromBackup(backupDir, targetDir string, backedUpFiles []string)**:
1. Iterates through list of backed-up files
2. For each file:
   - Reads file from backupDir
   - Writes file to targetDir (overwrites failed changes)
   - Preserves directory structure
3. On success, backup remains in `backup.cyberpatcher` for investigation

## Why This Design?

### Design Evolution

**Initial Design (Incorrect)**:
❌ Backup created in `main.go` **BEFORE** calling `ApplyPatch`
- **Problem**: Backup captured **unverified state** (potentially corrupted)
- **Problem**: If source was corrupted, backup was corrupted
- **Problem**: "Restoration" would restore corrupted state

**Improved Design (Better)**:
✅ Backup created in `applier.go` **AFTER** pre-verification passes
- **Fixed**: Backup captures **verified clean state**
- **Fixed**: Restoration truly restores clean version
- **Fixed**: Never backs up corrupted installations
- **Problem**: Full directory copy wasted disk space

**Current Design (Optimal)**:
✅ **Selective backup** created **AFTER** pre-verification passes
- **Guarantee**: Backup captures **verified clean state**
- **Guarantee**: Only backs up files being changed (OpModify + OpDelete)
- **Guarantee**: Mirror structure makes manual rollback intuitive
- **Guarantee**: Backup **preserved after success** for safety
- **Efficiency**: Minimal disk space usage (only changed files)
- **Quality of Life**: Simple drag-and-drop manual recovery

## Best Practices

### For Users
1. **Keep backup enabled** (default behavior): Backup is enabled by default for safety
2. **Ensure sufficient disk space**: Backup requires space for changed files only (much less than full installation)
3. **Don't modify backup.cyberpatcher directory**: Let the system manage it during patching
4. **Cleanup when safe**: Delete `backup.cyberpatcher` folder after confirming patch success
5. **Check error messages**: Pre-verification failures indicate problems before patching

### For Developers
1. **Never create backup before verification**: Always verify state is clean first
2. **Selective backup strategy**: Only back up OpModify and OpDelete operations
3. **Backup ownership**: Keep backup logic in applier package, not main
4. **Cross-platform paths**: Use `filepath.Join`, not string concatenation
5. **Preserve backup**: Keep backup after success AND failure for user safety
6. **Mirror structure**: Preserve exact directory structure for intuitive manual recovery

## Troubleshooting

**backup.cyberpatcher directory exists from previous patch:**
- Previous patch may have succeeded (backup preserved for safety)
- Previous patch may have failed (backup used for rollback)
- Safe to delete `target\backup.cyberpatcher` before applying new patch
- New patch will overwrite any existing backup anyway

**Not enough disk space for backup:**
- Free up space for changed files (much less than full installation)
- Check patch metadata to see how many files will be backed up
- Or use `--backup=false` flag (not recommended - disables rollback safety)

**Need to manually rollback:**
1. Files in `backup.cyberpatcher` mirror the original structure
2. Copy files from `backup.cyberpatcher\` to their original locations
3. Example: `backup.cyberpatcher\program.exe` → `program.exe`
4. Example: `backup.cyberpatcher\folder1\somefile.dll` → `folder1\somefile.dll`
5. Delete `backup.cyberpatcher` folder when done

**Automatic rollback fails:**
- Check disk space (needs space to restore files)
- Check file permissions on target directory
- Manually restore from `backup.cyberpatcher` as described above

**Pre-verification fails (no backup created):**
- This is **correct behavior** - don't backup corrupted state
- Fix the underlying issue:
  - Re-install clean source version
  - Use correct patch for your version
  - Check for file corruption or modifications

**Want to see what will be backed up:**
- Run applier with `--dry-run` flag
- Output shows which files will be backed up before patching

## Related Documentation

- [Backup System](backup-system.md) - Overview and quick reference for current backup system
- [Hash Verification](hash-verification.md) - Understanding pre/post verification
- [How It Works](how-it-works.md) - Complete patching workflow including backup
- [Applier Tool Guide](applier-guide.md) - Using the applier tool with backup options
- [Testing Guide](testing-guide.md) - Tests 23-27 validate backup system functionality
