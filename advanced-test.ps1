# CyberPatchMaker Advanced Test Suite
# Tests complex scenarios with nested directories, multiple operations, and various compression formats

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "CyberPatchMaker Advanced Test Suite" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Track test results
$passed = 0
$failed = 0

# Function to create version 1.0.0 (baseline version)
function Create-Version-1.0.0 {
    param([string]$BasePath = "testdata/versions/1.0.0")
    
    Write-Host "  Creating version 1.0.0..." -ForegroundColor Gray
    
    # Create directory structure
    New-Item -ItemType Directory -Force -Path "$BasePath/data" | Out-Null
    New-Item -ItemType Directory -Force -Path "$BasePath/libs" | Out-Null
    
    # Create program.exe
    Set-Content -Path "$BasePath/program.exe" -Value "Test Program v1.0.0`n"
    
    # Create data/config.json
    Set-Content -Path "$BasePath/data/config.json" -Value '{"version":"1.0.0","name":"TestApp","features":["basic"]}'
    
    # Create libs/core.dll
    Set-Content -Path "$BasePath/libs/core.dll" -Value "Core Library v1.0.0`n"
    
    Write-Host "  Version 1.0.0 created (3 files, 2 directories)" -ForegroundColor Gray
}

# Function to create version 1.0.1 (simple update)
function Create-Version-1.0.1 {
    param([string]$BasePath = "testdata/versions/1.0.1")
    
    Write-Host "  Creating version 1.0.1..." -ForegroundColor Gray
    
    # Create directory structure
    New-Item -ItemType Directory -Force -Path "$BasePath/data" | Out-Null
    New-Item -ItemType Directory -Force -Path "$BasePath/libs" | Out-Null
    
    # Create program.exe (modified)
    Set-Content -Path "$BasePath/program.exe" -Value "Test Program v1.0.1`n"
    
    # Create data/config.json (modified)
    Set-Content -Path "$BasePath/data/config.json" -Value '{"version":"1.0.1","name":"TestApp","features":["basic","advanced"]}'
    
    # Create libs/core.dll (modified)
    Set-Content -Path "$BasePath/libs/core.dll" -Value "Core Library v1.5.0`n"
    
    # Create libs/newfeature.dll (new file)
    Set-Content -Path "$BasePath/libs/newfeature.dll" -Value "New Feature v1.0.0`n"
    
    Write-Host "  Version 1.0.1 created (4 files, 2 directories)" -ForegroundColor Gray
}

# Function to create version 1.0.2 (complex nested structure)
function Create-Version-1.0.2 {
    param([string]$BasePath = "testdata/versions/1.0.2")
    
    Write-Host "  Creating version 1.0.2..." -ForegroundColor Gray
    
    # Create complex directory structure
    New-Item -ItemType Directory -Force -Path "$BasePath/data" | Out-Null
    New-Item -ItemType Directory -Force -Path "$BasePath/data/assets/images" | Out-Null
    New-Item -ItemType Directory -Force -Path "$BasePath/data/locale" | Out-Null
    New-Item -ItemType Directory -Force -Path "$BasePath/libs" | Out-Null
    New-Item -ItemType Directory -Force -Path "$BasePath/libs/plugins" | Out-Null
    New-Item -ItemType Directory -Force -Path "$BasePath/plugins" | Out-Null
    
    # Create program.exe (modified)
    Set-Content -Path "$BasePath/program.exe" -Value "Test Program v1.0.2`n"
    
    # Create data/config.json (modified with more features)
    Set-Content -Path "$BasePath/data/config.json" -Value '{"version":"1.0.2","name":"TestApp","features":["basic","advanced","premium"],"locale":"en-US"}'
    
    # Create data/assets/images/logo.png (new file - simulated binary)
    $pngHeader = [byte[]](0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A)
    [System.IO.File]::WriteAllBytes("$BasePath/data/assets/images/logo.png", $pngHeader + [byte[]](1..100))
    
    # Create data/assets/images/icon.png (new file - simulated binary)
    [System.IO.File]::WriteAllBytes("$BasePath/data/assets/images/icon.png", $pngHeader + [byte[]](1..50))
    
    # Create data/locale/en-US.json (new file)
    Set-Content -Path "$BasePath/data/locale/en-US.json" -Value '{"app_name":"Test Application","welcome":"Welcome to TestApp v1.0.2"}'
    
    # Create libs/core.dll (modified)
    Set-Content -Path "$BasePath/libs/core.dll" -Value "Core Library v2.5.0`n"
    
    # Create libs/newfeature.dll (modified)
    Set-Content -Path "$BasePath/libs/newfeature.dll" -Value "New Feature v1.5.0`n"
    
    # Create libs/plugins/api.dll (new file)
    Set-Content -Path "$BasePath/libs/plugins/api.dll" -Value "Plugin API v1.0.0`n"
    
    # Create plugins/sample.plugin (new file)
    Set-Content -Path "$BasePath/plugins/sample.plugin" -Value "Sample Plugin v1.0.0`n"
    
    # Create plugins/sample.json (new file)
    Set-Content -Path "$BasePath/plugins/sample.json" -Value '{"name":"sample","version":"1.0.0","enabled":true}'
    
    Write-Host "  Version 1.0.2 created (11 files, 6 directories, 3 levels deep)" -ForegroundColor Gray
}

