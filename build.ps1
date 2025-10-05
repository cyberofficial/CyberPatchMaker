#!/usr/bin/env pwsh
# CyberPatchMaker Build Script
# Builds CLI tools and GUI application

param(
    [switch]$Clean,
    [switch]$Verbose
)

$ErrorActionPreference = "Stop"

# Colors for output
function Write-Success { Write-Host $args -ForegroundColor Green }
function Write-Info { Write-Host $args -ForegroundColor Cyan }
function Write-Error { Write-Host $args -ForegroundColor Red }

Write-Info "=== CyberPatchMaker Build Script ==="
Write-Info ""

# Ensure TDM-GCC is in PATH for CGO (required for Fyne GUI)
if (Test-Path "C:\TDM-GCC-64\bin\gcc.exe") {
    $env:PATH = "C:\TDM-GCC-64\bin;" + $env:PATH
    Write-Info "✓ Using TDM-GCC for CGO compilation"
} else {
    Write-Info "⚠ TDM-GCC not found, using system GCC (may cause issues with GUI)"
}

# Create dist directory
$distDir = "dist"
if ($Clean -and (Test-Path $distDir)) {
    Write-Info "Cleaning dist directory..."
    Remove-Item -Recurse -Force $distDir
}

if (-not (Test-Path $distDir)) {
    New-Item -ItemType Directory -Path $distDir | Out-Null
    Write-Info "✓ Created dist directory"
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
Write-Info "[1/3] Building patch generator (CLI)..."
$generatorPath = Join-Path $distDir "patch-gen.exe"
& go build @buildFlags $generatorPath ./cmd/generator
if ($LASTEXITCODE -eq 0) {
    Write-Success "  ✓ patch-gen.exe"
} else {
    Write-Error "  ✗ Failed to build patch-gen.exe"
    exit 1
}

# Build CLI Applier
Write-Info "[2/3] Building patch applier (CLI)..."
$applierPath = Join-Path $distDir "patch-apply.exe"
& go build @buildFlags $applierPath ./cmd/applier
if ($LASTEXITCODE -eq 0) {
    Write-Success "  ✓ patch-apply.exe"
} else {
    Write-Error "  ✗ Failed to build patch-apply.exe"
    exit 1
}

# Build GUI
Write-Info "[3/3] Building patch GUI..."
$guiPath = Join-Path $distDir "patch-gui.exe"
& go build @buildFlags $guiPath ./cmd/patch-gui
if ($LASTEXITCODE -eq 0) {
    Write-Success "  ✓ patch-gui.exe"
} else {
    Write-Error "  ✗ Failed to build patch-gui.exe"
    exit 1
}

Write-Info ""
Write-Success "=== Build Complete ==="
Write-Info ""
Write-Info "Built files:"
Get-ChildItem $distDir -Filter *.exe | ForEach-Object {
    $size = "{0:N2} MB" -f ($_.Length / 1MB)
    Write-Info "  • $($_.Name) ($size)"
}

Write-Info ""
Write-Info "To run:"
Write-Info "  CLI Generator:  .\dist\patch-gen.exe --help"
Write-Info "  CLI Applier:    .\dist\patch-apply.exe --help"
Write-Info "  GUI:            .\dist\patch-gui.exe"
Write-Info ""
