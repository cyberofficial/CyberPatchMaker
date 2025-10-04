# CyberPatchMaker Test Suite
# This script tests the complete functionality of the patch system

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "CyberPatchMaker Test Suite" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Track test results
$passed = 0
$failed = 0

function Test-Step {
    param(
        [string]$Name,
        [scriptblock]$Test
    )
    
    Write-Host "Testing: $Name" -ForegroundColor Yellow
    try {
        & $Test
        Write-Host "✓ PASSED: $Name" -ForegroundColor Green
        $script:passed++
        return $true
    } catch {
        Write-Host "✗ FAILED: $Name" -ForegroundColor Red
        Write-Host "  Error: $_" -ForegroundColor Red
        $script:failed++
        return $false
    }
    Write-Host ""
}

# Test 1: Check if executables exist
Test-Step "Build executables" {
    Write-Host "  Building generator..." -ForegroundColor Gray
    go build ./cmd/generator 2>&1 | Out-Null
    if (-not (Test-Path "generator.exe")) {
        throw "Failed to build generator.exe"
    }
    
    Write-Host "  Building applier..." -ForegroundColor Gray
    go build ./cmd/applier 2>&1 | Out-Null
    if (-not (Test-Path "applier.exe")) {
        throw "Failed to build applier.exe"
    }
    
    Write-Host "  Both executables built successfully" -ForegroundColor Gray
}

# Test 2: Clean and create test directories
Test-Step "Setup test environment" {
    Write-Host "  Cleaning previous test data..." -ForegroundColor Gray
    if (Test-Path "testdata/test-output") {
        Remove-Item "testdata/test-output" -Recurse -Force
    }
    
    New-Item -ItemType Directory -Force -Path "testdata/test-output" | Out-Null
    New-Item -ItemType Directory -Force -Path "testdata/test-output/patches" | Out-Null
    New-Item -ItemType Directory -Force -Path "testdata/test-output/test-app" | Out-Null
    
    Write-Host "  Test directories created" -ForegroundColor Gray
}

# Test 3: Verify test data exists
Test-Step "Verify test data" {
    if (-not (Test-Path "testdata/versions/1.0.0")) {
        throw "Test version 1.0.0 not found"
    }
    if (-not (Test-Path "testdata/versions/1.0.1")) {
        throw "Test version 1.0.1 not found"
    }
    
    Write-Host "  Version 1.0.0 exists: $(Get-ChildItem 'testdata/versions/1.0.0' -Recurse -File | Measure-Object).Count files" -ForegroundColor Gray
    Write-Host "  Version 1.0.1 exists: $(Get-ChildItem 'testdata/versions/1.0.1' -Recurse -File | Measure-Object).Count files" -ForegroundColor Gray
}

# Test 4: Generate patch
Test-Step "Generate patch (1.0.0 → 1.0.1)" {
    Write-Host "  Running generator..." -ForegroundColor Gray
    $output = .\generator.exe --versions-dir .\testdata\versions --new-version 1.0.1 --output .\testdata\test-output\patches 2>&1
    
    if ($LASTEXITCODE -ne 0) {
        throw "Generator failed with exit code $LASTEXITCODE"
    }
    
    if (-not (Test-Path "testdata/test-output/patches/1.0.0-to-1.0.1.patch")) {
        throw "Patch file not created"
    }
    
    $patchSize = (Get-Item "testdata/test-output/patches/1.0.0-to-1.0.1.patch").Length
    Write-Host "  Patch generated: $patchSize bytes" -ForegroundColor Gray
}

# Test 5: Dry-run test
Test-Step "Dry-run patch application" {
    Write-Host "  Running applier in dry-run mode..." -ForegroundColor Gray
    $output = .\applier.exe --patch .\testdata\test-output\patches\1.0.0-to-1.0.1.patch --current-dir .\testdata\versions\1.0.0 --dry-run 2>&1
    
    if ($LASTEXITCODE -ne 0) {
        throw "Dry-run failed with exit code $LASTEXITCODE"
    }
    
    $outputStr = $output -join "`n"
    if ($outputStr -notmatch "DRY RUN MODE") {
        throw "Dry-run mode not detected in output"
    }
    if ($outputStr -notmatch "Key file verified") {
        throw "Key file verification not found in output"
    }
    
    Write-Host "  Dry-run completed successfully" -ForegroundColor Gray
}

# Test 6: Copy test version and apply patch
Test-Step "Apply patch to clean installation" {
    Write-Host "  Copying version 1.0.0 to test directory..." -ForegroundColor Gray
    New-Item -Path "testdata/test-output/test-app" -ItemType Directory -Force | Out-Null
    Get-ChildItem -Path "testdata/versions/1.0.0" | Copy-Item -Destination "testdata/test-output/test-app" -Recurse -Force
    
    Write-Host "  Applying patch..." -ForegroundColor Gray
    $output = .\applier.exe --patch .\testdata\test-output\patches\1.0.0-to-1.0.1.patch --current-dir .\testdata\test-output\test-app --verify 2>&1
    
    if ($LASTEXITCODE -ne 0) {
        throw "Patch application failed with exit code $LASTEXITCODE"
    }
    
    $outputStr = $output -join "`n"
    if ($outputStr -notmatch "Patch applied successfully") {
        throw "Success message not found in output"
    }
    
    Write-Host "  Patch applied successfully" -ForegroundColor Gray
}

