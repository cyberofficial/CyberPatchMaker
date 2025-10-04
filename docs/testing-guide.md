# Testing Guide

Guide to running and understanding CyberPatchMaker's test suite.

## Overview

CyberPatchMaker includes comprehensive test suites for both Windows (PowerShell) and Linux/macOS (Bash). These tests validate all core functionality including generation, application, verification, and error handling.

## Test Suites

### Windows Test Suite

**File:** `test.ps1`
**Shell:** PowerShell 5.1 or later
**Platform:** Windows only

### Linux/macOS Test Suite

**File:** `test.sh`
**Shell:** Bash 4.0 or later
**Platform:** Linux, macOS, WSL

## Running Tests

### Windows (PowerShell)

```powershell
# Run all tests
.\test.ps1

# Output:
Running CyberPatchMaker Test Suite...
==================================================
Test 1: Build tools... ✓ PASS
Test 2: Directory structure... ✓ PASS
Test 3: Generate patches (batch mode)... ✓ PASS
Test 4: Patch file existence... ✓ PASS
Test 5: Apply patch with verification... ✓ PASS
Test 6: Version verification... ✓ PASS
Test 7: Dry-run mode... ✓ PASS
Test 8: Pre-verification failure... ✓ PASS
Test 9: Compression options... ✓ PASS
Test 10: Cleanup... ✓ PASS
==================================================
All tests passed! (10/10)
```

---

### Linux/macOS (Bash)

```bash
# Make script executable (first time only)
chmod +x test.sh

# Run all tests
./test.sh

# Output:
Running CyberPatchMaker Test Suite...
==================================================
Test 1: Build tools... ✓ PASS
Test 2: Directory structure... ✓ PASS
Test 3: Generate patches (batch mode)... ✓ PASS
Test 4: Patch file existence... ✓ PASS
Test 5: Apply patch with verification... ✓ PASS
Test 6: Version verification... ✓ PASS
Test 7: Dry-run mode... ✓ PASS
Test 8: Pre-verification failure... ✓ PASS
Test 9: Compression options... ✓ PASS
Test 10: Cleanup... ✓ PASS
==================================================
All tests passed! (10/10)
```

---

## Test Descriptions

### Test 1: Build Tools

**Purpose:** Verify tools can be built from source

**What it does:**
1. Runs `go build` for generator tool
2. Runs `go build` for applier tool
3. Verifies executables were created

**Expected result:** Both tools build successfully

**Why it matters:** Ensures code compiles and is syntactically correct

---

### Test 2: Directory Structure

**Purpose:** Verify test data structure exists

**What it does:**
1. Checks `testdata/` directory exists
2. Checks version folders exist (1.0.0, 1.0.1, 1.0.2)
3. Checks test file exists in each version

**Expected result:** All directories and files present

**Why it matters:** Test data is properly set up for subsequent tests

**Test data structure:**
```
testdata/
├── 1.0.0/
│   └── test-app.txt (contains: "Version 1.0.0")
├── 1.0.1/
│   └── test-app.txt (contains: "Version 1.0.1")
└── 1.0.2/
    └── test-app.txt (contains: "Version 1.0.2")
```

---

### Test 3: Generate Patches (Batch Mode)

**Purpose:** Test patch generation in batch mode

**What it does:**
1. Creates test version directories (1.0.0, 1.0.1, 1.0.2)
2. Runs generator in batch mode: all versions → 1.0.2
3. Verifies command exits successfully

**Expected result:** Generator succeeds, exit code 0

**Why it matters:** Core patch generation functionality works

**Generated patches:**
- `1.0.0-to-1.0.2.patch`
- `1.0.1-to-1.0.2.patch`

---

### Test 4: Patch File Existence

**Purpose:** Verify patch files were created

**What it does:**
1. Checks `testdata/patches/1.0.0-to-1.0.2.patch` exists
2. Checks `testdata/patches/1.0.1-to-1.0.2.patch` exists

**Expected result:** Both patch files exist

**Why it matters:** Generator actually creates output files

---

### Test 5: Apply Patch with Verification

**Purpose:** Test patch application with full verification

**What it does:**
1. Creates test application directory from 1.0.0
2. Applies patch with `--verify` flag
3. Verifies command exits successfully

**Expected result:** Patch applies successfully, exit code 0

**Why it matters:** Core patch application functionality works

---

### Test 6: Version Verification

**Purpose:** Verify patch updated the version correctly

**What it does:**
1. Reads content of `test-app/test-app.txt`
2. Verifies it contains "Version 1.0.2"

**Expected result:** File contains expected version string

**Why it matters:** Patch actually modified files correctly

---

### Test 7: Dry-Run Mode

**Purpose:** Test dry-run preview functionality

**What it does:**
1. Creates fresh test application directory from 1.0.1
2. Runs applier with `--dry-run` flag
3. Verifies command exits successfully
4. Verifies version was NOT changed (still 1.0.1)

**Expected result:** 
- Command succeeds
- Version remains unchanged (1.0.1)

**Why it matters:** Dry-run doesn't modify files (preview only)

---

### Test 8: Pre-Verification Failure

**Purpose:** Test detection of modified installations

**What it does:**
1. Creates test application directory from 1.0.1
2. **Corrupts the test file** (modifies content)
3. Attempts to apply patch with verification
4. Verifies command FAILS (exit code ≠ 0)
5. Verifies corruption remains (file still modified)

**Expected result:**
- Command fails (corrupted installation detected)
- Backup is NOT created
- No changes made (file remains corrupted)

**Why it matters:** System properly rejects patches on corrupted installations

**This is the critical backup timing test!**

---

### Test 9: Compression Options

**Purpose:** Test different compression algorithms

**What it does:**
1. Generates patch with gzip compression
2. Verifies command exits successfully
3. Verifies patch file was created

**Expected result:** 
- Generator succeeds with gzip
- Patch file exists

**Why it matters:** Alternative compression algorithms work

---

### Test 10: Cleanup

**Purpose:** Clean up test artifacts

**What it does:**
1. Removes test application directory
2. Removes generated patches
3. Removes executables

**Expected result:** All test artifacts removed

**Why it matters:** Clean environment for next test run

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
      - name: Run tests
        run: .\test.ps1
        shell: pwsh

  test-linux:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - name: Run tests
        run: |
          chmod +x test.sh
          ./test.sh
```

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

## Writing New Tests

### Test Template

```bash
# Test N: Description
echo -n "Test N: Description... "

# Setup
# ... prepare test environment ...

# Execute
# ... run command ...
EXIT_CODE=$?

# Verify
if [ $EXIT_CODE -eq 0 ]; then
    # Additional checks
    if [ -f expected_file ]; then
        echo "✓ PASS"
    else
        echo "✗ FAIL"
        echo "  Expected file not found"
        exit 1
    fi
else
    echo "✗ FAIL"
    echo "  Command failed with exit code $EXIT_CODE"
    exit 1
fi
```

---

## Related Documentation

- [Quick Start](quick-start.md) - Getting started guide
- [Generator Guide](generator-guide.md) - Generator tool usage
- [Applier Guide](applier-guide.md) - Applier tool usage
- [Troubleshooting](troubleshooting.md) - Common issues
