# Quick Start Guide

Get up and running with CyberPatchMaker in 5 minutes.

## Prerequisites

- Go 1.24.0 installed
- Basic command-line knowledge

## Installation

### 1. Clone the Repository

```bash
git clone https://github.com/cyberofficial/CyberPatchMaker.git
cd CyberPatchMaker
```

### 2. Build the Tools

**Windows (PowerShell):**
```powershell
go build -o patch-gen.exe .\cmd\generator\
go build -o patch-apply.exe .\cmd\applier\
```

**Linux/macOS:**
```bash
go build -o patch-gen ./cmd/generator/
go build -o patch-apply ./cmd/applier/
```

### 3. Verify Installation

**Windows:**
```powershell
.\patch-gen.exe --help
.\patch-apply.exe --help
```

**Linux/macOS:**
```bash
./patch-gen --help
./patch-apply --help
```

## Your First Patch

### Scenario

You have two versions of your application:
- **Version 1.0.0**: 5GB application in `C:\MyApp\v1.0.0`
- **Version 1.0.1**: 5GB application in `C:\MyApp\v1.0.1`

You want to create a small patch file to upgrade 1.0.0 to 1.0.1.

### Step 1: Organize Your Versions

Create a versions directory:

```bash
mkdir versions
mkdir versions\1.0.0
mkdir versions\1.0.1
```

Copy your version folders into the structure:

```
versions/
├── 1.0.0/       # Copy your v1.0.0 here
│   ├── program.exe
│   └── ... (all your files)
└── 1.0.1/       # Copy your v1.0.1 here
    ├── program.exe
    └── ... (all your files)
```

### Step 2: Generate Patch

**Windows:**
```powershell
.\patch-gen.exe --versions-dir .\versions --new-version 1.0.1 --output .\patches
```

**Linux/macOS:**
```bash
./patch-gen --versions-dir ./versions --new-version 1.0.1 --output ./patches
```

This will:
- Scan version 1.0.1
- Auto-detect the key file (program.exe)
- Generate patch from 1.0.0 to 1.0.1
- Save to `patches/1.0.0-to-1.0.1.patch`

### Step 3: Test Patch (Dry-Run)

Create a test directory with version 1.0.0:

```bash
mkdir test-app
# Copy your 1.0.0 files to test-app
```

Run dry-run to see what would happen:

**Windows:**
```powershell
.\patch-apply.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir .\test-app --dry-run
```

**Linux/macOS:**
```bash
./patch-apply --patch ./patches/1.0.0-to-1.0.1.patch --current-dir ./test-app --dry-run
```

This shows you exactly what will be changed without modifying anything.

### Step 4: Apply Patch

If the dry-run looks good, apply the patch:

**Windows:**
```powershell
.\patch-apply.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir .\test-app --verify
```

**Linux/macOS:**
```bash
./patch-apply --patch ./patches/1.0.0-to-1.0.1.patch --current-dir ./test-app --verify
```

This will:
1. Check that you have version 1.0.0
2. Create a backup of your current files
3. Update the files to version 1.0.1
4. Check that everything updated correctly
5. Keep the backup for manual rollback if needed

### Step 5: Verify Success

Check your test-app directory - it should now be version 1.0.1!

## What's Next?

- Learn more about [Generator Tool](generator-guide.md)
- Learn more about [Applier Tool](applier-guide.md)
- Understand [How It Works](how-it-works.md)
- Read about [Backup System](backup-system.md)

## Common Issues

**"Key file not found"**: Make sure your application folder contains program.exe, game.exe, app.exe, or main.exe (the main program file)

**"Checksum mismatch"**: Your application files were changed after installation - patches only work on unmodified installations

**"Too many arguments"**: Make sure you're using the latest version of the tools

See [Troubleshooting](troubleshooting.md) for more help.

## Testing

Want to verify everything works? Run the comprehensive test suite:

**Windows:**
```powershell
.\advanced-test.ps1
```

This runs 59 automated tests to validate the entire system. Test data is automatically generated on first run, so there's no setup required!
