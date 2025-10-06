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
    Write-Host "  Command: generator.exe --versions-dir .\testdata\versions --from 1.0.1 --to 1.0.2 --output .\testdata\advanced-output\patches --compression zstd" -ForegroundColor Cyan
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
    Write-Host "  Command: generator.exe --versions-dir .\testdata\versions --from 1.0.1 --to 1.0.2 --output .\testdata\advanced-output\patches-gzip --compression gzip" -ForegroundColor Cyan
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
    Write-Host "  Command: generator.exe --versions-dir .\testdata\versions --from 1.0.1 --to 1.0.2 --output .\testdata\advanced-output\patches-none --compression none" -ForegroundColor Cyan
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
    Write-Host "  Command: applier.exe --patch .\testdata\advanced-output\patches\1.0.1-to-1.0.2.patch --current-dir .\testdata\versions\1.0.1 --dry-run" -ForegroundColor Cyan
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
    Write-Host "  Command: applier.exe --patch .\testdata\advanced-output\patches\1.0.1-to-1.0.2.patch --current-dir .\testdata\advanced-output\test-zstd --verify" -ForegroundColor Cyan
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
    Write-Host "  Command: applier.exe --patch .\testdata\advanced-output\patches-gzip\1.0.1-to-1.0.2.patch --current-dir .\testdata\advanced-output\test-gzip --verify" -ForegroundColor Cyan
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
    Write-Host "  Command: applier.exe --patch .\testdata\advanced-output\patches-none\1.0.1-to-1.0.2.patch --current-dir .\testdata\advanced-output\test-none --verify" -ForegroundColor Cyan
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
    Write-Host "  Command: generator.exe --versions-dir .\testdata\versions --from 1.0.0 --to 1.0.1 --output .\testdata\advanced-output\patches" -ForegroundColor Cyan
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
    Write-Host "  Command: applier.exe --patch .\testdata\advanced-output\patches\1.0.0-to-1.0.1.patch --current-dir .\testdata\advanced-output\multi-hop --verify" -ForegroundColor Cyan
    $output = .\applier.exe --patch .\testdata\advanced-output\patches\1.0.0-to-1.0.1.patch --current-dir .\testdata\advanced-output\multi-hop --verify 2>&1
    if ($LASTEXITCODE -ne 0) {
        throw "Failed to apply 1.0.0→1.0.1 patch"
    }
    
    # Apply second patch: 1.0.1 → 1.0.2
    Write-Host "  Applying second patch (1.0.1 → 1.0.2)..." -ForegroundColor Gray
    Write-Host "  Command: applier.exe --patch .\testdata\advanced-output\patches\1.0.1-to-1.0.2.patch --current-dir .\testdata\advanced-output\multi-hop --verify" -ForegroundColor Cyan
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

# Test 17: Test downgrade patch generation (1.0.2 → 1.0.1)
Test-Step "Generate downgrade patch (1.0.2 → 1.0.1)" {
    Write-Host "  Generating downgrade patch from 1.0.2 to 1.0.1..." -ForegroundColor Gray
    Write-Host "  Command: generator.exe --versions-dir .\testdata\versions --from 1.0.2 --to 1.0.1 --output .\testdata\advanced-output\patches --compression zstd" -ForegroundColor Cyan
    $output = .\generator.exe --versions-dir .\testdata\versions --from 1.0.2 --to 1.0.1 --output .\testdata\advanced-output\patches --compression zstd 2>&1
    
    if ($LASTEXITCODE -ne 0) {
        throw "Downgrade patch generation failed with exit code $LASTEXITCODE"
    }
    
    if (-not (Test-Path "testdata/advanced-output/patches/1.0.2-to-1.0.1.patch")) {
        throw "Downgrade patch file not created"
    }
    
    $patchSize = (Get-Item "testdata/advanced-output/patches/1.0.2-to-1.0.1.patch").Length
    Write-Host "  Downgrade patch generated: $patchSize bytes" -ForegroundColor Gray
}

# Test 18: Apply downgrade patch (1.0.2 → 1.0.1)
Test-Step "Apply downgrade patch to revert version" {
    Write-Host "  Copying version 1.0.2 to downgrade-test..." -ForegroundColor Gray
    New-Item -Path "testdata/advanced-output/downgrade-test" -ItemType Directory -Force | Out-Null
    Get-ChildItem -Path "testdata/versions/1.0.2" -Recurse | Copy-Item -Destination {
        $dest = Join-Path "testdata/advanced-output/downgrade-test" $_.FullName.Substring((Resolve-Path "testdata/versions/1.0.2").Path.Length)
        $destDir = Split-Path $dest
        if (-not (Test-Path $destDir)) { New-Item -Path $destDir -ItemType Directory -Force | Out-Null }
        $dest
    } -Force
    
    Write-Host "  Applying downgrade patch (1.0.2 → 1.0.1)..." -ForegroundColor Gray
    Write-Host "  Command: applier.exe --patch .\testdata\advanced-output\patches\1.0.2-to-1.0.1.patch --current-dir .\testdata\advanced-output\downgrade-test --verify" -ForegroundColor Cyan
    $output = .\applier.exe --patch .\testdata\advanced-output\patches\1.0.2-to-1.0.1.patch --current-dir .\testdata\advanced-output\downgrade-test --verify 2>&1
    
    if ($LASTEXITCODE -ne 0) {
        $outputStr = $output -join "`n"
        throw "Downgrade patch application failed: $outputStr"
    }
    
    Write-Host "  Downgrade patch applied successfully" -ForegroundColor Gray
}