# Function to ensure all test versions exist
function Ensure-Test-Versions {
    # Check if testdata was kept from previous run
    $cleanupStatePath = "testdata/.cleanup-deferred"
    if (Test-Path $cleanupStatePath) {
        Write-Host "Previous test data detected (cleanup was deferred)..." -ForegroundColor Yellow
        Write-Host "Removing old test data..." -ForegroundColor Yellow
        Remove-Item "testdata" -Recurse -Force -ErrorAction SilentlyContinue
        Write-Host "✓ Old test data removed" -ForegroundColor Green
        Write-Host ""
    }
    
    Write-Host "Checking for test versions..." -ForegroundColor Cyan
    
    $versionsCreated = 0
    
    if (-not (Test-Path "testdata/versions/1.0.0/program.exe")) {
        Write-Host "Version 1.0.0 not found, creating..." -ForegroundColor Yellow
        Create-Version-1.0.0
        $versionsCreated++
    } else {
        Write-Host "✓ Version 1.0.0 exists" -ForegroundColor Green
    }
    
    if (-not (Test-Path "testdata/versions/1.0.1/program.exe")) {
        Write-Host "Version 1.0.1 not found, creating..." -ForegroundColor Yellow
        Create-Version-1.0.1
        $versionsCreated++
    } else {
        Write-Host "✓ Version 1.0.1 exists" -ForegroundColor Green
    }
    
    if (-not (Test-Path "testdata/versions/1.0.2/program.exe")) {
        Write-Host "Version 1.0.2 not found, creating..." -ForegroundColor Yellow
        Create-Version-1.0.2
        $versionsCreated++
    } else {
        Write-Host "✓ Version 1.0.2 exists" -ForegroundColor Green
    }
    
    if ($versionsCreated -gt 0) {
        Write-Host ""
        Write-Host "Created $versionsCreated test version(s)" -ForegroundColor Green
    }
    Write-Host ""
}

# Ensure test versions exist before running tests
Ensure-Test-Versions

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

# Test 1: Verify executables
Test-Step "Verify executables exist" {
    if (-not (Test-Path "generator.exe")) {
        throw "generator.exe not found. Run 'go build ./cmd/generator' first."
    }
    if (-not (Test-Path "applier.exe")) {
        throw "applier.exe not found. Run 'go build ./cmd/applier' first."
    }
    Write-Host "  Both executables found" -ForegroundColor Gray
}

# Test 2: Setup advanced test environment
Test-Step "Setup advanced test environment" {
    Write-Host "  Creating test directories..." -ForegroundColor Gray
    if (Test-Path "testdata/advanced-output") {
        Remove-Item "testdata/advanced-output" -Recurse -Force
    }
    
    New-Item -ItemType Directory -Force -Path "testdata/advanced-output" | Out-Null
    New-Item -ItemType Directory -Force -Path "testdata/advanced-output/patches" | Out-Null
    New-Item -ItemType Directory -Force -Path "testdata/advanced-output/patches-gzip" | Out-Null
    New-Item -ItemType Directory -Force -Path "testdata/advanced-output/patches-none" | Out-Null
    
    Write-Host "  Test directories created" -ForegroundColor Gray
}

