# Advanced Test Suite - Detailed Results and Analysis

This document provides an in-depth explanation of the CyberPatchMaker Advanced Test Suite, including detailed test descriptions, command-line outputs, and analysis of what each test validates.

## Overview

**Test Suite:** `advanced-test.ps1`  
**Total Tests:** 59 comprehensive tests (58 standard + 1 optional 1GB test)  
**Test Categories:** Build, Generation, Application, Verification, Compression, Backup System, Advanced Scenarios, Custom Paths, Self-Contained Executables, File Exclusion, Silent Mode, Reverse Patches, Scan Cache, Simple Mode  
**Test Data:** Auto-generated on first run (versions 1.0.0, 1.0.1, 1.0.2)

**Recent Additions:**
- Tests 40-43: Backup exclusion, .cyberignore support, silent mode (automation), reverse patches
- Tests 44-50: Comprehensive scan cache testing (caching, custom directory, force rescan, performance, validation, invalidation)
- Tests 51-58: Simple Mode feature validation (simplified UI for end users, use case scenarios)

---

## Test Data Structure

Before running tests, the suite auto-generates three test versions with increasingly complex structures:

### Version 1.0.0 (Baseline)
**Structure:** 3 files, 2 directories
```
testdata/versions/1.0.0/
├── program.exe         # "Test Program v1.0.0\n"
├── data/
│   └── config.json     # {"version":"1.0.0","name":"TestApp","features":["basic"]}
└── libs/
    └── core.dll        # "Core Library v1.0.0\n"
```

### Version 1.0.1 (Simple Update)
**Structure:** 4 files, 2 directories  
**Changes:** Modified 3 files, added 1 new file
```
testdata/versions/1.0.1/
├── program.exe         # MODIFIED: "Test Program v1.0.1\n"
├── data/
│   └── config.json     # MODIFIED: Added "advanced" feature
└── libs/
    ├── core.dll        # MODIFIED: "Core Library v1.5.0\n"
    └── newfeature.dll  # NEW: "New Feature v1.0.0\n"
```

### Version 1.0.2 (Complex Structure)
**Structure:** 11 files, 6 directories, 3 levels deep  
**Changes:** Modified 4 files, added 7 new files, added 4 new directories
```
testdata/versions/1.0.2/
├── program.exe                   # MODIFIED: v1.0.2
├── data/
│   ├── config.json               # MODIFIED: Added "premium" feature
│   ├── assets/images/            # NEW: 3 levels deep
│   │   ├── logo.png              # NEW: Binary PNG file (108 bytes)
│   │   └── icon.png              # NEW: Binary PNG file (58 bytes)
│   └── locale/                   # NEW: Directory
│       └── en-US.json            # NEW: Localization file
├── libs/
│   ├── core.dll                  # MODIFIED: v2.5.0
│   ├── newfeature.dll            # MODIFIED: v1.5.0
│   └── plugins/                  # NEW: Directory
│       └── api.dll               # NEW: Plugin file
└── plugins/                      # NEW: Root-level directory
    ├── sample.plugin             # NEW: Plugin file
    └── sample.json               # NEW: Plugin config
```

---

## Test Results Summary

### Build Tests (Tests 1-2)

#### Test 1: Build Generator Tool
**Purpose:** Verify patch-gen.exe compiles successfully  
**Command:** `go build -o patch-gen.exe ./cmd/generator`

**Expected Output:**
```
Testing: Build generator tool... ✓ PASSED
```

**What This Tests:**
- Go build system is working correctly
- All dependencies are available
- Generator source code has no compilation errors
- Output binary is created in the correct location

**Success Criteria:**
- Exit code 0
- patch-gen.exe file exists
- File size > 0 bytes

---

#### Test 2: Build Applier Tool
**Purpose:** Verify patch-apply.exe compiles successfully  
**Command:** `go build -o patch-apply.exe ./cmd/applier`

**Expected Output:**
```
Testing: Build applier tool... ✓ PASSED
```

**What This Tests:**
- Applier source code compiles without errors
- All required packages are available
- Output binary is created successfully

**Success Criteria:**
- Exit code 0
- patch-apply.exe file exists
- File size > 0 bytes

---

### Version Verification (Test 3)

#### Test 3: Verify Test Versions Exist
**Purpose:** Confirm all three test versions are present and valid

**Expected Output:**
```
Testing: Verify test versions exist
  Version 1.0.0: 3 files
  Version 1.0.1: 4 files
  Version 1.0.2: 11 files (complex structure)
✓ PASSED: Verify test versions exist
```

**What This Tests:**
- Auto-generation created all three versions
- File counts match expected structure
- Directory paths are correct

**Success Criteria:**
- Version 1.0.0 exists with 3 files
- Version 1.0.1 exists with 4 files
- Version 1.0.2 exists with 11 files

---

### Patch Generation Tests (Tests 4-6)

#### Test 4: Generate Complex Patch (1.0.1 → 1.0.2) with zstd
**Purpose:** Test patch generation with high-performance zstd compression

**Command-Line:**
```
Command: patch-gen.exe --versions-dir .\testdata\versions --from 1.0.1 --to 1.0.2 --output .\testdata\advanced-output\patches --compression zstd
```

**Expected Output:**
```
Testing: Generate complex patch (1.0.1 → 1.0.2) with zstd
  Generating patch from 1.0.1 to 1.0.2 with zstd compression...
  Command: patch-gen.exe --versions-dir .\testdata\versions --from 1.0.1 --to 1.0.2 --output .\testdata\advanced-output\patches --compression zstd
  Patch generated (zstd): 1627 bytes
  Operations detected: Add=7, Modify=4, Delete=0
✓ PASSED: Generate complex patch (1.0.1 → 1.0.2) with zstd
```

**What This Tests:**
- Generator can compare complex nested directory structures
- Binary diff generation works correctly
- Zstd compression reduces patch size effectively
- Patch metadata includes operation counts

**Key Metrics:**
- **Patch Size:** ~1627 bytes (varies slightly)
- **Operations:** 7 additions, 4 modifications, 0 deletions
- **Compression:** Zstd provides excellent compression for binary diffs

**Technical Details:**
- **Added Files:** 7 new files across multiple directories
  - data/assets/images/logo.png
  - data/assets/images/icon.png
  - data/locale/en-US.json
  - libs/plugins/api.dll
  - plugins/sample.plugin
  - plugins/sample.json
  - And new directories created automatically

- **Modified Files:** 4 files updated with new content
  - program.exe (version string changed)
  - data/config.json (added features)
  - libs/core.dll (version bump)
  - libs/newfeature.dll (version bump)

---

#### Test 5: Generate Same Patch with gzip Compression
**Purpose:** Test alternative compression method for universal compatibility

**Command-Line:**
```
Command: patch-gen.exe --versions-dir .\testdata\versions --from 1.0.1 --to 1.0.2 --output .\testdata\advanced-output\patches-gzip --compression gzip
```

**Expected Output:**
```
Testing: Generate same patch with gzip compression
  Generating patch with gzip compression...
  Command: patch-gen.exe --versions-dir .\testdata\versions --from 1.0.1 --to 1.0.2 --output .\testdata\advanced-output\patches-gzip --compression gzip
  Patch generated (gzip): 1612 bytes
✓ PASSED: Generate same patch with gzip compression
```

**What This Tests:**
- Gzip compression integration works correctly
- Patch can be generated with alternative compression
- File operations are identical regardless of compression method

**Key Metrics:**
- **Patch Size:** ~1612 bytes
- **Comparison:** Slightly smaller than zstd for this data set
- **Compatibility:** Gzip is universally supported

---

#### Test 6: Generate Same Patch with No Compression
**Purpose:** Test uncompressed patches as a baseline

**Command-Line:**
```
Command: patch-gen.exe --versions-dir .\testdata\versions --from 1.0.1 --to 1.0.2 --output .\testdata\advanced-output\patches-none --compression none
```

**Expected Output:**
```
Testing: Generate same patch with no compression
  Generating patch with no compression...
  Command: patch-gen.exe --versions-dir .\testdata\versions --from 1.0.1 --to 1.0.2 --output .\testdata\advanced-output\patches-none --compression none
  Patch generated (none): 4712 bytes
✓ PASSED: Generate same patch with no compression
```

**What This Tests:**
- Patch system works without compression
- Provides baseline for compression effectiveness
- Useful for debugging or specialized scenarios

**Key Metrics:**
- **Patch Size:** ~4712 bytes
- **Size Increase:** ~289% larger than compressed versions
- **Use Case:** Debugging, network-optimized scenarios

---

### Compression Comparison (Test 7)

#### Test 7: Compare Compression Efficiency
**Purpose:** Analyze effectiveness of different compression algorithms

**Expected Output:**
```
Testing: Compare compression efficiency
  Compression comparison:
    zstd: 1627 bytes (100%)
    gzip: 1612 bytes (99.1%)
    none: 4712 bytes (289.6%)
  Compression is working correctly
✓ PASSED: Compare compression efficiency
```

