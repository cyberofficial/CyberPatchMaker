# Testing CLI Self-Contained Executables

This guide explains how to test the new CLI self-contained executable feature.

## What Was Implemented

### CLI Generator (`patch-gen.exe`)
- Added `--create-exe` flag to create self-contained CLI executables
- Embeds patch data into `patch-apply.exe` (CLI applier)
- Creates both `.patch` file and `.exe` file
- Works with all generation modes (single, batch, custom paths)

### CLI Applier (`patch-apply.exe`)
- Detects embedded patch data on startup
- Interactive console menu interface
- Options: Dry Run, Apply Patch, Toggle 1GB Bypass, Change Target Directory
- User-friendly console interface for non-technical users

## Testing Steps

### 1. Build Executables

Already done! The executables are built:
- `patch-gen.exe` - CLI generator
- `patch-apply.exe` - CLI applier

### 2. Generate Self-Contained Executable

A test executable has already been created:
```
testdata\patches\1.0.0-to-1.0.1.exe
```

To create more:
```powershell
# Single patch with executable
.\patch-gen.exe --from-dir "testdata\versions\1.0.0" --to-dir "testdata\versions\1.0.1" --output "testdata\patches" --create-exe

# Batch mode with executables
.\patch-gen.exe --versions-dir "testdata\versions" --new-version "1.0.2" --output "testdata\patches" --create-exe
```

### 3. Test Interactive Console Interface

#### Option 1: Test in Current Directory
```powershell
cd testdata\test-cli-apply
..\patches\1.0.0-to-1.0.1.exe
```

#### Option 2: Test with --ignore1gb Flag
```powershell
.\testdata\patches\1.0.0-to-1.0.1.exe --ignore1gb
```

### 4. Interactive Menu Testing

When you run the executable, you should see:

```
==============================================
  CyberPatchMaker - Self-Contained Patch
==============================================

=== Patch Information ===
From Version:     1.0.0
To Version:       1.0.1
Key File:         program.exe
Required Hash:    ...
Files Added:      1
Files Modified:   3
Files Deleted:    0

Target directory [current directory]:

==============================================
Options:
  1. Dry Run (simulate without changes)
  2. Apply Patch
  3. Toggle 1GB Bypass Mode (currently: Disabled)
  4. Change Target Directory
  5. Exit
==============================================
Select option [1-5]:
```

#### Test Each Menu Option:

**Option 1: Dry Run**
- Simulates patch without making changes
- Shows what would be modified/added/deleted
- Verifies key file and required files
- Safe to run multiple times

**Option 2: Apply Patch**
- Prompts for confirmation ("yes"/"no")
- Creates backup before applying
- Shows progress during application
- Displays success message when complete

**Option 3: Toggle 1GB Bypass**
- Toggles between Enabled/Disabled
- Shows warning when enabled
- Useful for large patches

**Option 4: Change Target Directory**
- Allows selecting a different directory
- Validates directory exists
- Updates target for subsequent operations

**Option 5: Exit**
- Cleanly exits the program

### 4. Verify Patch Application

After applying the patch, verify:

1. **Files were updated:**
   ```powershell
   ls testdata\test-cli-apply
   ```

2. **Backup was created:**
   ```powershell
   ls testdata\test-cli-apply\backup.cyberpatcher
   ```

3. **Backup includes deleted directories:**
   If the patch deletes directories, they should be backed up with all their contents:
   ```powershell
   tree /f testdata\test-cli-apply\backup.cyberpatcher
   ```

4. **Version file updated:**
   Check that files match version 1.0.1

### 6. Test Error Handling

#### Test 1: Run in Wrong Directory
```powershell
cd C:\
.\path\to\1.0.0-to-1.0.1.exe
```
Should prompt to select correct directory.

#### Test 2: Key File Mismatch
Copy executable to directory with wrong version - should fail verification in dry run.

#### Test 3: Cancel Patch Application
Select option 2, then type "no" when prompted - should cancel gracefully.

## Expected Results

### ✅ Success Criteria

1. **Generation:**
   - Both `.patch` and `.exe` files created
   - Executable size = applier size + patch size + 128 bytes
   - No errors during generation

2. **Interactive Menu:**
   - All menu options work correctly
   - Clear, user-friendly messages
   - Proper error handling

3. **Dry Run:**
   - Shows detailed simulation
   - No changes made to files
   - Accurate operation preview

4. **Patch Application:**
   - Backup created successfully
   - All files updated correctly
   - Success message displayed
   - Can be run multiple times

5. **1GB Bypass:**
   - Toggle works correctly
   - Warning displayed when enabled
   - State persists during session

## Comparison: GUI vs CLI Executables

| Feature | GUI Executable | CLI Executable |
|---------|----------------|----------------|
| Created by | `patch-gen-gui.exe --create-exe` | `patch-gen.exe --create-exe` |
| Uses | `patch-apply-gui.exe` | `patch-apply.exe` |
| Interface | Graphical windows | Interactive console |
| Dry Run | Button click | Menu option 1 |
| Apply | Button click | Menu option 2 |
| 1GB Bypass | Checkbox | Menu option 3 |
| Target Dir | Browse button | Menu option 4 |
| Best For | Non-technical users | Scripters, CLI users |

## Known Limitations

1. **Windows Only:** Currently only generates `.exe` files
2. **Interactive Only:** CLI executable requires user interaction (not for automation)
3. **Memory Usage:** Large patches (>1GB) require significant RAM
4. **Single Use:** Each executable is for one specific version transition

## Troubleshooting

### "Failed to create executable"
- Ensure `patch-apply.exe` is in the same directory as `patch-gen.exe`
- Check file permissions

### "Checksum mismatch"
- Executable may be corrupted
- Regenerate the executable

### "Directory not found"
- Verify target directory exists
- Use absolute paths if relative paths fail

### Menu not displaying correctly
- Ensure console supports Unicode
- Try running in Windows Terminal or PowerShell

## Next Steps

After manual testing, users can:

1. **Distribute executables** to end users
2. **Create batch of executables** for all version transitions
3. **Document user instructions** for running executables
4. **Set up automated generation** in CI/CD pipeline

## Documentation

For more information, see:
- [Self-Contained Executables Guide](docs/self-contained-executables.md)
- [Generator Guide](docs/generator-guide.md)
- [CLI Examples](docs/CLI-EXAMPLES.md)
- [Applier Guide](docs/applier-guide.md)

## Feedback

If you encounter issues:
1. Note the exact error message
2. Check console output for details
3. Verify file sizes and checksums
4. Try with a fresh test directory
