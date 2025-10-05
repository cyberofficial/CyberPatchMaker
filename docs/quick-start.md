# Quick Start Guide

Get up and running with CyberPatchMaker in 5 minutes.

## Prerequisites

- Go 1.21 or later installed
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
go build -o generator.exe .\cmd\generator\
go build -o applier.exe .\cmd\applier\
```

**Linux/macOS:**
```bash
go build -o generator ./cmd/generator/
go build -o applier ./cmd/applier/
```

### 3. Verify Installation

**Windows:**
```powershell
.\generator.exe --help
.\applier.exe --help
```

**Linux/macOS:**
```bash
./generator --help
./applier --help
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
.\generator.exe --versions-dir .\versions --new-version 1.0.1 --output .\patches
```

**Linux/macOS:**
```bash
./generator --versions-dir ./versions --new-version 1.0.1 --output ./patches
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
.\applier.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir .\test-app --dry-run
```

**Linux/macOS:**
```bash
./applier --patch ./patches/1.0.0-to-1.0.1.patch --current-dir ./test-app --dry-run
```

This shows you exactly what will be changed without modifying anything.

### Step 4: Apply Patch

If the dry-run looks good, apply the patch:

**Windows:**
```powershell
.\applier.exe --patch .\patches\1.0.0-to-1.0.1.patch --current-dir .\test-app --verify
```

**Linux/macOS:**
```bash
./applier --patch ./patches/1.0.0-to-1.0.1.patch --current-dir ./test-app --verify
```

This will:
1. Verify current version is 1.0.0
2. Create backup of current state
3. Apply the patch
4. Verify result matches version 1.0.1
5. Remove backup (or restore on failure)

### Step 5: Verify Success

Check your test-app directory - it should now be version 1.0.1!

## What's Next?

- Learn more about [Generator Tool](generator-guide.md)
- Learn more about [Applier Tool](applier-guide.md)
- Understand [How It Works](how-it-works.md)
- Read about [Safety Features](backup-rollback.md)

## Common Issues

**"Key file not found"**: Make sure your versions contain program.exe, game.exe, app.exe, or main.exe

**"Checksum mismatch"**: The target directory was modified - ensure you're applying the patch to a clean installation

**"Too many arguments"**: Check you're using the latest build of the tools

See [Troubleshooting](troubleshooting.md) for more help.

## Testing

Want to verify everything works? Run the comprehensive test suite:

**Windows:**
```powershell
.\advanced-test.ps1
```

This runs 28 automated tests to validate the entire system. Test data is automatically generated on first run, so there's no setup required!