**What This Tests:**
- All three compression methods produce valid patches
- Compression reduces patch size significantly
- Uncompressed patch is larger as expected

**Analysis:**
- **Zstd vs Gzip:** Very similar compression ratios for this dataset
- **Compression Ratio:** Both achieve ~65-66% size reduction
- **Uncompressed Baseline:** Shows compression saves ~289% in bandwidth
- **Recommendation:** Zstd for speed, gzip for compatibility

---

### Dry-Run Testing (Test 8)

#### Test 8: Dry-Run Complex Patch Application
**Purpose:** Test preview mode without making actual changes

**Command-Line:**
```
Command: patch-apply.exe --patch .\testdata\advanced-output\patches\1.0.1-to-1.0.2.patch --current-dir .\testdata\versions\1.0.1 --dry-run
```

**Expected Output:**
```
Testing: Dry-run complex patch application
  Running applier in dry-run mode...
  Command: patch-apply.exe --patch .\testdata\advanced-output\patches\1.0.1-to-1.0.2.patch --current-dir .\testdata\versions\1.0.1 --dry-run
  Dry-run completed successfully (exit code 0)
✓ PASSED: Dry-run complex patch application
```

**What This Tests:**
- Dry-run mode simulates patch application
- No actual files are modified
- Verification checks still run
- Exit code 0 indicates patch would succeed

**Use Cases:**
- Preview changes before applying
- Verify patch compatibility
- Test in production without risk
- Troubleshooting patch issues

---

### Patch Application Tests (Tests 9-11)

#### Test 9: Apply zstd Patch to Complex Directory Structure
**Purpose:** Test full patch application with zstd compression

**Command-Line:**
```
Command: patch-apply.exe --patch .\testdata\advanced-output\patches\1.0.1-to-1.0.2.patch --current-dir .\testdata\advanced-output\test-zstd --verify
```

**Expected Output:**
```
Testing: Apply zstd patch to complex directory structure
  Copying version 1.0.1 to test-zstd...
  Applying zstd patch...
  Command: patch-apply.exe --patch .\testdata\advanced-output\patches\1.0.1-to-1.0.2.patch --current-dir .\testdata\advanced-output\test-zstd --verify
  Zstd patch applied successfully
✓ PASSED: Apply zstd patch to complex directory structure
```

**What This Tests:**
- Zstd decompression works correctly
- Binary diffs are applied accurately
- New files are created in correct nested directories
- Modified files are updated properly
- SHA-256 verification passes

**Success Criteria:**
- Exit code 0
- All 11 files present after patching
- "Patch applied successfully" message appears
- No checksum mismatches

---

#### Test 10: Apply gzip Patch to Complex Directory Structure
**Purpose:** Test patch application with gzip compression

**Command-Line:**
```
Command: patch-apply.exe --patch .\testdata\advanced-output\patches-gzip\1.0.1-to-1.0.2.patch --current-dir .\testdata\advanced-output\test-gzip --verify
```

**Expected Output:**
```
Testing: Apply gzip patch to complex directory structure
  Copying version 1.0.1 to test-gzip...
  Applying gzip patch...
  Command: patch-apply.exe --patch .\testdata\advanced-output\patches-gzip\1.0.1-to-1.0.2.patch --current-dir .\testdata\advanced-output\test-gzip --verify
  Gzip patch applied successfully
✓ PASSED: Apply gzip patch to complex directory structure
```

**What This Tests:**
- Gzip decompression integration works
- Patch application is consistent across compression methods
- Verification succeeds with gzip-compressed patches

---

#### Test 11: Apply Uncompressed Patch to Complex Directory Structure
**Purpose:** Test patch application without compression

**Command-Line:**
```
Command: patch-apply.exe --patch .\testdata\advanced-output\patches-none\1.0.1-to-1.0.2.patch --current-dir .\testdata\advanced-output\test-none --verify
```

**Expected Output:**
```
Testing: Apply uncompressed patch to complex directory structure
  Copying version 1.0.1 to test-none...
  Applying uncompressed patch...
  Command: patch-apply.exe --patch .\testdata\advanced-output\patches-none\1.0.1-to-1.0.2.patch --current-dir .\testdata\advanced-output\test-none --verify
  Uncompressed patch applied successfully
✓ PASSED: Apply uncompressed patch to complex directory structure
```

**What This Tests:**
- Patch system works without compression
- Provides validation that compression is optional
- Verifies core patching logic is independent of compression

---

### Directory Structure Verification (Tests 12-14)

#### Test 12: Verify Complex Directory Structure After Patching
**Purpose:** Confirm nested directories were created correctly

**Expected Output:**
```
Testing: Verify complex directory structure after patching
  Checking nested directories...
  All nested directories created correctly
✓ PASSED: Verify complex directory structure after patching
```

**What This Tests:**
- **Directory Creation:** All 4 new directories exist
  - `data/assets/images/` (3 levels deep)
  - `data/locale/`
  - `libs/plugins/`
  - `plugins/` (root level)

**Verification Checks:**
```powershell
✓ data/assets/images/  # 3-level nested path
✓ data/locale/         # New directory
✓ libs/plugins/        # Nested in libs
✓ plugins/             # Root-level directory
```

---

#### Test 13: Verify New Files Added in Nested Paths
**Purpose:** Confirm all new files were added in correct locations

**Expected Output:**
```
Testing: Verify new files added in nested paths
  Checking new files...
  All new files added correctly
✓ PASSED: Verify new files added in nested paths
```

**What This Tests:**
- **File Additions:** All 7 new files exist in correct paths
  - `data/assets/images/logo.png` (deeply nested)
  - `data/assets/images/icon.png` (deeply nested)
  - `data/locale/en-US.json`
  - `libs/plugins/api.dll`
  - `plugins/sample.plugin`
  - `plugins/sample.json`

**Verification Checks:**
```powershell
✓ data/assets/images/logo.png  # 3-level nested binary file
✓ data/assets/images/icon.png  # 3-level nested binary file
✓ data/locale/en-US.json       # JSON localization
✓ libs/plugins/api.dll         # Plugin in nested dir
✓ plugins/sample.plugin        # Root-level plugin
✓ plugins/sample.json          # Plugin config
```

---

#### Test 14: Verify Modified Files Match Version 1.0.2
**Purpose:** Confirm modified files have exact expected content

**Expected Output:**
```
Testing: Verify modified files match version 1.0.2
  Comparing modified files with expected version...
  All modified files match expected version 1.0.2
✓ PASSED: Verify modified files match version 1.0.2
```

**What This Tests:**
- **Binary Diff Accuracy:** Modified files are byte-perfect
- **Content Verification:** Each modified file matches target version
  - `program.exe` → "Test Program v1.0.2\n"
  - `data/config.json` → Added "premium" feature
  - `libs/core.dll` → "Core Library v2.5.0\n"
  - `libs/newfeature.dll` → "New Feature v1.5.0\n"

**Technical Details:**
Uses `Compare-Object` to verify byte-for-byte accuracy between:
- Patched files in `test-zstd/`
- Reference files in `versions/1.0.2/`

Any difference would cause test failure, ensuring perfect patch application.

---

### Cross-Compression Verification (Test 15)

#### Test 15: Verify All Compression Methods Produce Identical Results
**Purpose:** Confirm compression choice doesn't affect patch results

**Expected Output:**
```
Testing: Verify all compression methods produce identical results
  Comparing results from different compression methods...
  All compression methods produced identical results (11 files each)
✓ PASSED: Verify all compression methods produce identical results
```

**What This Tests:**
- **Result Consistency:** All three compression methods produce identical output
- **File Count:** Same number of files (11) in each result directory
- **Content Accuracy:** program.exe content is identical across all three
- **Compression Independence:** Compression only affects patch size, not results

**Verification Checks:**
```powershell
✓ test-zstd:  11 files  (zstd compression)
✓ test-gzip:  11 files  (gzip compression)
✓ test-none:  11 files  (no compression)
✓ Content matches across all three methods
```

**Analysis:**
This proves that compression is a transport/storage optimization and doesn't affect the actual patch operations or results. Users can choose compression based on their needs (speed vs compatibility) without worrying about different outcomes.

---

### Multi-Hop Patching (Test 16)

#### Test 16: Test Multi-Hop Patching Scenario
**Purpose:** Verify sequential patching works correctly (1.0.0 → 1.0.1 → 1.0.2)

**Command-Line Sequence:**
```
1. Generate 1.0.0→1.0.1 patch:
   Command: patch-gen.exe --versions-dir .\testdata\versions --from 1.0.0 --to 1.0.1 --output .\testdata\advanced-output\patches

2. Apply first patch:
   Command: patch-apply.exe --patch .\testdata\advanced-output\patches\1.0.0-to-1.0.1.patch --current-dir .\testdata\advanced-output\multi-hop --verify

3. Apply second patch:
   Command: patch-apply.exe --patch .\testdata\advanced-output\patches\1.0.1-to-1.0.2.patch --current-dir .\testdata\advanced-output\multi-hop --verify
```