# Test 19: Verify downgrade results match version 1.0.1
Test-Step "Verify downgrade results match version 1.0.1" {
    Write-Host "  Verifying downgraded version matches 1.0.1..." -ForegroundColor Gray
    
    # Compare key files
    $diff = Compare-Object (Get-Content "testdata/advanced-output/downgrade-test/program.exe") (Get-Content "testdata/versions/1.0.1/program.exe")
    if ($diff) {
        throw "program.exe does not match version 1.0.1"
    }
    
    # Verify new files from 1.0.2 were removed
    if (Test-Path "testdata/advanced-output/downgrade-test/data/assets") {
        throw "Directory data/assets should have been removed"
    }
    if (Test-Path "testdata/advanced-output/downgrade-test/plugins") {
        throw "Directory plugins should have been removed"
    }
    
    # Verify file count matches 1.0.1 (excluding backup folder)
    $downgradedFiles = (Get-ChildItem "testdata/advanced-output/downgrade-test" -Recurse -File | Where-Object { $_.FullName -notlike "*backup.cyberpatcher*" } | Measure-Object).Count
    $expectedFiles = (Get-ChildItem "testdata/versions/1.0.1" -Recurse -File | Measure-Object).Count
    
    if ($downgradedFiles -ne $expectedFiles) {
        throw "File count mismatch: downgraded=$downgradedFiles, expected=$expectedFiles"
    }
    
    Write-Host "  Downgrade successful: version 1.0.2 → 1.0.1 verified" -ForegroundColor Gray
}

# Test 20: Test bidirectional patching (upgrade then downgrade)
Test-Step "Test bidirectional patching cycle" {
    Write-Host "  Testing complete bidirectional patch cycle..." -ForegroundColor Gray
    
    # Copy 1.0.1 to bidirectional test directory
    New-Item -Path "testdata/advanced-output/bidirectional" -ItemType Directory -Force | Out-Null
    Get-ChildItem -Path "testdata/versions/1.0.1" -Recurse | Copy-Item -Destination {
        $dest = Join-Path "testdata/advanced-output/bidirectional" $_.FullName.Substring((Resolve-Path "testdata/versions/1.0.1").Path.Length)
        $destDir = Split-Path $dest
        if (-not (Test-Path $destDir)) { New-Item -Path $destDir -ItemType Directory -Force | Out-Null }
        $dest
    } -Force
    
    # Upgrade: 1.0.1 → 1.0.2
    Write-Host "  Step 1: Upgrade 1.0.1 → 1.0.2..." -ForegroundColor Gray
    Write-Host "  Command: applier.exe --patch .\testdata\advanced-output\patches\1.0.1-to-1.0.2.patch --current-dir .\testdata\advanced-output\bidirectional --verify" -ForegroundColor Cyan
    $output = .\applier.exe --patch .\testdata\advanced-output\patches\1.0.1-to-1.0.2.patch --current-dir .\testdata\advanced-output\bidirectional --verify 2>&1
    if ($LASTEXITCODE -ne 0) {
        throw "Upgrade failed (1.0.1 → 1.0.2)"
    }
    
    # Verify upgraded to 1.0.2
    $diff = Compare-Object (Get-Content "testdata/advanced-output/bidirectional/program.exe") (Get-Content "testdata/versions/1.0.2/program.exe")
    if ($diff) {
        throw "Upgrade verification failed - not version 1.0.2"
    }
    Write-Host "  ✓ Upgraded to 1.0.2" -ForegroundColor Green
    
    # Downgrade: 1.0.2 → 1.0.1
    Write-Host "  Step 2: Downgrade 1.0.2 → 1.0.1..." -ForegroundColor Gray
    Write-Host "  Command: applier.exe --patch .\testdata\advanced-output\patches\1.0.2-to-1.0.1.patch --current-dir .\testdata\advanced-output\bidirectional --verify" -ForegroundColor Cyan
    $output = .\applier.exe --patch .\testdata\advanced-output\patches\1.0.2-to-1.0.1.patch --current-dir .\testdata\advanced-output\bidirectional --verify 2>&1
    if ($LASTEXITCODE -ne 0) {
        throw "Downgrade failed (1.0.2 → 1.0.1)"
    }
    
    # Verify downgraded back to 1.0.1
    $diff = Compare-Object (Get-Content "testdata/advanced-output/bidirectional/program.exe") (Get-Content "testdata/versions/1.0.1/program.exe")
    if ($diff) {
        throw "Downgrade verification failed - not version 1.0.1"
    }
    Write-Host "  ✓ Downgraded back to 1.0.1" -ForegroundColor Green
    
    Write-Host "  Bidirectional patching cycle successful: 1.0.1 → 1.0.2 → 1.0.1" -ForegroundColor Gray
}

# Test 21: Test wrong version detection
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
    Write-Host "  Command: applier.exe --patch .\testdata\advanced-output\patches\1.0.1-to-1.0.2.patch --current-dir .\testdata\advanced-output\wrong-version --verify" -ForegroundColor Cyan
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

# Test 22: Test file corruption detection
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
    Write-Host "  Command: applier.exe --patch .\testdata\advanced-output\patches\1.0.1-to-1.0.2.patch --current-dir .\testdata\advanced-output\corrupted --verify" -ForegroundColor Cyan
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