# Test 7: Verify patched files match expected version
Test-Step "Verify patched files match version 1.0.1" {
    Write-Host "  Comparing patched files with expected version..." -ForegroundColor Gray
    
    # Check program.exe
    $diff = Compare-Object (Get-Content "testdata/test-output/test-app/program.exe") (Get-Content "testdata/versions/1.0.1/program.exe")
    if ($diff) {
        throw "program.exe does not match expected version"
    }
    
    # Check config.json
    $diff = Compare-Object (Get-Content "testdata/test-output/test-app/data/config.json") (Get-Content "testdata/versions/1.0.1/data/config.json")
    if ($diff) {
        throw "data/config.json does not match expected version"
    }
    
    # Check core.dll
    $diff = Compare-Object (Get-Content "testdata/test-output/test-app/libs/core.dll") (Get-Content "testdata/versions/1.0.1/libs/core.dll")
    if ($diff) {
        throw "libs/core.dll does not match expected version"
    }
    
    # Check new file exists
    if (-not (Test-Path "testdata/test-output/test-app/libs/newfeature.dll")) {
        throw "New file libs/newfeature.dll was not added"
    }
    
    # Check new file content
    $diff = Compare-Object (Get-Content "testdata/test-output/test-app/libs/newfeature.dll") (Get-Content "testdata/versions/1.0.1/libs/newfeature.dll")
    if ($diff) {
        throw "libs/newfeature.dll does not match expected version"
    }
    
    Write-Host "  All files match expected version 1.0.1" -ForegroundColor Gray
}

# Test 8: Test rejection of modified files
Test-Step "Verify patch rejection for modified installation" {
    Write-Host "  Creating modified installation..." -ForegroundColor Gray
    # Copy 1.0.0 contents to modified-app
    New-Item -Path "testdata/test-output/modified-app" -ItemType Directory -Force | Out-Null
    Get-ChildItem -Path "testdata/versions/1.0.0" | Copy-Item -Destination "testdata/test-output/modified-app" -Recurse -Force
    
    Write-Host "  Modifying key file..." -ForegroundColor Gray
    Add-Content -Path "testdata/test-output/modified-app/program.exe" -Value "CORRUPTED"
    
    Write-Host "  Attempting to apply patch (should fail)..." -ForegroundColor Gray
    $output = .\applier.exe --patch .\testdata\test-output\patches\1.0.0-to-1.0.1.patch --current-dir .\testdata\test-output\modified-app --verify 2>&1
    
    if ($LASTEXITCODE -eq 0) {
        throw "Patch should have been rejected but succeeded"
    }
    
    $outputStr = $output -join "`n"
    if ($outputStr -notmatch "checksum mismatch") {
        throw "Checksum mismatch error not found in output"
    }
    if ($outputStr -notmatch "Backup restored successfully") {
        throw "Backup restoration message not found in output"
    }
    
    # Verify file was restored
    $diff = Compare-Object (Get-Content "testdata/test-output/modified-app/program.exe") (Get-Content "testdata/versions/1.0.0/program.exe")
    if ($diff) {
        throw "Modified file was not restored from backup"
    }
    
    Write-Host "  Patch correctly rejected and backup restored" -ForegroundColor Gray
}

# Test 9: Test with different compression
Test-Step "Generate patch with gzip compression" {
    Write-Host "  Generating patch with gzip..." -ForegroundColor Gray
    New-Item -Path "testdata/test-output/patches-gzip" -ItemType Directory -Force | Out-Null
    $output = .\generator.exe --versions-dir .\testdata\versions --from 1.0.0 --to 1.0.1 --output .\testdata\test-output\patches-gzip --compression gzip 2>&1
    
    if ($LASTEXITCODE -ne 0) {
        throw "Generator failed with exit code $LASTEXITCODE"
    }
    
    if (-not (Test-Path "testdata/test-output/patches-gzip/1.0.0-to-1.0.1.patch")) {
        throw "Gzip patch file not created"
    }
    
    $patchSize = (Get-Item "testdata/test-output/patches-gzip/1.0.0-to-1.0.1.patch").Length
    Write-Host "  Gzip patch generated: $patchSize bytes" -ForegroundColor Gray
}

# Test 10: Apply gzip patch
Test-Step "Apply gzip-compressed patch" {
    Write-Host "  Copying version 1.0.0..." -ForegroundColor Gray
    if (Test-Path "testdata/test-output/gzip-app") {
        Remove-Item "testdata/test-output/gzip-app" -Recurse -Force
    }
    # Copy 1.0.0 contents to gzip-app
    New-Item -Path "testdata/test-output/gzip-app" -ItemType Directory -Force | Out-Null
    Get-ChildItem -Path "testdata/versions/1.0.0" | Copy-Item -Destination "testdata/test-output/gzip-app" -Recurse -Force
    
    Write-Host "  Applying gzip patch..." -ForegroundColor Gray
    $output = .\applier.exe --patch .\testdata\test-output\patches-gzip\1.0.0-to-1.0.1.patch --current-dir .\testdata\test-output\gzip-app --verify 2>&1
    
    if ($LASTEXITCODE -ne 0) {
        throw "Gzip patch application failed with exit code $LASTEXITCODE"
    }
    
    # Verify result matches expected version
    $diff = Compare-Object (Get-Content "testdata/test-output/gzip-app/program.exe") (Get-Content "testdata/versions/1.0.1/program.exe")
    if ($diff) {
        throw "Gzip patched files do not match expected version"
    }
    
    Write-Host "  Gzip patch applied successfully" -ForegroundColor Gray
}

# Final summary
Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Test Results" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Passed: $passed" -ForegroundColor Green
Write-Host "Failed: $failed" -ForegroundColor Red
Write-Host ""

if ($failed -eq 0) {
    Write-Host "✓ All tests passed! CyberPatchMaker is working correctly." -ForegroundColor Green
    exit 0
} else {
    Write-Host "✗ Some tests failed. Please check the output above." -ForegroundColor Red
    exit 1
}