**Expected Output:**
```
Testing: Test multi-hop patching scenario
  Testing 1.0.0 → 1.0.1 → 1.0.2 patch chain...
  Command: patch-gen.exe --versions-dir .\testdata\versions --from 1.0.0 --to 1.0.1 --output .\testdata\advanced-output\patches
  Starting from version 1.0.0...
  Applying first patch (1.0.0 → 1.0.1)...
  Command: patch-apply.exe --patch .\testdata\advanced-output\patches\1.0.0-to-1.0.1.patch --current-dir .\testdata\advanced-output\multi-hop --verify
  Applying second patch (1.0.1 → 1.0.2)...
  Command: patch-apply.exe --patch .\testdata\advanced-output\patches\1.0.1-to-1.0.2.patch --current-dir .\testdata\advanced-output\multi-hop --verify
  Multi-hop patching successful: 1.0.0 → 1.0.1 → 1.0.2
✓ PASSED: Test multi-hop patching scenario
```

**What This Tests:**
- **Sequential Patching:** Patches can be applied in sequence
- **Version Tracking:** Key file verification works at each step
- **Cumulative Changes:** Final result matches direct 1.0.0→1.0.2 patch
- **No State Corruption:** Each patch builds on previous state correctly

**Use Cases:**
- Incremental updates when direct patch isn't available
- Controlled rollout: v1 → v2 → v3 instead of v1 → v3
- Reduced patch generation: Only need adjacent version patches
- Flexibility in deployment scenarios

**Technical Validation:**
After applying both patches sequentially, the test compares the final `program.exe` content against the reference `1.0.2/program.exe`. Perfect match confirms multi-hop patching works correctly.

---

### Error Detection Tests (Tests 17-18)

#### Test 17: Verify Patch Rejection for Wrong Source Version
**Purpose:** Confirm system detects and rejects patches applied to wrong versions

**Command-Line:**
```
Command: patch-apply.exe --patch .\testdata\advanced-output\patches\1.0.1-to-1.0.2.patch --current-dir .\testdata\advanced-output\wrong-version --verify
```

**Expected Output:**
```
Testing: Verify patch rejection for wrong source version
  Testing patch rejection (applying 1.0.1→1.0.2 to 1.0.0)...
  Command: patch-apply.exe --patch .\testdata\advanced-output\patches\1.0.1-to-1.0.2.patch --current-dir .\testdata\advanced-output\wrong-version --verify
  Patch correctly rejected for wrong source version
✓ PASSED: Verify patch rejection for wrong source version
```

**What This Tests:**
- **Key File Verification:** System checks program.exe hash before patching
- **Version Mismatch Detection:** 1.0.0's key file doesn't match 1.0.1's expected hash
- **Safe Failure:** Patch rejected with non-zero exit code
- **Error Message:** Clear error about checksum mismatch or wrong version

**Scenario:**
1. Copy version 1.0.0 to test directory
2. Try to apply 1.0.1→1.0.2 patch (requires 1.0.1 source)
3. Applier detects hash mismatch in program.exe
4. Patch rejected, no changes made

**Error Message Examples:**
- "checksum mismatch"
- "verification failed"
- "wrong version"

**Why This Matters:**
Prevents catastrophic failures from applying wrong patches. Without this check, applying a patch to the wrong version could corrupt the installation beyond repair.

---

#### Test 18: Verify Detection of Corrupted Files in Source
**Purpose:** Confirm system detects file corruption before attempting patch

**Command-Line:**
```
Command: patch-apply.exe --patch .\testdata\advanced-output\patches\1.0.1-to-1.0.2.patch --current-dir .\testdata\advanced-output\corrupted --verify
```

**Expected Output:**
```
Testing: Verify detection of corrupted files in source
  Testing corrupted file detection...
  Corrupting libs/core.dll...
  Command: patch-apply.exe --patch .\testdata\advanced-output\patches\1.0.1-to-1.0.2.patch --current-dir .\testdata\advanced-output\corrupted --verify
  Corrupted file correctly detected
✓ PASSED: Verify detection of corrupted files in source
```

**What This Tests:**
- **Pre-Patch Verification:** All required files are hashed before patching
- **Corruption Detection:** Modified file detected via SHA-256 mismatch
- **Safe Failure:** Patch rejected before any changes made
- **Error Message:** Clear indication of which file failed verification

**Scenario:**
1. Copy version 1.0.1 to test directory
2. Corrupt `libs/core.dll` by appending "CORRUPTED DATA"
3. Try to apply 1.0.1→1.0.2 patch
4. Pre-verification detects hash mismatch in core.dll
5. Patch rejected immediately

**Why This Matters:**
Ensures patches are only applied to pristine installations. Prevents:
- Applying patches to partially-updated versions
- Building on corrupted files
- Creating unpredictable results
- Data loss from bad patch application

**Error Message Examples:**
- "checksum mismatch"
- "verification failed"
- "core.dll hash does not match"

---

### Backup System Verification (Test 19)

#### Test 19: Verify Backup System Works Correctly
**Purpose:** Confirm automatic backup creation during patch application

**Command-Line:**
```
Command: patch-apply.exe --patch .\testdata\advanced-output\patches\1.0.1-to-1.0.2.patch --current-dir .\testdata\advanced-output\backup-test --verify
```

**Expected Output:**
```
Testing: Verify backup system works correctly
  Testing backup and rollback functionality...
  Command: patch-apply.exe --patch .\testdata\advanced-output\patches\1.0.1-to-1.0.2.patch --current-dir .\testdata\advanced-output\backup-test --verify
  Backup was created during patch application
  Backup system verified
✓ PASSED: Verify backup system works correctly
```

**What This Tests:**
- **Automatic Backup:** System creates backup before modifying files
- **Backup Message:** User is informed about backup creation
- **Safety Net:** Rollback possible if something goes wrong
- **Successful Patch:** Despite backup, patch completes successfully

**Backup Lifecycle:**
1. Pre-verification passes
2. Backup created (files copied to safe location)
3. Patch operations applied
4. Post-verification passes
5. Backup cleaned up (or kept based on settings)

**Use Cases:**
- Automatic rollback if patch fails
- Manual rollback if user wants to revert
- Safety during critical updates
- Disaster recovery

---

### Performance Verification (Test 20)

#### Test 20: Verify Patch Generation Performance
**Purpose:** Measure patch generation speed and ensure reasonable performance

**Command-Line:**
```
Command: patch-gen.exe --versions-dir .\testdata\versions --from 1.0.0 --to 1.0.2 --output .\testdata\advanced-output\patches
```

**Expected Output:**
```
Testing: Verify patch generation performance
  Measuring patch generation time...
  Command: patch-gen.exe --versions-dir .\testdata\versions --from 1.0.0 --to 1.0.2 --output .\testdata\advanced-output\patches
  Patch generation completed in 0.03 seconds
  Performance is acceptable
✓ PASSED: Verify patch generation performance
```

**What This Tests:**
- **Generation Speed:** Patch created in reasonable time
- **Baseline Performance:** Establishes expected timing
- **Performance Regression:** Detects if future changes slow down generation

**Performance Metrics:**
- **Expected:** < 1 second for test versions (small files)
- **Typical Result:** 0.02-0.05 seconds
- **Warning Threshold:** > 30 seconds (indicates performance issue)

**Analysis:**
The 1.0.0→1.0.2 patch is more complex than 1.0.1→1.0.2:
- More files changed (starting from baseline)
- Larger version gap
- Still completes in milliseconds

**Real-World Performance:**
For production use with larger applications (GBs):
- Scanning: 1-5 seconds per GB
- Diffing: 0.1-1 seconds per MB of changes
- Compression: 0.5-2 seconds per MB
- Total: Usually under 1 minute for 5GB applications with 50MB changes

---

### Backup System Verification (Tests 21-28)

**Purpose:** Validate the selective backup system that creates safety nets during patching operations.

**Test Coverage:**
- Test 21: Empty directory handling in version manifests
- Test 22: Incremental patching workflow
- Tests 23-27: Backup system functionality
- Test 28: Performance under realistic conditions

---

#### Test 21: Empty Directory Handling

**Objective:** Verify that version generation correctly handles empty directories in the directory tree.

**Setup:**
- Create version folder with empty directories
- Generate version manifest

**Test Actions:**
```powershell
# Generator handles empty directories automatically during patch generation
patch-gen.exe --versions-dir .\testdata\versions --from 1.0.0 --to 1.0.1 --output .\patches
```

**Expected Results:**
- ✓ Empty directories are recorded in patch manifest
- ✓ Directory structure is preserved
- ✓ No errors during patch generation
- ✓ Empty directories can be recreated during patching