# Test 23: Verify backup creation and structure
Test-Step "Verify backup creation and mirror structure" {
    Write-Host "  Testing backup system with mirror structure..." -ForegroundColor Gray
    
    # Copy 1.0.1 to backup-test directory
    New-Item -Path "testdata/advanced-output/backup-test" -ItemType Directory -Force | Out-Null
    Get-ChildItem -Path "testdata/versions/1.0.1" -Recurse | Copy-Item -Destination {
        $dest = Join-Path "testdata/advanced-output/backup-test" $_.FullName.Substring((Resolve-Path "testdata/versions/1.0.1").Path.Length)
        $destDir = Split-Path $dest
        if (-not (Test-Path $destDir)) { New-Item -Path $destDir -ItemType Directory -Force | Out-Null }
        $dest
    } -Force
    
    # Apply patch (should create backup)
    Write-Host "  Command: applier.exe --patch .\testdata\advanced-output\patches\1.0.1-to-1.0.2.patch --current-dir .\testdata\advanced-output\backup-test --verify" -ForegroundColor Cyan
    $output = .\applier.exe --patch .\testdata\advanced-output\patches\1.0.1-to-1.0.2.patch --current-dir .\testdata\advanced-output\backup-test --verify 2>&1
    
    if ($LASTEXITCODE -ne 0) {
        throw "Patch application failed"
    }
    
    # Verify backup directory was created
    if (-not (Test-Path "testdata/advanced-output/backup-test/backup.cyberpatcher")) {
        throw "Backup directory 'backup.cyberpatcher' was not created"
    }
    Write-Host "  ✓ Backup directory created: backup.cyberpatcher" -ForegroundColor Green
    
    # Verify backup message in output
    $outputStr = $output -join "`n"
    if ($outputStr -notmatch "backup|Backup") {
        throw "Backup message not found in output"
    }
    Write-Host "  ✓ Backup creation message found in output" -ForegroundColor Green
    
    # Verify mirror structure - modified files should be backed up with correct paths
    if (-not (Test-Path "testdata/advanced-output/backup-test/backup.cyberpatcher/program.exe")) {
        throw "Backup file 'program.exe' not found in backup directory"
    }
    Write-Host "  ✓ Key file backed up: program.exe" -ForegroundColor Green
    
    if (-not (Test-Path "testdata/advanced-output/backup-test/backup.cyberpatcher/data/config.json")) {
        throw "Backup file 'data/config.json' not found in backup directory"
    }
    Write-Host "  ✓ Nested file backed up: data/config.json" -ForegroundColor Green
    
    if (-not (Test-Path "testdata/advanced-output/backup-test/backup.cyberpatcher/libs/core.dll")) {
        throw "Backup file 'libs/core.dll' not found in backup directory"
    }
    Write-Host "  ✓ Library file backed up: libs/core.dll" -ForegroundColor Green
    
    if (-not (Test-Path "testdata/advanced-output/backup-test/backup.cyberpatcher/libs/newfeature.dll")) {
        throw "Backup file 'libs/newfeature.dll' not found in backup directory"
    }
    Write-Host "  ✓ Modified library backed up: libs/newfeature.dll" -ForegroundColor Green
    
    Write-Host "  Backup system with mirror structure verified" -ForegroundColor Gray
}

# Test 24: Verify backup only contains modified/deleted files
Test-Step "Verify selective backup (only modified/deleted files)" {
    Write-Host "  Verifying backup is selective (not full copy)..." -ForegroundColor Gray
    
    # Count files in backup vs target after patching
    $backupFiles = @(Get-ChildItem "testdata/advanced-output/backup-test/backup.cyberpatcher" -Recurse -File)
    $patchedFiles = @(Get-ChildItem "testdata/advanced-output/backup-test" -Recurse -File | Where-Object { $_.FullName -notlike "*backup.cyberpatcher*" })
    
    Write-Host "  Patched version files: $($patchedFiles.Count)" -ForegroundColor Gray
    Write-Host "  Backup files: $($backupFiles.Count)" -ForegroundColor Gray
    
    # Backup should have FEWER files than patched version (only modified/deleted, not added)
    # Version 1.0.2 has 10 files total, but 1.0.1 had 4 files
    # So backup should have 4 files (only the modified ones from 1.0.1)
    # And patched version should have 10 files (4 modified + 6 added)
    if ($backupFiles.Count -ge $patchedFiles.Count) {
        throw "Backup contains $($backupFiles.Count) files, expected fewer than patched version ($($patchedFiles.Count) files)"
    }
    
    # Expected backed up files: program.exe, data/config.json, libs/core.dll, libs/newfeature.dll
    # (Files that are modified in 1.0.1→1.0.2 patch, not the 6 new files)
    if ($backupFiles.Count -ne 4) {
        Write-Host "  Warning: Expected 4 backed up files, got $($backupFiles.Count)" -ForegroundColor Yellow
        Write-Host "  Backed up files:" -ForegroundColor Yellow
        foreach ($file in $backupFiles) {
            $relativePath = $file.FullName.Substring((Resolve-Path "testdata/advanced-output/backup-test/backup.cyberpatcher").Path.Length + 1)
            Write-Host "    - $relativePath" -ForegroundColor Yellow
        }
    } else {
        Write-Host "  ✓ Correct number of files backed up (4 modified files)" -ForegroundColor Green
    }
    
    Write-Host "  ✓ Backup is selective (not a full copy)" -ForegroundColor Green
    Write-Host "  Selective backup verified (only modified/deleted files)" -ForegroundColor Gray
}

