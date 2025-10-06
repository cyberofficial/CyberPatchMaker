# Key File System Deep Dive

## Overview

The Key File System is a critical security feature in CyberPatchMaker that prevents users from applying patches to the wrong version or entirely different application. This document provides a comprehensive deep dive into how key files work, their detection algorithm, and advanced usage scenarios.

## Purpose and Benefits

### Why Key Files?

Without key file verification, several problematic scenarios could occur:

1. **Wrong Version**: User tries to apply 1.0.1→1.0.3 patch to version 1.0.2
2. **Different Application**: User tries to apply "GameX" patch to "GameY"
3. **Corrupted Installation**: Modified or incomplete installation accepts patch
4. **Renamed Directories**: User moves/renames folders and loses version tracking

Key files solve these problems by using a designated executable's SHA-256 hash as the **unique version identifier**, independent of:
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

## Key File Detection Algorithm

### Manual Selection Process

In the GUI, when registering a version or generating patches, you manually select the key file:

#### File Selection

```
The GUI displays all files in the version directory:
- All file types are shown (not just executables)
- Files listed from root directory (non-recursive for key file selection)
- You can choose any file: .exe, .dll, .so, .bin, .dat, .ini, etc.
```

#### Auto-Selection Logic

```
If only one file exists in the directory:
- That file is automatically selected as the key file
- Works with any file type
- Saves time for simple version structures
```

**Rationale**: Key file should be a stable identifier that exists across all versions

#### Best Practices for Key File Selection

```
Priority order (case-insensitive):
1. program.exe, program, Program
2. app.exe, app, App
3. main.exe, main, Main
4. [ApplicationName].exe (if known)
5. game.exe, game (for games)
6. launcher.exe, launcher
7. Other executables (alphabetical)
```

**Rationale**: Common naming conventions for main executables

#### Step 4: Size Filtering

```
Filter out likely non-main executables:
- Too small: < 100KB (probably utility/script)
- Too large: > 500MB (probably data/asset file)
- Installers: Names containing "install", "setup", "uninstall"
- Updaters: Names containing "update", "patch", "launcher"
```

**Rationale**: Main program typically 100KB-500MB

#### Step 5: User Confirmation

```
Present top candidate to user:
- Show detected file path
- Show file size
- Request confirmation or allow manual selection
```

### Detection Pseudo-Code

```go
func detectKeyFile(versionPath string) (KeyFileInfo, error) {
    // Step 1: Find all executables
    executables := scanForExecutables(versionPath)
    if len(executables) == 0 {
        return KeyFileInfo{}, ErrNoExecutablesFound
    }
    
    // Step 2: Score by location
    scored := scoreByLocation(executables)
    
    // Step 3: Apply naming priority
    scored = applyNamingPriority(scored)
    
    // Step 4: Filter by size
    filtered := filterBySize(scored, 100*KB, 500*MB)
    
    // Step 5: Get top candidate
    if len(filtered) == 0 {
        return KeyFileInfo{}, ErrNoSuitableCandidates
    }
    
    candidate := filtered[0]
    
    // Calculate hash
    hash, err := calculateSHA256(candidate.path)
    if err != nil {
        return KeyFileInfo{}, err
    }
    
    return KeyFileInfo{
        Path:     makeRelative(candidate.path, versionPath),
        Checksum: hash,
        Size:     candidate.size,
    }, nil
}
```

## Key File Verification Workflow

### During Version Registration

```
1. User: "Register version 1.0.3 at D:\releases\1.0.3"
2. System: Scan D:\releases\1.0.3 for executables
3. System: Detect candidate → "program.exe" (at root, 15MB)
4. System: Calculate SHA-256 → "a1b2c3d4e5f6..."
5. System: Store in manifest:
   {
     "key_file": {
       "path": "program.exe",
       "checksum": "a1b2c3d4e5f6...",
       "size": 15728640
     }
   }
```

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
   
3. Calculate current hash:
   - SHA-256(C:\MyApp\program.exe) → "abc123..."
   
4. Compare:
   if currentHash != requiredHash {
       REJECT PATCH → "Key file mismatch. This patch requires version 1.0.0"
   }
   
5. If match → Proceed with patch application
```

## Advanced Scenarios

### Multi-Executable Applications

**Problem**: Application has multiple executables (game.exe, launcher.exe, server.exe)

**Solution**: Key file is typically the **main executable**, not utility executables.

**Detection Strategy**:
1. Prioritize by naming (game.exe > launcher.exe)
2. Prioritize by size (larger = likely main program)
3. Allow user to manually select if ambiguous

**Example**:
```
MyGame/
├── launcher.exe      (2MB)   ← Utility
├── game.exe          (250MB) ← Main executable (KEY FILE)
└── tools/
    └── editor.exe    (50MB)  ← Tool
```

### Custom Key File Paths

**Problem**: Main executable is in non-standard location

**Example**:
```
Application/
├── data/
└── bin/
    └── main.exe      ← Main executable here
