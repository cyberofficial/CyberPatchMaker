#!/usr/bin/env pwsh
# CyberPatchMaker Build Script
# Builds CLI tools (GUI components removed as deprecated)

param(
    [switch]$Clean,
    [switch]$Verbose,
    [switch]$i,   # Increment patch version
    [switch]$ii,  # Increment minor version (resets patch to 0)
    [switch]$iii  # Increment major version (resets minor and patch to 0)
)

$ErrorActionPreference = "Stop"

# Colors for output
function Write-Success { Write-Host $args -ForegroundColor Green }
function Write-Info { Write-Host $args -ForegroundColor Cyan }
function Write-Error { Write-Host $args -ForegroundColor Red }

Write-Info "=== CyberPatchMaker Build Script ==="
Write-Info ""

# Handle version increment if requested
$versionFile = "internal\core\version\version.go"

if ($iii) {
    Write-Info "Incrementing major version (resetting minor and patch to 0)..."
    
    # Read version file
    $content = Get-Content $versionFile -Raw
    
    # Extract current major number
    if ($content -match 'Major = (\d+)') {
        $currentMajor = [int]$matches[1]
        $newMajor = $currentMajor + 1
        
        # Replace major number
        $content = $content -replace "Major = $currentMajor", "Major = $newMajor"
        
        # Reset minor and patch to 0
        $content = $content -replace 'Minor = \d+', 'Minor = 0'
        $content = $content -replace 'Patch = \d+', 'Patch = 0'
        
        # Write back to file
        Set-Content $versionFile -Value $content -NoNewline
        
        Write-Success "[OK] Incremented major version: $currentMajor -> $newMajor (minor and patch reset to 0)"
    } else {
        Write-Error "Failed to parse major version from $versionFile"
        exit 1
    }
    Write-Info ""
}
elseif ($ii) {
    Write-Info "Incrementing minor version (resetting patch to 0)..."
    
    # Read version file
    $content = Get-Content $versionFile -Raw
    
    # Extract current minor number
    if ($content -match 'Minor = (\d+)') {
        $currentMinor = [int]$matches[1]
        $newMinor = $currentMinor + 1
        
        # Replace minor number
        $content = $content -replace "Minor = $currentMinor", "Minor = $newMinor"
        
        # Reset patch to 0
        $content = $content -replace 'Patch = \d+', 'Patch = 0'
        
        # Write back to file
        Set-Content $versionFile -Value $content -NoNewline
        
        Write-Success "[OK] Incremented minor version: $currentMinor -> $newMinor (patch reset to 0)"
    } else {
        Write-Error "Failed to parse minor version from $versionFile"
        exit 1
    }
    Write-Info ""
}
elseif ($i) {
    Write-Info "Incrementing patch version..."
    
    # Read version file
    $content = Get-Content $versionFile -Raw
    
    # Extract current patch number
    if ($content -match 'Patch = (\d+)') {
        $currentPatch = [int]$matches[1]
        $newPatch = $currentPatch + 1
        
        # Replace patch number
        $content = $content -replace "Patch = $currentPatch", "Patch = $newPatch"
        
        # Write back to file
        Set-Content $versionFile -Value $content -NoNewline
        
        Write-Success "[OK] Incremented patch version: $currentPatch -> $newPatch"
    } else {
        Write-Error "Failed to parse patch version from $versionFile"
        exit 1
    }
    Write-Info ""
}

# Get current version for directory naming
$versionContent = Get-Content $versionFile -Raw
$major = if ($versionContent -match 'Major = (\d+)') { $matches[1] } else { "0" }
$minor = if ($versionContent -match 'Minor = (\d+)') { $matches[1] } else { "0" }
$patch = if ($versionContent -match 'Patch = (\d+)') { $matches[1] } else { "0" }
$version = "$major.$minor.$patch"

Write-Info "Building version: $version"
Write-Info ""

# Ensure TDM-GCC is in PATH for CGO (required for Fyne GUI)
if (Test-Path "C:\TDM-GCC-64\bin\gcc.exe") {
    $env:PATH = "C:\TDM-GCC-64\bin;" + $env:PATH
    Write-Info "[OK] Using TDM-GCC for CGO compilation"
} else {
    Write-Info "[WARN] TDM-GCC not found, using system GCC (may cause issues with GUI)"
}

# Create dist directory with version subdirectory
$distDir = "dist"
$versionDir = Join-Path $distDir $version

if ($Clean -and (Test-Path $distDir)) {
    Write-Info "Cleaning dist directory..."
    Remove-Item -Recurse -Force $distDir
}

if (-not (Test-Path $distDir)) {
    New-Item -ItemType Directory -Path $distDir | Out-Null
}

if (-not (Test-Path $versionDir)) {
    New-Item -ItemType Directory -Path $versionDir | Out-Null
    Write-Info "[OK] Created version directory: $versionDir"
}

# Build flags
$buildFlags = @("-o")
if ($Verbose) {
    $buildFlags += "-v"
}

Write-Info ""
Write-Info "Building components..."
Write-Info ""

# Build CLI Generator
Write-Info "[1/2] Building patch generator (CLI)..."
$generatorPath = Join-Path $versionDir "patch-gen.exe"
& go build @buildFlags $generatorPath ./cmd/generator
if ($LASTEXITCODE -eq 0) {
    Write-Success "  [OK] patch-gen.exe"
} else {
    Write-Error "  [FAIL] Failed to build patch-gen.exe"
    exit 1
}

# Build CLI Applier
Write-Info "[2/2] Building patch applier (CLI)..."
$applierPath = Join-Path $versionDir "patch-apply.exe"
& go build @buildFlags $applierPath ./cmd/applier
if ($LASTEXITCODE -eq 0) {
    Write-Success "  [OK] patch-apply.exe"
} else {
    Write-Error "  [FAIL] Failed to build patch-apply.exe"
    exit 1
}

Write-Info ""
Write-Success "=== Build Complete ==="
Write-Info ""
Write-Info "Version: $version"
Write-Info "Built files in $versionDir :"
Get-ChildItem $versionDir -Filter *.exe | ForEach-Object {
    $size = "{0:N2} MB" -f ($_.Length / 1MB)
    Write-Info "  - $($_.Name) ($size)"
}

Write-Info ""
Write-Info "To run:"
Write-Info "  CLI Generator:      .\dist\$version\patch-gen.exe --help"
Write-Info "  CLI Applier:        .\dist\$version\patch-apply.exe --help"
Write-Info ""
