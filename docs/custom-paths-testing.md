# Custom Paths Testing Guide

## Overview

This document provides comprehensive testing procedures for the new custom directory paths feature in CyberPatchMaker. This feature allows patch generation between versions located on different drives, network locations, or arbitrary file system paths.

## Feature Description

### CLI Support
- New flags: `--from-dir` and `--to-dir`
- Accepts full paths to version directories
- Extracts version numbers from directory names using `filepath.Base()`
- Maintains backward compatibility with `--versions-dir` mode

### GUI Support
- Checkbox: "Use Custom Paths (different drives/locations)"
- Independent directory pickers for source and target versions
- Disables batch mode when custom paths are enabled
- Maintains backward compatibility with legacy single-directory mode

## Test Scenarios

### Scenario 1: CLI - Same Drive, Different Paths

**Purpose**: Verify CLI works with versions in different directories on the same drive

**Setup**:
```powershell
# Create test versions in different locations
New-Item -ItemType Directory -Path "C:\test-versions\release-1.0.0" -Force
New-Item -ItemType Directory -Path "C:\test-versions\release-1.0.0\data" -Force
"config v1" | Out-File "C:\test-versions\release-1.0.0\data\config.json"
"program v1" | Out-File "C:\test-versions\release-1.0.0\program.exe"

New-Item -ItemType Directory -Path "C:\test-builds\build-1.0.1" -Force
New-Item -ItemType Directory -Path "C:\test-builds\build-1.0.1\data" -Force
"config v2" | Out-File "C:\test-builds\build-1.0.1\data\config.json"
"program v2" | Out-File "C:\test-builds\build-1.0.1\program.exe"

New-Item -ItemType Directory -Path "C:\test-patches" -Force
```

**Command**:
```powershell
.\bin\patch-gen.exe --from-dir "C:\test-versions\release-1.0.0" --to-dir "C:\test-builds\build-1.0.1" --output "C:\test-patches"
```

**Expected Results**:
- ✅ Patch file created: `C:\test-patches\release-1.0.0-to-build-1.0.1.patch`
- ✅ Console output shows extracted version numbers
- ✅ No errors in patch generation
- ✅ Patch file size > 0 bytes

**Cleanup**:
```powershell
Remove-Item -Recurse -Force "C:\test-versions", "C:\test-builds", "C:\test-patches"
```

---

### Scenario 2: CLI - Cross-Drive Paths

**Purpose**: Verify CLI works with versions on different drives (primary use case)

**Prerequisites**: Access to at least two drives (e.g., C: and D:)

**Setup**:
```powershell
# Version 1.0.0 on C: drive
New-Item -ItemType Directory -Path "C:\app-releases\1.0.0" -Force
New-Item -ItemType Directory -Path "C:\app-releases\1.0.0\libs" -Force
"app v1.0.0" | Out-File "C:\app-releases\1.0.0\app.exe"
"library v1" | Out-File "C:\app-releases\1.0.0\libs\core.dll"

# Version 1.0.1 on D: drive
New-Item -ItemType Directory -Path "D:\backups\app-1.0.1" -Force
New-Item -ItemType Directory -Path "D:\backups\app-1.0.1\libs" -Force
"app v1.0.1" | Out-File "D:\backups\app-1.0.1\app.exe"
"library v2" | Out-File "D:\backups\app-1.0.1\libs\core.dll"

New-Item -ItemType Directory -Path "C:\patches" -Force
```

**Command**:
```powershell
.\bin\patch-gen.exe --from-dir "C:\app-releases\1.0.0" --to-dir "D:\backups\app-1.0.1" --output "C:\patches"
```

**Expected Results**:
- ✅ Patch file created: `C:\patches\1.0.0-to-app-1.0.1.patch`
- ✅ Console shows scanning both drives
- ✅ Patch contains operations for modified files
- ✅ No errors related to cross-drive access

**Cleanup**:
```powershell
Remove-Item -Recurse -Force "C:\app-releases", "D:\backups\app-1.0.1", "C:\patches"
```

---

### Scenario 3: CLI - Backward Compatibility (Legacy Mode)

**Purpose**: Verify existing --versions-dir workflow still works

**Setup**:
```powershell
New-Item -ItemType Directory -Path "C:\versions\1.0.0" -Force
"app v1.0.0" | Out-File "C:\versions\1.0.0\app.exe"

New-Item -ItemType Directory -Path "C:\versions\1.0.1" -Force
"app v1.0.1" | Out-File "C:\versions\1.0.1\app.exe"

New-Item -ItemType Directory -Path "C:\patches" -Force
```

**Command**:
```powershell
.\bin\patch-gen.exe --versions-dir "C:\versions" --from "1.0.0" --to "1.0.1" --output "C:\patches"
```

**Expected Results**:
- ✅ Patch file created: `C:\patches\1.0.0-to-1.0.1.patch`
- ✅ Same behavior as before custom paths feature
- ✅ No breaking changes

**Cleanup**:
```powershell
Remove-Item -Recurse -Force "C:\versions", "C:\patches"
```

