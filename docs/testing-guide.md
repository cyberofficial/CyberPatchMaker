# Testing Guide

Guide to running and understanding CyberPatchMaker's comprehensive test suite.

## Overview

CyberPatchMaker includes a comprehensive test suite with 28 tests that validate all core functionality including generation, application, verification, error handling, backup system, and advanced scenarios like multi-hop patching, bidirectional patching, downgrade testing, compression formats, and automatic rollback.

**Key Feature:** Test data is automatically generated on first run - no bloat files committed to the repository!

## Test Suite

### Advanced Test Suite

**File:** `advanced-test.ps1`
**Shell:** PowerShell 5.1 or later
**Platform:** Windows (PowerShell)
**Tests:** 24 comprehensive tests
**Test Data:** Auto-generated on first run (1.0.0, 1.0.1, 1.0.2)
**Command Visibility:** Shows exact command-line for each operation (displayed in cyan)
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
Running CyberPatchMaker Advanced Test Suite
========================================

Testing: Generate complex patch (1.0.1 → 1.0.2) with zstd
  Generating patch from 1.0.1 to 1.0.2 with zstd compression...
  Command: patch-gen.exe --versions-dir .\testdata\versions --from 1.0.1 --to 1.0.2 --output .\testdata\advanced-output\patches --compression zstd
  Patch generated (zstd): 1627 bytes
✓ PASSED: Generate complex patch (1.0.1 → 1.0.2) with zstd

Testing: Apply zstd patch to complex directory structure
  Copying version 1.0.1 to test-zstd...
  Applying zstd patch...
  Command: patch-apply.exe --patch .\testdata\advanced-output\patches\1.0.1-to-1.0.2.patch --current-dir .\testdata\advanced-output\test-zstd --verify
  Zstd patch applied successfully
✓ PASSED: Apply zstd patch to complex directory structure

...
(All 28 tests with command visibility)
========================================
Advanced Test Results
========================================
Passed: 28
Failed: 0

✓ All advanced tests passed!

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
ls testdata/patches/

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

The advanced test suite includes 24 comprehensive tests organized into several categories:

### Core Functionality Tests (Tests 1-6)
1. **Build Generator Tool** - Verifies patch-gen.exe compiles correctly
2. **Build Applier Tool** - Verifies patch-apply.exe compiles correctly
3. **Auto-Generate Test Versions** - Creates test data (1.0.0, 1.0.1, 1.0.2) if missing
4. **Generate Patches (Batch)** - Tests batch patch generation from all versions to 1.0.2
5. **Patch File Verification** - Confirms all expected patch files were created
6. **Apply Simple Patch** - Tests basic patch application (1.0.0 → 1.0.1)

### Verification Tests (Tests 7-9)
7. **Verify Patch Results** - Confirms all files match expected checksums after patching
8. **Pre-Verification Test** - Ensures system detects source version before patching
9. **Wrong Version Detection** - Verifies system rejects patches for wrong versions

### Compression Tests (Tests 10-12)
10. **Zstd Compression** - Tests high-performance zstd compression
11. **Gzip Compression** - Tests universal gzip compression
12. **No Compression** - Tests uncompressed patches

### Advanced Scenario Tests (Tests 13-16)
13. **Complex Directory Structure** - Tests 3-level nested directories (1.0.1 → 1.0.2)
14. **Multi-Hop Patching** - Tests sequential patching (1.0.0 → 1.0.1 → 1.0.2)
15. **File Corruption Detection** - Verifies system detects corrupted files before patching
16. **Dry-Run Mode** - Tests preview mode without making changes

### Bidirectional Patching Tests (Tests 17-20)
17. **Generate Downgrade Patch** - Tests downgrade patch generation (1.0.2 → 1.0.1)
18. **Apply Downgrade Patch** - Tests downgrade patch application and verification
19. **Verify Downgrade Results** - Confirms downgraded version matches expected state
20. **Bidirectional Cycle** - Tests complete upgrade/downgrade cycle (1.0.1 ↔ 1.0.2)

### Error Handling & Safety Tests (Tests 21-24)
21. **Wrong Version Detection** - Verifies patches are rejected for wrong versions
22. **Backup System** - Verifies automatic backup creation and restoration
23. **Performance Benchmark** - Measures patch generation speed
24. **Verify All Operations** - Final comprehensive check of all test versions

### Test Data Structure

**Auto-generated test versions:**

**Version 1.0.0** (Baseline - 3 files, 2 directories):
3. Checks test file exists in each version

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

### Success Output

```
Test 1: Build tools... ✓ PASS
```

**Meaning:** Test passed all checks

---

### Failure Output

```
Test 5: Apply patch with verification... ✗ FAIL
  Expected exit code 0, got 1
```

**Meaning:** Test failed, explanation provided

**What to do:**
1. Read the error message
2. Check if test expectations are correct
3. Debug the failing component
4. Fix the issue
5. Re-run tests

---

### Detailed Failure Output

If a test fails, you may see additional details:

