# Self-Contained Patch Executables

> **✨ NEW FEATURE**
> 
> Create standalone `.exe` files that embed patch data, making distribution simpler for end users.

## Overview

Self-contained patch executables combine the patch applier and patch data into a single `.exe` file. End users simply download one file, drop it into their application folder, and double-click to apply the patch - no additional tools or setup required.

## How It Works

### Traditional Workflow
```
User downloads:
├── patch-1.0.0-to-1.0.1.patch    ← Patch file
└── patch-apply-gui.exe            ← Applier tool

User must:
1. Download both files
2. Run patch-apply-gui.exe
3. Browse for .patch file
4. Select target directory
5. Click apply
```

### Self-Contained Workflow
```
User downloads:
└── patch-1.0.0-to-1.0.1.exe      ← Everything in one file

User must:
1. Download one file
2. Drop into game/app folder
3. Double-click
4. Click apply (patch is already loaded!)
```

## Creating Self-Contained Executables

### Using the GUI Generator

1. Open `patch-gen-gui.exe`
2. Configure your patch as normal:
   - Select versions directory
   - Choose from/to versions (or use batch mode)
   - Set compression and options
3. **Check the "Create self-contained executable" checkbox** ☑
4. Click "Generate Patch"
5. Both files are created:
   - `patch-1.0.0-to-1.0.1.patch` (standard patch file)
   - `patch-1.0.0-to-1.0.1.exe` (self-contained executable)

### Batch Mode

When using batch mode with the self-contained option enabled:
```
Input:
- versions/1.0.0/
- versions/1.0.1/
- versions/1.0.2/
- Target version: 1.0.3

Output (with checkbox enabled):
patches/
├── 1.0.0-to-1.0.3.patch
├── 1.0.0-to-1.0.3.exe ← Self-contained
├── 1.0.1-to-1.0.3.patch
├── 1.0.1-to-1.0.3.exe ← Self-contained
├── 1.0.2-to-1.0.3.patch
└── 1.0.2-to-1.0.3.exe ← Self-contained
```

## Technical Details

### File Structure

A self-contained executable consists of three parts:

```
┌─────────────────────────────┐
│ patch-apply-gui.exe         │  ← Base applier (~50 MB)
├─────────────────────────────┤
│ Compressed Patch Data       │  ← Your patch (varies)
├─────────────────────────────┤
│ 128-byte Header             │  ← Metadata at end
└─────────────────────────────┘
```

### Header Format (128 bytes)

Located at the end of the file:

| Offset | Size | Field | Description |
|--------|------|-------|-------------|
| 0-7 | 8 bytes | Magic | "CPMPATCH" identifier |
| 8-11 | 4 bytes | Version | Format version (currently 1) |
| 12-19 | 8 bytes | Stub Size | Size of applier executable |
| 20-27 | 8 bytes | Data Offset | Where patch data starts |
| 28-35 | 8 bytes | Data Size | Size of patch data |
| 36-51 | 16 bytes | Compression | Type: "zstd", "gzip", or "none" |
| 52-83 | 32 bytes | Checksum | SHA-256 of patch data |
| 84-127 | 44 bytes | Reserved | For future use |

### Detection Process

When `patch-apply-gui.exe` starts:
1. Reads last 128 bytes of itself
2. Checks for magic bytes "CPMPATCH"
3. If found:
   - Validates checksum
   - Extracts patch data
   - Decompresses if needed
   - Loads patch into GUI automatically
4. If not found:
   - Runs in normal mode (browse for .patch file)

## File Sizes

### Size Breakdown

Base executable: **~50 MB** (includes GUI framework)

Example patch sizes:
- Small update (10 MB changed): **50.01 MB** total
- Medium update (50 MB changed): **55 MB** total (with zstd)
- Large update (200 MB changed): **80 MB** total (with zstd)

### Size Optimization Tips