---

### Scenario 4: GUI - Custom Paths Mode (Same Drive)

**Purpose**: Verify GUI custom paths mode works correctly

**Manual Steps**:
1. Launch GUI: `.\bin\patch-gui.exe`
2. Check "Use Custom Paths (different drives/locations)"
3. Observe:
   - ✅ From Directory and To Directory entries are enabled
   - ✅ Versions Directory entry is disabled
   - ✅ From/To version dropdowns are disabled
   - ✅ Batch Mode checkbox is disabled
4. Click "Browse" next to "From Directory"
5. Select `C:\test-versions\release-1.0.0`
6. Observe:
   - ✅ Path appears in From Directory entry
7. Click "Browse" next to "To Directory"
8. Select `C:\test-builds\build-1.0.1`
9. Observe:
   - ✅ Path appears in To Directory entry
10. Click "Browse" next to "Output Directory"
11. Select `C:\test-patches`
12. Observe:
   - ✅ Generate Patch button becomes enabled
13. Click "Generate Patch"
14. Observe:
   - ✅ Log shows "=== CUSTOM PATHS MODE ==="
   - ✅ Log shows extracted version numbers
   - ✅ Patch generation succeeds
   - ✅ Success dialog appears
   - ✅ Patch file created with correct name

**Setup** (before GUI test):
```powershell
New-Item -ItemType Directory -Path "C:\test-versions\release-1.0.0\data" -Force
"config v1" | Out-File "C:\test-versions\release-1.0.0\data\config.json"
"program v1" | Out-File "C:\test-versions\release-1.0.0\program.exe"

New-Item -ItemType Directory -Path "C:\test-builds\build-1.0.1\data" -Force
"config v2" | Out-File "C:\test-builds\build-1.0.1\data\config.json"
"program v2" | Out-File "C:\test-builds\build-1.0.1\program.exe"

New-Item -ItemType Directory -Path "C:\test-patches" -Force
```

---

### Scenario 5: GUI - Custom Paths Mode (Cross-Drive)

**Purpose**: Verify GUI works with cross-drive selection

**Prerequisites**: Access to C: and D: drives

**Manual Steps**:
1. Launch GUI: `.\bin\patch-gui.exe`
2. Check "Use Custom Paths (different drives/locations)"
3. Select From Directory: `C:\app-releases\1.0.0`
4. Select To Directory: `D:\backups\app-1.0.1`
5. Select Output Directory: `C:\patches`
6. Click "Generate Patch"
7. Verify:
   - ✅ Log shows paths from both drives
   - ✅ Patch generation succeeds
   - ✅ Patch file created

**Setup** (use same setup as Scenario 2)

---

### Scenario 6: GUI - Legacy Mode (Backward Compatibility)

**Purpose**: Verify GUI legacy mode still works after custom paths implementation

**Manual Steps**:
1. Launch GUI: `.\bin\patch-gui.exe`
2. Ensure "Use Custom Paths" is UNCHECKED
3. Observe:
   - ✅ Versions Directory entry is enabled
   - ✅ From/To version dropdowns are enabled
   - ✅ Batch Mode checkbox is enabled
   - ✅ Custom path entries are disabled
4. Click "Browse" next to "Versions Directory"
5. Select `C:\versions`
6. Click "Scan Versions"
7. Observe:
   - ✅ From/To version dropdowns populate with versions
8. Select From: `1.0.0`
9. Select To: `1.0.1`
10. Select Output Directory: `C:\patches`
11. Click "Generate Patch"
12. Verify:
   - ✅ Log shows "=== LEGACY MODE ==="
   - ✅ Patch generation succeeds
   - ✅ Same behavior as before

**Setup** (use same setup as Scenario 3)

---

### Scenario 7: GUI - Mode Toggle Behavior

**Purpose**: Verify switching between modes works correctly

**Manual Steps**:
1. Launch GUI
2. Select Versions Directory: `C:\versions`
3. Scan Versions
4. Select From: `1.0.0`, To: `1.0.1`
5. Check "Use Custom Paths"
6. Observe:
   - ✅ Legacy controls disabled
   - ✅ Custom path entries enabled
   - ✅ Generate button disabled (no custom paths selected yet)
7. Uncheck "Use Custom Paths"
8. Observe:
   - ✅ Legacy controls re-enabled
   - ✅ Custom path entries disabled
   - ✅ Previous selections (1.0.0, 1.0.1) still present
   - ✅ Generate button enabled again

---

### Scenario 8: CLI - Version Number Extraction

**Purpose**: Verify correct version number extraction from various path formats

**Test Cases**:

**Case A: Simple version number**
```powershell
.\bin\patch-gen.exe --from-dir "C:\releases\1.0.0" --to-dir "C:\releases\1.0.1" --output "C:\patches"
# Expected patch name: 1.0.0-to-1.0.1.patch
```

