# GUI Usage Guide

## Patch Generator GUI

The Patch Generator GUI provides a user-friendly interface for creating binary patches between software versions.

### Features

#### Version Selection
- **Versions Directory**: Select the parent directory containing all your version folders
- **Batch Mode**: Generate patches from ALL existing versions to a new target version
  - When enabled: Only target version needs to be selected
  - When disabled: Both source and target versions are required
  - Automatically discovers and processes all version folders
- **Key Files**: Specify the main executable files for version verification
  - **From Key File**: Key file name in source versions (default: `program.exe`)
  - **To Key File**: Key file name in target version (default: `program.exe`)
  - Can be different if executable was renamed between versions
  - Used to verify patch is being applied to correct version
  - Common examples: `app.exe`, `game.exe`, `program.exe`
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
- **Diff Threshold (KB)**: Minimum file size for binary diff generation
  - Files smaller than this are included as full files
  - Default: 1 KB
  - Increase for faster generation of small file changes

#### Patch Generation

**Normal Mode (Single Patch):**
1. Select your versions directory (containing subdirectories like `1.0.0`, `1.0.1`, etc.)
2. Ensure "Batch Mode" is **unchecked**
3. Enter the key file names for source and target versions (e.g., `program.exe`)
4. Choose source version from the "From Version" dropdown
5. Choose target version from the "To Version" dropdown
6. Select output directory for the patch file
7. Choose compression method and level
8. Configure advanced options (verify, skip identical, diff threshold)
9. Click "Generate Patch"
10. Monitor progress in the log output

**Batch Mode (Multiple Patches):**
1. Select your versions directory (containing subdirectories like `1.0.0`, `1.0.1`, etc.)
2. **Check** the "Batch Mode" checkbox
3. Enter the key file names (From Key File will be used for all source versions)
4. Choose only the target version from "To Version" dropdown
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
9. Advanced Options: All defaults (verify ✓, skip identical ✓, threshold: 1)
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
- Typically the main executable (`.exe`, `.bin`, etc.)

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
   - Adjust diff threshold for small vs large file updates
8. **Batch mode benefits**:
   - Generate all upgrade paths at once
   - Consistent compression settings across all patches
   - Faster than manual single-patch generation

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

## Applying Patches (CLI)

While the GUI focuses on patch generation, patches are applied using the CLI tool:

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
