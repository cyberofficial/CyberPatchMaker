# Testing Guide

Guide to running and understanding CyberPatchMaker's comprehensive test suite.

## Overview

CyberPatchMaker includes a comprehensive PowerShell test suite with 69 tests (67 standard + 2 optional) that validate all core functionality including generation, application, verification, error handling, backup system, and advanced scenarios like multi-hop patching, bidirectional patching, downgrade testing, compression formats, and manual rollback from backup.

**Key Feature:** Test data is automatically generated on first run - no bloat files committed to the repository!

## Test Suite

### Advanced Test Suite

**File:** `advanced-test.ps1`
**Shell:** PowerShell 5.1 or later
**Platform:** Windows (PowerShell)
**Tests:** 69 comprehensive tests (67 standard + 2 optional)
**Test Data:** Auto-generated on first run (1.0.0, 1.0.1, 1.0.2)
**Optional Test Flags:** `-run1gbtest`, `-runlargefile`
**Command Visibility:** Shows the exact command-line for CLI operations (displayed in cyan)
**Bidirectional Testing:** Includes upgrade/downgrade cycle verification

## Running Tests

### Windows (PowerShell)

```powershell
# Run the comprehensive test suite
.\advanced-test.ps1

# First run output (auto-generates test data):
Checking for test versions...
Version 1.0.0 not found, creating...
  Creating version 1.0.0...
  Version 1.0.0 created (3 files, 2 directories)
Version 1.0.1 not found, creating...
  Creating version 1.0.1...
  Version 1.0.1 created (4 files, 2 directories)
Version 1.0.2 not found, creating...
  Creating version 1.0.2...
  Version 1.0.2 created (11 files, 6 directories, 3 levels deep)

Created 3 test version(s)

# Then runs all tests with command visibility:
CyberPatchMaker Advanced Test Suite
========================================

Testing: Generate complex patch (1.0.1 -> 1.0.2) with zstd
  Generating patch from 1.0.1 to 1.0.2 with zstd compression...
  Command: patch-gen.exe --versions-dir .\testdata\versions --from 1.0.1 --to 1.0.2 --output .\testdata\advanced-output\patches --compression zstd
  Patch generated (zstd): 1627 bytes
[OK] PASSED: Generate complex patch (1.0.1 -> 1.0.2) with zstd

Testing: Apply zstd patch to complex directory structure
  Copying version 1.0.1 to test-zstd...
  Applying zstd patch...
  Command: patch-apply.exe --patch .\testdata\advanced-output\patches\1.0.1-to-1.0.2.patch --current-dir .\testdata\advanced-output\test-zstd --verify
  Zstd patch applied successfully
[OK] PASSED: Apply zstd patch to complex directory structure

...
(All 67 standard tests with command visibility)
========================================
Advanced Test Results
========================================
Passed: 67
Failed: 0

[OK] All 67 advanced tests passed!

Would you like to clean up test data now? (Y/N): Y
Cleaning up test data...
✓ Test data removed successfully
```

**Note:** The test script automatically generates test versions on first run. Subsequent runs will use the existing test data unless you delete the `testdata/versions/` directory.

### Test Data Cleanup Management

The test suite includes an intelligent cleanup system to help manage test data:

**Interactive Cleanup Prompt:**
- After all tests complete, you'll be prompted to clean up test data
- **Press Y** to immediately delete the `testdata/` directory
- **Press N** to keep test data for inspection

**Auto-Delete Behavior:**
- If you choose to keep test data (N), a `.cleanup-deferred` state file is created
- On the next test run, the system automatically detects deferred cleanup
- Previous test data is automatically deleted before creating fresh test data
- This ensures clean test runs while giving you time to inspect results

**Example Workflow:**

```powershell
# First run - Choose to keep data for inspection
.\advanced-test.ps1
# ... tests complete ...
# Would you like to clean up test data now? (Y/N): N
# Test data kept in testdata/ directory (cleanup deferred to next run)

# Inspect test data manually
ls testdata/versions/
ls testdata/advanced-output/

# Second run - Auto cleanup of old data
.\advanced-test.ps1
# Previous test data detected (cleanup was deferred)...
# Removing old test data...
# ✓ Old test data removed
# ... tests run with fresh data ...
# Would you like to clean up test data now? (Y/N): Y
# Cleaning up test data...
# ✓ Test data removed successfully
```

