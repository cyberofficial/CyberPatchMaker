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
Write-Info "[1/5] Building patch generator (CLI)..."
$generatorPath = Join-Path $distDir "patch-gen.exe"
& go build @buildFlags $generatorPath ./cmd/generator
if ($LASTEXITCODE -eq 0) {
    Write-Success "  ✓ patch-gen.exe"
} else {
    Write-Error "  ✗ Failed to build patch-gen.exe"
    exit 1
}

# Build CLI Applier
Write-Info "[2/5] Building patch applier (CLI)..."
$applierPath = Join-Path $distDir "patch-apply.exe"
& go build @buildFlags $applierPath ./cmd/applier
if ($LASTEXITCODE -eq 0) {
    Write-Success "  ✓ patch-apply.exe"
} else {
    Write-Error "  ✗ Failed to build patch-apply.exe"
    exit 1
}

# Build Generator GUI
Write-Info "[3/5] Building patch generator GUI..."
$genGuiPath = Join-Path $distDir "patch-gen-gui.exe"
& go build @buildFlags $genGuiPath ./cmd/patch-gui
if ($LASTEXITCODE -eq 0) {
    Write-Success "  ✓ patch-gen-gui.exe"
} else {
    Write-Error "  ✗ Failed to build patch-gen-gui.exe"
    exit 1
}

# Build Applier GUI
Write-Info "[4/5] Building patch applier GUI..."
$appGuiPath = Join-Path $distDir "patch-apply-gui.exe"
& go build @buildFlags $appGuiPath ./cmd/applier-gui
if ($LASTEXITCODE -eq 0) {
    Write-Success "  ✓ patch-apply-gui.exe"
} else {
    Write-Error "  ✗ Failed to build patch-apply-gui.exe"
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
Write-Info "  CLI Generator:      .\dist\patch-gen.exe --help"
Write-Info "  CLI Applier:        .\dist\patch-apply.exe --help"
Write-Info "  Generator GUI:      .\dist\patch-gen-gui.exe"
Write-Info "  Applier GUI:        .\dist\patch-apply-gui.exe"
Write-Info ""
