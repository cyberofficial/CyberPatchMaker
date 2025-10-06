# Advanced Test Script Updates

## Overview
The `advanced-test.ps1` script has been updated to test the new CLI self-contained executable creation feature and the 1GB bypass mode.

## Changes Made

### 1. New Parameter: `-1gbtest`
The script now accepts an optional `-1gbtest` switch parameter to enable 1GB bypass testing with large patches.

**Usage:**
```powershell
# Run normal tests (38 tests)
.\advanced-test.ps1

# Run with 1GB bypass test (39 tests)
.\advanced-test.ps1 -1gbtest
```

### 2. Executable Name Updates
All references to `generator.exe` and `applier.exe` have been updated to `patch-gen.exe` and `patch-apply.exe` throughout the script (Tests 1-34).

### 3. New Test Cases

#### Test 35: CLI Executable Creation
- Tests the `--create-exe` flag with CLI generator
- Verifies both patch file and executable are created
- Validates executable size (should be larger than patch)
- Uses custom paths mode (`--from-dir`, `--to-dir`)

#### Test 36: Verify CLI Executable Structure
- Reads the last 128 bytes (header) from the executable
- Verifies magic bytes ("CPMPATCH")
- Checks format version (should be 1)
- Ensures proper embedded patch structure

#### Test 37: Batch Mode with CLI Executables
- Tests batch mode with `--create-exe` flag
- Generates multiple patches to version 1.0.2
- Verifies all patches and executables are created:
  - `1.0.0-to-1.0.2.patch` + `.exe`
  - `1.0.1-to-1.0.2.patch` + `.exe`
- Displays file sizes

#### Test 38: 1GB Bypass Test (Conditional)
**Only runs when `-1gbtest` flag is set**

Creates large test versions with ~1.1GB of data:
- Creates two 550MB random binary files in version 1.0.0-large
- Creates modified versions in version 1.0.1-large
- Generates large patch with `--create-exe` flag
- Uses zstd compression level 4
- Verifies patch and executable creation
- Checks if patch exceeds 1GB (will require `--ignore1gb` flag)

**Note:** This test creates ~2.2GB of test data and may take several minutes to complete.

#### Test 39: Verify CLI Applier Usage
- Verifies that executables embed the CLI applier (not GUI)
- Calculates applier size: `exe_size - patch_size - 128_byte_header`
- CLI applier should be < 10 MB
- GUI applier would be > 40 MB
- Ensures correct applier type is used

### 4. Updated Summary
The final summary now shows:
- **38 tests** for normal mode
- **39 tests** for 1GB bypass mode (`-1gbtest`)
- Added new features to verification list:
  - CLI self-contained executable creation (`--create-exe`)
  - CLI executable structure verification (header, magic bytes)
  - Batch mode with executable creation
  - CLI applier verification (not GUI)
  - 1GB bypass mode with large patches (when using `-1gbtest`)

## Test Output Example

### Normal Mode (38 tests)
```
✓ All 38 advanced tests passed!

Advanced Features Verified:
  • Complex nested directory structures
  • Multiple compression formats (zstd, gzip, none)
  • Multi-hop patching (1.0.0 → 1.0.1 → 1.0.2)
  • Bidirectional patching (upgrade and downgrade)
  • Downgrade patches (1.0.2 → 1.0.1 rollback)
  • Complete bidirectional cycle (1.0.1 ↔ 1.0.2)
  • Wrong version detection
  • File corruption detection
  • Backup system with mirror structure
  • Selective backup (only modified/deleted files)
  • Backup preservation after successful patching
  • Manual rollback from backup
  • Backup with complex nested paths
  • Deep file path operations
  • Custom paths mode (--from-dir, --to-dir)
  • Version extraction from directory names
  • Custom paths with complex nested structures
  • All compression formats with custom paths
  • Error handling for invalid custom paths
  • Backward compatibility with legacy mode
  • CLI self-contained executable creation (--create-exe)
  • CLI executable structure verification (header, magic bytes)
  • Batch mode with executable creation
  • CLI applier verification (not GUI)
```

### 1GB Bypass Mode (39 tests)
Same as above, plus:
```
  • 1GB bypass mode with large patches (>1GB)
```

## Performance Notes

### Normal Mode
- Runs 38 tests in ~30-60 seconds (depending on system)
- Creates ~50-100 MB of test data

### 1GB Bypass Mode
- Runs 39 tests in ~5-10 minutes (depending on system)
- Creates ~2.2 GB of test data
- Test 38 involves:
  - Creating two 550MB random files
  - Copying and modifying them
  - Compressing with zstd level 4
  - Creating self-contained executable

**Recommendation:** Only use `-1gbtest` when specifically testing large patch functionality.

## Running the Tests

### Standard Testing
```powershell
# Build the executables first
.\build.ps1

# Run advanced tests
.\advanced-test.ps1
```

### 1GB Bypass Testing
```powershell
# Build the executables first
.\build.ps1

# Run advanced tests with 1GB bypass mode
.\advanced-test.ps1 -1gbtest
```

## What's Tested

### CLI Executable Creation
- Single patch with `--create-exe`
- Batch mode with `--create-exe`
- Executable structure validation
- CLI applier verification (not GUI)

### 1GB Bypass Mode (with `-1gbtest`)
- Large patch creation (>1GB)
- Self-contained executable with large patch
- Compression effectiveness
- Bypass mode flag behavior

## Future Enhancements
- Add test for running the CLI executable interactively
- Add test for `--ignore1gb` flag with applier
- Add test for dry-run mode in CLI executable
- Add test for custom directory selection in interactive mode
