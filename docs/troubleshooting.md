# Troubleshooting Guide

Common issues and solutions for CyberPatchMaker.

## Quick Diagnostics

Before diving into specific issues, run these quick checks:

```bash
# Check tool versions
go version          # Should be 1.24.0 or later

# Build tools
go build ./cmd/generator
go build ./cmd/applier

# Run tests
.\advanced-test.ps1  # Windows only
```

If all 58 tests pass, your setup is correct! Test data is auto-generated on first run.

---

## Generator Issues

### Error: "versions directory not found"

**Symptom:**
```
Error: versions directory not found: ./versions
```

**Cause:** Directory doesn't exist or path is wrong

**Solutions:**
1. Check directory exists: `ls ./versions` or `dir .\versions`
2. Use absolute path: `--versions-dir C:\full\path\to\versions`
3. Check current directory: `pwd`
4. Create directory: `mkdir versions`

---

### Error: "no versions found in directory"

**Symptom:**
```
Error: no versions found in directory: ./versions
```

**Cause:** Directory exists but contains no version folders

**Solutions:**
1. Check directory contents: `ls ./versions`
2. Verify version folders exist (1.0.0, 1.0.1, etc.)
3. Version folders must contain at least one file
4. Check folder naming (must be valid version numbers)

Example fix:
```bash
# Create version folders
mkdir versions/1.0.0
mkdir versions/1.0.1

# Add files to each version
cp -r /path/to/app/* versions/1.0.0/
cp -r /path/to/updated-app/* versions/1.0.1/
```

---

### Error: "key file not found in version"

**Symptom:**
```
Error: key file not found in version 1.0.0
Searched for: program.exe, game.exe, app.exe, main.exe
```

**Cause:** No recognized executable found in version folder

**Solutions:**
1. Add a key file to the version:
   - Name it `program.exe`, `game.exe`, `app.exe`, or `main.exe`
   - Must be a real file (not empty)
2. Ensure version folder isn't empty
3. Check file permissions (must be readable)

Example:
```bash
# Copy main executable as key file
cp myapp.exe versions/1.0.0/program.exe
```

---

### Error: "failed to generate binary diff"

**Symptom:**
```
Error: failed to generate binary diff for file: data/large-file.bin
```

**Cause:** File too large or binary diff algorithm failed

**Solutions:**
1. **Check disk space**: Binary diffing needs temporary space
2. **Check memory**: Very large files may need more RAM
3. **Check file corruption**: Verify both versions of the file
4. **Skip problematic files**: Use `--compression none` for faster testing

---

### Error: "failed to create patch file"

**Symptom:**
```
Error: failed to create patch file: patches/1.0.0-to-1.0.1.patch
```

**Cause:** Can't write to output directory

**Solutions:**
1. Check output directory exists: `mkdir patches`
2. Check write permissions
3. Check disk space: `df -h .` (Linux) or check drive properties (Windows)
4. Check path is valid (no invalid characters)
5. Close any programs that might have the file open

Windows:
```powershell
# Check and create output directory
New-Item -ItemType Directory -Force -Path .\patches
```

Linux:
```bash
# Check disk space
df -h .

# Create output directory with correct permissions
mkdir -p patches
chmod 755 patches
```

---

### Slow Performance

**Symptom:** Patch generation takes a very long time

**Causes & Solutions:**

**Large files:**
- Binary diffing large files (1GB+) is slow
- Use `--compression none` for faster testing
- Consider splitting large files if possible

**Many files:**
- Scanning 100,000+ files takes time
- This is normal for large projects
- Use faster storage (SSD vs HDD)

**Network paths:**
- Reading/writing over network is slow
- Copy versions to local disk first
- Use local output directory

**Compression:**
- High compression levels (4) are slow
- Use level 3 (default) or 2 for faster generation
- Use `--compression gzip` for compatibility

Performance comparison:
```
Compression None:   Fast generation,  Large patches
Compression Level 2: Medium generation, Medium patches
Compression Level 3: Medium generation, Small patches (DEFAULT)
Compression Level 4: Slow generation,   Tiny patches
```

---

## Applier Issues

### Error: "patch file not found"

**Symptom:**
```
Error: patch file not found: ./patches/1.0.0-to-1.0.1.patch
```

**Cause:** File doesn't exist or path is wrong

**Solutions:**
1. Check file exists: `ls ./patches/*.patch`
2. Use absolute path: `--patch C:\full\path\to\patch.patch`
3. Check file extension (must be `.patch`)
4. Check file wasn't moved or deleted

---

### Error: "current directory not found"

**Symptom:**
```
Error: current directory not found: ./myapp
```

**Cause:** Installation directory doesn't exist or path is wrong