**Validation:**
- Patch manifest contains directory entries
- Directory hierarchy is accurate
- Empty directory paths are preserved

---

#### Test 22: Incremental Patching

**Objective:** Test sequential patch application (version chain patching).

**Setup:**
- Apply patches in sequence: 1.0.0 → 1.0.1 → 1.0.2
- Verify each step maintains integrity

**Test Actions:**
```powershell
# Apply first patch
patch-apply.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir .\test-incremental --verify

# Apply second patch  
patch-apply.exe --patch .\patches\1.0.1-to-1.0.2.patch --current-dir .\test-incremental --verify
```

**Expected Results:**
- ✓ First patch applies successfully
- ✓ Version becomes 1.0.1 after first patch
- ✓ Second patch applies successfully
- ✓ Final version is 1.0.2
- ✓ Final state matches direct 1.0.0→1.0.2 patch result

**Validation:**
- Version tracking works at each step
- Cumulative changes are correct
- File integrity maintained throughout chain
- Final SHA-256 hashes match expected 1.0.2 state

---

#### Test 23: Backup Creation During Patching

**Objective:** Verify that backup is automatically created before applying patches.

**Setup:**
- Clean target directory with no existing backup
- Prepare patch for application

**Test Actions:**
```powershell
patch-apply.exe --patch .\patches\1.0.1-to-1.0.2.patch --current-dir .\test-backup --verify

# Check backup existence
Test-Path .\test-backup\backup.cyberpatcher
```

**Expected Results:**
- ✓ `backup.cyberpatcher/` directory is created
- ✓ Backup is created before any modifications
- ✓ User is notified about backup creation
- ✓ Patch continues after backup creation

**Validation:**
- Backup directory exists at `./test-backup/backup.cyberpatcher/`
- Backup was created before patching started
- Console output shows backup creation message
- Patch completed successfully

**Behavior:**
The applier automatically creates a backup directory before making any changes. This provides a safety net if the patch fails or if manual rollback is needed later.

---

#### Test 24: Backup Preservation After Success

**Objective:** Verify that backup is kept after successful patching for manual rollback capability.

**Setup:**
- Apply patch successfully
- Verify backup state after completion

**Test Actions:**
```powershell
patch-apply.exe --patch .\patches\1.0.1-to-1.0.2.patch --current-dir .\test-backup --verify

# After successful patch, check backup
Test-Path .\test-backup\backup.cyberpatcher
```

**Expected Results:**
- ✓ Patch applies successfully
- ✓ Backup directory still exists after success
- ✓ Backup contents remain intact
- ✓ User can manually rollback if needed

**Validation:**
- `backup.cyberpatcher/` directory still exists
- Backup files are unchanged
- User has option for manual rollback
- No automatic cleanup after success

**Design Rationale:**
Backups are preserved after successful patches because:
1. Users might discover issues later that require rollback
2. Manual inspection of changes is possible
3. No automatic cleanup reduces risk
4. Users control when to remove backups

---

#### Test 25: Selective Backup Content Verification

**Objective:** Verify that backup only contains modified and deleted files, not added files.

**Setup:**
- Apply patch that adds, modifies, and deletes files
- Inspect backup contents

**Test Actions:**
```powershell
patch-apply.exe --patch .\patches\1.0.1-to-1.0.2.patch --current-dir .\test-backup --verify

# Check backup contents
Get-ChildItem -Recurse .\test-backup\backup.cyberpatcher
```

**Expected Results:**
- ✓ Modified files ARE in backup (OpModify)
- ✓ Deleted files ARE in backup (OpDelete)
- ✓ Added files are NOT in backup (OpAdd)
- ✓ Backup is much smaller than full version

**Validation:**
- Check for modified files: `data/config.json` (if modified)
- Check for deleted files: (any deleted files should be present)
- Verify new files are NOT in backup: `plugins/sample.plugin` (new in 1.0.2)
- Backup size is minimal (only changed files)

**Patch 1.0.1→1.0.2 Operations:**
- **OpModify:** `data/config.json` (modified) → **IN BACKUP**
- **OpAdd:** `plugins/` directory + files (new) → **NOT IN BACKUP**
- **OpAdd:** `data/assets/` directory (new) → **NOT IN BACKUP**
- **OpAdd:** `data/locale/` directory (new) → **NOT IN BACKUP**
- **OpAdd:** `libs/plugins/` directory (new) → **NOT IN BACKUP**

**Result:** Backup only contains ~100 bytes (config.json), not the full ~4KB of new files.

**Why Selective Backup:**
- Added files don't need backup (they don't exist in original)
- Modified files need backup (original state preservation)
- Deleted files need backup (can be restored if needed)
- Saves disk space (only backs up what can be lost)

---

#### Test 26: Manual Rollback from Backup

**Objective:** Verify that users can manually restore files from the backup directory.

**Setup:**
- Apply patch successfully (backup exists)
- Simulate needing to rollback specific file

**Test Actions:**
```powershell
# After patching, simulate rollback need
# User discovers issue and wants to restore original config.json

# Manual rollback using mirror structure
Copy-Item .\test-backup\backup.cyberpatcher\data\config.json `
          .\test-backup\data\config.json -Force
```

**Expected Results:**
- ✓ Backup file can be copied directly to original location
- ✓ Mirror structure makes paths intuitive
- ✓ Original file is restored successfully
- ✓ No complex path mapping needed

**Validation:**
- File restored matches pre-patch state
- Hash verification confirms original content
- Mirror structure paths work 1:1
- User can restore any backed-up file easily

**Mirror Structure Advantage:**
```
Original location:  ./test-backup/data/config.json
Backup location:    ./test-backup/backup.cyberpatcher/data/config.json
                                  ^^^^^^^^^^^^^^^^^^^^^ just add this prefix
```

The mirror structure means users can intuitively find backed-up files without complex path translation.

---

#### Test 27: Backup Directory Structure Verification

**Objective:** Verify that backup preserves exact directory hierarchy (mirror structure).

**Setup:**
- Apply patch with nested directory changes
- Inspect backup directory structure

**Test Actions:**
```powershell
patch-apply.exe --patch .\patches\1.0.1-to-1.0.2.patch --current-dir .\test-backup --verify

# Check directory hierarchy
Get-ChildItem -Recurse .\test-backup\backup.cyberpatcher | Select-Object FullName
```

**Expected Results:**
- ✓ Backup preserves exact directory paths
- ✓ Nested directories are created as needed
- ✓ Empty directories are preserved if files were deleted
- ✓ Directory structure mirrors original installation

**Validation:**
- Compare directory trees:
  - Original: `./test-backup/data/config.json`
  - Backup: `./test-backup/backup.cyberpatcher/data/config.json`
- All intermediate directories exist in backup
- No path flattening or restructuring
- Mirror structure is exact

**Example Structure:**
```
test-backup/
├── data/
│   ├── config.json (modified by patch)
│   ├── assets/ (added by patch)
│   └── locale/ (added by patch)
├── libs/
│   └── plugins/ (added by patch)
├── plugins/ (added by patch)
└── backup.cyberpatcher/
    └── data/
        └── config.json (backup of original)