**State File:**
- **Location:** `testdata/.cleanup-deferred`
- **Purpose:** Tracks that cleanup was deferred
- **Behavior:** Triggers auto-delete on next run
- **Cleanup:** Automatically removed when test data is deleted

---

## Test Suite Overview

The advanced test suite includes 69 comprehensive tests organized into several categories (numbers follow execution order in the script; optional tests only run with their flags):

### Build & Setup Tests (Tests 1-3)
1. **Verify executables exist** - Confirms patch-gen.exe and patch-apply.exe are present
2. **Setup advanced test environment** - Creates the test output directory structure
3. **Verify test versions exist** - Confirms all three test versions (1.0.0, 1.0.1, 1.0.2) are present

### Patch Generation Tests (Tests 4-8)
4. **Generate complex patch (1.0.1 -> 1.0.2) with zstd** - zstd compression
5. **Generate same patch with gzip compression** - gzip compression
6. **Generate same patch with no compression** - uncompressed
7. **Compare compression efficiency** - Verify all three methods produce valid patches
8. **Dry-run complex patch application** - Preview mode without changes

### Patch Application Tests (Tests 9-15)
9. **Apply zstd patch to complex directory structure** - Apply zstd-compressed patch
10. **Apply gzip patch to complex directory structure** - Apply gzip-compressed patch
11. **Apply uncompressed patch to complex directory structure** - Apply uncompressed patch
12. **Verify complex directory structure after patching** - Verify file tree integrity
13. **Verify new files were added correctly** - Check added files in subdirectories
14. **Verify modified files match version 1.0.2** - Check modified files
15. **Verify all compression methods produce identical results** - Cross-compression validation

### Multi-Hop & Downgrade Tests (Tests 16-20)
16. **Test multi-hop patching scenario** - Sequential patching (1.0.0 -> 1.0.1 -> 1.0.2)
17. **Generate downgrade patch (1.0.2 -> 1.0.1)** - Downgrade patch generation
18. **Apply downgrade patch to revert version** - Downgrade patch application
19. **Verify downgrade results match version 1.0.1** - Verify downgrade correctness
20. **Test bidirectional patching cycle** - Complete upgrade/downgrade cycle

### Error Detection Tests (Tests 21-22)
21. **Verify patch rejection for wrong source version** - Wrong version rejection
22. **Verify detection of corrupted files in source** - Corruption detection

### Backup System Tests (Tests 23-28)
23. **Verify backup creation and mirror structure** - Backup with mirror directory structure
24. **Verify selective backup (only modified/deleted files)** - Only changed files backed up
25. **Verify backup is preserved after successful patching** - Backup preserved post-patch
26. **Verify manual rollback from backup works** - Manual restore from backup
27. **Verify backup handles complex nested paths** - Deep path backup
28. **Verify deleted directories are backed up with all contents** - Deleted directory backup

### Performance & Custom Paths Tests (Tests 29-36)
29. **Verify patch generation performance** - Speed benchmarking
30. **Test custom paths mode with different directories** - Custom --from-dir/--to-dir
31. **Apply patch generated with custom paths mode** - Apply custom paths patch
32. **Test custom paths with complex nested directories** - Complex custom paths
33. **Verify version number extraction from directory names** - Version from dir name
34. **Test compression options with custom paths** - Compression with custom paths
35. **Test error handling for non-existent directories** - Error handling
36. **Verify backward compatibility with legacy --versions-dir mode** - Legacy mode

### Self-Contained Executable Tests (Tests 37-41)
37. **Test CLI self-contained executable creation** - --create-exe flag
38. **Verify CLI executable structure and header** - Magic bytes verification
39. **Test batch mode with CLI executable creation** - Batch exe creation
40. **Test 1GB bypass with large patch creation** *(optional: -run1gbtest)* - Large patches
41. **Verify CLI executables use CLI applier (not GUI)** - CLI applier verification