**Solutions:**
1. Check directory exists: `ls ./myapp`
2. Use absolute path: `--current-dir C:\Program Files\MyApp`
3. Check spelling and case (Linux is case-sensitive)
4. Verify it's a directory (not a file)

---

### Error: "pre-verification failed: key file checksum mismatch"

**Symptom:**
```
Pre-verification failed: key file checksum mismatch
Expected: abc123...
Got:      xyz789...
```

**Cause:** Key file has been modified or installation is corrupted

**Solutions:**

**If this is expected:**
1. You might be applying the wrong patch
2. Check patch info: which version does it expect?
3. Apply correct patch for your current version

**If this is unexpected:**
1. Your installation may be corrupted
2. Re-install the original version
3. Verify installation integrity
4. Check for malware or disk errors

Example - verify you have the right version:
```bash
# Check what version you have
cat ./myapp/program.exe | grep -a "Version"

# Check what version the patch expects
./applier --patch patches/1.0.0-to-1.0.1.patch --help
# Look for "From Version: X.X.X"
```

---

### Error: "pre-verification failed: required file missing"

**Symptom:**
```
Pre-verification failed: required file missing
File: data/config.json
```

**Cause:** Installation is incomplete or files were deleted

**Solutions:**
1. Re-install the original version (cleanly)
2. Restore missing files from backup
3. Verify all files are present before patching
4. Check if files were quarantined by antivirus

---

### Error: "pre-verification failed: file checksum mismatch"

**Symptom:**
```
Pre-verification failed: file checksum mismatch
File: data/config.json
Expected: def456...
Got:      ghi789...
```

**Cause:** File has been modified

**Solutions:**

**If modification is expected (you edited configs):**
1. Backup your changes: `cp data/config.json data/config.json.backup`
2. Re-install original version
3. Apply patch
4. Re-apply your changes manually

**If modification is unexpected:**
1. Installation may be corrupted
2. Re-install the original version
3. Check for malware or disk errors
4. Verify file permissions

---

### Error: "insufficient disk space"

**Symptom:**
```
Error: insufficient disk space
Required: 500 MB
Available: 200 MB
```

**Cause:** Not enough free disk space for selective backup + patch

**Note:** The selective backup system only backs up modified/deleted files, requiring minimal extra space (e.g., ~50MB for changed files instead of 5GB full copy).

**Solutions:**
1. Free up disk space:
   - Delete temporary files
   - Empty recycle bin / trash
   - Remove old `backup.cyberpatcher` folders from previous patches
   - Move files to another drive
2. Disable automatic backup (RISKY!):
   ```bash
   # Only for testing - NOT recommended!
   ./applier --patch patch.patch --current-dir ./app --backup=false
   ```
   **WARNING:** Without backup, you cannot automatically rollback on failure!

---

### Error: "failed to create backup"

**Symptom:**
```
Error: failed to create backup directory: ./myapp/backup.cyberpatcher
```

**Cause:** Can't create `backup.cyberpatcher` subdirectory inside target directory

**Solutions:**
1. Check write permissions in target directory (`--current-dir`)
2. Check disk space (selective backup needs minimal space)
3. Check another process isn't using `backup.cyberpatcher` directory
4. Close any programs accessing the installation
5. Run as administrator (Windows) or with sudo (Linux)

Windows:
```powershell
# Run as administrator
Start-Process -FilePath ".\patch-apply.exe" -ArgumentList "--patch",".\patch.patch","--current-dir",".\myapp","--verify" -Verb RunAs
```

Linux:
```bash
# Check permissions
ls -la .

# Run with elevated privileges if needed
sudo ./applier --patch ./patch.patch --current-dir ./myapp --verify
```

---

### Error: "post-verification failed"

**Symptom:**
```
Post-verification failed: file checksum mismatch
File: program.exe
Expected: abc999...
Got:      abc123...
```

**Cause:** Patch application failed or was interrupted

**Solutions:**
1. **Manual restore** from selective backup:
   ```bash
   # Restore backed up files from mirror structure
   cp -r ./myapp/backup.cyberpatcher/* ./myapp/
   
   # Delete any files that were added (not in backup)
   # Then remove backup folder after confirming restoration
   rm -rf ./myapp/backup.cyberpatcher
   ```
   
   **Note:** The mirror structure makes rollback intuitive - just copy files back to their exact original paths.

2. **Try again** after restoration:
   - Check disk space (selective backup needs minimal space)
   - Close all programs accessing the installation
   - Verify patch file integrity (re-download if needed)
3. **Report the issue**: This shouldn't happen - may be a bug

---

### Error: "restoration failed"

**Symptom:**
```
Error: restoration failed - backup may be corrupted
```