```

**Note:** Only `data/config.json` is backed up because it's the only modified file. New additions like `plugins/`, `assets/`, and `locale/` are NOT in backup because they're new (OpAdd operations).

---

#### Test 28: Performance Under Realistic Conditions

**Objective:** Measure backup system performance with larger, more realistic patch scenarios.

**Setup:**
- Generate patches with various sizes and complexities
- Measure backup creation and patch application time

**Test Actions:**
```powershell
# Measure patch application with backup creation
Measure-Command {
    patch-apply.exe --patch .\patches\1.0.0-to-1.0.2.patch --current-dir .\test-performance --verify
}
```

**Expected Results:**
- ✓ Backup creation adds minimal overhead (< 100ms for test data)
- ✓ Total patch time remains under performance threshold (< 30 seconds)
- ✓ Backup system doesn't significantly impact user experience
- ✓ Performance scales reasonably with patch size

**Validation:**
- Record total execution time
- Compare with non-backup patch time (should be similar)
- Verify backup overhead is acceptable
- Performance meets production requirements

**Typical Performance (Test Data):**
- Patch 1.0.1→1.0.2 (4 operations, ~4KB changes):
  - Without backup: ~0.02 seconds
  - With backup: ~0.025 seconds
  - Overhead: ~5ms (25% increase, but still instantaneous)

**Real-World Expectations (5GB Application):**
- Full scan: 1-5 seconds
- Backup creation: 0.1-0.5 seconds (only modified files)
- Patching: 5-30 seconds (depends on changes)
- Total: Usually under 1 minute

**Performance Targets:**
- ✓ Backup creation < 10% of total patch time
- ✓ Total time < 30 seconds for typical updates
- ✓ User experience remains smooth
- ✓ No noticeable lag for small patches

---

## Summary of Advanced Features Verified

### ✓ Complex Nested Directory Structures
- 3-level deep paths (`data/assets/images/`)
- Multiple new directories at various levels
- Root-level and nested directory creation
- Empty directory handling

### ✓ Multiple Compression Formats
- **Zstd:** High performance, excellent compression (~1627 bytes)
- **Gzip:** Universal compatibility, good compression (~1612 bytes)
- **None:** Baseline/debugging, uncompressed (~4712 bytes)
- All three produce identical patching results

### ✓ Multi-Hop Patching
- Sequential patching (1.0.0 → 1.0.1 → 1.0.2)
- Cumulative changes work correctly
- Version tracking at each step
- Final result matches direct patch

### ✓ Wrong Version Detection
- Key file hash verification prevents wrong patches
- Clear error messages for version mismatches
- No changes made when wrong version detected
- Safe failure with non-zero exit code

### ✓ File Corruption Detection
- Pre-patch SHA-256 verification of all files
- Corrupted files detected before patching starts
- Clear indication of which file is corrupted
- Prevents building on corrupted installations

### ✓ Backup System Functionality (Tests 21-28)
- Automatic backup creation before patching
- Selective backup (only modified/deleted files)
- Mirror directory structure for intuitive restoration
- Backup preservation after successful patches
- Manual rollback capability with simple file copying
- Empty directory handling in manifests
- Incremental patching with version chain integrity
- Performance overhead < 10% of total patch time
- User notification and safety net for critical updates

### ✓ Performance Benchmarks
- Generation speed measured and verified
- Acceptable performance confirmed (< 30 seconds threshold)
- Typical performance: 0.02-0.05 seconds for test data
- Performance regression detection

### ✓ Deep File Path Operations
- Binary files handled correctly (PNG images)
- Text files modified accurately
- JSON files parsed and validated
- DLL files treated as binary data

---

## Command-Line Pattern Reference

### Generator Command Pattern
```powershell
patch-gen.exe --versions-dir <path> --from <version> --to <version> --output <path> [--compression <type>]
```

**Options:**
- `--versions-dir`: Root directory containing all version folders
- `--from`: Source version number (e.g., "1.0.1")
- `--to`: Target version number (e.g., "1.0.2")
- `--output`: Directory where patch file will be created
- `--compression`: Optional compression type (zstd, gzip, none)

**Example:**
```powershell
patch-gen.exe --versions-dir .\testdata\versions --from 1.0.1 --to 1.0.2 --output .\patches --compression zstd
```

---

### Applier Command Pattern
```powershell
patch-apply.exe --patch <patch-file> --current-dir <directory> [--verify] [--dry-run]
```

**Options:**
- `--patch`: Path to the .patch file to apply
- `--current-dir`: Directory containing the current version to update
- `--verify`: Enable SHA-256 verification (recommended)
- `--dry-run`: Simulate patch without making changes

**Example:**
```powershell
patch-apply.exe --patch .\patches\1.0.1-to-1.0.2.patch --current-dir .\app --verify
```

---

## Test Data Cleanup

After all tests complete, the suite prompts for cleanup:

```
========================================
Test Data Cleanup
========================================

Test data is located in: .\testdata\

Would you like to clean up test data now? (Y/N):
```

**Options:**
- **Y (Yes):** Immediately deletes all test data
- **N (No):** Keeps test data for inspection, creates `.cleanup-deferred` marker

**Auto-Cleanup Behavior:**
If you choose "N", the next test run will automatically:
1. Detect the `.cleanup-deferred` marker
2. Remove old test data
3. Generate fresh test data
4. Run all tests

This allows you to inspect test results without permanently accumulating test data.

---

## Interpreting Test Failures

### Build Failures (Tests 1-2)
**Symptoms:**
```
Testing: Build generator tool... ✗ FAILED
  patch-gen.exe compilation failed
```

**Common Causes:**
- Go version too old (need 1.21+)
- Missing dependencies
- Syntax errors in code
- Build environment issues

**Solutions:**
1. Check Go version: `go version`
2. Update dependencies: `go mod tidy`
3. Verify source code compiles: `go build ./...`
4. Check error messages in build output

---

### Generation Failures (Tests 4-6)
**Symptoms:**
```
Testing: Generate complex patch... ✗ FAILED
  Generator failed with exit code 1
```

**Common Causes:**
- Version directories missing or corrupted
- Insufficient disk space
- Permission issues
- Corrupted test data

**Solutions:**
1. Verify version directories exist
2. Check disk space
3. Run with administrator privileges
4. Delete and regenerate test data

---

### Application Failures (Tests 9-11)
**Symptoms:**
```
Testing: Apply zstd patch... ✗ FAILED
  Patch application failed: verification error
```

**Common Causes:**
- Patch file corrupted
- Source version doesn't match patch requirements
- Insufficient disk space
- Permission issues on target directory

**Solutions:**
1. Regenerate patch files
2. Verify source version is correct
3. Check disk space
4. Verify write permissions on target directory

---

### Verification Failures (Tests 12-15)
**Symptoms:**
```
Testing: Verify complex directory structure... ✗ FAILED
  Nested directory data/assets/images not created
```

**Common Causes:**
- Patch application incomplete
- Directory creation failed
- Path too long (Windows MAX_PATH issue)

**Solutions:**
1. Check patch application completed successfully
2. Verify no errors during patch application
3. On Windows, enable long path support

---

## Scan Cache Tests (Tests 44-50)

### Test 44: Scan Cache - Basic Functionality
**Purpose:** Verify scan cache creation and loading with `--savescans` flag  
**Commands:**
```bash
# First generation: Create cache
patch-gen --versions-dir ./versions --from 1.0.0 --to 1.0.1 --output ./patches --savescans

# Second generation: Load from cache
patch-gen --versions-dir ./versions --from 1.0.0 --to 1.0.1 --output ./patches --savescans
```

**Expected Output:**
```
Testing: Verify scan cache basic functionality with --savescans
  ✓ Cache directory created: .data
  ✓ Cache files created: 2 files
  ✓ Cache save message found in output
  ✓ Cache hit: Loaded scan from cache
✓ PASSED
```

**What This Tests:**
- Cache directory creation (`.data/`)
- Cache file generation for each version
- Cache save messages in generator output
- Cache hit detection on subsequent runs
- Cache provides instant version loading

---

### Test 45: Scan Cache - Custom Directory
**Purpose:** Verify `--scandata` flag for custom cache location  
**Command:**
```bash
patch-gen --versions-dir ./versions --from 1.0.0 --to 1.0.1 --output ./patches --savescans --scandata ./testdata/my-custom-cache
```

**Expected Output:**
```
Testing: Verify scan cache custom directory with --scandata
  ✓ Custom cache directory created: .\testdata\my-custom-cache
  ✓ Cache files created in custom directory: 2 files
✓ PASSED
```

**What This Tests:**
- Custom cache directory creation
- Cache files stored in specified location
- Useful for shared cache or specific storage

---

### Test 46: Scan Cache - Force Rescan
**Purpose:** Verify `--rescan` flag bypasses cache  
**Commands:**
```bash
# Create initial cache
patch-gen --versions-dir ./versions --from 1.0.0 --to 1.0.1 --output ./patches --savescans

# Force rescan with --rescan flag
patch-gen --versions-dir ./versions --from 1.0.0 --to 1.0.1 --output ./patches --savescans --rescan
```

**Expected Output:**
```
Testing: Verify force rescan with --rescan flag
  ✓ Force rescan mode enabled
  ✓ Cache not loaded (rescanned as expected)
  ✓ Cache files updated with fresh scan
✓ PASSED
```

**What This Tests:**
- `--rescan` flag forces fresh directory scan
- Cache is bypassed even when available
- Cache files are updated with new scan data
- Useful when files changed and need to rebuild cache

---

### Test 47: Scan Cache - Performance Benefit
**Purpose:** Measure performance improvement from caching  
**Commands:**
```bash
# First scan (no cache)
Measure-Command { patch-gen ... }

# Second scan (with cache)
Measure-Command { patch-gen ... --savescans }
```

**Expected Output:**
```
Testing: Verify scan cache performance improvement
  First scan time: 37 ms
  Second scan time: 30 ms
  Cache impact: saved 7 ms
  ✓ Cache used in second run
✓ PASSED
```

**What This Tests:**
- Cache provides measurable performance improvement
- Small projects: 5-10ms saved
- Large projects: 15+ minutes → instant (massive improvement)
- Example: War Thunder (34,650 files) - 15 min → instant

---

### Test 48: Scan Cache - Custom Paths Mode
**Purpose:** Verify cache works with `--from-dir` and `--to-dir`  
**Commands:**
```bash
# Custom paths with cache
patch-gen --from-dir ./testdata/custom-paths/1.0.0 --to-dir ./testdata/custom-paths/1.0.1 --output ./patches --savescans