### File Exclusion & Silent Mode Tests (Tests 42-47)
42. **Verify backup.cyberpatcher directories are excluded** - Backup exclusion
43. **Verify .cyberignore file pattern matching** - Cyberignore patterns
44. **Verify self-contained executable --silent flag for automation** - Silent mode
45. **Verify silent mode generates timestamped log files** - Log file generation
46. **Verify generator --silent flag embeds silent mode in created executables** - Embedded silent
47. **Verify --crp flag creates reverse patches for downgrades** - Reverse patches

### Scan Cache Tests (Tests 48-54)
48. **Verify scan cache basic functionality with --savescans** - Cache creation
49. **Verify scan cache custom directory with --scandata** - Custom cache dir
50. **Verify force rescan with --rescan flag** - Force rescan
51. **Verify scan cache performance improvement** - Performance comparison
52. **Verify scan cache works with custom paths mode** - Cache + custom paths
53. **Verify scan cache file structure and content** - Cache file validation
54. **Verify scan cache invalidation on file changes** - Cache invalidation

### Simple Mode Tests (Tests 55-59)
55. **Verify patch generation with Simple Mode enabled** - SimpleMode field in patch structure
56. **Verify simplified applier interface for Simple Mode patches** - Simplified applier interface
57. **Verify complete Simple Mode workflow (generator -> applier)** - End-to-end workflow
58. **Verify Simple Mode documentation and feature completeness** - Feature validation
59. **Verify Simple Mode addresses real-world use cases** - Use case scenarios

**Note:** The `SimpleMode` field and `runSimpleMode()` function exist in the codebase but no generator code path currently sets `SimpleMode = true`. These tests verify the infrastructure exists.

### .cyberignore Advanced & Large File Tests (Tests 60-61)
60. **Verify .cyberignore absolute path pattern support** - Absolute path patterns
61. **Verify large file chunked processing and memory optimization** *(optional: -runlargefile)* - Chunked processing

### Optimization Tests (Tests 62-69)
62. **Optimization: Key file auto-detection in versions-dir mode** - Auto-detection
63. **Optimization: Key file auto-detection in from-dir/to-dir mode** - Custom paths detection
64. **Optimization: Key file override with --key-file flag** - Custom key file
65. **Optimization: zstd compression level 4** - Level mapping
66. **Optimization: gzip compression level 3** - Level mapping
67. **Optimization: Streaming LoadPatch with zstd** - Magic-byte detection
68. **Optimization: Streaming LoadPatch with gzip** - Magic-byte detection
69. **Optimization: Forward and reverse patch roundtrip** - --crp roundtrip

### Test Data Structure

**Auto-generated test versions:**

**Version 1.0.0** (Baseline - 3 files, 2 directories):

**Expected result:** All directories and files present

```
testdata/versions/1.0.0/
├── program.exe         # "Test Program v1.0.0\n"
├── data/
│   └── config.json     # JSON config file
└── libs/
    └── core.dll        # "Core Library v1.0.0\n"
```

**Version 1.0.1** (Simple update - 4 files, 2 directories):
```
testdata/versions/1.0.1/
├── program.exe         # Modified: "Test Program v1.0.1\n"
├── data/
│   └── config.json     # Modified: updated features
└── libs/
    ├── core.dll        # Modified: "Core Library v1.5.0\n"
    └── newfeature.dll  # NEW: "New Feature v1.0.0\n"
```

**Version 1.0.2** (Complex structure - 11 files, 6 directories, 3 levels deep):
```
testdata/versions/1.0.2/
├── program.exe                   # Modified: v1.0.2
├── data/
│   ├── config.json               # Modified: added features
│   ├── assets/images/            # NEW: 3 levels deep
│   │   ├── logo.png              # NEW: binary PNG file
│   │   └── icon.png              # NEW: binary PNG file
│   └── locale/                   # NEW: directory
│       └── en-US.json            # NEW: localization file
├── libs/
│   ├── core.dll                  # Modified: v2.5.0
│   ├── newfeature.dll            # Modified: v1.5.0
│   └── plugins/                  # NEW: directory
│       └── api.dll               # NEW: plugin file
└── plugins/                      # NEW: root-level directory
    ├── sample.plugin             # NEW: plugin file
    └── sample.json               # NEW: plugin config
```

---

## Understanding Test Output