**Cause:** Backup was corrupted or couldn't be restored

**Solutions:**
1. **Check if backup exists**: `ls ./myapp/backup.cyberpatcher`
2. **Manual restoration** from selective backup:
   ```bash
   # Restore backed up files using mirror structure
   cp -r ./myapp/backup.cyberpatcher/* ./myapp/
   
   # Delete any files that were added by the patch (not in backup)
   # Then remove backup folder
   rm -rf ./myapp/backup.cyberpatcher
   ```
   
   **Advantage:** Mirror structure makes identifying backed up files easy - they're at exact original paths within `backup.cyberpatcher`.

3. **If backup is corrupted**: Re-install original version from distribution
4. **Report the issue**: Backup corruption shouldn't happen

---

### Slow Performance

**Symptom:** Patch application takes a very long time

**Causes & Solutions:**

**Large installation (100GB+):**
- Pre-verification scans all files (can take 10+ minutes)
- This is normal for large installations
- Use faster storage (SSD vs HDD)

**Network paths:**
- Applying patches to network drives is slow
- Copy installation to local disk first
- Apply patch locally
- Copy back to network

**Verification overhead:**
- Verification uses SHA-256 to check file hashes before and after patching
- Use `--verify=false` to skip verification for faster patching
- Only disable verification if you're confident installation is valid

**Antivirus scanning:**
- Antivirus may scan every modified file
- Temporarily disable real-time scanning
- Add installation directory to exclusions

Performance comparison:
```
--verify (default: true):  Safe,    Standard (SHA-256 verification)
--verify=false:            Risky,   Faster   (no verification, not recommended)
```

---

## Permission Issues

### Windows: "Access denied"

**Symptom:**
```
Error: failed to write file: Access is denied
```

**Solutions:**
1. **Run as administrator**:
   - Right-click executable → "Run as administrator"
   - Or use PowerShell:
   ```powershell
   Start-Process -FilePath ".\patch-apply.exe" -Verb RunAs -ArgumentList "--patch",".\patch.patch","--current-dir",".\myapp","--verify"
   ```
2. **Check file/folder permissions**:
   - Right-click → Properties → Security
   - Ensure your user has "Full Control"
3. **Disable read-only**:
   - Right-click → Properties → uncheck "Read-only"
4. **Close programs**: Close any programs using the installation

---

### Linux: "Permission denied"

**Symptom:**
```
Error: failed to write file: permission denied
```

**Solutions:**
1. **Check ownership**:
   ```bash
   ls -la ./myapp
   # Should show your username, not root
   ```
2. **Fix ownership** (if owned by root):
   ```bash
   sudo chown -R $USER:$USER ./myapp
   ```
3. **Check permissions**:
   ```bash
   # Installation directory should be writable
   chmod 755 ./myapp
   chmod 644 ./myapp/*
   ```
4. **Use sudo** (if necessary):
   ```bash
   sudo ./applier --patch ./patch.patch --current-dir /opt/myapp --verify
   ```

---

## Corruption Issues

### Detecting Corruption

**Run verification manually:**
```bash
# This will scan installation and report any issues
./applier --patch ./patches/current-to-next.patch \
          --current-dir ./myapp \
          --dry-run \
          --verify
```

If pre-verification fails, your installation is corrupted.

---

### Recovering from Corruption

**Option 1: Restore from selective backup**
```bash
# If you have a backup.cyberpatcher directory
cp -r ./myapp/backup.cyberpatcher/* ./myapp/

# Delete any files added by the patch (not in backup)
# Then remove backup folder after confirming restoration
rm -rf ./myapp/backup.cyberpatcher
```

**Backup System Advantage:** Selective backup with mirror structure makes recovery intuitive:
- Backed up files are at exact original paths within `backup.cyberpatcher`
- Just copy them back to restore original state
- Added files (not in backup) need manual deletion

**Option 2: Re-install original version**
```bash
# Download original version
# Install to ./myapp

# Then apply patch
./applier --patch ./patch.patch --current-dir ./myapp --verify
```

**Option 3: Repair specific files** (mirror structure makes this easy)
```bash
# If you know which files are corrupted, use mirror paths
cp ./myapp/backup.cyberpatcher/program.exe ./myapp/
cp ./myapp/backup.cyberpatcher/data/config.json ./myapp/data/
cp ./myapp/backup.cyberpatcher/libs/core.dll ./myapp/libs/

# Mirror structure = intuitive restoration with exact paths preserved!
```

---

## Build Issues

### Error: "go: no such file or directory"

**Symptom:**
```
go: command not found
```

**Cause:** Go is not installed or not in PATH