# Load from cache
patch-gen --from-dir ./testdata/custom-paths/1.0.0 --to-dir ./testdata/custom-paths/1.0.1 --output ./patches --savescans
```

**Expected Output:**
```
Testing: Verify scan cache works with custom paths mode
  ✓ Cache created with custom paths mode
  ✓ Cache hit with custom paths
✓ PASSED
```

**What This Tests:**
- Cache compatibility with custom paths mode
- Cache matches directories regardless of access mode
- Location hash ensures unique cache per path

---

### Test 49: Scan Cache - File Structure Validation
**Purpose:** Verify cache JSON structure and content  
**Validation:**
```powershell
$cacheFile = Get-Content ".data/scan_1.0.0_*.json" | ConvertFrom-Json
# Verify: version, location, manifest, key_file, cached_at
```

**Expected Output:**
```
Testing: Verify scan cache file structure and content
  ✓ Cache has version field: 1.0.0
  ✓ Cache has location field
  ✓ Cache has manifest field
  ✓ Cache manifest has files array: 3 files
  ✓ Cache file entries have complete metadata (path, checksum, size)
  ✓ Cache has key file info
  ✓ Cache has creation timestamp
✓ PASSED
```

**What This Tests:**
- Cache is valid JSON format
- Contains version, location, manifest
- Manifest has complete file metadata (path, checksum, size)
- Includes key file hash for validation
- Has `cached_at` timestamp field
- Cache file naming: `scan_<version>_<hash>.json`

---

### Test 50: Scan Cache - Invalidation on Changes
**Purpose:** Verify cache invalidation when key file changes  
**Commands:**
```bash
# Create cache
patch-gen --versions-dir ./versions --from 1.0.0 --to 1.0.1 --output ./patches --savescans

# Modify key file (simulate version change)
echo "modified" >> ./testdata/versions/1.0.0/program.exe

# Try to use cache (should detect invalidation)
patch-gen --versions-dir ./versions --from 1.0.0 --to 1.0.1 --output ./patches --savescans
```

**Expected Output:**
```
Testing: Verify scan cache invalidation on file changes
  ✓ Initial cache created
  ✓ Cache invalidation detected
✓ PASSED
```

**What This Tests:**
- Cache validates key file hash before use
- Falls back to fresh scan if validation fails
- Prevents using stale/incorrect cache data
- Ensures cache accuracy and reliability

---

## Custom Paths and Advanced Features Tests (Tests 29-43)

### Test 29: Apply Custom Paths Patch
**Purpose:** Verify patch application works with custom directory paths (--from-dir, --to-dir)

**Command-Line:**
```
Command: patch-gen.exe --from-dir .\testdata\advanced-output\custom-paths\1.0.1 --to-dir .\testdata\advanced-output\custom-paths\1.0.2 --output .\testdata\advanced-output\patches --compression zstd
Command: patch-apply.exe --patch .\testdata\advanced-output\patches\1.0.1-to-1.0.2.patch --current-dir .\testdata\advanced-output\custom-apply --verify
```

**Expected Output:**
```
Testing: Apply custom paths patch
  Generating patch from custom directories...
  Command: patch-gen.exe --from-dir .\testdata\advanced-output\custom-paths\1.0.1 --to-dir .\testdata\advanced-output\custom-paths\1.0.2 --output .\testdata\advanced-output\patches --compression zstd
  Applying custom paths patch...
  Command: patch-apply.exe --patch .\testdata\advanced-output\patches\1.0.1-to-1.0.2.patch --current-dir .\testdata\advanced-output\custom-apply --verify
  Custom paths patch applied successfully
✓ PASSED: Apply custom paths patch
```

**What This Tests:**
- Generator can work with arbitrary directory paths (--from-dir, --to-dir)
- Version numbers are extracted from directory names automatically
- Patch application works with custom source directories
- No dependency on versions/ subdirectory structure

**Use Cases:**
- Patching applications installed in custom locations
- Working with non-standard directory structures
- Integration with existing deployment workflows
- Flexibility for different installation patterns

---

#### Test 30: Custom Paths with Complex Nested Structure
**Purpose:** Verify custom paths mode handles complex directory hierarchies

**Command-Line:**
```
Command: patch-gen.exe --from-dir .\testdata\advanced-output\custom-complex\1.0.1 --to-dir .\testdata\advanced-output\custom-complex\1.0.2 --output .\testdata\advanced-output\patches --compression zstd
Command: patch-apply.exe --patch .\testdata\advanced-output\patches\1.0.1-to-1.0.2.patch --current-dir .\testdata\advanced-output\custom-complex-apply --verify
```

**Expected Output:**
```
Testing: Custom paths with complex nested structure
  Creating complex nested directory structure...
  Generating patch with custom paths and complex structure...
  Command: patch-gen.exe --from-dir .\testdata\advanced-output\custom-complex\1.0.1 --to-dir .\testdata\advanced-output\custom-complex\1.0.2 --output .\testdata\advanced-output\patches --compression zstd
  Applying patch to complex custom structure...
  Command: patch-apply.exe --patch .\testdata\advanced-output\patches\1.0.1-to-1.0.2.patch --current-dir .\testdata\advanced-output\custom-complex-apply --verify
  Complex custom paths patch applied successfully
✓ PASSED: Custom paths with complex nested structure
```

**What This Tests:**
- Deep directory hierarchies (3+ levels) work with custom paths
- Complex nested structures are handled correctly
- Path resolution works with arbitrary directory depths
- No limitations on directory complexity

**Technical Details:**
Creates structures like:
```
custom-complex/1.0.1/
├── app/
│   ├── bin/
│   │   ├── core/
│   │   │   └── modules/
│   │   │       └── plugins/
│   │   │           └── extensions/
│   │   │               └── advanced.dll
│   │   └── main.exe
│   └── config/
│       └── settings.json
└── data/
    └── user/
        └── profiles/
            └── default.json
```

---

#### Test 31: Custom Paths - Version Number Extraction
**Purpose:** Verify automatic version number extraction from custom directory names

**Expected Output:**
```
Testing: Custom paths - version number extraction
  Testing version extraction from various directory patterns...
  ✓ Extracted version '2.1.3' from 'MyApp_v2.1.3_release'
  ✓ Extracted version '1.0.0' from 'version-1.0.0'
  ✓ Extracted version '3.2.1-beta' from 'app-3.2.1-beta'
  ✓ Extracted version '1.5.0' from 'software_1.5.0_final'
  Version extraction working correctly
✓ PASSED: Custom paths - version number extraction
```

**What This Tests:**
- Regex-based version extraction from directory names
- Handles various naming patterns (underscores, hyphens, dots)
- Supports semantic versioning with pre-release tags
- Robust parsing of version strings in directory names

**Supported Patterns:**
- `MyApp_v2.1.3_release` → `2.1.3`
- `version-1.0.0` → `1.0.0`
- `app-3.2.1-beta` → `3.2.1-beta`
- `software_1.5.0_final` → `1.5.0`

---

#### Test 32: Custom Paths - Compression Options
**Purpose:** Verify all compression formats work with custom paths mode

**Command-Line:**
```
Command: patch-gen.exe --from-dir .\testdata\advanced-output\custom-compress\1.0.1 --to-dir .\testdata\advanced-output\custom-compress\1.0.2 --output .\testdata\advanced-output\patches --compression zstd
Command: patch-gen.exe --from-dir .\testdata\advanced-output\custom-compress\1.0.1 --to-dir .\testdata\advanced-output\custom-compress\1.0.2 --output .\testdata\advanced-output\patches-gzip --compression gzip
Command: patch-gen.exe --from-dir .\testdata\advanced-output\custom-compress\1.0.1 --to-dir .\testdata\advanced-output\custom-compress\1.0.2 --output .\testdata\advanced-output\patches-none --compression none
```

**Expected Output:**
```
Testing: Custom paths - compression options
  Testing all compression formats with custom paths...
  ✓ Zstd compression with custom paths: 1627 bytes
  ✓ Gzip compression with custom paths: 1612 bytes
  ✓ No compression with custom paths: 4712 bytes
  All compression formats work with custom paths
✓ PASSED: Custom paths - compression options
```

**What This Tests:**
- All compression options (zstd, gzip, none) work with --from-dir/--to-dir
- Compression ratios are consistent with standard mode
- Custom paths don't affect compression effectiveness
- Patch generation succeeds with all compression types

---

#### Test 33: Custom Paths - Error Handling (Non-existent Directory)
**Purpose:** Verify proper error handling for invalid custom directory paths

**Command-Line:**
```
Command: patch-gen.exe --from-dir .\non-existent-source --to-dir .\non-existent-target --output .\testdata\advanced-output\patches
```

**Expected Output:**
```
Testing: Custom paths - error handling (non-existent directory)
  Testing error handling for invalid custom paths...
  Command: patch-gen.exe --from-dir .\non-existent-source --to-dir .\non-existent-target --output .\testdata\advanced-output\patches
  Generator correctly failed with invalid directory error
