# Key File System Deep Dive

## Overview

The Key File System is a critical security feature in CyberPatchMaker that prevents users from applying patches to the wrong version or entirely different application. This document provides a comprehensive deep dive into how key files work, how they are specified, and advanced usage scenarios.

## Purpose and Benefits

### Why Key Files?

Without key file verification, several problematic scenarios could occur:

1. **Wrong Version**: User tries to apply 1.0.1→1.0.3 patch to version 1.0.2
2. **Different Application**: User tries to apply "GameX" patch to "GameY"
3. **Corrupted Installation**: Modified or incomplete installation accepts patch
4. **Renamed Directories**: User moves/renames folders and loses version tracking

Key files solve these problems by using a designated file's SHA-256 hash as the **unique version identifier**, independent of:
- Directory names or locations
- File system paths
- Registry entries or metadata
- User-provided version strings

### Security Benefits

- **Tamper Detection**: Any modification to the key file changes its hash
- **Version Authentication**: Cryptographic proof of exact version
- **Cross-Platform**: Works consistently across Windows, macOS, Linux
- **No External Dependencies**: Self-contained verification

## Key File Concept

### What is a Key File?

A **key file** is any designated file (typically the main program or a critical component) that serves as the version's fingerprint. Properties:

- **Path**: Relative path from version root (e.g., `program.exe`, `bin/app.exe`, `core.dll`, `data.bin`)
- **Checksum**: SHA-256 hash of the file's binary content
- **Size**: File size in bytes (secondary verification)
- **File Type**: Can be any file type - executables, libraries, data files, configuration files, etc.

### Version Identification

```
Version 1.0.0:
  Key File: program.exe
  Checksum: a1b2c3d4e5f6...
  
Version 1.0.1:
  Key File: program.exe
  Checksum: x9y8z7w6v5u4...  ← Different hash = different version
```

Even if the filename is identical (`program.exe`), the **hash uniquely identifies each version**.

## Key File Selection

### Auto-Detection from Common Names

When no `--key-file` flag is provided, the generator automatically looks for a key file by checking for the following common names in the version directory root (in order):

1. `program.exe`
2. `game.exe`
3. `app.exe`
4. `main.exe`

If one of these files is found, it is used as the key file. If none are found, patch generation fails with an error:

```
Error: could not find key file (program.exe, game.exe, app.exe, or main.exe)
Hint: Use --key-file to specify a custom key file
```

### Manual Override with --key-file

The `--key-file` flag allows specifying any file as the key file, regardless of its name or type:

```bash
# Generator: specify custom key file
patch-gen.exe --from-dir "D:\releases\1.0.0" --to-dir "D:\releases\1.0.3" --key-file "myapp.exe" --output "patches"

# Applier: override key file at apply time (if renamed or moved)
patch-apply.exe --patch patch.patch --current-dir C:\MyApp --key-file "renamed.exe"
```

### Interactive Menu (Applier)

When running a self-contained executable in interactive mode, the menu offers a "Specify Custom Key File" option (option 5) that allows the user to provide a custom key file path without restarting.

## Key File Verification Workflow

### During Patch Generation

Version registration happens automatically when you generate patches. The system:

```
1. User runs: patch-gen.exe --from-dir "D:\releases\1.0.0" --to-dir "D:\releases\1.0.3" --output "patches"

2. System determines key file:
   - If --key-file provided → use that file
   - Otherwise → auto-detect by checking program.exe, game.exe, app.exe, main.exe

3. System scans source directory (1.0.0):
   - Calculate SHA-256 of key file → "a1b2c3d4e5f6..."
   - Create full file manifest

4. System scans target directory (1.0.3):
   - Calculate SHA-256 of key file → "xyz789..."
   - Create full file manifest

5. System embeds key file info in patch metadata
```

**Note**: There is no separate "register" command. Versions are registered automatically during patch generation.

### During Patch Generation