Each test is a `Test-Step` invocation. The helper prints the test name, then a pass/fail line, and tracks the counts automatically:

### Success Output

```
Testing: Verify executables exist
[OK] PASSED: Verify executables exist
```

**Meaning:** Test passed all checks

---

### Failure Output

```
Testing: Verify complex directory structure after patching
[FAIL] FAILED: Verify complex directory structure after patching
  Error: <exception message thrown by the test>
```

**Meaning:** Test failed, explanation provided

**What to do:**
1. Read the error message
2. Check if test expectations are correct
3. Debug the failing component
4. Fix the issue
5. Re-run tests

---

## Manual Testing

### Testing Patch Generation

```bash
# Create test versions
mkdir -p testdata/manual/1.0.0
mkdir -p testdata/manual/1.0.1
echo "Version 1.0.0" > testdata/manual/1.0.0/app.exe
echo "Version 1.0.1" > testdata/manual/1.0.1/app.exe

# Generate patch
./patch-gen.exe --versions-dir testdata/manual \
            --new-version 1.0.1 \
            --output testdata/manual/patches

# Verify patch exists
ls testdata/manual/patches/
# Should show: 1.0.0-to-1.0.1.patch
```

---

### Testing Patch Application

```bash
# Create test application
mkdir -p testdata/manual/test-app
cp testdata/manual/1.0.0/app.exe testdata/manual/test-app/

# Apply patch
./patch-apply.exe --patch testdata/manual/patches/1.0.0-to-1.0.1.patch \
          --current-dir testdata/manual/test-app \
          --verify

# Verify version updated
cat testdata/manual/test-app/app.exe
# Should show: Version 1.0.1
```

---

### Testing Dry-Run

```bash
# Reset to version 1.0.0
rm -rf testdata/manual/test-app
mkdir -p testdata/manual/test-app
cp testdata/manual/1.0.0/app.exe testdata/manual/test-app/

# Dry-run (no changes)
./patch-apply.exe --patch testdata/manual/patches/1.0.0-to-1.0.1.patch \
          --current-dir testdata/manual/test-app \
          --dry-run

# Verify version unchanged
cat testdata/manual/test-app/app.exe
# Should still show: Version 1.0.0
```

---

### Testing Pre-Verification

```bash
# Create corrupted installation
mkdir -p testdata/manual/corrupted
echo "Corrupted Version" > testdata/manual/corrupted/app.exe

# Try to apply patch (should fail)
./patch-apply.exe --patch testdata/manual/patches/1.0.0-to-1.0.1.patch \
          --current-dir testdata/manual/corrupted \
          --verify

# Should see error: "key file checksum mismatch"
# No backup should be created
ls testdata/manual/corrupted/backup.cyberpatcher
# Should show: directory not found
```

---

### Testing Downgrade Patches

**Generating a downgrade patch:**

```bash
# Generate downgrade patch (1.0.1 -> 1.0.0)
./patch-gen.exe --versions-dir testdata/manual \
            --from 1.0.1 \
            --to 1.0.0 \
            --output testdata/manual/patches

# Verify downgrade patch exists
ls testdata/manual/patches/
# Should show: 1.0.1-to-1.0.0.patch
```

**Applying a downgrade patch:**

```bash
# Create test installation with version 1.0.1
mkdir -p testdata/manual/test-downgrade
cp testdata/manual/1.0.1/app.exe testdata/manual/test-downgrade/

# Apply downgrade patch
./patch-apply.exe --patch testdata/manual/patches/1.0.1-to-1.0.0.patch \
          --current-dir testdata/manual/test-downgrade \
          --verify

# Verify version downgraded to 1.0.0
cat testdata/manual/test-downgrade/app.exe
# Should show: Version 1.0.0
```

**Testing bidirectional patching cycle:**

```bash
# Start with version 1.0.0
mkdir -p testdata/manual/test-bidirectional
cp testdata/manual/1.0.0/app.exe testdata/manual/test-bidirectional/

# Upgrade to 1.0.1
./patch-apply.exe --patch testdata/manual/patches/1.0.0-to-1.0.1.patch \
          --current-dir testdata/manual/test-bidirectional \
          --verify

cat testdata/manual/test-bidirectional/app.exe
# Should show: Version 1.0.1

# Downgrade back to 1.0.0
./patch-apply.exe --patch testdata/manual/patches/1.0.1-to-1.0.0.patch \
          --current-dir testdata/manual/test-bidirectional \
          --verify

cat testdata/manual/test-bidirectional/app.exe
# Should show: Version 1.0.0 (back to original)
```

