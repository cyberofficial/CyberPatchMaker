# GUI Usage Guide

## Patch Generator GUI

The Patch Generator GUI provides a user-friendly interface for creating binary patches between software versions.

### Features

#### Version Selection
- **Versions Directory**: Select the parent directory containing all your version folders
- **Key File**: Specify the main executable file used for version verification (default: `program.exe`)
  - This file must exist in both source and target versions
  - Used to verify patch is being applied to correct version
  - Common examples: `app.exe`, `game.exe`, `program.exe`
- **From Version**: Select the source version to patch from
- **To Version**: Select the target version to patch to

#### Output Configuration
- **Output Directory**: Choose where to save the generated patch file
- **Compression**: Select compression method
  - **zstd** (default): Fast compression with good ratio
  - **gzip**: Universal compatibility
  - **none**: No compression (larger file size)

#### Patch Generation
1. Select your versions directory (containing subdirectories like `1.0.0`, `1.0.1`, etc.)
2. Enter the key file name (e.g., `program.exe`) - this must exist in all versions
3. Choose source and target versions from the dropdowns
4. Select output directory for the patch file
5. Choose compression method
6. Click "Generate Patch"
7. Monitor progress in the log output

### Example Workflow

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
2. Key File: program.exe
3. From Version: 1.0.0
4. To Version: 1.0.2
5. Output Directory: E:\MyApp\patches
6. Compression: zstd (default)
7. Click "Generate Patch"

Result: E:\MyApp\patches\1.0.0-to-1.0.2.patch
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
3. **Check logs**: Monitor the log output for detailed progress
4. **Test patches**: Always test generated patches before distribution
5. **Compression choice**: 
   - Use `zstd` for best speed/size balance (recommended)
   - Use `gzip` for maximum compatibility
   - Use `none` only for debugging or very small updates

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