1. **Use compression**: Always enable zstd compression
   ```
   Without compression: 50 MB + 10 MB = 60 MB
   With zstd: 50 MB + 2 MB = 52 MB
   ```

2. **Skip identical files**: Enable "Skip identical files" option
   - Reduces patch data size
   - Only includes actual changes

3. **Higher compression level**: Use level 3 or 4 for production
   - Level 3 (default): Good balance
   - Level 4: Smallest size, slower generation

## Distribution

### Recommended Distribution Structure

```
MyGame_Update_v1.0.3/
├── README.txt                           ← Instructions for users
├── patch-1.0.0-to-1.0.3.exe            ← For users on 1.0.0
├── patch-1.0.1-to-1.0.3.exe            ← For users on 1.0.1
├── patch-1.0.2-to-1.0.3.exe            ← For users on 1.0.2
└── advanced/                            ← Optional: for power users
    ├── patch-apply-gui.exe              ← Standalone applier
    ├── patch-1.0.0-to-1.0.3.patch       ← Standard patch files
    ├── patch-1.0.1-to-1.0.3.patch
    └── patch-1.0.2-to-1.0.3.patch
```

### README.txt Example

```
MyGame Update v1.0.3

EASY INSTALL (Recommended):
1. Find which version you currently have installed
2. Download the matching patch file:
   - patch-1.0.0-to-1.0.3.exe if you're on v1.0.0
   - patch-1.0.1-to-1.0.3.exe if you're on v1.0.1
   - patch-1.0.2-to-1.0.3.exe if you're on v1.0.2
3. Place the .exe in your game folder (same location as MyGame.exe)
4. Double-click the patch .exe
5. Click "Apply Patch"

ADVANCED:
See the /advanced folder for manual patching with separate patch files.

Need help? Visit: https://support.mygame.com/updates
```

## User Experience

### What End Users See

1. **Download**: One `.exe` file matching their version
2. **Run**: Double-click the executable
3. **GUI Opens**: Shows patch information automatically:
   ```
   ┌─────────────────────────────────────────────┐
   │ Apply Patch: 1.0.0 → 1.0.1                  │
   ├─────────────────────────────────────────────┤
   │ [Embedded Patch Data]                       │
   │                                             │
   │ Target: D:\Games\MyGame\              [...]│
   │                                             │
   │ From: 1.0.0      Files to Add:    15       │
   │ To:   1.0.1      Files to Modify: 8        │
   │ Key:  game.exe   Files to Delete: 2        │
   │                                             │
   │ Log:                                        │
   │ ✓ Self-contained patch loaded               │
   │   From version: 1.0.0                       │
   │   To version: 1.0.1                         │
   │   Target directory: D:\Games\MyGame\        │
   │                                             │
   │   Click 'Apply Patch' when ready...        │
   │                                             │
   │ [Dry Run]  [Apply Patch]  [Close]          │
   └─────────────────────────────────────────────┘
   ```
4. **Apply**: Click "Apply Patch" button
5. **Done**: Patch applied successfully

### User Benefits

- ✅ **Simple**: Only one file to download
- ✅ **No tools needed**: Everything embedded
- ✅ **No confusion**: Can't select wrong patch file
- ✅ **Portable**: Single file can be shared easily
- ✅ **Safe**: Still includes all verification and backup features

## Advantages vs Traditional Patches

### For End Users

| Feature | Traditional | Self-Contained |
|---------|-------------|----------------|
| Files to download | 2 (applier + patch) | 1 (exe only) |
| Setup complexity | Medium | None |
| Confusion possible | Yes (wrong patch file) | No (embedded) |
| File management | Must keep organized | Single file |
| Portability | Multi-file | Single file |

### For Developers

| Feature | Traditional | Self-Contained |
|---------|-------------|----------------|
| Distribution | Multiple files | Clean, single file |
| Support burden | Higher (setup issues) | Lower (simpler) |
| User errors | More common | Rare |
| Bandwidth | Lower (~2-10 MB) | Higher (~50+ MB) |
| Hosting cost | Lower | Higher |