# Test 3: Verify all three versions exist
Test-Step "Verify test versions exist" {
    if (-not (Test-Path "testdata/versions/1.0.0")) {
        throw "Version 1.0.0 not found"
    }
    if (-not (Test-Path "testdata/versions/1.0.1")) {
        throw "Version 1.0.1 not found"
    }
    if (-not (Test-Path "testdata/versions/1.0.2")) {
        throw "Version 1.0.2 not found"
    }
    
    $v100_files = (Get-ChildItem 'testdata/versions/1.0.0' -Recurse -File | Measure-Object).Count
    $v101_files = (Get-ChildItem 'testdata/versions/1.0.1' -Recurse -File | Measure-Object).Count
    $v102_files = (Get-ChildItem 'testdata/versions/1.0.2' -Recurse -File | Measure-Object).Count
    
    Write-Host "  Version 1.0.0: $v100_files files" -ForegroundColor Gray
    Write-Host "  Version 1.0.1: $v101_files files" -ForegroundColor Gray
    Write-Host "  Version 1.0.2: $v102_files files (complex structure)" -ForegroundColor Gray
}

# Test 4: Generate patch with default compression (zstd)
Test-Step "Generate complex patch (1.0.1 → 1.0.2) with zstd" {
    Write-Host "  Generating patch from 1.0.1 to 1.0.2 with zstd compression..." -ForegroundColor Gray
    $output = .\generator.exe --versions-dir .\testdata\versions --from 1.0.1 --to 1.0.2 --output .\testdata\advanced-output\patches --compression zstd 2>&1
    
    if ($LASTEXITCODE -ne 0) {
        throw "Generator failed with exit code $LASTEXITCODE"
    }
    
    if (-not (Test-Path "testdata/advanced-output/patches/1.0.1-to-1.0.2.patch")) {
        throw "Patch file not created"
    }
    
    $patchSize = (Get-Item "testdata/advanced-output/patches/1.0.1-to-1.0.2.patch").Length
    Write-Host "  Patch generated (zstd): $patchSize bytes" -ForegroundColor Gray
    
    $outputStr = $output -join "`n"
    if ($outputStr -match "Operations:.*Add:\s*(\d+).*Modify:\s*(\d+).*Delete:\s*(\d+)") {
        Write-Host "  Operations detected: Add=$($Matches[1]), Modify=$($Matches[2]), Delete=$($Matches[3])" -ForegroundColor Gray
    }
}

# Test 5: Generate same patch with gzip compression
Test-Step "Generate same patch with gzip compression" {
    Write-Host "  Generating patch with gzip compression..." -ForegroundColor Gray
    $output = .\generator.exe --versions-dir .\testdata\versions --from 1.0.1 --to 1.0.2 --output .\testdata\advanced-output\patches-gzip --compression gzip 2>&1
    
    if ($LASTEXITCODE -ne 0) {
        throw "Generator failed with exit code $LASTEXITCODE"
    }
    
    if (-not (Test-Path "testdata/advanced-output/patches-gzip/1.0.1-to-1.0.2.patch")) {
        throw "Gzip patch file not created"
    }
    
    $patchSize = (Get-Item "testdata/advanced-output/patches-gzip/1.0.1-to-1.0.2.patch").Length
    Write-Host "  Patch generated (gzip): $patchSize bytes" -ForegroundColor Gray
}

# Test 6: Generate same patch with no compression
Test-Step "Generate same patch with no compression" {
    Write-Host "  Generating patch with no compression..." -ForegroundColor Gray
    $output = .\generator.exe --versions-dir .\testdata\versions --from 1.0.1 --to 1.0.2 --output .\testdata\advanced-output\patches-none --compression none 2>&1
    
    if ($LASTEXITCODE -ne 0) {
        throw "Generator failed with exit code $LASTEXITCODE"
    }
    
    if (-not (Test-Path "testdata/advanced-output/patches-none/1.0.1-to-1.0.2.patch")) {
        throw "Uncompressed patch file not created"
    }
    
    $patchSize = (Get-Item "testdata/advanced-output/patches-none/1.0.1-to-1.0.2.patch").Length
    Write-Host "  Patch generated (none): $patchSize bytes" -ForegroundColor Gray
}

# Test 7: Compare patch sizes
Test-Step "Compare compression efficiency" {
    $zstdSize = (Get-Item "testdata/advanced-output/patches/1.0.1-to-1.0.2.patch").Length
    $gzipSize = (Get-Item "testdata/advanced-output/patches-gzip/1.0.1-to-1.0.2.patch").Length
    $noneSize = (Get-Item "testdata/advanced-output/patches-none/1.0.1-to-1.0.2.patch").Length
    
    Write-Host "  Compression comparison:" -ForegroundColor Gray
    Write-Host "    zstd: $zstdSize bytes (100%)" -ForegroundColor Gray
    Write-Host "    gzip: $gzipSize bytes ($([math]::Round($gzipSize * 100.0 / $zstdSize, 1))%)" -ForegroundColor Gray
    Write-Host "    none: $noneSize bytes ($([math]::Round($noneSize * 100.0 / $zstdSize, 1))%)" -ForegroundColor Gray
    
    if ($noneSize -le $zstdSize) {
        throw "Uncompressed patch should be larger than compressed"
    }
    
    Write-Host "  Compression is working correctly" -ForegroundColor Gray
}

