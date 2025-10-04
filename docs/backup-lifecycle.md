# Backup Lifecycle

Understanding how CyberPatchMaker manages backups is crucial for understanding the safety guarantees of the system.

## Overview

CyberPatchMaker creates backups of your installation **only when necessary** and at **exactly the right time** to ensure maximum data integrity and safety.

## Critical Timing: When Backups Are Created

Backups are created **AFTER pre-verification passes** but **BEFORE any operations are applied**.

This timing is critical because:
- ✅ Pre-verification ensures the backup captures a **verified clean state**
- ✅ Backup exists before any modifications, enabling restoration if operations fail
- ❌ Never backs up corrupted or unverified state

## The Three Scenarios

### Scenario 1: Successful Patch Application ✅

**Flow:**
```
1. Load patch and display information
2. Pre-verification (verify all files match source version)
   ✓ Key file hash matches
   ✓ All required files match expected hashes
3. Create backup ← Captures VERIFIED CLEAN STATE
4. Apply operations (add/modify/delete files)
5. Post-verification (verify all files match target version)
   ✓ All modified files match expected hashes
6. Remove backup (success, no longer needed)
7. Success message
```

**Result:** Installation upgraded from 1.0.0 to 1.0.1, backup cleaned up

**Example Output:**
```
Applying patch from 1.0.0 to 1.0.1...
Verifying current version...
Pre-patch verification successful

Creating backup...
Backup created at: ./myapp.backup

Applying 20 operations...
  Modified: program.exe
  Modified: data/config.json
  Added: libs/newfeature.dll

Post-patch verification successful

Removing backup...

=== Patch Applied Successfully ===
Version updated from 1.0.0 to 1.0.1
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

**Result:** Installation remains in its current state (corrupted), no changes made

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

---

### Scenario 3: Operation/Post-Verification Failure ❌ (Mid-Patch Failure)

**Flow:**
```
1. Load patch and display information
2. Pre-verification (verify all files match source version)
   ✓ Key file hash matches
   ✓ All required files match expected hashes
3. Create backup ← Captures VERIFIED CLEAN STATE
4. Apply operations (add/modify/delete files)
   ✗ OPERATION FAILS (permission error, disk full, corrupted diff, etc.)
   or
5. Post-verification (verify all files match target version)
   ✗ Modified file hash DOES NOT MATCH expected
6. Backup still exists (not removed due to error)
7. Restore from backup ← Restore VERIFIED CLEAN STATE
8. Exit with error code
```

**Result:** Installation restored to original clean state from backup

**Example Output:**
```
Applying patch from 1.0.0 to 1.0.1...
Verifying current version...
Pre-patch verification successful

Creating backup...
Backup created at: ./myapp.backup

Applying 20 operations...
  Modified: program.exe
  Modified: data/config.json
  Error: failed to write file: permission denied

Restoring from backup...
Backup restored successfully

Error: patch application failed
```

**Why Restoration Works:**
- Backup was created from **verified clean state** (after pre-verification)
- Restoring from backup **guarantees** return to clean version 1.0.0
- User can retry after fixing the issue (e.g., file permissions, disk space)

---

## Backup Storage

### Location
Backup is stored at: `<targetDir>.backup`

Example:
- Target directory: `C:\MyApp\`
- Backup location: `C:\MyApp.backup\`

### Contents
The backup contains a **complete recursive copy** of the entire directory tree:
- All files at all levels
- All subdirectories
- Complete directory hierarchy preserved
- All file permissions preserved (where supported)

### Cleanup
- **On success**: Backup is automatically removed
- **On failure**: Backup is preserved for manual recovery or automatic restoration

## Implementation Details

### Code Location
Backup management is implemented in `internal/core/patcher/applier.go`:

```go
// In ApplyPatch function (after pre-verification):
if createBackup {
    fmt.Println("\nCreating backup...")
    backupDir := targetDir + ".backup"
    if err := a.createBackup(targetDir, backupDir); err != nil {
        return fmt.Errorf("failed to create backup: %w", err)
    }
    fmt.Printf("Backup created at: %s\n", backupDir)
}

// ... apply operations ...

// In ApplyPatch function (after post-verification success):
if createBackup {
    fmt.Println("Removing backup...")
    backupDir := targetDir + ".backup"
    if err := os.RemoveAll(backupDir); err != nil {
        fmt.Printf("Warning: failed to remove backup: %v\n", err)
    }
}
```

### Backup Methods

**createBackup(srcDir, backupDir string)**:
1. Removes existing backup if present
2. Creates new backup directory
3. Recursively copies all files using `copyDir`

**copyDir(src, dst string)**:
1. Reads directory entries
2. For each entry:
   - If directory: create directory and recurse
   - If file: copy file using `utils.CopyFile`
3. Uses `filepath.Join` for cross-platform paths

## Why This Design?

### Previous Design (Incorrect)
❌ Backup created in `main.go` **BEFORE** calling `ApplyPatch`
- **Problem**: Backup captured **unverified state** (potentially corrupted)
- **Problem**: If source was corrupted, backup was corrupted
- **Problem**: "Restoration" would restore corrupted state

### Current Design (Correct)
✅ Backup created in `applier.go` **AFTER** pre-verification passes
- **Guarantee**: Backup captures **verified clean state**
- **Guarantee**: Restoration truly restores clean version
- **Guarantee**: Never backs up corrupted installations

## Best Practices

### For Users
1. **Always use --verify flag**: Enables pre/post verification and automatic backup
2. **Ensure sufficient disk space**: Backup requires space equal to installation size
3. **Don't modify backup directory**: Let the system manage it
4. **Check error messages**: Pre-verification failures indicate problems before patching

### For Developers
1. **Never create backup before verification**: Always verify state is clean first
2. **Backup ownership**: Keep backup logic in applier package, not main
3. **Cross-platform paths**: Use `filepath.Join`, not string concatenation
4. **Cleanup on success**: Remove backup to free disk space
5. **Preserve on failure**: Keep backup for investigation or manual recovery

## Troubleshooting

**Backup directory exists but patch fails:**
- Previous patch may have failed
- Safe to delete `<targetDir>.backup` and retry

**Not enough disk space for backup:**
- Free up space equal to installation size
- Or use `--no-backup` flag (not recommended)

**Backup restoration fails:**
- Check disk space
- Check file permissions
- Manually copy files from `.backup` to target directory

**Pre-verification fails (no backup created):**
- This is **correct behavior** - don't backup corrupted state
- Fix the underlying issue:
  - Re-install clean source version
  - Use correct patch for your version
  - Check for file corruption

## Related Documentation

- [Hash Verification](hash-verification.md) - Understanding pre/post verification
- [Error Handling](error-handling.md) - What errors mean and how to fix them
- [Safety Features](how-it-works.md#safety-features) - Other safety mechanisms
- [Applier Tool Guide](applier-guide.md) - Using the applier tool