```
1. Load source manifest (1.0.0) → key_file: "program.exe" (hash: abc123...)
2. Load target manifest (1.0.3) → key_file: "program.exe" (hash: xyz789...)
3. Embed in patch:
   {
     "from_key_file": {
       "path": "program.exe",
       "checksum": "abc123...",
       "size": 14680064
     },
     "to_key_file": {
       "path": "program.exe",
       "checksum": "xyz789...",
       "size": 15728640
     }
   }
```

### During Patch Application

```
1. Read patch metadata:
   - Required key file: "program.exe"
   - Required hash: "abc123..."

2. Find key file in target directory:
   - Look at: C:\MyApp\program.exe
   - Or use --key-file override if provided

3. Calculate current hash:
   - SHA-256(C:\MyApp\program.exe) → "abc123..."

4. Compare:
   if currentHash != requiredHash {
       REJECT PATCH → "key file checksum mismatch: expected abc123..., got xyz789..."
   }

5. If match → Proceed with patch application
```

## Advanced Scenarios

### Multi-Executable Applications

**Problem**: Application has multiple executables (game.exe, launcher.exe, server.exe)

**Solution**: Use the `--key-file` flag to explicitly specify which file should serve as the key file. Only files matching the auto-detected names (`program.exe`, `game.exe`, `app.exe`, `main.exe`) will be found automatically.

**Example**:
```
MyGame/
├── launcher.exe      (2MB)
├── game.exe          (250MB) ← Auto-detected as key file
└── tools/
    └── editor.exe    (50MB)
```

### Custom Key File Paths

**Problem**: Main file is in a non-standard location or has a non-standard name

**Example**:
```
Application/
├── data/
└── bin/
    └── myapp.exe     ← Main executable here
```

**Solution**: Use the `--key-file` flag to specify the custom file name:
```bash
patch-gen.exe --from-dir "./1.0.0" --to-dir "./1.0.1" --key-file "bin/myapp.exe" --output "patches"
```

### Platform-Specific Key Files

**Problem**: Cross-platform application with different executables per platform

**Example**:
```
MyApp/
├── windows/
│   └── app.exe       ← Windows key file
├── macos/
│   └── app           ← macOS key file
└── linux/
    └── app           ← Linux key file
```

**Solution**: Generate separate patches per platform using `--from-dir` and `--to-dir`:

```powershell
# Windows patches
patch-gen.exe --from-dir "./1.0.0/windows" --to-dir "./1.0.1/windows" --output "./patches/windows"

# macOS patches
patch-gen.exe --from-dir "./1.0.0/macos" --to-dir "./1.0.1/macos" --output "./patches/macos"

# Linux patches
patch-gen.exe --from-dir "./1.0.0/linux" --to-dir "./1.0.1/linux" --output "./patches/linux"
```

**Note**: Use `--key-file` to specify the correct key file for each platform (e.g., `--key-file app.exe` for Windows, `--key-file app` for macOS/Linux).

### No Suitable Key File

**Problem**: Application has none of the auto-detected file names and no `--key-file` is specified

**Example**:
```
ScriptApp/
├── main.py           ← Entry point
├── config.json
└── modules/
```

**Solution**: Use `--key-file` to specify any file as the key file. The key file can be any file type - executables, libraries, data files, configuration files, etc.

## Edge Cases and Troubleshooting

### Case 1: No Key File Found

**Symptom**: Patch generation fails with "could not find key file (program.exe, game.exe, app.exe, or main.exe)" or manifest creation fails with "no files provided for manifest"

**Causes**:
- Directory contains only data/script files
- Key file has a non-standard name
- Key file is in a subdirectory

**Solutions**:
1. Use `--key-file` to specify the correct file name
2. Verify the file exists in the version directory

### Case 2: Multiple Suitable Candidates

**Symptom**: None of the auto-detected names match, or you want to use a specific file among many

**Example**:
```
MyApp/
├── client.exe    (150MB)
├── server.exe    (180MB)
└── admin.exe     (120MB)
```

**Solutions**:
1. Use `--key-file` to explicitly specify which file to use as the key file
2. Only `program.exe`, `game.exe`, `app.exe`, and `main.exe` are auto-detected

### Case 3: Key File Modified