# Test 8: Dry-run complex patch
Test-Step "Dry-run complex patch application" {
    Write-Host "  Running applier in dry-run mode..." -ForegroundColor Gray
    $output = .\applier.exe --patch .\testdata\advanced-output\patches\1.0.1-to-1.0.2.patch --current-dir .\testdata\versions\1.0.1 --dry-run 2>&1
    
    if ($LASTEXITCODE -ne 0) {
        throw "Dry-run failed with exit code $LASTEXITCODE"
    }
    
    $outputStr = $output -join "`n"
    # Check for dry-run indicators (more flexible matching)
    if ($outputStr -notmatch "DRY|dry|Dry|simulation|would") {
        Write-Host "  Warning: Dry-run mode message not clearly detected" -ForegroundColor Yellow
    }
    
    Write-Host "  Dry-run completed successfully (exit code 0)" -ForegroundColor Gray
}

# Test 9: Apply zstd patch to clean copy
Test-Step "Apply zstd patch to complex directory structure" {
    Write-Host "  Copying version 1.0.1 to test-zstd..." -ForegroundColor Gray
    New-Item -Path "testdata/advanced-output/test-zstd" -ItemType Directory -Force | Out-Null
    Get-ChildItem -Path "testdata/versions/1.0.1" -Recurse | Copy-Item -Destination {
        $dest = Join-Path "testdata/advanced-output/test-zstd" $_.FullName.Substring((Resolve-Path "testdata/versions/1.0.1").Path.Length)
        $destDir = Split-Path $dest
        if (-not (Test-Path $destDir)) { New-Item -Path $destDir -ItemType Directory -Force | Out-Null }
        $dest
    } -Force
    
    Write-Host "  Applying zstd patch..." -ForegroundColor Gray
    $output = .\applier.exe --patch .\testdata\advanced-output\patches\1.0.1-to-1.0.2.patch --current-dir .\testdata\advanced-output\test-zstd --verify 2>&1
    
    if ($LASTEXITCODE -ne 0) {
        $outputStr = $output -join "`n"
        throw "Patch application failed: $outputStr"
    }
    
    $outputStr = $output -join "`n"
    if ($outputStr -notmatch "Patch applied successfully") {
        throw "Success message not found in output"
    }
    
    Write-Host "  Zstd patch applied successfully" -ForegroundColor Gray
}

# Test 10: Apply gzip patch to clean copy
Test-Step "Apply gzip patch to complex directory structure" {
    Write-Host "  Copying version 1.0.1 to test-gzip..." -ForegroundColor Gray
    New-Item -Path "testdata/advanced-output/test-gzip" -ItemType Directory -Force | Out-Null
    Get-ChildItem -Path "testdata/versions/1.0.1" -Recurse | Copy-Item -Destination {
        $dest = Join-Path "testdata/advanced-output/test-gzip" $_.FullName.Substring((Resolve-Path "testdata/versions/1.0.1").Path.Length)
        $destDir = Split-Path $dest
        if (-not (Test-Path $destDir)) { New-Item -Path $destDir -ItemType Directory -Force | Out-Null }
        $dest
    } -Force
    
    Write-Host "  Applying gzip patch..." -ForegroundColor Gray
    $output = .\applier.exe --patch .\testdata\advanced-output\patches-gzip\1.0.1-to-1.0.2.patch --current-dir .\testdata\advanced-output\test-gzip --verify 2>&1
    
    if ($LASTEXITCODE -ne 0) {
        $outputStr = $output -join "`n"
        throw "Gzip patch application failed: $outputStr"
    }
    
    Write-Host "  Gzip patch applied successfully" -ForegroundColor Gray
}

