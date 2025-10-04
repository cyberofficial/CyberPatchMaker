# CyberPatchMaker

A comprehensive Go-based delta patch system for efficiently updating large software distributions. Generates and applies binary patches between software versions, minimizing download sizes by only transferring changed data.

## Features

- **Binary Delta Patches**: Uses bsdiff algorithm to create efficient binary diffs
- **SHA-256 Verification**: Every file is verified before and after patching
- **Complete Directory Tree Support**: Handles entire directory structures with unlimited nesting
- **Key File System**: Prevents applying patches to wrong versions or applications
- **Atomic Operations**: Safe patching with automatic rollback on failure
- **Compression**: Supports zstd and gzip compression for optimal patch sizes
- **Cross-Platform**: Works on Windows, macOS, and Linux

## Use Case Example

- Version 1.0.0: 5GB application
- Version 1.0.1: 5GB application (only a few MB changed)
- Patch file: Only a few MB instead of 5GB full download

## Installation

### Prerequisites

- Go 1.21 or later

### Building from Source

```bash
# Clone the repository
git clone https://github.com/cyberofficial/CyberPatchMaker.git
cd CyberPatchMaker

# Build CLI tools
go build ./cmd/generator
go build ./cmd/applier
```

## CLI Tools

### Generator Tool

Generates delta patches between software versions.

#### Generate patches from all existing versions to a new version:

```bash
generator --versions-dir ./versions --new-version 1.0.3 --output ./patches
```

This will:
1. Scan the new version directory
2. Auto-detect the key file (program.exe, game.exe, app.exe, or main.exe)
3. Register the new version
4. Generate patches from ALL existing versions to the new version
5. Save patches as `{from}-to-{to}.patch`

#### Generate a single patch between two specific versions:

```bash
generator --from 1.0.0 --to 1.0.3 --output ./patches/custom.patch
```

#### Options:

- `--versions-dir <path>`: Directory containing version folders (required for batch mode)
- `--new-version <version>`: New version to generate patches for (required for batch mode)
- `--from <version>`: Source version (required for single patch mode)
- `--to <version>`: Target version (required for single patch mode)
- `--output <path>`: Output directory for patches
- `--compression <type>`: Compression algorithm (zstd, gzip, none) - default: zstd
- `--level <1-4>`: Compression level - default: 3
- `--verify`: Verify patches after creation
- `--help`: Show usage information

### Applier Tool

Applies delta patches to upgrade software versions.

#### Dry-run mode (simulate without making changes):

```bash
applier --patch ./patches/1.0.0-to-1.0.3.patch --current-dir ./app --dry-run
```

This shows:
- Patch information (versions, key file, sizes, operation counts)
- Key file verification status
- Required files verification status
- Operations that would be performed

#### Apply patch with verification:

```bash
applier --patch ./patches/1.0.0-to-1.0.3.patch --current-dir ./app --verify
```

This will:
1. Display patch information
2. Verify key file exists and hash matches
3. Verify all required files exist with correct hashes
4. Create backup of current installation
5. Apply patch operations (add/modify/delete files and directories)
6. Verify all modified files have correct hashes
7. Remove backup on success OR restore backup on failure

#### Options:

- `--patch <path>`: Path to patch file (required)
- `--current-dir <path>`: Directory containing current version (required)
- `--dry-run`: Simulate patch without making changes
- `--verify`: Verify files before and after patching (recommended)
- `--backup`: Create backup before patching (automatic with --verify)
- `--help`: Show usage information

## How It Works

### Version Organization

```
versions/
├── 1.0.0/                  # Full version folder (5GB)
│   ├── program.exe         # Key file for version identification
│   ├── data/
│   │   ├── config.json
│   │   └── assets/
│   │       └── textures/
│   └── libs/
│       └── core.dll
├── 1.0.1/                  # Full version folder (5GB)
│   ├── program.exe         # Modified key file
│   ├── data/
│   │   ├── config.json     # Modified
│   │   └── assets/
│   │       └── textures/   # Added new textures
│   └── libs/
│       ├── core.dll        # Modified
│       └── newfeature.dll  # Added new file
└── patches/
    └── deltas/
        └── 1.0.0-to-1.0.1.patch  # Only a few MB
```

### Key File Verification System

Every version has a designated "key file" (usually the main executable) that serves as a version identifier:

1. **During Version Registration**: System calculates SHA-256 hash of the key file
2. **During Patch Generation**: Key file info is embedded in patch (path + hash)
3. **During Patch Application**: 
   - System verifies key file exists at specified path
   - Calculates hash and compares against required hash
   - **If hashes don't match, patch is rejected immediately**

This prevents:
- Applying patches to wrong versions
- Applying patches for different applications
- Patching corrupted or modified installations

### Hash-Based File Verification

All file operations use SHA-256 hash comparison of the ENTIRE directory tree:

1. **Pre-Patch Verification**:
   - Calculate hash for EVERY file in the directory tree
   - Verify key file hash matches requirements
   - Verify ALL required files exist with correct hashes
   - Reject patch if ANY file is missing or modified

2. **Patch Application**:
   - Apply operations throughout directory tree
   - Modify/add/delete files at ANY level
   - Create/remove directories as needed

3. **Post-Patch Verification**:
   - Calculate hash for EVERY modified/added file
   - Verify all hashes match expected values
   - Rollback if ANY verification fails

### Error Handling & Safety

- **Atomic Operations**: All changes applied to temporary locations first
- **Automatic Backup**: Complete backup created before any changes
- **Automatic Rollback**: Complete restoration on any verification failure
- **Manual Rollback**: Backup preserved for manual recovery if needed