✓ PASSED: Custom paths - error handling (non-existent directory)
```

**What This Tests:**
- Clear error messages for non-existent directories
- Graceful failure instead of crashes
- User-friendly error reporting
- Prevents silent failures with invalid paths

**Error Messages:**
- "Source directory does not exist"
- "Target directory does not exist"
- "Cannot access directory: permission denied"

---

#### Test 34: Backward Compatibility - Legacy Mode Still Works
**Purpose:** Verify traditional versions/ directory structure still works

**Command-Line:**
```
Command: patch-gen.exe --versions-dir .\testdata\versions --from 1.0.1 --to 1.0.2 --output .\testdata\advanced-output\patches
Command: patch-apply.exe --patch .\testdata\advanced-output\patches\1.0.1-to-1.0.2.patch --current-dir .\testdata\advanced-output\legacy-test --verify
```

**Expected Output:**
```
Testing: Backward compatibility - legacy mode still works
  Testing traditional versions directory structure...
  Command: patch-gen.exe --versions-dir .\testdata\versions --from 1.0.1 --to 1.0.2 --output .\testdata\advanced-output\patches
  Applying patch with legacy mode...
  Command: patch-apply.exe --patch .\testdata\advanced-output\patches\1.0.1-to-1.0.2.patch --current-dir .\testdata\advanced-output\legacy-test --verify
  Legacy mode patch applied successfully
✓ PASSED: Backward compatibility - legacy mode still works
```

**What This Tests:**
- Original --versions-dir mode still functions
- Existing workflows continue to work
- No breaking changes to legacy usage
- Backward compatibility maintained

---

#### Test 35: CLI Executable Creation
**Purpose:** Verify --create-exe flag creates self-contained executable patches

**Command-Line:**
```
Command: patch-gen.exe --versions-dir .\testdata\versions --from 1.0.1 --to 1.0.2 --output .\testdata\advanced-output\executables --create-exe --compression zstd
```

**Expected Output:**
```
Testing: CLI executable creation
  Creating self-contained executable patch...
  Command: patch-gen.exe --versions-dir .\testdata\versions --from 1.0.1 --to 1.0.2 --output .\testdata\advanced-output\executables --create-exe --compression zstd
  Self-contained executable created: 1.0.1-to-1.0.2.exe
✓ PASSED: CLI executable creation
```

**What This Tests:**
- --create-exe flag creates executable wrapper
- Executable contains embedded patch data
- CLI generator can create self-contained patches
- Output file has .exe extension

**Executable Features:**
- Self-contained (no external dependencies)
- Embeds patch data directly
- Can be distributed as single file
- Works on target systems without CyberPatchMaker installation

---

#### Test 36: Verify CLI Executable Structure
**Purpose:** Verify self-contained executable has correct internal structure

**Expected Output:**
```
Testing: Verify CLI executable structure
  Analyzing executable structure...
  ✓ Executable size: 18543 bytes
  ✓ Magic bytes detected: CYBERPATCH
  ✓ Header structure valid
  ✓ Embedded patch data found
  ✓ Executable is self-contained
✓ PASSED: Verify CLI executable structure
```

**What This Tests:**
- Executable has correct header format (128 bytes)
- Contains "CYBERPATCH" magic bytes
- Embedded patch data is intact
- File size includes both executable code and patch data

**Internal Structure:**
```
Offset 0-127: Executable header with metadata
Offset 128+: Embedded patch data (compressed)
Magic Bytes: "CYBERPATCH" at known offset
```

---

#### Test 37: Batch Mode with CLI Executables
**Purpose:** Verify batch creation of multiple executable patches

**Command-Line:**
```
Command: patch-gen.exe --versions-dir .\testdata\versions --from 1.0.0 --to 1.0.1 --output .\testdata\advanced-output\batch-exe --create-exe
Command: patch-gen.exe --versions-dir .\testdata\versions --from 1.0.1 --to 1.0.2 --output .\testdata\advanced-output\batch-exe --create-exe
```

**Expected Output:**
```
Testing: Batch mode with CLI executables
  Creating multiple executable patches...
  ✓ Executable 1: 1.0.0-to-1.0.1.exe created
  ✓ Executable 2: 1.0.1-to-1.0.2.exe created
  ✓ Both executables have valid structure
✓ PASSED: Batch mode with CLI executables
```

**What This Tests:**
- Multiple executables can be created in sequence
- Each executable is independent and valid
- Batch processing works correctly
- No conflicts between multiple executable generations

---

#### Test 38: 1GB Bypass Test (only if -run1gbtest flag is set)
**Purpose:** Verify handling of large patches (>1GB) with bypass mode

**Command-Line:**
```
Command: patch-gen.exe --versions-dir .\testdata\large-versions --from 1.0.0 --to 2.0.0 --output .\testdata\advanced-output\large-patches --create-exe
```

**Expected Output:**
```
Testing: 1GB bypass test
  Creating large patch (>1GB)...
  ✓ Large patch generated successfully
  ✓ Bypass mode activated for large files
  ✓ Executable created despite size
✓ PASSED: 1GB bypass test
```

**What This Tests:**
- Large patch handling (>1GB limit)
- Bypass mode for oversized patches
- Executable creation works with large embedded data
- Performance with large datasets

**Note:** This test only runs when -run1gbtest flag is provided to advanced-test.ps1

---

#### Test 39: Verify All Executables Use CLI Applier
**Purpose:** Confirm all generated executables use the CLI applier internally

**Expected Output:**
```
Testing: Verify all executables use CLI applier
  Testing executable applier selection...
  ✓ Executable uses CLI applier (not GUI)
  ✓ No GUI dependencies in executable
  ✓ Command-line interface available
✓ PASSED: Verify all executables use CLI applier
```

**What This Tests:**
- Executables embed CLI applier code
- No GUI dependencies in self-contained executables
- CLI interface is available when running executable
- Consistent behavior across all executable patches

---

#### Test 40: Backup Directory Exclusion
**Purpose:** Verify backup.cyberpatcher directory is excluded from patches

**Command-Line:**
```
Command: patch-gen.exe --versions-dir .\testdata\versions --from 1.0.1 --to 1.0.2 --output .\testdata\advanced-output\patches
```

**Expected Output:**
```
Testing: Backup directory exclusion
  Testing backup.cyberpatcher exclusion...
  ✓ Backup directory exists in source
  ✓ Backup directory not included in patch
  ✓ Patch operations exclude backup.cyberpatcher
✓ PASSED: Backup directory exclusion
```

**What This Tests:**
- backup.cyberpatcher directories are ignored during scanning
- Backup files don't get included in patches
- Prevents recursive backup inclusion
- Clean patch generation without backup pollution

**Why This Matters:**
- Prevents infinite recursion (backups containing backups)
- Keeps patches focused on actual application changes
- Reduces patch size by excluding temporary files
- Maintains clean version differences

---

#### Test 41: .cyberignore File Support
**Purpose:** Verify .cyberignore file excludes specified files/patterns from patches

**Setup:** Create .cyberignore file with patterns:
```
*.log
temp/
*.tmp
cache/
```

**Command-Line:**
```
Command: patch-gen.exe --versions-dir .\testdata\versions --from 1.0.1 --to 1.0.2 --output .\testdata\advanced-output\patches
```

**Expected Output:**
```
Testing: .cyberignore file support
  Testing file exclusion patterns...
  ✓ .cyberignore file found and parsed
  ✓ Log files excluded (*.log)
  ✓ Temp directories excluded (temp/)
  ✓ Cache files excluded (cache/)
  ✓ Ignored files not in patch manifest
✓ PASSED: .cyberignore file support
```

**What This Tests:**
- .cyberignore file is read and parsed
- Wildcard patterns work (*.log, *.tmp)
- Directory patterns work (temp/, cache/)
- Exact filename matches work
- Ignored files are completely excluded from patches

**Pattern Types:**
- `*.log` - Wildcard file extension
- `temp/` - Directory exclusion
- `cache/` - Directory exclusion
- `debug.txt` - Exact filename

---

#### Test 42: Self-Contained Executable Silent Mode
**Purpose:** Verify --silent flag works with self-contained executables

**Command-Line:**
```
Command: .\testdata\advanced-output\executables\1.0.1-to-1.0.2.exe --silent
```

**Expected Output:**
```
Testing: Self-contained executable silent mode
  Testing --silent flag with executable...
  ✓ Executable ran in silent mode
  ✓ No console output during execution
  ✓ Exit code 0 (success) or 1 (failure)
  ✓ Automatic log file generated
✓ PASSED: Self-contained executable silent mode
```

**What This Tests:**
- --silent flag suppresses all console output
- Executable still performs patch application
- Returns appropriate exit codes (0=success, 1=failure)
- Generates log file for audit trail

**Silent Mode Behavior:**
- No progress messages
- No user prompts
- Exit code indicates success/failure
- Log file contains full execution details

---

#### Test 43: Silent Mode Log File Generation
**Purpose:** Verify automatic log file creation in silent mode

**Expected Output:**
```
Testing: Silent mode log file generation
  Testing automatic log generation...
  ✓ Log file created: cyberpatch-20231201-143022.log
  ✓ Log contains execution details
  ✓ Log includes timestamp and version info
  ✓ Log shows success/failure status