```
Test 8: Pre-verification failure... ✗ FAIL
  Expected exit code ≠ 0, got 0
  Expected file to remain corrupted, but it was restored
```

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
./generator --versions-dir testdata/manual \
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
./applier --patch testdata/manual/patches/1.0.0-to-1.0.1.patch \
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
./applier --patch testdata/manual/patches/1.0.0-to-1.0.1.patch \
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
./applier --patch testdata/manual/patches/1.0.0-to-1.0.1.patch \
          --current-dir testdata/manual/corrupted \
          --verify

# Should see error: "key file checksum mismatch"
# No backup should be created
ls testdata/manual/corrupted.backup
# Should show: directory not found
```

---

### Testing Downgrade Patches

**Generating a downgrade patch:**

```bash
# Generate downgrade patch (1.0.1 → 1.0.0)
./generator --versions-dir testdata/manual \
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
./applier --patch testdata/manual/patches/1.0.1-to-1.0.0.patch \
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
./applier --patch testdata/manual/patches/1.0.0-to-1.0.1.patch \
          --current-dir testdata/manual/test-bidirectional \
          --verify

cat testdata/manual/test-bidirectional/app.exe
# Should show: Version 1.0.1

# Downgrade back to 1.0.0
./applier --patch testdata/manual/patches/1.0.1-to-1.0.0.patch \
          --current-dir testdata/manual/test-bidirectional \
          --verify

cat testdata/manual/test-bidirectional/app.exe
# Should show: Version 1.0.0 (back to original)
```

> **Note:** For comprehensive downgrade documentation, see the [Downgrade Guide](downgrade-guide.md).

---

## Continuous Integration

### GitHub Actions Example

```yaml
name: Test CyberPatchMaker

on: [push, pull_request]

jobs:
  test-windows:
    runs-on: windows-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - name: Run advanced test suite
        run: .\advanced-test.ps1
        shell: pwsh
```

**Note:** Test data is auto-generated on first run, so no additional setup is required

---

## Test Data Management

### Creating New Test Versions

```bash
# Create new version directory
mkdir testdata/1.0.3

# Add test files
echo "Version 1.0.3" > testdata/1.0.3/test-app.txt
echo "New feature data" > testdata/1.0.3/feature.txt

# Generate patches
./generator --versions-dir testdata \
            --new-version 1.0.3 \
            --output testdata/patches
```

---

### Cleaning Test Data

```bash
# Remove generated patches
rm -rf testdata/patches/*.patch

# Remove test applications
rm -rf test-app
rm -rf testdata/1.0.0-test
rm -rf testdata/1.0.1-test

# Remove executables (rebuild from source)
rm generator applier
```

---

## Troubleshooting Test Failures

### Build Failures (Test 1)

**Symptom:** Test 1 fails with compilation errors

**Solutions:**
1. Check Go version: `go version` (need 1.21+)
2. Verify code compiles: `go build ./...`
3. Check for syntax errors
4. Update dependencies: `go mod tidy`

---

### Directory Structure Failures (Test 2)

**Symptom:** Test 2 fails to find testdata

**Solutions:**
1. Run tests from project root
2. Check testdata/ directory exists
3. Check version folders exist (1.0.0, 1.0.1, 1.0.2)
4. Check test-app.txt files exist in each version

---

### Generation Failures (Test 3)

**Symptom:** Generator command fails

**Solutions:**
1. Check generator executable exists
2. Verify testdata structure is correct
3. Check disk space
4. Run generator manually with verbose output
5. Check error messages in test output

---

### Application Failures (Test 5)

**Symptom:** Applier command fails

**Solutions:**
1. Check applier executable exists
2. Verify patch file exists
3. Check test-app directory is correct
4. Verify test-app has correct version
5. Run applier manually with verbose output

---

### Pre-Verification Test Failures (Test 8)

**Symptom:** Test 8 passes when it should detect corruption

**Solutions:**
1. Verify corruption step actually modifies file
2. Check that verification is enabled
3. Check that pre-verification rejects corrupted installations
4. Verify NO backup is created on failure
5. This is the critical backup timing test!

---

## Adding New Tests

To add new tests to the advanced test suite:

1. **Edit advanced-test.ps1**
2. **Add new test function** following the pattern:
   ```powershell
   function Test-NewFeature {
       Write-Host "Test N: New feature description... " -NoNewline
       
       # Setup
       # ... prepare test environment ...
       
       # Execute
       $result = # ... run command ...
       
       # Verify
       if ($result) {
           Write-Host "✓ PASS" -ForegroundColor Green
           return $true
       } else {
           Write-Host "✗ FAIL" -ForegroundColor Red
           Write-Host "  Error description"
           return $false
       }
   }
   ```
3. **Add test to main execution block**
4. **Update test count** in summary section
5. **Test thoroughly** before committing

---

## Related Documentation

- [Quick Start](quick-start.md) - Getting started guide
- [Generator Guide](generator-guide.md) - Generator tool usage
- [Applier Guide](applier-guide.md) - Applier tool usage
- [Troubleshooting](troubleshooting.md) - Common issues