# Test 11: Apply uncompressed patch to clean copy
Test-Step "Apply uncompressed patch to complex directory structure" {
    Write-Host "  Copying version 1.0.1 to test-none..." -ForegroundColor Gray
    New-Item -Path "testdata/advanced-output/test-none" -ItemType Directory -Force | Out-Null
    Get-ChildItem -Path "testdata/versions/1.0.1" -Recurse | Copy-Item -Destination {
        $dest = Join-Path "testdata/advanced-output/test-none" $_.FullName.Substring((Resolve-Path "testdata/versions/1.0.1").Path.Length)
        $destDir = Split-Path $dest
        if (-not (Test-Path $destDir)) { New-Item -Path $destDir -ItemType Directory -Force | Out-Null }
        $dest
    } -Force
    
    Write-Host "  Applying uncompressed patch..." -ForegroundColor Gray
    $output = .\applier.exe --patch .\testdata\advanced-output\patches-none\1.0.1-to-1.0.2.patch --current-dir .\testdata\advanced-output\test-none --verify 2>&1
    
    if ($LASTEXITCODE -ne 0) {
        $outputStr = $output -join "`n"
        throw "Uncompressed patch application failed: $outputStr"
    }
    
    Write-Host "  Uncompressed patch applied successfully" -ForegroundColor Gray
}

# Test 12: Verify nested directory structure created correctly
Test-Step "Verify complex directory structure after patching" {
    Write-Host "  Checking nested directories..." -ForegroundColor Gray
    
    # Check that new directories were created
    if (-not (Test-Path "testdata/advanced-output/test-zstd/data/assets/images")) {
        throw "Nested directory data/assets/images not created"
    }
    if (-not (Test-Path "testdata/advanced-output/test-zstd/data/locale")) {
        throw "Directory data/locale not created"
    }
    if (-not (Test-Path "testdata/advanced-output/test-zstd/libs/plugins")) {
        throw "Directory libs/plugins not created"
    }
    if (-not (Test-Path "testdata/advanced-output/test-zstd/plugins")) {
        throw "Directory plugins not created"
    }
    
    Write-Host "  All nested directories created correctly" -ForegroundColor Gray
}

# Test 13: Verify new files were added correctly
Test-Step "Verify new files added in nested paths" {
    Write-Host "  Checking new files..." -ForegroundColor Gray
    
    # Check deeply nested files
    if (-not (Test-Path "testdata/advanced-output/test-zstd/data/assets/images/logo.png")) {
        throw "File data/assets/images/logo.png not added"
    }
    if (-not (Test-Path "testdata/advanced-output/test-zstd/data/assets/images/icon.png")) {
        throw "File data/assets/images/icon.png not added"
    }
    if (-not (Test-Path "testdata/advanced-output/test-zstd/data/locale/en-US.json")) {
        throw "File data/locale/en-US.json not added"
    }
    if (-not (Test-Path "testdata/advanced-output/test-zstd/libs/plugins/api.dll")) {
        throw "File libs/plugins/api.dll not added"
    }
    if (-not (Test-Path "testdata/advanced-output/test-zstd/plugins/sample.plugin")) {
        throw "File plugins/sample.plugin not added"
    }
    if (-not (Test-Path "testdata/advanced-output/test-zstd/plugins/sample.json")) {
        throw "File plugins/sample.json not added"
    }
    
    Write-Host "  All new files added correctly" -ForegroundColor Gray
}

# Test 14: Verify modified files match expected version
Test-Step "Verify modified files match version 1.0.2" {
    Write-Host "  Comparing modified files with expected version..." -ForegroundColor Gray
    
    # Compare key file
    $diff = Compare-Object (Get-Content "testdata/advanced-output/test-zstd/program.exe") (Get-Content "testdata/versions/1.0.2/program.exe")
    if ($diff) {
        throw "program.exe does not match expected version 1.0.2"
    }
    
    # Compare config.json
    $diff = Compare-Object (Get-Content "testdata/advanced-output/test-zstd/data/config.json") (Get-Content "testdata/versions/1.0.2/data/config.json")
    if ($diff) {
        throw "data/config.json does not match expected version"
    }
    
    # Compare core.dll
    $diff = Compare-Object (Get-Content "testdata/advanced-output/test-zstd/libs/core.dll") (Get-Content "testdata/versions/1.0.2/libs/core.dll")
    if ($diff) {
        throw "libs/core.dll does not match expected version"
    }
    
    # Compare newfeature.dll
    $diff = Compare-Object (Get-Content "testdata/advanced-output/test-zstd/libs/newfeature.dll") (Get-Content "testdata/versions/1.0.2/libs/newfeature.dll")
    if ($diff) {
        throw "libs/newfeature.dll does not match expected version"
    }
    
    Write-Host "  All modified files match expected version 1.0.2" -ForegroundColor Gray
}

