# GUI Usage Guide

> **🧪 EXPERIMENTAL FEATURE - IN DEVELOPMENT**
> 
> The GUI tools are currently experimental and not recommended for production use.
> For production environments, please use the fully-supported CLI tools:
> - [Generator CLI Guide](generator-guide.md)
> - [Applier CLI Guide](applier-guide.md)
>
> The information below describes the current development state of the GUI.

## Patch Generator GUI

The Patch Generator GUI provides a user-friendly interface for creating binary patches between software versions.

### Features

#### Version Selection
- **Versions Directory**: Select the parent directory containing all your version folders
- **Batch Mode**: Generate patches from ALL existing versions to a new target version
  - When enabled: Only target version needs to be selected
  - When disabled: Both source and target versions are required
  - Automatically discovers and processes all version folders
- **Key Files**: Specify any file for version verification (doesn't have to be executable)
  - **From Key File**: Key file name in source versions (e.g., `program.exe`, `main.dll`, `config.ini`)
  - **To Key File**: Key file name in target version (e.g., `program.exe`, `main.dll`, `config.ini`)
  - Can be different if the key file was renamed between versions
  - Used to verify patch is being applied to correct version
  - Can be any file type: executables (.exe), libraries (.dll), data files, etc.
  - Common examples: `app.exe`, `game.exe`, `core.dll`, `launcher.bin`
- **From Version**: Select the source version to patch from (disabled in batch mode)
- **To Version**: Select the target version to patch to

#### Output Configuration
- **Output Directory**: Choose where to save the generated patch file(s)
- **Compression**: Select compression method
  - **zstd** (default): Fast compression with good ratio
  - **gzip**: Universal compatibility
  - **none**: No compression (larger file size)
- **Compression Level**: Fine-tune compression strength
  - **zstd**: Levels 1-4 (3 = default, balanced)
  - **gzip**: Levels 1-9 (3 = default, balanced)
  - Higher levels = better compression but slower
  - Lower levels = faster but larger patches

#### Advanced Options
- **Verify patches after creation**: Automatically validate generated patches (recommended)
  - Ensures patch integrity before distribution
  - Catches generation errors early
  - Default: Enabled
- **Skip binary-identical files**: Skip files with identical content
  - Improves performance by avoiding unnecessary diffs
  - Reduces patch size for files that haven't changed
  - Default: Enabled
- **Create self-contained executable**: Generate standalone `.exe` with embedded patch data
  - Creates both `.patch` file and self-contained `.exe` file
  - End users get single file that includes applier + patch data
  - Executable size: ~50 MB + patch size
  - Perfect for non-technical users (just double-click to apply)
  - See [Self-Contained Executables Guide](self-contained-executables.md) for details
  - Default: Disabled

#### Patch Generation

**Normal Mode (Single Patch):**
1. Select your versions directory (containing subdirectories like `1.0.0`, `1.0.1`, etc.)
2. Ensure "Batch Mode" is **unchecked**
3. Choose source version from the "From Version" dropdown
4. Select the key file for source version (can be any file: .exe, .dll, .bin, etc.)
5. Choose target version from the "To Version" dropdown
6. Select the key file for target version (can be any file type)
6. Select output directory for the patch file
7. Choose compression method and level
8. Configure advanced options (verify, skip identical)
9. Click "Generate Patch"
10. Monitor progress in the log output

**Batch Mode (Multiple Patches):**
1. Select your versions directory (containing subdirectories like `1.0.0`, `1.0.1`, etc.)
2. **Check** the "Batch Mode" checkbox
3. Choose only the target version from "To Version" dropdown
4. Select key files (From Key File will be used for all source versions, can be any file type)
5. Select output directory for patch files
6. Choose compression method and level
7. Configure advanced options
8. Click "Generate Patch"
9. System will automatically generate patches from ALL existing versions to the target
10. Monitor progress in the log output - shows each patch being generated

### Example Workflows

#### Normal Mode: Single Patch

```
Directory Structure:
E:\MyApp\versions\
├── 1.0.0\
│   ├── program.exe  ← Key file
│   ├── data\
│   └── libs\
├── 1.0.1\
│   ├── program.exe  ← Key file
│   ├── data\
│   └── libs\
└── 1.0.2\
    ├── program.exe  ← Key file
    ├── data\
    └── libs\

Steps:
1. Versions Directory: E:\MyApp\versions
2. Batch Mode: Unchecked
3. From Key File: program.exe
4. To Key File: program.exe
5. From Version: 1.0.0
6. To Version: 1.0.2
7. Output Directory: E:\MyApp\patches
8. Compression: zstd, Level: 3
9. Advanced Options: All defaults (verify ✓, skip identical ✓)
10. Click "Generate Patch"

Result: E:\MyApp\patches\1.0.0-to-1.0.2.patch
```

#### Batch Mode: Multiple Patches

```
Same Directory Structure as above

Steps:
1. Versions Directory: E:\MyApp\versions
2. Batch Mode: ✓ Checked
3. From Key File: program.exe (used for all source versions)
4. To Key File: program.exe
5. From Version: (disabled in batch mode)
6. To Version: 1.0.2
7. Output Directory: E:\MyApp\patches
8. Compression: zstd, Level: 3
9. Advanced Options: All defaults
10. Click "Generate Patch"

Results:
- E:\MyApp\patches\1.0.0-to-1.0.2.patch
- E:\MyApp\patches\1.0.1-to-1.0.2.patch

Note: Batch mode automatically discovers 1.0.0 and 1.0.1, 
      generates patches from each to 1.0.2
```

### Key File Requirements

The key file is critical for patch safety:
- **Must exist** in both source and target version directories
- **Must be at the same relative path** in both versions
- Used to verify patch is applied to correct version
- Prevents applying wrong patches to wrong applications
- **Can be any file type**: executables (`.exe`, `.bin`), libraries (`.dll`, `.so`), data files (`.dat`, `.ini`), etc.
- Just needs to be a stable identifier file that exists in all versions

### Log Output

The log area shows detailed information:
- Version scanning progress
- File counts and directory structures
- Patch generation steps
- Success/error messages
- File paths being processed

### Tips

1. **Organize versions**: Keep all versions in subdirectories under one parent folder
2. **Consistent key files**: Use the same main executable name across all versions
   - If executable name changes, specify different From/To key files
3. **Use batch mode** when releasing new versions:
   - Automatically generates patches from all previous versions
   - Saves time compared to generating each patch individually
   - Ensures users can upgrade from any version
4. **Check logs**: Monitor the log output for detailed progress
5. **Test patches**: Always test generated patches before distribution
6. **Compression tuning**: 
   - Use `zstd` for best speed/size balance (recommended)
   - Level 3 is optimal for most cases (default)
   - Use `gzip` for maximum compatibility
   - Use `none` only for debugging or very small updates
   - Higher levels = slower but smaller patches
7. **Advanced options**:
   - Keep "Verify" enabled for production patches (catches errors early)
   - "Skip identical" improves performance (recommended)
8. **Self-contained executables**:
   - Enable the checkbox for end-user friendly distribution
   - Creates both `.patch` (advanced) and `.exe` (simple) files
   - Perfect for non-technical users - just double-click and apply
   - Increases file size by ~50 MB per patch (includes GUI applier)
   - See [Self-Contained Executables Guide](self-contained-executables.md) for details
9. **Batch mode benefits**:
   - Generate all upgrade paths at once
   - Consistent compression settings across all patches
   - Faster than manual single-patch generation
   - Works with self-contained executables (creates .exe for each patch)

### Error Handling

Common errors and solutions:

**"Versions directory does not exist"**
- Verify the path is correct
- Check you have read permissions

**"No versions found"**
- Ensure subdirectories exist in the versions folder
- Directory names should match version numbers

**"Failed to register version"**
- Key file might not exist in the version directory
- Check key file name spelling

**"From and To versions must be different"**
- Select different source and target versions

**"Failed to scan version"**
- Check file permissions
- Verify directory is not corrupted

## Batch Script Generator

The Batch Script Generator tab provides a tool for creating Windows batch scripts (.bat files) that end users can double-click to apply patches easily. This eliminates the need for end users to use the command line.

### Features

#### Batch Script Configuration
- **Patch File**: Select the `.patch` file to create a script for
- **Target Directory**: Optional - specify the directory to patch (default: script location)
- **Script Name**: Customize the batch script filename (default: `apply_patch.bat`)

#### Options
- **Include dry-run option**: Creates a safe test run script that won't make changes
  - Useful for testing patches before actual deployment
  - Users can verify patch compatibility without risk
- **Disable verification (--verify=false)**: Skip hash verification
  - Faster but less safe
  - Only use if verification is causing issues
- **Disable backup (--backup=false)**: Skip backup creation
  - Saves disk space
  - Not recommended for production use

#### Custom Instructions
- Add custom messages for end users
- Instructions appear in the script as comments and on-screen messages
- Useful for providing context or special instructions

#### Preview
- Real-time preview of generated batch script
- See exactly what the script will look like before generating
- Preview updates automatically as you change options

### Generated Batch Script Features

The generated batch scripts include:
- **Clear user interface**: Friendly prompts and messages
- **Error handling**: Detects and reports failures clearly
- **Success confirmation**: Shows clear success/failure messages
- **Pause prompts**: Allows users to see results before closing
- **Custom instructions**: Your messages displayed to users
- **Professional appearance**: Branded with CyberPatchMaker

### Example Workflow

```
Steps:
1. Open CyberPatchMaker GUI
2. Switch to "Batch Script Generator" tab
3. Click "Browse" next to "Patch File"
4. Select: E:\MyApp\patches\1.0.0-to-1.0.2.patch
5. Leave "Target Directory" empty (will use script location)
6. Keep script name: apply_patch.bat
7. Add custom instructions:
   "Please close the application before running this patch."
   "The update will take approximately 30 seconds."
8. Keep all options unchecked (safe defaults)
9. Review the preview
10. Click "Generate Batch Script"

Result: E:\MyApp\patches\apply_patch.bat

The script will be saved in the same directory as the patch file.
End users can simply double-click apply_patch.bat to apply the patch.
```

### Generated Script Example

```batch
@echo off
REM CyberPatchMaker Patch Application Script
REM Generated by CyberPatchMaker GUI
REM
REM Please close the application before running this patch.
REM The update will take approximately 30 seconds.
REM

title CyberPatchMaker - Applying Patch

echo ========================================
echo CyberPatchMaker - Patch Application
echo ========================================
echo.
echo Please close the application before running this patch.
echo The update will take approximately 30 seconds.
echo.
echo This script will apply the patch to your installation.
echo.
pause

echo.
echo Applying patch...
echo.

applier.exe --patch "1.0.0-to-1.0.2.patch" --current-dir "%~dp0"

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ========================================
    echo PATCH APPLIED SUCCESSFULLY!
    echo ========================================
    echo.
    echo Your installation has been updated.
    echo.
) else (
    echo.
    echo ========================================
    echo PATCH APPLICATION FAILED!
    echo ========================================
    echo.
    echo Error code: %ERRORLEVEL%
    echo.
    echo Please check the error messages above.
    echo If verification failed, your installation may have been modified.
    echo If backup was enabled, you can restore from the .backup folder.
    echo.
)

echo.
pause
```

### Distribution Workflow

1. **Generate patches** using the Patch Generator tab
2. **Generate batch scripts** using the Batch Script Generator tab
3. **Package for distribution**:
   ```
   MyApp_Update_1.0.2\
   ├── 1.0.0-to-1.0.2.patch
   ├── 1.0.1-to-1.0.2.patch
   ├── applier.exe          ← Include the CLI applier
   ├── apply_from_1.0.0.bat ← Batch script for 1.0.0 users
   └── apply_from_1.0.1.bat ← Batch script for 1.0.1 users
   ```
4. **Distribute** the package to users
5. **User instructions**: "Double-click the batch script that matches your current version"

### Tips

1. **Multiple scripts**: Generate separate batch scripts for each patch file
   - Name them clearly: `apply_from_1.0.0.bat`, `apply_from_1.0.1.bat`
   - Users can easily identify which script to use

2. **Include applier.exe**: Always package `applier.exe` with your batch scripts
   - Batch script calls `applier.exe` from the same directory
   - Users don't need CyberPatchMaker installed

3. **Custom instructions**: Use custom instructions to:
   - Warn users to close the application
   - Provide backup instructions
   - Set expectations (time required, disk space needed)
   - Include support contact information

4. **Dry-run scripts**: Consider providing two versions:
   - `test_patch.bat` with dry-run enabled (safe testing)
   - `apply_patch.bat` for actual application

5. **Target directory**:
   - Leave empty if patch should be applied in the same location
   - Specify a path for fixed installation locations
   - Use `%~dp0` (script directory) for relative paths

6. **Test scripts**: Always test generated batch scripts before distribution
   - Run on a test system
   - Verify error handling works correctly
   - Confirm success/failure messages are clear

### Error Handling in Scripts

Generated scripts include comprehensive error handling:
- **Error code display**: Shows Windows error level
- **Clear failure messages**: Explains what went wrong
- **Backup reminders**: Tells users how to restore if needed
- **Pause before exit**: Users can read error messages

### Dry-Run Mode

When "Include dry-run option" is checked:
- Script performs a safe test run
- No changes are made to files
- Verifies patch can be applied successfully
- Different success message indicates test mode
- Users can confirm compatibility before real application

### Benefits

1. **User-Friendly**: No command-line knowledge required
2. **Professional**: Branded, clear interface
3. **Safe**: Built-in error handling and verification
4. **Customizable**: Add your own instructions and branding
5. **Portable**: Scripts can be distributed with patches
6. **Flexible**: Support for dry-run, custom paths, and options

## Applying Patches (CLI)

While the GUI can generate batch scripts for end users, patches can also be applied directly using the CLI tool:

```powershell
.\dist\patch-apply.exe apply `
    --patch "patches\1.0.0-to-1.0.2.patch" `
    --target "C:\Program Files\MyApp" `
    --verify
```

See [applier-guide.md](applier-guide.md) for detailed instructions on applying patches.

## Building the GUI

The GUI requires CGO and a compatible GCC compiler:

```powershell
# Build all tools including GUI
.\build.ps1

# Or build GUI only
$env:PATH = "C:\TDM-GCC-64\bin;" + $env:PATH
go build -o dist/patch-gui.exe ./cmd/patch-gui
```

**Requirements:**
- Go 1.21 or later
- TDM-GCC or compatible C compiler for CGO
- Fyne dependencies (automatically handled by `go build`)

## Troubleshooting

**GUI won't start**
- Verify all dependencies are installed
- Check you have graphics drivers (OpenGL)
- Try running from command line to see error messages

**Patch generation fails**
- Verify key file exists in both versions
- Check you have write permissions to output directory
- Ensure sufficient disk space
- Review log output for specific errors

**Generated patch is large**
- Try different compression methods
- Verify versions are actually different
- Check if binary files are being compressed (some formats don't compress well)

## Next Steps

- [Quick Start Guide](quick-start.md)
- [CLI Reference](cli-reference.md)
- [Understanding Patches](how-it-works.md)
- [Testing Guide](testing-guide.md)
