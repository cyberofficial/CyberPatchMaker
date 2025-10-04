#!/bin/bash
# CyberPatchMaker Test Suite
# This script tests the complete functionality of the patch system

echo "========================================"
echo "CyberPatchMaker Test Suite"
echo "========================================"
echo ""

# Track test results
passed=0
failed=0

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
GRAY='\033[0;37m'
NC='\033[0m' # No Color

test_step() {
    local name="$1"
    echo -e "${YELLOW}Testing: $name${NC}"
    
    if eval "$2"; then
        echo -e "${GREEN}✓ PASSED: $name${NC}"
        ((passed++))
        echo ""
        return 0
    else
        echo -e "${RED}✗ FAILED: $name${NC}"
        ((failed++))
        echo ""
        return 1
    fi
}

# Test 1: Check if executables exist
test_step "Build executables" '
    echo -e "  ${GRAY}Building generator...${NC}"
    go build ./cmd/generator > /dev/null 2>&1 || { echo "Failed to build generator"; exit 1; }
    [ -f "generator" ] || { echo "generator executable not found"; exit 1; }
    
    echo -e "  ${GRAY}Building applier...${NC}"
    go build ./cmd/applier > /dev/null 2>&1 || { echo "Failed to build applier"; exit 1; }
    [ -f "applier" ] || { echo "applier executable not found"; exit 1; }
    
    echo -e "  ${GRAY}Both executables built successfully${NC}"
'

# Test 2: Clean and create test directories
test_step "Setup test environment" '
    echo -e "  ${GRAY}Cleaning previous test data...${NC}"
    rm -rf testdata/test-output
    
    mkdir -p testdata/test-output/patches
    mkdir -p testdata/test-output/test-app
    
    echo -e "  ${GRAY}Test directories created${NC}"
'

# Test 3: Verify test data exists
test_step "Verify test data" '
    [ -d "testdata/versions/1.0.0" ] || { echo "Test version 1.0.0 not found"; exit 1; }
    [ -d "testdata/versions/1.0.1" ] || { echo "Test version 1.0.1 not found"; exit 1; }
    
    file_count_100=$(find testdata/versions/1.0.0 -type f | wc -l)
    file_count_101=$(find testdata/versions/1.0.1 -type f | wc -l)
    
    echo -e "  ${GRAY}Version 1.0.0 exists: $file_count_100 files${NC}"
    echo -e "  ${GRAY}Version 1.0.1 exists: $file_count_101 files${NC}"
'

# Test 4: Generate patch
test_step "Generate patch (1.0.0 → 1.0.1)" '
    echo -e "  ${GRAY}Running generator...${NC}"
    ./generator --versions-dir ./testdata/versions --new-version 1.0.1 --output ./testdata/test-output/patches > /dev/null 2>&1 || { echo "Generator failed"; exit 1; }
    
    [ -f "testdata/test-output/patches/1.0.0-to-1.0.1.patch" ] || { echo "Patch file not created"; exit 1; }
    
    patch_size=$(stat -f%z "testdata/test-output/patches/1.0.0-to-1.0.1.patch" 2>/dev/null || stat -c%s "testdata/test-output/patches/1.0.0-to-1.0.1.patch" 2>/dev/null)
    echo -e "  ${GRAY}Patch generated: $patch_size bytes${NC}"
'

# Test 5: Dry-run test
test_step "Dry-run patch application" '
    echo -e "  ${GRAY}Running applier in dry-run mode...${NC}"
    output=$(./applier --patch ./testdata/test-output/patches/1.0.0-to-1.0.1.patch --current-dir ./testdata/versions/1.0.0 --dry-run 2>&1)
    
    echo "$output" | grep -q "DRY RUN MODE" || { echo "Dry-run mode not detected in output"; exit 1; }
    echo "$output" | grep -q "Key file verified" || { echo "Key file verification not found in output"; exit 1; }
    
    echo -e "  ${GRAY}Dry-run completed successfully${NC}"
'

# Test 6: Copy test version and apply patch
test_step "Apply patch to clean installation" '
    echo -e "  ${GRAY}Copying version 1.0.0 to test directory...${NC}"
    mkdir -p testdata/test-output/test-app
    cp -r testdata/versions/1.0.0/. testdata/test-output/test-app/
    
    echo -e "  ${GRAY}Applying patch...${NC}"
    output=$(./applier --patch ./testdata/test-output/patches/1.0.0-to-1.0.1.patch --current-dir ./testdata/test-output/test-app --verify 2>&1)
    [ $? -eq 0 ] || { echo "Patch application failed"; exit 1; }
    
    echo "$output" | grep -q "Patch applied successfully" || { echo "Success message not found in output"; exit 1; }
    
    echo -e "  ${GRAY}Patch applied successfully${NC}"
'