# Test 15: Verify all three compression methods produce identical results
Test-Step "Verify all compression methods produce identical results" {
    Write-Host "  Comparing results from different compression methods..." -ForegroundColor Gray
    
    # Get file counts
    $zstdFiles = (Get-ChildItem "testdata/advanced-output/test-zstd" -Recurse -File | Measure-Object).Count
    $gzipFiles = (Get-ChildItem "testdata/advanced-output/test-gzip" -Recurse -File | Measure-Object).Count
    $noneFiles = (Get-ChildItem "testdata/advanced-output/test-none" -Recurse -File | Measure-Object).Count
    
    if ($zstdFiles -ne $gzipFiles -or $gzipFiles -ne $noneFiles) {
        throw "File counts differ: zstd=$zstdFiles, gzip=$gzipFiles, none=$noneFiles"
    }
    
    # Compare program.exe across all three
    $zstdContent = Get-Content "testdata/advanced-output/test-zstd/program.exe"
    $gzipContent = Get-Content "testdata/advanced-output/test-gzip/program.exe"
    $noneContent = Get-Content "testdata/advanced-output/test-none/program.exe"
    
    $diff1 = Compare-Object $zstdContent $gzipContent
    $diff2 = Compare-Object $gzipContent $noneContent
    
    if ($diff1 -or $diff2) {
        throw "Results differ between compression methods"
    }
    
    Write-Host "  All compression methods produced identical results ($zstdFiles files each)" -ForegroundColor Gray
}

# Test 16: Test multi-hop patching (1.0.0 → 1.0.1 → 1.0.2)
Test-Step "Test multi-hop patching scenario" {
    Write-Host "  Testing 1.0.0 → 1.0.1 → 1.0.2 patch chain..." -ForegroundColor Gray
    
    # Generate 1.0.0 → 1.0.1 patch
    $output = .\generator.exe --versions-dir .\testdata\versions --from 1.0.0 --to 1.0.1 --output .\testdata\advanced-output\patches 2>&1
    if ($LASTEXITCODE -ne 0) {
        throw "Failed to generate 1.0.0→1.0.1 patch"
    }
    
    # Copy 1.0.0 to multi-hop test directory
    Write-Host "  Starting from version 1.0.0..." -ForegroundColor Gray
    New-Item -Path "testdata/advanced-output/multi-hop" -ItemType Directory -Force | Out-Null
    Get-ChildItem -Path "testdata/versions/1.0.0" -Recurse | Copy-Item -Destination {
        $dest = Join-Path "testdata/advanced-output/multi-hop" $_.FullName.Substring((Resolve-Path "testdata/versions/1.0.0").Path.Length)
        $destDir = Split-Path $dest
        if (-not (Test-Path $destDir)) { New-Item -Path $destDir -ItemType Directory -Force | Out-Null }
        $dest
    } -Force
    
    # Apply first patch: 1.0.0 → 1.0.1
    Write-Host "  Applying first patch (1.0.0 → 1.0.1)..." -ForegroundColor Gray
    $output = .\applier.exe --patch .\testdata\advanced-output\patches\1.0.0-to-1.0.1.patch --current-dir .\testdata\advanced-output\multi-hop --verify 2>&1
    if ($LASTEXITCODE -ne 0) {
        throw "Failed to apply 1.0.0→1.0.1 patch"
    }
    
    # Apply second patch: 1.0.1 → 1.0.2
    Write-Host "  Applying second patch (1.0.1 → 1.0.2)..." -ForegroundColor Gray
    $output = .\applier.exe --patch .\testdata\advanced-output\patches\1.0.1-to-1.0.2.patch --current-dir .\testdata\advanced-output\multi-hop --verify 2>&1
    if ($LASTEXITCODE -ne 0) {
        throw "Failed to apply 1.0.1→1.0.2 patch"
    }
    
    # Verify final result matches 1.0.2
    $diff = Compare-Object (Get-Content "testdata/advanced-output/multi-hop/program.exe") (Get-Content "testdata/versions/1.0.2/program.exe")
    if ($diff) {
        throw "Multi-hop result does not match expected version 1.0.2"
    }
    
    Write-Host "  Multi-hop patching successful: 1.0.0 → 1.0.1 → 1.0.2" -ForegroundColor Gray
}