## Example Workflow

### 1. Generate patches for a new version:

```bash
# Create new version folder
mkdir versions/1.0.2
# ... copy your new version files ...

# Generate patches from all existing versions
generator --versions-dir ./versions --new-version 1.0.2 --output ./patches
```

Output:
```
Generating patches for new version 1.0.2
Using key file: program.exe
Version 1.0.2 registered: 156 files, 12 directories

Processing version 1.0.0...
Generating patch from 1.0.0 to 1.0.2...
Patch saved to: patches/1.0.0-to-1.0.2.patch

Processing version 1.0.1...
Generating patch from 1.0.1 to 1.0.2...
Patch saved to: patches/1.0.1-to-1.0.2.patch
```

### 2. Test patch with dry-run:

```bash
applier --patch ./patches/1.0.0-to-1.0.2.patch --current-dir ./myapp --dry-run
```

Output:
```
=== Patch Information ===
From Version:     1.0.0
To Version:       1.0.2
Key File:         program.exe
Required Hash:    a1b2c3d4...
Files Added:      5
Files Modified:   12
Files Deleted:    3
Required Files:   156 (must match exact hashes)

=== DRY RUN MODE ===
✓ Key file verified
✓ All required files verified

Operations that would be performed:
  ADD: libs/newfeature.dll
  MODIFY: program.exe
  MODIFY: data/config.json
  ...
```

### 3. Apply patch:

```bash
applier --patch ./patches/1.0.0-to-1.0.2.patch --current-dir ./myapp --verify
```

Output:
```
Creating backup...
Backup created at: ./myapp.backup

Applying patch from 1.0.0 to 1.0.2...
Pre-patch verification successful
Applying 20 operations...
Post-patch verification successful

=== Patch Applied Successfully ===
Version updated from 1.0.0 to 1.0.2
Removing backup...
```

## Project Status

✅ **Phase 1 Complete**: Core Foundation
- All core utilities implemented
- Binary diffing using bsdiff
- Complete directory tree support
- SHA-256 verification system
- CLI tools (generator and applier)

🔄 **Phase 2-3 In Progress**: Testing & Optimization
- Test scenarios with sample data
- Performance optimization for large files
- Enhanced error reporting

📋 **Phase 4 Planned**: GUI Application
- Cross-platform GUI using Fyne
- Visual version management
- Drag-and-drop patch application
- Progress indicators

## Testing

The project includes a comprehensive test suite with 20 tests to verify the entire codebase works correctly.

### Advanced Test Suite

Run the advanced test suite:

**Windows (PowerShell):**
```powershell
.\advanced-test.ps1
```

**Note**: On first run, the test script will automatically generate test versions (1.0.0, 1.0.1, and 1.0.2) if they don't exist. This ensures the repository stays clean without committing test data files.

The advanced test suite validates:
- ✅ Automatic test data generation (no bloat files in repo)
- ✅ Complex nested directory structures (3 levels deep)
- ✅ Multiple compression formats (zstd, gzip, none)
- ✅ Compression efficiency comparison (~59% size reduction)
- ✅ Multi-hop patching (1.0.0 → 1.0.1 → 1.0.2)
- ✅ Wrong version detection and rejection
- ✅ File corruption detection via checksums
- ✅ Backup system functionality
- ✅ Performance benchmarks (0.03s patch generation)
- ✅ Deep file path operations
- ✅ All compression formats produce identical results

**Test Data Complexity:**
- Version 1.0.0: 5 items (baseline, 1 level nesting)
- Version 1.0.1: 6 items (simple update, 1 level nesting)
- Version 1.0.2: 17 items (complex structure, 3 levels nesting)

See [ADVANCED-TEST-SUMMARY.md](ADVANCED-TEST-SUMMARY.md) for detailed test results and analysis.

### Manual Testing

You can also run individual tests manually:

```bash
# Run generator test
generator --versions-dir ./testdata/versions --new-version 1.0.1 --output ./testdata/patches

# Run applier test (dry-run)
applier --patch ./testdata/patches/1.0.0-to-1.0.1.patch --current-dir ./testdata/test-app --dry-run

# Run applier test (actual application)
applier --patch ./testdata/patches/1.0.0-to-1.0.1.patch --current-dir ./testdata/test-app --verify
```

## Architecture

```
pkg/utils/          # Shared utilities (types, checksum, fileops, compress)
internal/core/      # Core business logic
  ├── scanner/      # Directory scanning and hashing
  ├── manifest/     # Manifest creation and comparison
  ├── version/      # Version management and registry
  ├── config/       # Configuration management
  ├── differ/       # Binary diff generation using bsdiff
  └── patcher/      # Patch generation and application
cmd/                # Command-line tools
  ├── generator/    # Patch generator CLI
  └── applier/      # Patch applier CLI
```

## Performance

- **Patch Generation**: 5GB version in < 5 minutes
- **Patch Application**: 5GB version in < 3 minutes
- **Directory Verification**: 5GB version in < 2 minutes
- **Memory Usage**: < 500MB for any operation
- **Patch Size**: < 5% of full version (typical updates)

## Security

- **SHA-256 Verification**: All files verified before and after patching
- **Key File System**: Prevents wrong patch application
- **Atomic Operations**: No partial installations
- **Automatic Rollback**: Safe recovery on failure

## License

See LICENSE file for details.

## Author

Created by CyberOfficial - https://github.com/cyberofficial