✓ PASSED: Silent mode log file generation
```

**What This Tests:**
- Automatic log file naming (timestamp-based)
- Log file contains complete execution trace
- Timestamp format: YYYYMMDD-HHMMSS
- Log includes all operations and results

**Log File Contents:**
```
CyberPatchMaker Silent Mode Log
Timestamp: 2023-12-01 14:30:22
Version: 1.0.1 → 1.0.2
Command: --silent
Status: SUCCESS
Operations: 7 added, 4 modified, 0 deleted
Duration: 0.15 seconds
```

---

### Simple Mode Feature Tests (Tests 51-56)

#### Test 51: Simple Mode - Patch Generation with SimpleMode Flag
**Purpose:** Verify SimpleMode field is set in patch structure when enabled

**Command-Line:**
```
Command: patch-gen.exe --versions-dir .\testdata\versions --from 1.0.1 --to 1.0.2 --output .\testdata\advanced-output\patches --create-exe
```

**Expected Output:**
```
Testing: Simple mode - patch generation with SimpleMode flag
  Testing SimpleMode field in patch structure...
  ✓ Patch generated with SimpleMode=true
  ✓ SimpleMode field found in patch header
  ✓ GUI generator enables SimpleMode via checkbox
  ✓ CLI generator enables SimpleMode via --simple-mode flag
✓ PASSED: Simple mode - patch generation with SimpleMode flag
```

**What This Tests:**
- SimpleMode boolean field in Patch struct (types.go)
- GUI generator checkbox sets SimpleMode=true
- CLI --simple-mode flag sets SimpleMode=true
- Patch header contains SimpleMode metadata

**Technical Details:**
- SimpleMode field added to Patch structure
- GUI generator has "Enable Simple Mode for End Users" checkbox
- CLI generator has --simple-mode flag
- Field is embedded in patch header for applier detection

---

#### Test 52: Simple Mode - GUI Applier Simplified Interface
**Purpose:** Verify GUI applier shows simplified interface when SimpleMode=true

**Expected Output:**
```
Testing: Simple mode - GUI applier simplified interface
  Testing simplified GUI interface...
  ✓ GUI detects SimpleMode field in patch
  ✓ Simplified interface shown (3 options only)
  ✓ Advanced options hidden (compression, verification checkboxes)
  ✓ User sees: Simple message + basic backup option + 3-choice menu
✓ PASSED: Simple mode - GUI applier simplified interface
```

**What This Tests:**
- GUI applier enableSimpleMode() method
- Patch.SimpleMode field detection
- Interface simplification when SimpleMode=true
- Hidden advanced options (compression, verification)

**Simplified Interface:**
- Simple message: "You are about to patch X to Y"
- Basic backup option (default: Yes)
- 3-choice menu: "Dry Run (1)", "Apply Patch (2)", "Exit (3)"
- No technical settings visible

---

#### Test 53: Simple Mode - End-to-End Workflow
**Purpose:** Verify complete Simple Mode workflow from generator to end user

**Command-Line:**
```
Command: patch-gen.exe --versions-dir .\testdata\versions --from 1.0.1 --to 1.0.2 --output .\testdata\advanced-output\simple-workflow --create-exe
```

**Expected Output:**
```
Testing: Simple mode - end-to-end workflow
  Testing complete Simple Mode workflow...
  ✓ Self-contained exe created with SimpleMode=true
  ✓ End user runs exe and sees simple 3-option menu
  ✓ GUI version: Simple message + basic options only
  ✓ CLI version: 3 options - Dry Run (1), Apply Patch (2), Exit (3)
  ✓ Advanced options hidden/auto-enabled for safety
✓ PASSED: Simple mode - end-to-end workflow
```

**What This Tests:**
- Complete workflow: GUI generator → self-contained exe → end user
- Patch creator enables Simple Mode checkbox
- End user receives exe with embedded SimpleMode=true
- End user sees simplified interface (GUI or CLI)
- No advanced settings exposed to end user

**Workflow Steps:**
1. **Patch Creator:** Uses GUI generator, enables "Enable Simple Mode for End Users"
2. **Patch Creator:** Creates self-contained exe with SimpleMode=true
3. **End User:** Runs exe, sees simple 3-option menu
4. **End User:** Can test with Dry Run, then Apply Patch
5. **End User:** Simple, clear choices - no technical complexity

---

#### Test 54: Simple Mode - Feature Documentation Validation
**Purpose:** Verify Simple Mode feature is properly documented and implemented

**Expected Output:**
```
Testing: Simple mode - feature documentation validation
  Validating Simple Mode implementation...
  ✓ Documentation files exist (simple-mode-guide.md, generator-guide.md, etc.)
  ✓ SimpleMode field in types.go Patch struct
  ✓ GUI generator simpleModeForUsers checkbox
  ✓ GUI applier enableSimpleMode method
  ✓ CLI applier runSimpleMode function
  ✓ Feature mentioned in README.md
✓ PASSED: Simple mode - feature documentation validation
```

**What This Tests:**
- All required documentation files exist
- Code implementation is complete across all components
- Feature is properly documented for users
- Implementation follows design specifications

**Validated Components:**
- **Documentation:** simple-mode-guide.md, generator-guide.md, applier-guide.md, gui-usage.md
- **Types:** SimpleMode bool field in Patch struct
- **GUI Generator:** simpleModeCheck widget.Check with "Simple Mode for End Users" text
- **GUI Applier:** enableSimpleMode() method detects patch.SimpleMode
- **CLI Applier:** runSimpleMode() function for simplified CLI interface
- **README:** Feature prominently mentioned

---

#### Test 55: Simple Mode - Use Case Scenarios
**Purpose:** Verify Simple Mode addresses real-world distribution scenarios

**Expected Output:**
```
Testing: Simple mode - use case scenarios
  Validating Simple Mode use cases...
  ✓ Software vendor scenario: Non-technical customers get simple exe
  ✓ IT department scenario: Employees run exe from shared drive
  ✓ Game modder scenario: Users see simple interface, reduces support burden
  ✓ Automation scenario: --silent flag for CI/CD (fully automatic)
✓ PASSED: Simple mode - use case scenarios
```

**What This Tests:**
- Real-world applicability of Simple Mode
- Different user personas and use cases
- Problem-solving for distribution challenges
- Integration with automation workflows

**Validated Use Cases:**

**Use Case 1: Software Vendor Updates**
- Vendor creates patch with Simple Mode enabled
- Customers receive self-contained exe with simple 3-option menu
- Customers see: 'You are about to patch X to Y' message
- Customers choose: Dry Run (test), Apply Patch, or Exit
- No technical knowledge required

**Use Case 2: IT Department Internal Updates**
- IT creates patches with Simple Mode for all versions
- Employees run exe from their version folder
- Simple 3-option interface prevents confusion
- Backup option available (default: Yes)

**Use Case 3: Game/App Modders**
- Modders enable Simple Mode for user-friendly patching
- Users can test with 'Dry Run' before applying
- Reduces support burden (fewer confused users)

**Use Case 4: Automation Scripts (Silent Mode)**
- CLI applier with --silent flag applies patch automatically
- No user interaction required (fully automatic)
- Returns exit code 0 on success, 1 on failure
- Perfect for CI/CD pipelines or deployment scripts

---

#### Test 56: Automatic Rollback on Failed Patch Application
**Purpose:** Verify automatic rollback when patch application fails

**Command-Line:**
```
Command: patch-apply.exe --patch .\testdata\advanced-output\patches\1.0.1-to-1.0.2.patch --current-dir .\testdata\advanced-output\rollback-test --verify
```

**Expected Output:**
```
Testing: Automatic rollback on failed patch application
  Testing automatic rollback when patch fails...
  ✓ Patch application failed as expected (corrupted target file)
  ✓ Automatic rollback was triggered
  ✓ Original files restored from backup
  ✓ Backup directory preserved for inspection
✓ PASSED: Automatic rollback on failed patch application
```

**What This Tests:**
- Automatic rollback on patch failure
- Backup restoration when verification fails
- Corrupted file detection during patching
- Backup preservation after rollback

**Failure Scenario:**
1. Apply patch to directory with corrupted target file
2. Patch application starts and creates backup
3. Verification fails due to file corruption
4. Automatic rollback triggered
5. Original files restored from backup.cyberpatcher/
6. Backup directory kept for manual inspection

**Error Messages:**
- "automatically restoring from backup"
- "Restored X files from backup"
- Clear indication of rollback operation

**Why This Matters:**
- Prevents partial updates that could break applications
- Automatic recovery from patch failures
- Maintains system stability during updates
- Provides safety net for critical deployments