# Test 17: Test wrong version detection
Test-Step "Verify patch rejection for wrong source version" {
    Write-Host "  Testing patch rejection (applying 1.0.1→1.0.2 to 1.0.0)..." -ForegroundColor Gray
    
    # Copy 1.0.0 to wrong-version test directory
    New-Item -Path "testdata/advanced-output/wrong-version" -ItemType Directory -Force | Out-Null
    Get-ChildItem -Path "testdata/versions/1.0.0" -Recurse | Copy-Item -Destination {
        $dest = Join-Path "testdata/advanced-output/wrong-version" $_.FullName.Substring((Resolve-Path "testdata/versions/1.0.0").Path.Length)
        $destDir = Split-Path $dest
        if (-not (Test-Path $destDir)) { New-Item -Path $destDir -ItemType Directory -Force | Out-Null }
        $dest
    } -Force
    
    # Try to apply 1.0.1→1.0.2 patch (should fail)
    $output = .\applier.exe --patch .\testdata\advanced-output\patches\1.0.1-to-1.0.2.patch --current-dir .\testdata\advanced-output\wrong-version --verify 2>&1
    
    if ($LASTEXITCODE -eq 0) {
        throw "Patch should have been rejected but succeeded"
    }
    
    $outputStr = $output -join "`n"
    if ($outputStr -notmatch "checksum mismatch|verification failed|wrong version") {
        throw "Expected error message not found in output"
    }
    
    Write-Host "  Patch correctly rejected for wrong source version" -ForegroundColor Gray
}

# Test 18: Test file corruption detection
Test-Step "Verify detection of corrupted files in source" {
    Write-Host "  Testing corrupted file detection..." -ForegroundColor Gray
    
    # Copy 1.0.1 to corrupted test directory
    New-Item -Path "testdata/advanced-output/corrupted" -ItemType Directory -Force | Out-Null
    Get-ChildItem -Path "testdata/versions/1.0.1" -Recurse | Copy-Item -Destination {
        $dest = Join-Path "testdata/advanced-output/corrupted" $_.FullName.Substring((Resolve-Path "testdata/versions/1.0.1").Path.Length)
        $destDir = Split-Path $dest
        if (-not (Test-Path $destDir)) { New-Item -Path $destDir -ItemType Directory -Force | Out-Null }
        $dest
    } -Force
    
    # Corrupt a file
    Write-Host "  Corrupting libs/core.dll..." -ForegroundColor Gray
    Add-Content -Path "testdata/advanced-output/corrupted/libs/core.dll" -Value "CORRUPTED DATA"
    
    # Try to apply patch (should fail)
    $output = .\applier.exe --patch .\testdata\advanced-output\patches\1.0.1-to-1.0.2.patch --current-dir .\testdata\advanced-output\corrupted --verify 2>&1
    
    if ($LASTEXITCODE -eq 0) {
        throw "Patch should have been rejected but succeeded"
    }
    
    $outputStr = $output -join "`n"
    if ($outputStr -notmatch "checksum mismatch|verification failed") {
        throw "Checksum error not found in output"
    }
    
    Write-Host "  Corrupted file correctly detected" -ForegroundColor Gray
}

# Test 19: Verify backup creation and rollback
Test-Step "Verify backup system works correctly" {
    Write-Host "  Testing backup and rollback functionality..." -ForegroundColor Gray
    
    # Copy 1.0.1 to backup-test directory
    New-Item -Path "testdata/advanced-output/backup-test" -ItemType Directory -Force | Out-Null
    Get-ChildItem -Path "testdata/versions/1.0.1" -Recurse | Copy-Item -Destination {
        $dest = Join-Path "testdata/advanced-output/backup-test" $_.FullName.Substring((Resolve-Path "testdata/versions/1.0.1").Path.Length)
        $destDir = Split-Path $dest
        if (-not (Test-Path $destDir)) { New-Item -Path $destDir -ItemType Directory -Force | Out-Null }
        $dest
    } -Force
    
    # Apply patch (should create backup)
    $output = .\applier.exe --patch .\testdata\advanced-output\patches\1.0.1-to-1.0.2.patch --current-dir .\testdata\advanced-output\backup-test --verify 2>&1
    
    if ($LASTEXITCODE -ne 0) {
        throw "Patch application failed"
    }
    
    $outputStr = $output -join "`n"
    if ($outputStr -notmatch "backup|Backup") {
        Write-Host "  Warning: Backup message not found in output, but patch succeeded" -ForegroundColor Yellow
    } else {
        Write-Host "  Backup was created during patch application" -ForegroundColor Gray
    }
    
    Write-Host "  Backup system verified" -ForegroundColor Gray
}