**Symptom**: Patch application fails with "key file checksum mismatch: expected ..., got ..."

**Causes**:
- User modified executable (mod, crack, etc.)
- Executable corrupted during download/copy
- Wrong version installed
- Anti-virus quarantined and restored executable

**Solutions**:
1. Re-download clean version
2. Verify with original installer
3. Check anti-virus logs

### Case 4: Renamed Key File

**Symptom**: Patch fails to find key file at expected path

**Example**:
```
Original:
  MyApp/program.exe

User renamed:
  MyApp/game.exe  ← Patch looks for "program.exe"
```

**Solutions**:
- **Prevention**: Document that renaming the key file breaks patching
- **Recovery**: Restore original filename before patching
- **Override**: Use `--key-file` flag in the applier to specify the new name
- **Interactive**: Use "Specify Custom Key File" (option 5) in the interactive menu

## Performance Considerations

### Hash Calculation

- Hash calculated once during version registration
- Cached in manifest (no recalculation needed)
- Only recalculated during patch verification

## Security Considerations

### Hash Collision Resistance

**SHA-256 Properties**:
- 256-bit hash space (2^256 possible values)
- Collision probability: negligible (1 in 2^128)
- Pre-image resistance: computationally infeasible to find matching file

**Practical Security**:
- No known SHA-256 collisions for real-world executables
- Quantum computers would need ~10^20 operations (not practical)
- More secure than MD5/SHA-1 (both have known collisions)

### Tampering Detection

**Attack Scenario**: Attacker modifies executable to inject malware

**Protection**:
```
Original:
  program.exe → SHA-256: abc123... (clean)

Tampered:
  program.exe → SHA-256: xyz789... (modified)

Patch Check:
  Expected: abc123...
  Got:      xyz789...
  RESULT: PATCH REJECTED ✓
```

**Conclusion**: Any byte-level modification to key file changes hash, preventing patch application.

### Key File Spoofing

**Attack Scenario**: Attacker creates fake version with crafted key file

**Protection**:
- Patches include BOTH source and target key file hashes
- Attacker cannot generate executable with specific SHA-256 hash (pre-image resistance)
- Even if attacker matches source hash, target hash must also match

**Mitigation**: Users should obtain patches only from trusted sources.

## Best Practices

### For Developers

1. **Consistent Naming**: Use consistent key file names across versions
2. **Use Standard Names**: Name the key file `program.exe`, `game.exe`, `app.exe`, or `main.exe` for auto-detection
3. **Avoid Renaming**: Don't rename the key file between versions
4. **Document Key File**: Clearly document which file is the key file

### For Users

1. **Don't Modify**: Never modify the key file manually
2. **Restore Original**: If modified, restore original before patching
3. **Verify Source**: Obtain patches only from trusted/official sources
4. **Check Errors**: Read error messages carefully - they explain key file mismatches
5. **Custom Key File**: Use `--key-file` if the key file was renamed or moved

### For System Administrators

1. **Standard Names**: Use one of `program.exe`, `game.exe`, `app.exe`, or `main.exe` for auto-detection
2. **Manual Override**: Use `--key-file` when the key file has a non-standard name
3. **Document Overrides**: Record any custom key file specifications
4. **Verify Manifests**: Periodically verify manifests match actual installations

## Summary

The Key File System is a **critical security feature** that:

- **Prevents wrong patch application** (version mismatch detection)
- **Detects tampering** (any modification changes hash)
- **Works cross-platform** (SHA-256 is universal)
- **Requires no external dependencies** (self-contained verification)
- **Provides clear error messages** (users understand why patch failed)

By using SHA-256 hashes of designated files as version identifiers, CyberPatchMaker ensures **cryptographic-strength verification** that patches are applied to exactly the correct version.

## See Also

- [Version Management](version-management.md) - Managing versions and manifests
- [Hash Verification](hash-verification.md) - Deep dive into SHA-256 verification system
- [Architecture](architecture.md) - Overall system architecture
- [Troubleshooting](troubleshooting.md) - Common key file error scenarios