# Test 7: Verify patched files match expected version
test_step "Verify patched files match version 1.0.1" '
    echo -e "  ${GRAY}Comparing patched files with expected version...${NC}"
    
    # Check program.exe
    diff testdata/test-output/test-app/program.exe testdata/versions/1.0.1/program.exe > /dev/null || { echo "program.exe does not match"; exit 1; }
    
    # Check config.json
    diff testdata/test-output/test-app/data/config.json testdata/versions/1.0.1/data/config.json > /dev/null || { echo "data/config.json does not match"; exit 1; }
    
    # Check core.dll
    diff testdata/test-output/test-app/libs/core.dll testdata/versions/1.0.1/libs/core.dll > /dev/null || { echo "libs/core.dll does not match"; exit 1; }
    
    # Check new file exists
    [ -f "testdata/test-output/test-app/libs/newfeature.dll" ] || { echo "New file libs/newfeature.dll was not added"; exit 1; }
    
    # Check new file content
    diff testdata/test-output/test-app/libs/newfeature.dll testdata/versions/1.0.1/libs/newfeature.dll > /dev/null || { echo "libs/newfeature.dll does not match"; exit 1; }
    
    echo -e "  ${GRAY}All files match expected version 1.0.1${NC}"
'

# Test 8: Test rejection of modified files
test_step "Verify patch rejection for modified installation" '
    echo -e "  ${GRAY}Creating modified installation...${NC}"
    mkdir -p testdata/test-output/modified-app
    cp -r testdata/versions/1.0.0/. testdata/test-output/modified-app/
    
    echo -e "  ${GRAY}Modifying key file...${NC}"
    echo "CORRUPTED" >> testdata/test-output/modified-app/program.exe
    
    echo -e "  ${GRAY}Attempting to apply patch (should fail)...${NC}"
    output=$(./applier --patch ./testdata/test-output/patches/1.0.0-to-1.0.1.patch --current-dir ./testdata/test-output/modified-app --verify 2>&1)
    
    [ $? -ne 0 ] || { echo "Patch should have been rejected but succeeded"; exit 1; }
    
    echo "$output" | grep -q "checksum mismatch" || { echo "Checksum mismatch error not found"; exit 1; }
    echo "$output" | grep -q "Restoring from backup" || { echo "Backup restoration message not found"; exit 1; }
    
    # Verify file was restored
    diff testdata/test-output/modified-app/program.exe testdata/versions/1.0.0/program.exe > /dev/null || { echo "Modified file was not restored from backup"; exit 1; }
    
    echo -e "  ${GRAY}Patch correctly rejected and backup restored${NC}"
'

# Test 9: Test with different compression
test_step "Generate patch with gzip compression" '
    echo -e "  ${GRAY}Generating patch with gzip...${NC}"
    ./generator --versions-dir ./testdata/versions --from 1.0.0 --to 1.0.1 --output ./testdata/test-output/patches/gzip-test.patch --compression gzip > /dev/null 2>&1 || { echo "Generator failed"; exit 1; }
    
    [ -f "testdata/test-output/patches/gzip-test.patch" ] || { echo "Gzip patch file not created"; exit 1; }
    
    patch_size=$(stat -f%z "testdata/test-output/patches/gzip-test.patch" 2>/dev/null || stat -c%s "testdata/test-output/patches/gzip-test.patch" 2>/dev/null)
    echo -e "  ${GRAY}Gzip patch generated: $patch_size bytes${NC}"
'

# Test 10: Apply gzip patch
test_step "Apply gzip-compressed patch" '
    echo -e "  ${GRAY}Copying version 1.0.0...${NC}"
    [ -d "testdata/test-output/gzip-app" ] && rm -rf testdata/test-output/gzip-app
    mkdir -p testdata/test-output/gzip-app
    cp -r testdata/versions/1.0.0/. testdata/test-output/gzip-app/
    
    echo -e "  ${GRAY}Applying gzip patch...${NC}"
    ./applier --patch ./testdata/test-output/patches/gzip-test.patch --current-dir ./testdata/test-output/gzip-app --verify > /dev/null 2>&1 || { echo "Gzip patch application failed"; exit 1; }
    
    # Verify result matches expected version
    diff testdata/test-output/gzip-app/program.exe testdata/versions/1.0.1/program.exe > /dev/null || { echo "Gzip patched files do not match expected version"; exit 1; }
    
    echo -e "  ${GRAY}Gzip patch applied successfully${NC}"
'

# Final summary
echo ""
echo "========================================"
echo "Test Results"
echo "========================================"
echo -e "${GREEN}Passed: $passed${NC}"
echo -e "${RED}Failed: $failed${NC}"
echo ""

if [ $failed -eq 0 ]; then
    echo -e "${GREEN}✓ All tests passed! CyberPatchMaker is working correctly.${NC}"
    exit 0
else
    echo -e "${RED}✗ Some tests failed. Please check the output above.${NC}"
    exit 1
fi