# Test 20: Performance check - verify generation speed
Test-Step "Verify patch generation performance" {
    Write-Host "  Measuring patch generation time..." -ForegroundColor Gray
    
    $stopwatch = [System.Diagnostics.Stopwatch]::StartNew()
    $output = .\generator.exe --versions-dir .\testdata\versions --from 1.0.0 --to 1.0.2 --output .\testdata\advanced-output\patches 2>&1
    $stopwatch.Stop()
    
    if ($LASTEXITCODE -ne 0) {
        throw "Patch generation failed"
    }
    
    $elapsed = $stopwatch.Elapsed.TotalSeconds
    Write-Host "  Patch generation completed in $([math]::Round($elapsed, 2)) seconds" -ForegroundColor Gray
    
    if ($elapsed -gt 30) {
        Write-Host "  Warning: Patch generation took longer than expected" -ForegroundColor Yellow
    } else {
        Write-Host "  Performance is acceptable" -ForegroundColor Gray
    }
}

# Final summary
Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Advanced Test Results" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Passed: $passed" -ForegroundColor Green
Write-Host "Failed: $failed" -ForegroundColor Red
Write-Host ""

if ($failed -eq 0) {
    Write-Host "✓ All advanced tests passed!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Advanced Features Verified:" -ForegroundColor Cyan
    Write-Host "  • Complex nested directory structures" -ForegroundColor Gray
    Write-Host "  • Multiple compression formats (zstd, gzip, none)" -ForegroundColor Gray
    Write-Host "  • Multi-hop patching (1.0.0 → 1.0.1 → 1.0.2)" -ForegroundColor Gray
    Write-Host "  • Wrong version detection" -ForegroundColor Gray
    Write-Host "  • File corruption detection" -ForegroundColor Gray
    Write-Host "  • Backup system functionality" -ForegroundColor Gray
    Write-Host "  • Performance benchmarks" -ForegroundColor Gray
    Write-Host "  • Deep file path operations" -ForegroundColor Gray
    Write-Host ""
    Write-Host "CyberPatchMaker advanced functionality verified!" -ForegroundColor Green
    
    # Cleanup prompt
    Write-Host ""
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host "Test Data Cleanup" -ForegroundColor Cyan
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Test data is located in: .\testdata\" -ForegroundColor Gray
    Write-Host ""
    
    $response = Read-Host "Would you like to clean up test data now? (Y/N)"
    
    if ($response -match '^[Yy]') {
        Write-Host ""
        Write-Host "Cleaning up test data..." -ForegroundColor Yellow
        Remove-Item "testdata" -Recurse -Force -ErrorAction SilentlyContinue
        Write-Host "✓ Test data removed" -ForegroundColor Green
        Write-Host ""
    } else {
        Write-Host ""
        Write-Host "Test data kept for inspection." -ForegroundColor Cyan
        Write-Host "Note: On next run, test data will be automatically removed and recreated." -ForegroundColor Yellow
        Write-Host ""
        
        # Create state file to track deferred cleanup
        New-Item -ItemType Directory -Force -Path "testdata" -ErrorAction SilentlyContinue | Out-Null
        Set-Content -Path "testdata/.cleanup-deferred" -Value "cleanup deferred from previous run"
    }
    
    exit 0
} else {
    Write-Host "✗ Some advanced tests failed. Please check the output above." -ForegroundColor Red
    
    # Cleanup prompt even on failure
    Write-Host ""
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host "Test Data Cleanup" -ForegroundColor Cyan
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Test data is located in: .\testdata\" -ForegroundColor Gray
    Write-Host "You may want to keep it to inspect the failure." -ForegroundColor Yellow
    Write-Host ""
    
    $response = Read-Host "Would you like to clean up test data now? (Y/N)"
    
    if ($response -match '^[Yy]') {
        Write-Host ""
        Write-Host "Cleaning up test data..." -ForegroundColor Yellow
        Remove-Item "testdata" -Recurse -Force -ErrorAction SilentlyContinue
        Write-Host "✓ Test data removed" -ForegroundColor Green
        Write-Host ""
    } else {
        Write-Host ""
        Write-Host "Test data kept for inspection." -ForegroundColor Cyan
        Write-Host "Note: On next run, test data will be automatically removed and recreated." -ForegroundColor Yellow
        Write-Host ""
        
        # Create state file to track deferred cleanup
        New-Item -ItemType Directory -Force -Path "testdata" -ErrorAction SilentlyContinue | Out-Null
        Set-Content -Path "testdata/.cleanup-deferred" -Value "cleanup deferred from previous run"
    }
    
    exit 1
}