> **Note:** For comprehensive downgrade documentation, see the [Downgrade Guide](downgrade-guide.md).

---

## Optional Test Flags

The test suite supports optional flags for testing advanced features:

```powershell
# Test large patches >1GB (requires significant disk space and time)
.\advanced-test.ps1 -run1gbtest

# Test chunked processing for 1.5GB file (memory optimization)
.\advanced-test.ps1 -runlargefile

# Combine flags
.\advanced-test.ps1 -run1gbtest -runlargefile
```

---

## Test Data Management

### Creating New Test Versions

```bash
# Create new version directory
mkdir testdata/versions/1.0.3

# Add a key file (required - auto-detected: program.exe, game.exe, app.exe, or main.exe)
echo "Version 1.0.3" > testdata/versions/1.0.3/program.exe

# Add test files
echo "New feature data" > testdata/versions/1.0.3/feature.txt

# Generate patches
./patch-gen.exe --versions-dir testdata/versions \
            --new-version 1.0.3 \
            --output testdata/patches
```

---

### Cleaning Test Data

```bash
# Remove all test data created by the suite
# (testdata/versions, testdata/advanced-output, testdata/versions-custom, testdata/.data)
rm -rf testdata

# Remove executables (rebuild from source)
rm patch-gen.exe patch-apply.exe
```

---

## Troubleshooting Test Failures

### Build Failures

**Symptom:** Build step fails with compilation errors

**Solutions:**
1. Check Go version: `go version` (need 1.24.0+)
2. Verify code compiles: `go build ./...`
3. Check for syntax errors
4. Update dependencies: `go mod tidy`

---

### Directory Structure Failures

**Symptom:** Setup fails to find testdata

**Solutions:**
1. Run tests from project root
2. Check testdata/ directory exists
3. Check version folders exist (1.0.0, 1.0.1, 1.0.2)
4. Check program.exe exists in each version directory

---

### Generation Failures

**Symptom:** Generator command fails

**Solutions:**
1. Check generator executable exists
2. Verify testdata structure is correct
3. Check disk space
4. Run generator manually with verbose output
5. Check error messages in test output

---

### Application Failures

**Symptom:** Applier command fails

**Solutions:**
1. Check applier executable exists
2. Verify patch file exists
3. Check test-app directory is correct
4. Verify test-app has correct version
5. Run applier manually with verbose output

---

### Pre-Verification Failures

**Symptom:** Corruption detection test passes when it should report failure

**Solutions:**
1. Verify corruption step actually modifies file
2. Check that verification is enabled
3. Check that pre-verification rejects corrupted installations
4. Remember that verification runs before any backup is created (see the Backup System docs for the timing)

---

## Adding New Tests

To add new tests to the advanced test suite:

1. **Edit advanced-test.ps1**
2. **Add a new Test-Step block** following the existing pattern:
   ```powershell
   Test-Step "Verify new feature" {
       # Setup
       # ... prepare test environment ...

       # Execute + Verify (throw on failure)
       $result = & .\patch-apply.exe --patch ... --current-dir ... 2>&1
       if ($LASTEXITCODE -ne 0) {
           throw "Patch application failed: $result"
       }

       # Additional assertions - throw with a descriptive message on failure
       if (-not (Test-Path "expected\file")) {
           throw "Expected file missing"
       }
   }
   ```
   The `Test-Step` helper prints the test name, catches any thrown error, prints `[OK] PASSED` / `[FAIL] FAILED`, and updates the passed/failed/total counters automatically.
3. **Test thoroughly** before committing

---

## Related Documentation

- [Quick Start](quick-start.md) - Getting started guide
- [Generator Guide](generator-guide.md) - Generator tool usage
- [Applier Guide](applier-guide.md) - Applier tool usage
- [Troubleshooting](troubleshooting.md) - Common issues