**Case B: Prefixed version number**
```powershell
.\bin\patch-gen.exe --from-dir "C:\releases\v1.0.0" --to-dir "C:\releases\v1.0.1" --output "C:\patches"
# Expected patch name: v1.0.0-to-v1.0.1.patch
```

**Case C: Descriptive directory names**
```powershell
.\bin\patch-gen.exe --from-dir "C:\releases\release-1.0.0" --to-dir "C:\builds\build-1.0.1" --output "C:\patches"
# Expected patch name: release-1.0.0-to-build-1.0.1.patch
```

**Case D: Date-based directory names**
```powershell
.\bin\patch-gen.exe --from-dir "C:\releases\2025-01-15-v1" --to-dir "C:\releases\2025-01-20-v2" --output "C:\patches"
# Expected patch name: 2025-01-15-v1-to-2025-01-20-v2.patch
```

**Expected Results**:
- ✅ All cases generate patches with correct names
- ✅ Version numbers match directory names
- ✅ No errors in extraction logic

---

### Scenario 9: Error Handling - Invalid Paths

**Purpose**: Verify proper error handling for invalid paths

**Test Cases**:

**Case A: Non-existent from directory**
```powershell
.\bin\patch-gen.exe --from-dir "C:\does-not-exist\1.0.0" --to-dir "C:\versions\1.0.1" --output "C:\patches"
```
Expected: ❌ Error message about missing directory

**Case B: Non-existent to directory**
```powershell
.\bin\patch-gen.exe --from-dir "C:\versions\1.0.0" --to-dir "C:\does-not-exist\1.0.1" --output "C:\patches"
```
Expected: ❌ Error message about missing directory

**Case C: Same directory for from and to**
```powershell
.\bin\patch-gen.exe --from-dir "C:\versions\1.0.0" --to-dir "C:\versions\1.0.0" --output "C:\patches"
```
Expected: ❌ Error about same version

**Expected Results**:
- ✅ Clear error messages
- ✅ No crashes
- ✅ Helpful guidance for users

---

### Scenario 10: Compression Options with Custom Paths

**Purpose**: Verify compression works with custom paths

**Commands**:
```powershell
# zstd compression (default)
.\bin\patch-gen.exe --from-dir "C:\v1" --to-dir "C:\v2" --output "C:\patches" --compression zstd

# gzip compression
.\bin\patch-gen.exe --from-dir "C:\v1" --to-dir "C:\v2" --output "C:\patches" --compression gzip

# no compression
.\bin\patch-gen.exe --from-dir "C:\v1" --to-dir "C:\v2" --output "C:\patches" --compression none
```

**Expected Results**:
- ✅ All compression methods work
- ✅ Patch sizes vary based on compression
- ✅ Patches are valid and applicable

---

## Validation Checklist

After completing all scenarios, verify:

### CLI Validation
- [ ] `--from-dir` and `--to-dir` flags work correctly
- [ ] Version numbers extracted from paths correctly
- [ ] Cross-drive paths work (C: to D:)
- [ ] Backward compatibility maintained (`--versions-dir` mode)
- [ ] Help text updated with new flags
- [ ] Error handling for invalid paths
- [ ] All compression options work

### GUI Validation
- [ ] Custom paths checkbox toggle works
- [ ] Directory pickers open and select correctly
- [ ] Custom path entries show selected paths
- [ ] Generate button validation works (enables/disables)
- [ ] Batch mode disabled in custom paths mode
- [ ] Legacy mode still works correctly
- [ ] Mode switching preserves state appropriately
- [ ] Log output shows correct mode
- [ ] Success dialog appears on completion

### Integration Validation
- [ ] Patches generated by CLI can be applied
- [ ] Patches generated by GUI can be applied
- [ ] Cross-drive patches work end-to-end
- [ ] Version verification works correctly
- [ ] File hash validation succeeds

## Known Limitations

1. **Batch Mode**: Not supported with custom paths mode (by design)
2. **Version Number Extraction**: Uses directory name as version number (must be valid)
3. **Network Paths**: Should work but may have different performance characteristics

## Troubleshooting

### Issue: "Directory does not exist"
- **Cause**: Invalid path or insufficient permissions
- **Solution**: Verify path exists and user has read access

### Issue: "Same version detected"
- **Cause**: From and to directories have the same name
- **Solution**: Ensure directory names are different

### Issue: "Generate button disabled in GUI"
- **Cause**: Missing required fields
- **Solution**: In custom mode, all three fields required (from, to, output)

## Success Criteria

The custom paths feature is considered fully functional when:

1. ✅ All CLI scenarios pass
2. ✅ All GUI scenarios pass
3. ✅ Error handling works correctly
4. ✅ Backward compatibility maintained
5. ✅ Cross-drive support confirmed
6. ✅ Documentation updated
7. ✅ No regressions in existing functionality

## Next Steps

After testing completes successfully:
1. Update main README.md with custom paths examples
2. Update CLI-EXAMPLES.md with --from-dir/--to-dir usage
3. Create user-facing guide for GUI custom paths mode
4. Consider adding network path examples to documentation
5. Consider adding unit tests for version extraction logic