## When to Use

### ✅ Use Self-Contained Executables When:

- Target audience is non-technical users
- Simplicity is more important than bandwidth
- You want to minimize support requests
- Distribution platform allows large files
- Users have reasonable internet speeds
- You want one-click update experience

### ❌ Use Traditional Patches When:

- Bandwidth/hosting costs are a concern
- Target users have slow internet
- Patch files are very small (< 5 MB)
- Users are technical and prefer flexibility
- You need multiple applier versions
- Storage space is limited

## Troubleshooting

### "Failed to create executable: failed to read applier executable"

**Problem**: Generator can't find `patch-apply-gui.exe`

**Solution**: 
- Ensure `patch-apply-gui.exe` is in the same directory as `patch-gen-gui.exe`
- Check file hasn't been renamed or deleted
- Verify read permissions on applier file

### "Executable created but won't run"

**Problem**: Self-contained exe fails to launch

**Solution**:
- Check Windows doesn't block downloaded exe (right-click → Properties → Unblock)
- Verify file isn't corrupted (check file size)
- Try running from command line to see error messages
- Ensure user has execution permissions

### "Checksum mismatch (data corrupted)"

**Problem**: Embedded patch data fails validation

**Solution**:
- Re-download the executable (file may be corrupted)
- Check antivirus didn't quarantine or modify file
- Verify download completed successfully
- Re-generate the self-contained exe

### "No patches generated in batch mode with checkbox"

**Problem**: Self-contained executables not created in batch mode

**Solution**:
- Verify "Create self-contained executable" checkbox is enabled
- Check `patch-apply-gui.exe` exists in generator directory
- Look for errors in log output
- Ensure sufficient disk space (50 MB × number of patches)

## Best Practices

### Generation

1. **Always test**: Create and test self-contained exe before distribution
2. **Both formats**: Generate both traditional and self-contained versions
   - Offer self-contained as "Easy Install"
   - Provide traditional as "Advanced"
3. **Compression**: Always use zstd compression to minimize final size
4. **Batch mode**: Use batch mode to create all version paths at once

### Distribution

1. **Clear naming**: Use descriptive names
   ```
   Good: patch-1.0.0-to-1.0.3-INSTALLER.exe
   Better: MyGame_Update_v1.0.3_from_v1.0.0.exe
   ```

2. **User instructions**: Provide clear README explaining:
   - Which version user needs
   - Where to place the file
   - How to run it
   - What to expect

3. **Mirror files**: Host on multiple locations
   - Primary download (fast)
   - Mirror/backup (reliability)

4. **Checksums**: Provide SHA-256 checksums
   ```
   MyGame_Update_v1.0.3_from_v1.0.0.exe
   SHA256: a3d5f6b8c2e1...
   ```

### Support

1. **Version check**: Remind users to verify their current version first
2. **Close app**: Instruct users to close application before patching
3. **Backup note**: Remind that automatic backups are created
4. **Rollback**: Explain how to restore from backup if needed

## Technical Limitations

### Current Limitations

1. **Windows Only**: Currently only generates `.exe` files for Windows
2. **Fixed Base Size**: Base applier is ~50 MB regardless of patch size
3. **No Streaming**: Entire file must be downloaded before use
4. **Single Compression**: Can't mix compression methods in one exe

### Future Enhancements

Potential future improvements:
- Linux/Mac support (.AppImage, .app bundles)
- Progress bar during embedded patch extraction
- Custom branding/icons for generated executables
- Compression of the applier executable itself
- Delta updates for self-contained exes

## Related Documentation

- [Generator Guide](generator-guide.md) - Creating patches
- [GUI Usage](gui-usage.md) - Using the generator GUI
- [Applier Guide](applier-guide.md) - Applying patches
- [How It Works](how-it-works.md) - Understanding the patch system
- [Compression Guide](compression-guide.md) - Optimizing patch sizes
