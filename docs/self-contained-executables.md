# Self-Contained Patch Executables

> **NEW FEATURE**
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
3. **Check the "Create self-contained executable" checkbox**
4. Click "Generate Patch"
5. Both files are created:
   - `patch-1.0.0-to-1.0.1.patch` (standard patch file)
   - `patch-1.0.0-to-1.0.1.exe` (self-contained GUI executable)

### Using the CLI Generator

You can also create self-contained executables from the command line:

```powershell
# Generate a single patch with self-contained executable
patch-gen --from-dir "C:\releases\1.0.0" --to-dir "C:\releases\1.0.1" --output patches --create-exe

# Using versions directory
patch-gen --versions-dir "C:\versions" --from "1.0.0" --to "1.0.1" --output patches --create-exe

# Batch mode with executables
patch-gen --versions-dir "C:\versions" --new-version "1.0.3" --output patches --create-exe
```

**CLI vs GUI Executables:**
- GUI-created executables use `patch-apply-gui.exe` (graphical interface)
- CLI-created executables use `patch-apply.exe` (console interface)
- Both work identically for patching, but have different user interfaces

**CLI Self-Contained Features:**
- Interactive console menu with options
- Dry-run simulation before applying
- 1GB bypass toggle in the console
- Manual target directory selection
- Same patch format and verification as GUI version

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
2. Parses header structure
3. **Security validations** (fail-safe design):
   - Validates format version (only v1 supported)
   - Checks for magic bytes "CPMPATCH"
   - Verifies `DataOffset == StubSize` (no gaps)
   - Validates `StubSize + DataSize + HEADER_SIZE == fileSize` (exact match)
   - Ensures offsets are within file bounds
   - Limits patch size to max 1 GB (prevents memory exhaustion)
4. If all validations pass:
   - Extracts patch data from validated offset
   - Verifies SHA-256 checksum
   - Decompresses if needed
   - Loads patch into GUI automatically
5. If any validation fails:
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

#### GUI Self-Contained Executable (Created by GUI)

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

#### CLI Self-Contained Executable (Created by CLI)

1. **Download**: One `.exe` file matching their version
2. **Run**: Double-click the executable
3. **Console Opens**: Shows interactive menu:
   ```
   ==============================================
     CyberPatchMaker - Self-Contained Patch
   ==============================================

   === Patch Information ===
   From Version:     1.0.0
   To Version:       1.0.1
   Key File:         program.exe
   Required Hash:    a3f5b2c1d4e6f8...
   Files Added:      15
   Files Modified:   8
   Files Deleted:    2

   Target directory [D:\Games\MyGame]:

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
4. **Choose Option**: Select 1 for dry run or 2 to apply
5. **Apply**: Confirm with "yes" when prompted
6. **Done**: Patch applied successfully

**Choosing Between GUI and CLI:**
- **GUI** (`patch-gen-gui.exe --create-exe`): Best for non-technical users who prefer visual interfaces
- **CLI** (`patch-gen.exe --create-exe`): Best for automation, scripting, or users comfortable with consoles

### User Benefits

- **Simple**: Only one file to download
- **No tools needed**: Everything embedded
- **No confusion**: Can't select wrong patch file
- **Portable**: Single file can be shared easily
- **Safe**: Still includes all verification and backup features

### Automation Mode (Silent Flag)

Self-contained executables support a `--silent` flag for **fully automated patching** without user interaction:

```powershell
# Automated silent patching
1.2.4-to-1.2.5.exe --silent

# Silent mode with explicit target directory
1.2.4-to-1.2.5.exe --silent --current-dir C:\MyApp
```

**Features:**
- **No prompts**: Applies patch automatically without asking
- **Default settings**: Uses verify=true and backup=true
- **Exit codes**: Returns 0 on success, 1 on failure
- **Minimal output**: Only essential status messages
- **Perfect for**:
  - Automated deployments via scripts
  - CI/CD pipelines
  - Mass deployments across machines
  - Task Scheduler / cron jobs
  - Unattended updates

**Example: PowerShell deployment script**
```powershell
# Deploy patch to multiple machines
$servers = @("Server1", "Server2", "Server3")

foreach ($server in $servers) {
    Write-Host "Updating $server..."
    & "\\share\1.2.4-to-1.2.5.exe" --silent --current-dir "\\$server\C$\MyApp"
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ Success" -ForegroundColor Green
    } else {
        Write-Host "✗ Failed" -ForegroundColor Red
    }
}
```

**Example: Task Scheduler automation**
```powershell
# Create scheduled task for automated patching at 2 AM
$action = New-ScheduledTaskAction `
    -Execute "C:\Patches\1.2.4-to-1.2.5.exe" `
    -Argument "--silent --current-dir C:\MyApp"

$trigger = New-ScheduledTaskTrigger -At 2:00AM -Daily

Register-ScheduledTask `
    -TaskName "MyApp Auto-Update" `
    -Action $action `
    -Trigger $trigger `
    -User "SYSTEM"
```

See [Applier Guide - Automation Mode](applier-guide.md#automation-mode-silent-flag) for more details.

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

### Use Self-Contained Executables When:

- Target audience is non-technical users
- Simplicity is more important than bandwidth
- You want to minimize support requests
- Distribution platform allows large files
- Users have reasonable internet speeds
- You want one-click update experience

### Use Traditional Patches When:

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

### "Patch size exceeds 1GB limit"

**Problem**: Self-contained executable fails to load because patch data is over 1GB

**Solution**:
- Use CLI flag: Run executable with `patch-1.0.0-to-1.0.1.exe --ignore1gb`
- Or: Enable "Ignore 1GB limit" checkbox in the GUI before applying
- Note: Requires sufficient RAM to load large patch into memory
- Consider: If patch is very large, traditional separate .patch file may be better

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