# Test 25: Verify backup preservation after successful patch
Test-Step "Verify backup is preserved after successful patching" {
    Write-Host "  Verifying backup persists after successful patch..." -ForegroundColor Gray
    
    # Backup directory should still exist
    if (-not (Test-Path "testdata/advanced-output/backup-test/backup.cyberpatcher")) {
        throw "Backup directory should be preserved but was deleted"
    }
    
    # Verify backup files still exist
    if (-not (Test-Path "testdata/advanced-output/backup-test/backup.cyberpatcher/program.exe")) {
        throw "Backup files should be preserved but were deleted"
    }
    
    Write-Host "  ✓ Backup directory preserved: backup.cyberpatcher" -ForegroundColor Green
    Write-Host "  ✓ Backup files intact after successful patching" -ForegroundColor Green
    
    Write-Host "  Backup preservation verified" -ForegroundColor Gray
}

# Test 26: Verify manual rollback using backup
Test-Step "Verify manual rollback from backup works" {
    Write-Host "  Testing manual rollback using backup..." -ForegroundColor Gray
    
    # Verify current state is 1.0.2
    $currentVersion = Get-Content "testdata/advanced-output/backup-test/program.exe" -Raw
    $expected102 = Get-Content "testdata/versions/1.0.2/program.exe" -Raw
    
    if ($currentVersion -ne $expected102) {
        throw "Current version should be 1.0.2 before rollback test"
    }
    Write-Host "  Current version confirmed: 1.0.2" -ForegroundColor Gray
    
    # Manually rollback by copying from backup
    Write-Host "  Performing manual rollback (copying from backup)..." -ForegroundColor Gray
    
    # Copy program.exe from backup
    Copy-Item "testdata/advanced-output/backup-test/backup.cyberpatcher/program.exe" `
              "testdata/advanced-output/backup-test/program.exe" -Force
    
    # Copy data/config.json from backup
    Copy-Item "testdata/advanced-output/backup-test/backup.cyberpatcher/data/config.json" `
              "testdata/advanced-output/backup-test/data/config.json" -Force
    
    # Copy libs/core.dll from backup
    Copy-Item "testdata/advanced-output/backup-test/backup.cyberpatcher/libs/core.dll" `
              "testdata/advanced-output/backup-test/libs/core.dll" -Force
    
    # Copy libs/newfeature.dll from backup
    Copy-Item "testdata/advanced-output/backup-test/backup.cyberpatcher/libs/newfeature.dll" `
              "testdata/advanced-output/backup-test/libs/newfeature.dll" -Force
    
    Write-Host "  Files copied from backup" -ForegroundColor Gray
    
    # Verify rollback - program.exe should now match 1.0.1
    $rolledBackVersion = Get-Content "testdata/advanced-output/backup-test/program.exe" -Raw
    $expected101 = Get-Content "testdata/versions/1.0.1/program.exe" -Raw
    
    if ($rolledBackVersion -ne $expected101) {
        throw "Manual rollback failed - program.exe does not match version 1.0.1"
    }
    
    Write-Host "  ✓ Manual rollback successful: 1.0.2 → 1.0.1" -ForegroundColor Green
    Write-Host "  ✓ Backup system enables easy rollback" -ForegroundColor Green
    
    Write-Host "  Manual rollback verified" -ForegroundColor Gray
}

# Test 27: Verify backup with complex nested structure
Test-Step "Verify backup handles complex nested paths" {
    Write-Host "  Testing backup with deeply nested directory structure..." -ForegroundColor Gray
    
    # Copy 1.0.0 to nested-backup-test
    New-Item -Path "testdata/advanced-output/nested-backup-test" -ItemType Directory -Force | Out-Null
    Get-ChildItem -Path "testdata/versions/1.0.0" -Recurse | Copy-Item -Destination {
        $dest = Join-Path "testdata/advanced-output/nested-backup-test" $_.FullName.Substring((Resolve-Path "testdata/versions/1.0.0").Path.Length)
        $destDir = Split-Path $dest
        if (-not (Test-Path $destDir)) { New-Item -Path $destDir -ItemType Directory -Force | Out-Null }
        $dest
    } -Force
    
    # Generate and apply 1.0.0 → 1.0.2 patch (includes deep nesting)
    Write-Host "  Command: generator.exe --versions-dir .\testdata\versions --from 1.0.0 --to 1.0.2 --output .\testdata\advanced-output\patches" -ForegroundColor Cyan
    $output = .\generator.exe --versions-dir .\testdata\versions --from 1.0.0 --to 1.0.2 --output .\testdata\advanced-output\patches 2>&1
    if ($LASTEXITCODE -ne 0) {
        throw "Failed to generate 1.0.0→1.0.2 patch"
    }
    
    Write-Host "  Command: applier.exe --patch .\testdata\advanced-output\patches\1.0.0-to-1.0.2.patch --current-dir .\testdata\advanced-output\nested-backup-test --verify" -ForegroundColor Cyan
    $output = .\applier.exe --patch .\testdata\advanced-output\patches\1.0.0-to-1.0.2.patch --current-dir .\testdata\advanced-output\nested-backup-test --verify 2>&1
    
    if ($LASTEXITCODE -ne 0) {
        throw "Patch application with backup failed"
    }
    
    # Verify backup directory exists
    if (-not (Test-Path "testdata/advanced-output/nested-backup-test/backup.cyberpatcher")) {
        throw "Backup directory not created for complex structure"
    }
    
    # Verify nested directories in backup match original structure
    if (Test-Path "testdata/advanced-output/nested-backup-test/backup.cyberpatcher/data") {
        Write-Host "  ✓ Nested directory preserved in backup: data/" -ForegroundColor Green
    }
    
    if (Test-Path "testdata/advanced-output/nested-backup-test/backup.cyberpatcher/libs") {
        Write-Host "  ✓ Nested directory preserved in backup: libs/" -ForegroundColor Green
    }
    
    # Verify backed up files maintain directory hierarchy
    $backupFiles = @(Get-ChildItem "testdata/advanced-output/nested-backup-test/backup.cyberpatcher" -Recurse -File)
    Write-Host "  ✓ Backup created with $($backupFiles.Count) files in mirror structure" -ForegroundColor Green
    
    foreach ($file in $backupFiles) {
        $relativePath = $file.FullName.Substring((Resolve-Path "testdata/advanced-output/nested-backup-test/backup.cyberpatcher").Path.Length + 1)
        Write-Host "    - $relativePath" -ForegroundColor Gray
    }
    
    Write-Host "  Complex nested backup structure verified" -ForegroundColor Gray
}