```

**Solution**: Detection algorithm checks `bin/` directory with high priority.

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

**Solution**: Register separate versions per platform:
```bash
# Windows version
patch-gen register --version 1.0.0-windows \
                   --path ./windows \
                   --key-file app.exe

# macOS version
patch-gen register --version 1.0.0-macos \
                   --path ./macos \
                   --key-file app

# Linux version
patch-gen register --version 1.0.0-linux \
                   --path ./linux \
                   --key-file app
```

### No Suitable Executable

**Problem**: Application is script-based (Python, JavaScript) with no executable

**Example**:
```
ScriptApp/
├── main.py           ← Entry point (not executable)
├── config.json
└── modules/
```

**Current Limitation**: CyberPatchMaker requires at least one executable file. Script-only apps need a wrapper executable.

## Edge Cases and Troubleshooting

### Case 1: No Executable Found

**Symptom**: Registration fails with "No executables found in version directory"

**Causes**:
- Directory contains only data/script files
- Executables lack execution permissions (Unix)
- Executables in deeply nested subdirectories

**Solutions**:
1. Verify executable exists and is accessible
2. Check file permissions (Unix: `chmod +x`)

### Case 2: Multiple Suitable Candidates

**Symptom**: Ambiguous key file detection - multiple executables seem equally valid

**Example**:
```
MyApp/
├── client.exe    (150MB)
├── server.exe    (180MB)
└── admin.exe     (120MB)
```

**Solutions**:
1. Review detection ranking - usually picks largest/primary
2. Manually select during registration (current UI)

### Case 3: Key File Modified

**Symptom**: Patch application fails with "Key file hash mismatch"

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
- **Prevention**: Document that renaming executables breaks patching
- **Recovery**: Restore original filename before patching
- **Alternative**: Reinstall original version, then apply patch

## Performance Considerations

### Hash Calculation Cost

**SHA-256 Performance**:
- Small executables (<10MB): < 100ms
- Medium executables (10-100MB): 100ms - 1s
- Large executables (100-500MB): 1-5s

**Optimization**:
- Hash calculated once during registration
- Cached in manifest (no recalculation needed)
- Only recalculated during patch verification

### Disk I/O Impact

**Reading Executable**:
- Buffered I/O (64KB chunks)
- Sequential reads (optimal for HDDs)
- Memory-mapped for large files (>100MB)

**Best Practices**:
- Avoid network paths for key files (slow hash calculation)
- Use SSD for version storage when possible
- Cache manifests to avoid repeated scans

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

1. **Consistent Naming**: Use consistent executable names across versions
2. **Standard Locations**: Place main executable at root or in `bin/`
3. **Avoid Renaming**: Don't rename main executable between versions
4. **Document Key File**: Clearly document which file is the main executable

### For Users

1. **Don't Modify**: Never modify/patch the main executable manually
2. **Restore Original**: If modified, restore original before patching
3. **Verify Source**: Obtain patches only from trusted/official sources
4. **Check Errors**: Read error messages carefully - they explain key file mismatches

### For System Administrators

1. **Automated Detection**: Let CyberPatchMaker detect key file automatically
2. **Manual Override**: Use manual selection only when necessary
3. **Document Overrides**: Record any custom key file specifications
4. **Verify Manifests**: Periodically verify manifests match actual installations

## Comparison with Other Systems

### Visual Patch (Indigo Rose)

**CyberPatchMaker Advantages**:
- Automatic key file detection (Visual Patch requires manual specification)
- Cross-platform hash verification (Visual Patch is Windows-only)
- Complete directory tree verification (Visual Patch checks fewer files)

### Git Delta Compression

**CyberPatchMaker Advantages**:
- Binary-specific diff algorithm (Git uses generic compression)
- Application-specific key file concept (Git uses repository hash)
- User-friendly version identification (Git uses commit SHAs)

### Windows Update

**CyberPatchMaker Advantages**:
- Works for any application (Windows Update only for OS)
- User control over versions (Windows Update is automatic)
- Explicit key file verification (Windows Update uses internal mechanisms)

## Summary

The Key File System is a **critical security feature** that:

✅ **Prevents wrong patch application** (version mismatch detection)
✅ **Detects tampering** (any modification changes hash)
✅ **Works cross-platform** (SHA-256 is universal)
✅ **Requires no external dependencies** (self-contained verification)
✅ **Provides clear error messages** (users understand why patch failed)

By using SHA-256 hashes of designated executables as version identifiers, CyberPatchMaker ensures **cryptographic-strength verification** that patches are applied to exactly the correct version.

## See Also

- [Version Management](version-management.md) - Managing versions and manifests
- [Hash Verification](hash-verification.md) - Deep dive into SHA-256 verification system
- [Architecture](architecture.md) - Overall system architecture
- [Troubleshooting](troubleshooting.md) - Common key file error scenarios