**Solutions:**
1. **Install Go**: Download from https://golang.org/dl/
2. **Verify installation**: `go version`
3. **Add to PATH** (if installed but not found):

Windows (PowerShell):
```powershell
$env:PATH += ";C:\Go\bin"
```

Linux/macOS:
```bash
export PATH=$PATH:/usr/local/go/bin
```

---

### Error: "package not found"

**Symptom:**
```
go build: cannot find package "github.com/klauspost/compress/zstd"
```

**Cause:** Dependencies not downloaded

**Solutions:**
```bash
# Download all dependencies
go mod download

# Tidy up dependencies
go mod tidy

# Rebuild
go build ./cmd/generator
go build ./cmd/applier
```

---

### Error: "syntax error"

**Symptom:**
```
syntax error: unexpected newline, expecting comma or }
```

**Cause:** Code syntax error

**Solutions:**
1. Check the file mentioned in error message
2. Look at line number provided
3. Fix syntax error
4. Run `go fmt` to auto-format code:
   ```bash
   go fmt ./...
   ```

---

## Platform-Specific Issues

### Windows: "The system cannot find the path specified"

**Cause:** Path contains invalid characters or is too long

**Solutions:**
1. **Use shorter paths**: Move installation closer to drive root
2. **Avoid special characters**: No `<>:"|?*` in paths
3. **Use quotes**: Wrap paths in quotes if they contain spaces
   ```powershell
   .\patch-apply.exe --patch ".\patches\my patch.patch" --current-dir ".\My App" --verify
   ```

---

### Windows: UNC Paths

**Symptom:** Network paths don't work

**Solutions:**
```powershell
# Map network drive first
net use Z: \\server\share

# Use mapped drive
.\patch-apply.exe --patch Z:\patches\patch.patch --current-dir Z:\app --verify

# Or use UNC path (may be slower)
.\patch-apply.exe --patch \\server\share\patches\patch.patch --current-dir \\server\share\app --verify
```

---

### Linux: Case Sensitivity

**Symptom:** File not found but you can see it

**Cause:** Linux is case-sensitive, Windows is not

**Solutions:**
```bash
# Wrong:
./applier --patch ./Patches/patch.patch --current-dir ./MyApp

# Correct:
./applier --patch ./patches/patch.patch --current-dir ./myapp
```

---

### Linux: Executable Permissions

**Symptom:** "Permission denied" when running tools

**Solutions:**
```bash
# Make tools executable
chmod +x generator applier

# Verify
ls -la generator applier
# Should show: -rwxr-xr-x (x means executable)
```

---

### macOS: "Developer cannot be verified"

**Symptom:** macOS blocks the executable

**Solutions:**
1. **Allow in System Preferences**:
   - System Preferences → Security & Privacy
   - Click "Allow Anyway" for the blocked executable
2. **Or remove quarantine attribute**:
   ```bash
   xattr -d com.apple.quarantine generator applier
   ```

---

## Getting Help

### Information to Provide

When reporting issues, include:

1. **Platform**: Windows, Linux, macOS
2. **Go version**: `go version`
3. **Command used**: Full command with all flags
4. **Error message**: Complete error output
5. **Test results**: `.\advanced-test.ps1` output
6. **File sizes**: Size of versions and patches
7. **Disk space**: Available space on relevant drives

Example:
```
Platform: Windows 11
Go version: go1.24.0 windows/amd64
Command: .\patch-apply.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir .\myapp --verify
Error: Pre-verification failed: key file checksum mismatch
Test results: 20/20 passing
Sizes: Version 1.0.0 = 5.2GB, Version 1.0.1 = 5.3GB, Patch = 50MB
Disk space: 100GB free on C: drive
```

---

### Collecting Logs

To save output to a file:

Windows:
```powershell
.\patch-apply.exe --patch .\patch.patch --current-dir .\app --verify 2>&1 | Tee-Object -FilePath log.txt
```

Linux:
```bash
./applier --patch ./patch.patch --current-dir ./app --verify 2>&1 | tee log.txt
```

---

### Reporting Bugs

**Where to report:**
- GitHub Issues: https://github.com/cyberofficial/CyberPatchMaker/issues

**What to include:**
1. Clear description of the problem
2. Steps to reproduce
3. Expected behavior
4. Actual behavior
5. System information (see "Information to Provide" above)
6. Test results
7. Any error messages or logs

---

## Related Documentation

- [Quick Start](quick-start.md) - Getting started guide
- [Generator Guide](generator-guide.md) - Generator tool usage
- [Applier Guide](applier-guide.md) - Applier tool usage
- [Testing Guide](testing-guide.md) - Running tests
- [CLI Reference](cli-reference.md) - Command reference