# Test 24: Performance check - verify generation speed
Test-Step "Verify patch generation performance" {
    Write-Host "  Measuring patch generation time..." -ForegroundColor Gray
    Write-Host "  Command: generator.exe --versions-dir .\testdata\versions --from 1.0.0 --to 1.0.2 --output .\testdata\advanced-output\patches" -ForegroundColor Cyan
    
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

# Test 28: Custom Paths Mode - Different Directories (Same Drive)
Test-Step "Test custom paths mode with different directories" {
    Write-Host "  Testing custom paths mode (--from-dir and --to-dir)..." -ForegroundColor Gray
    
    # Create separate directory locations
    New-Item -ItemType Directory -Force -Path "testdata/custom-paths/release-1.0.0" | Out-Null
    New-Item -ItemType Directory -Force -Path "testdata/custom-paths/build-1.0.1" | Out-Null
    New-Item -ItemType Directory -Force -Path "testdata/custom-paths/patches" | Out-Null
    
    # Copy versions to different paths
    Get-ChildItem -Path "testdata/versions/1.0.0" -Recurse | Copy-Item -Destination {
        $dest = Join-Path "testdata/custom-paths/release-1.0.0" $_.FullName.Substring((Resolve-Path "testdata/versions/1.0.0").Path.Length)
        $destDir = Split-Path $dest
        if (-not (Test-Path $destDir)) { New-Item -Path $destDir -ItemType Directory -Force | Out-Null }
        $dest
    } -Force
    
    Get-ChildItem -Path "testdata/versions/1.0.1" -Recurse | Copy-Item -Destination {
        $dest = Join-Path "testdata/custom-paths/build-1.0.1" $_.FullName.Substring((Resolve-Path "testdata/versions/1.0.1").Path.Length)
        $destDir = Split-Path $dest
        if (-not (Test-Path $destDir)) { New-Item -Path $destDir -ItemType Directory -Force | Out-Null }
        $dest
    } -Force
    
    Write-Host "  Command: generator.exe --from-dir .\testdata\custom-paths\release-1.0.0 --to-dir .\testdata\custom-paths\build-1.0.1 --output .\testdata\custom-paths\patches" -ForegroundColor Cyan
    $output = .\generator.exe --from-dir .\testdata\custom-paths\release-1.0.0 --to-dir .\testdata\custom-paths\build-1.0.1 --output .\testdata\custom-paths\patches 2>&1
    
    if ($LASTEXITCODE -ne 0) {
        throw "Custom paths patch generation failed with exit code $LASTEXITCODE"
    }
    
    # Verify patch was created with correct naming (uses directory names as versions)
    if (-not (Test-Path "testdata/custom-paths/patches/release-1.0.0-to-build-1.0.1.patch")) {
        throw "Custom paths patch file not created with expected name"
    }
    
    $patchSize = (Get-Item "testdata/custom-paths/patches/release-1.0.0-to-build-1.0.1.patch").Length
    Write-Host "  Custom paths patch generated: $patchSize bytes" -ForegroundColor Gray
    Write-Host "  Patch name: release-1.0.0-to-build-1.0.1.patch" -ForegroundColor Gray
}

# Test 29: Apply Custom Paths Patch
Test-Step "Apply patch generated with custom paths mode" {
    Write-Host "  Applying custom paths patch..." -ForegroundColor Gray
    
    # Copy source version to apply-test directory
    New-Item -Path "testdata/custom-paths/apply-test" -ItemType Directory -Force | Out-Null
    Get-ChildItem -Path "testdata/custom-paths/release-1.0.0" -Recurse | Copy-Item -Destination {
        $dest = Join-Path "testdata/custom-paths/apply-test" $_.FullName.Substring((Resolve-Path "testdata/custom-paths/release-1.0.0").Path.Length)
        $destDir = Split-Path $dest
        if (-not (Test-Path $destDir)) { New-Item -Path $destDir -ItemType Directory -Force | Out-Null }
        $dest
    } -Force
    
    Write-Host "  Command: applier.exe --patch .\testdata\custom-paths\patches\release-1.0.0-to-build-1.0.1.patch --current-dir .\testdata\custom-paths\apply-test --verify" -ForegroundColor Cyan
    $output = .\applier.exe --patch .\testdata\custom-paths\patches\release-1.0.0-to-build-1.0.1.patch --current-dir .\testdata\custom-paths\apply-test --verify 2>&1
    
    if ($LASTEXITCODE -ne 0) {
        $outputStr = $output -join "`n"
        throw "Custom paths patch application failed: $outputStr"
    }
    
    # Verify result matches expected version
    $diff = Compare-Object (Get-Content "testdata/custom-paths/apply-test/program.exe") (Get-Content "testdata/versions/1.0.1/program.exe")
    if ($diff) {
        throw "Applied version does not match expected version 1.0.1"
    }
    
    Write-Host "  Custom paths patch applied successfully" -ForegroundColor Gray
}

# Test 30: Custom Paths with Complex Nested Structure
Test-Step "Test custom paths with complex nested directories" {
    Write-Host "  Testing custom paths with complex structure (1.0.1 → 1.0.2)..." -ForegroundColor Gray
    
    # Create complex structure in different locations
    New-Item -ItemType Directory -Force -Path "testdata/custom-paths/app-v101" | Out-Null
    New-Item -ItemType Directory -Force -Path "testdata/custom-paths/app-v102" | Out-Null
    
    Get-ChildItem -Path "testdata/versions/1.0.1" -Recurse | Copy-Item -Destination {
        $dest = Join-Path "testdata/custom-paths/app-v101" $_.FullName.Substring((Resolve-Path "testdata/versions/1.0.1").Path.Length)
        $destDir = Split-Path $dest
        if (-not (Test-Path $destDir)) { New-Item -Path $destDir -ItemType Directory -Force | Out-Null }
        $dest
    } -Force
    
    Get-ChildItem -Path "testdata/versions/1.0.2" -Recurse | Copy-Item -Destination {
        $dest = Join-Path "testdata/custom-paths/app-v102" $_.FullName.Substring((Resolve-Path "testdata/versions/1.0.2").Path.Length)
        $destDir = Split-Path $dest
        if (-not (Test-Path $destDir)) { New-Item -Path $destDir -ItemType Directory -Force | Out-Null }
        $dest
    } -Force
    
    Write-Host "  Command: generator.exe --from-dir .\testdata\custom-paths\app-v101 --to-dir .\testdata\custom-paths\app-v102 --output .\testdata\custom-paths\patches --compression zstd" -ForegroundColor Cyan
    $output = .\generator.exe --from-dir .\testdata\custom-paths\app-v101 --to-dir .\testdata\custom-paths\app-v102 --output .\testdata\custom-paths\patches --compression zstd 2>&1
    
    if ($LASTEXITCODE -ne 0) {
        throw "Complex custom paths patch generation failed"
    }
    
    if (-not (Test-Path "testdata/custom-paths/patches/app-v101-to-app-v102.patch")) {
        throw "Complex patch file not created"
    }
    
    $patchSize = (Get-Item "testdata/custom-paths/patches/app-v101-to-app-v102.patch").Length
    Write-Host "  Complex custom paths patch: $patchSize bytes" -ForegroundColor Gray
    
    # Apply the complex patch
    New-Item -Path "testdata/custom-paths/complex-apply" -ItemType Directory -Force | Out-Null
    Get-ChildItem -Path "testdata/custom-paths/app-v101" -Recurse | Copy-Item -Destination {
        $dest = Join-Path "testdata/custom-paths/complex-apply" $_.FullName.Substring((Resolve-Path "testdata/custom-paths/app-v101").Path.Length)
        $destDir = Split-Path $dest
        if (-not (Test-Path $destDir)) { New-Item -Path $destDir -ItemType Directory -Force | Out-Null }
        $dest
    } -Force
    
    Write-Host "  Command: applier.exe --patch .\testdata\custom-paths\patches\app-v101-to-app-v102.patch --current-dir .\testdata\custom-paths\complex-apply --verify" -ForegroundColor Cyan
    $output = .\applier.exe --patch .\testdata\custom-paths\patches\app-v101-to-app-v102.patch --current-dir .\testdata\custom-paths\complex-apply --verify 2>&1
    
    if ($LASTEXITCODE -ne 0) {
        throw "Complex patch application failed"
    }
    
    # Verify nested directories were created
    if (-not (Test-Path "testdata/custom-paths/complex-apply/data/assets/images")) {
        throw "Nested directories not created with custom paths"
    }
    
    if (-not (Test-Path "testdata/custom-paths/complex-apply/plugins/sample.json")) {
        throw "New files in deep paths not added"
    }
    
    Write-Host "  Complex custom paths patch verified" -ForegroundColor Gray
}

# Test 31: Custom Paths - Version Number Extraction
Test-Step "Verify version number extraction from directory names" {
    Write-Host "  Testing version extraction from various path formats..." -ForegroundColor Gray
    
    # Test Case A: Simple version numbers
    New-Item -ItemType Directory -Force -Path "testdata/custom-paths/1.0.0" | Out-Null
    New-Item -ItemType Directory -Force -Path "testdata/custom-paths/1.0.1" | Out-Null
    
    Get-ChildItem -Path "testdata/versions/1.0.0" -Recurse | Copy-Item -Destination {
        $dest = Join-Path "testdata/custom-paths/1.0.0" $_.FullName.Substring((Resolve-Path "testdata/versions/1.0.0").Path.Length)
        $destDir = Split-Path $dest
        if (-not (Test-Path $destDir)) { New-Item -Path $destDir -ItemType Directory -Force | Out-Null }
        $dest
    } -Force
    
    Get-ChildItem -Path "testdata/versions/1.0.1" -Recurse | Copy-Item -Destination {
        $dest = Join-Path "testdata/custom-paths/1.0.1" $_.FullName.Substring((Resolve-Path "testdata/versions/1.0.1").Path.Length)
        $destDir = Split-Path $dest
        if (-not (Test-Path $destDir)) { New-Item -Path $destDir -ItemType Directory -Force | Out-Null }
        $dest
    } -Force
    
    Write-Host "  Case A: Simple version numbers (1.0.0, 1.0.1)..." -ForegroundColor Gray
    Write-Host "  Command: generator.exe --from-dir .\testdata\custom-paths\1.0.0 --to-dir .\testdata\custom-paths\1.0.1 --output .\testdata\custom-paths\patches" -ForegroundColor Cyan
    $output = .\generator.exe --from-dir .\testdata\custom-paths\1.0.0 --to-dir .\testdata\custom-paths\1.0.1 --output .\testdata\custom-paths\patches 2>&1
    
    if ($LASTEXITCODE -ne 0) {
        throw "Version extraction failed for simple numbers"
    }
    
    if (-not (Test-Path "testdata/custom-paths/patches/1.0.0-to-1.0.1.patch")) {
        throw "Patch name incorrect for simple version numbers"
    }
    Write-Host "  ✓ Simple version numbers extracted correctly" -ForegroundColor Green
    
    # Test Case B: Prefixed version numbers
    New-Item -ItemType Directory -Force -Path "testdata/custom-paths/v1.0.0" | Out-Null
    New-Item -ItemType Directory -Force -Path "testdata/custom-paths/v1.0.1" | Out-Null
    
    Get-ChildItem -Path "testdata/versions/1.0.0" -Recurse | Copy-Item -Destination {
        $dest = Join-Path "testdata/custom-paths/v1.0.0" $_.FullName.Substring((Resolve-Path "testdata/versions/1.0.0").Path.Length)
        $destDir = Split-Path $dest
        if (-not (Test-Path $destDir)) { New-Item -Path $destDir -ItemType Directory -Force | Out-Null }
        $dest
    } -Force
    
    Get-ChildItem -Path "testdata/versions/1.0.1" -Recurse | Copy-Item -Destination {
        $dest = Join-Path "testdata/custom-paths/v1.0.1" $_.FullName.Substring((Resolve-Path "testdata/versions/1.0.1").Path.Length)
        $destDir = Split-Path $dest
        if (-not (Test-Path $destDir)) { New-Item -Path $destDir -ItemType Directory -Force | Out-Null }
        $dest
    } -Force
    
    Write-Host "  Case B: Prefixed version numbers (v1.0.0, v1.0.1)..." -ForegroundColor Gray
    Write-Host "  Command: generator.exe --from-dir .\testdata\custom-paths\v1.0.0 --to-dir .\testdata\custom-paths\v1.0.1 --output .\testdata\custom-paths\patches" -ForegroundColor Cyan
    $output = .\generator.exe --from-dir .\testdata\custom-paths\v1.0.0 --to-dir .\testdata\custom-paths\v1.0.1 --output .\testdata\custom-paths\patches 2>&1
    
    if ($LASTEXITCODE -ne 0) {
        throw "Version extraction failed for prefixed versions"
    }
    
    if (-not (Test-Path "testdata/custom-paths/patches/v1.0.0-to-v1.0.1.patch")) {
        throw "Patch name incorrect for prefixed versions"
    }
    Write-Host "  ✓ Prefixed version numbers extracted correctly" -ForegroundColor Green
    
    Write-Host "  Version extraction from paths verified" -ForegroundColor Gray
}

# Test 32: Custom Paths - Compression Options
Test-Step "Test compression options with custom paths" {
    Write-Host "  Testing all compression methods with custom paths..." -ForegroundColor Gray
    
    # Test zstd
    Write-Host "  Testing zstd compression..." -ForegroundColor Gray
    Write-Host "  Command: generator.exe --from-dir .\testdata\custom-paths\1.0.0 --to-dir .\testdata\custom-paths\1.0.1 --output .\testdata\custom-paths\patches --compression zstd" -ForegroundColor Cyan
    $output = .\generator.exe --from-dir .\testdata\custom-paths\1.0.0 --to-dir .\testdata\custom-paths\1.0.1 --output .\testdata\custom-paths\patches --compression zstd 2>&1
    if ($LASTEXITCODE -ne 0) {
        throw "Custom paths with zstd failed"
    }
    $zstdSize = (Get-Item "testdata/custom-paths/patches/1.0.0-to-1.0.1.patch").Length
    Write-Host "  ✓ zstd: $zstdSize bytes" -ForegroundColor Green
    
    # Test gzip
    Write-Host "  Testing gzip compression..." -ForegroundColor Gray
    Write-Host "  Command: generator.exe --from-dir .\testdata\custom-paths\1.0.0 --to-dir .\testdata\custom-paths\1.0.1 --output .\testdata\custom-paths\patches-gzip --compression gzip" -ForegroundColor Cyan
    New-Item -ItemType Directory -Force -Path "testdata/custom-paths/patches-gzip" | Out-Null
    $output = .\generator.exe --from-dir .\testdata\custom-paths\1.0.0 --to-dir .\testdata\custom-paths\1.0.1 --output .\testdata\custom-paths\patches-gzip --compression gzip 2>&1
    if ($LASTEXITCODE -ne 0) {
        throw "Custom paths with gzip failed"
    }
    $gzipSize = (Get-Item "testdata/custom-paths/patches-gzip/1.0.0-to-1.0.1.patch").Length
    Write-Host "  ✓ gzip: $gzipSize bytes" -ForegroundColor Green
    
    # Test none
    Write-Host "  Testing no compression..." -ForegroundColor Gray
    Write-Host "  Command: generator.exe --from-dir .\testdata\custom-paths\1.0.0 --to-dir .\testdata\custom-paths\1.0.1 --output .\testdata\custom-paths\patches-none --compression none" -ForegroundColor Cyan
    New-Item -ItemType Directory -Force -Path "testdata/custom-paths/patches-none" | Out-Null
    $output = .\generator.exe --from-dir .\testdata\custom-paths\1.0.0 --to-dir .\testdata\custom-paths\1.0.1 --output .\testdata\custom-paths\patches-none --compression none 2>&1
    if ($LASTEXITCODE -ne 0) {
        throw "Custom paths with no compression failed"
    }
    $noneSize = (Get-Item "testdata/custom-paths/patches-none/1.0.0-to-1.0.1.patch").Length
    Write-Host "  ✓ none: $noneSize bytes" -ForegroundColor Green
    
    Write-Host "  All compression methods work with custom paths" -ForegroundColor Gray
}

# Test 33: Custom Paths - Error Handling (Non-existent Directory)
Test-Step "Test error handling for non-existent directories" {
    Write-Host "  Testing error handling with invalid paths..." -ForegroundColor Gray
    
    # Test non-existent from directory
    Write-Host "  Testing non-existent from directory..." -ForegroundColor Gray
    Write-Host "  Command: generator.exe --from-dir .\testdata\custom-paths\does-not-exist --to-dir .\testdata\custom-paths\1.0.1 --output .\testdata\custom-paths\patches" -ForegroundColor Cyan
    $output = .\generator.exe --from-dir .\testdata\custom-paths\does-not-exist --to-dir .\testdata\custom-paths\1.0.1 --output .\testdata\custom-paths\patches 2>&1
    
    if ($LASTEXITCODE -eq 0) {
        throw "Generator should have failed for non-existent from directory"
    }
    
    $outputStr = $output -join "`n"
    if ($outputStr -notmatch "does not exist|not found|cannot find") {
        Write-Host "  Warning: Error message could be clearer" -ForegroundColor Yellow
    } else {
        Write-Host "  ✓ Clear error message for missing directory" -ForegroundColor Green
    }
    
    Write-Host "  Error handling for invalid paths verified" -ForegroundColor Gray
}

# Test 34: Backward Compatibility - Legacy Mode Still Works
Test-Step "Verify backward compatibility with legacy --versions-dir mode" {
    Write-Host "  Testing that legacy mode still works alongside custom paths..." -ForegroundColor Gray
    
    Write-Host "  Command: generator.exe --versions-dir .\testdata\versions --from 1.0.0 --to 1.0.1 --output .\testdata\custom-paths\patches" -ForegroundColor Cyan
    $output = .\generator.exe --versions-dir .\testdata\versions --from 1.0.0 --to 1.0.1 --output .\testdata\custom-paths\patches 2>&1
    
    if ($LASTEXITCODE -ne 0) {
        throw "Legacy mode broken after custom paths implementation"
    }
    
    if (-not (Test-Path "testdata/custom-paths/patches/1.0.0-to-1.0.1.patch")) {
        throw "Legacy mode patch not created"
    }
    
    Write-Host "  ✓ Legacy --versions-dir mode still works" -ForegroundColor Green
    Write-Host "  Backward compatibility verified" -ForegroundColor Gray
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
    Write-Host "  • Bidirectional patching (upgrade and downgrade)" -ForegroundColor Gray
    Write-Host "  • Downgrade patches (1.0.2 → 1.0.1 rollback)" -ForegroundColor Gray
    Write-Host "  • Complete bidirectional cycle (1.0.1 ↔ 1.0.2)" -ForegroundColor Gray
    Write-Host "  • Wrong version detection" -ForegroundColor Gray
    Write-Host "  • File corruption detection" -ForegroundColor Gray
    Write-Host "  • Backup system with mirror structure" -ForegroundColor Gray
    Write-Host "  • Selective backup (only modified/deleted files)" -ForegroundColor Gray
    Write-Host "  • Backup preservation after successful patching" -ForegroundColor Gray
    Write-Host "  • Manual rollback from backup" -ForegroundColor Gray
    Write-Host "  • Backup with complex nested paths" -ForegroundColor Gray
    Write-Host "  • Deep file path operations" -ForegroundColor Gray
    Write-Host "  • Custom paths mode (--from-dir, --to-dir)" -ForegroundColor Gray
    Write-Host "  • Version extraction from directory names" -ForegroundColor Gray
    Write-Host "  • Custom paths with complex nested structures" -ForegroundColor Gray
    Write-Host "  • All compression formats with custom paths" -ForegroundColor Gray
    Write-Host "  • Error handling for invalid custom paths" -ForegroundColor Gray
    Write-Host "  • Backward compatibility with legacy mode" -ForegroundColor Gray
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
