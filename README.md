# CyberPatchMaker

**Update your software smarter, not harder.**

CyberPatchMaker creates tiny update files for large applications. Instead of downloading a whole 5GB app again, download just a few megabytes of changes.

## Why Use CyberPatchMaker?

Imagine you have a 5GB application. You make some small changes and release version 1.0.1. Without CyberPatchMaker, users download the entire 5GB again. With CyberPatchMaker, they download only a few MB patch file that contains just the changes.

**Real-world example:**
- Version 1.0.0: 5GB application
- Version 1.0.1: 5GB application (only a few files changed)
- **Traditional update:** Download 5GB again
- **With CyberPatchMaker:** Download 5MB patch file

Perfect for:
- Game developers releasing updates
- Software companies with large applications
- Anyone who wants to save bandwidth and user time

## Key Features

- **Safe & Reliable** - Automatic verification and rollback if anything goes wrong
- **Efficient** - Smart compression reduces patch sizes by ~60%
- **Bidirectional** - Generate both upgrade and downgrade patches for version flexibility
- **Cross-Platform** - Works on Windows, macOS, and Linux
- **Developer-Friendly** - Simple command-line tools
- **Production-Ready** - Built-in backup and recovery systems
- **Self-Contained Executables** - Create standalone `.exe` files for easy end-user distribution
- **Smart File Exclusion** - Use `.cyberignore` to exclude sensitive files and reduce patch size
- **Scan Caching** - Cache directory scans for instant patch generation (15+ min → <1 sec for large projects) **(New!)**

## Quick Start

### Installation

**Requirements:** Go 1.21 or later

```bash
# Clone and build
git clone https://github.com/cyberofficial/CyberPatchMaker.git
cd CyberPatchMaker
go build -o patch-gen ./cmd/generator
go build -o patch-apply ./cmd/applier
```

**Detailed setup:** See [Development Setup Guide](docs/development-setup.md)

## Basic Usage

### Creating a Patch

Generate patches for a new version of your software:

```bash
patch-gen --versions-dir ./versions --new-version 1.0.3 --output ./patches
```

This automatically creates patch files from all previous versions to version 1.0.3.

### Applying a Patch

Update your application with a patch file:

```bash
# Test first (dry-run)
patch-apply --patch ./patches/1.0.0-to-1.0.3.patch --current-dir ./app --dry-run

# Apply the update
patch-apply --patch ./patches/1.0.0-to-1.0.3.patch --current-dir ./app --verify
```

The `--verify` flag ensures everything is checked before and after patching, with automatic rollback if anything goes wrong.

**More examples and options:**
- [Generator Guide](docs/generator-guide.md) - All patch creation options
- [Applier Guide](docs/applier-guide.md) - All patch application options
- [CLI Reference](docs/cli-reference.md) - Complete command reference
- [CLI Examples](docs/CLI-EXAMPLES.md) - Common usage patterns
- [Downgrade Guide](docs/downgrade-guide.md) - Rollback to previous versions

### Excluding Files with .cyberignore (New!)

Control which files are included in patches using a `.cyberignore` file (similar to `.gitignore`):

```
:: Place in your version directory
:: Lines starting with :: are comments

:: Ignore sensitive files
*.key
*.crt
config/secrets.json

:: Ignore logs and temporary files
*.log
*.tmp
logs/

:: Ignore user data
saves/
user_config.json
```

The generator automatically excludes matching files from patches. Perfect for keeping API keys, certificates, and user data out of updates!

**Complete guide:** [.cyberignore File Guide](docs/cyberignore-guide.md)

### Downgrade Patches (Rolling Back)

Need to rollback to a previous version? CyberPatchMaker supports bidirectional patches:

```bash
# Generate downgrade patch
patch-gen --from 1.0.3 --to 1.0.2 --versions-dir ./versions --output ./patches/downgrade

# Apply downgrade to rollback
patch-apply --patch ./patches/downgrade/1.0.3-to-1.0.2.patch --current-dir ./app --verify
```

This allows users to safely revert to an earlier version if needed.

**Complete downgrade documentation:** [Downgrade Guide](docs/downgrade-guide.md)

## How It Works (Simple Version)

CyberPatchMaker compares two versions of your software and creates a small patch file containing only the differences. When users apply the patch, it safely updates their installation with built-in verification and rollback protection.

**The process:**
1. **Generate:** Compare old version with new version → Create small patch file
2. **Distribute:** Share the tiny patch file instead of the full application
3. **Apply:** Users run the patch → Their software updates safely

**Safety features:**
- Verifies files before patching (catches corruption early)
- Creates automatic backup
- Verifies files after patching
- Automatic rollback if anything fails

**Want technical details?**
- [How It Works](docs/how-it-works.md) - Deep dive into the internals
- [Architecture](docs/architecture.md) - System design and components
- [Hash Verification](docs/hash-verification.md) - Security and verification
- [Key File System](docs/key-file-system.md) - Version identification
- [Backup Lifecycle](docs/backup-lifecycle.md) - Backup and recovery process

## Complete Example

Here's a typical workflow from start to finish:

```bash
# 1. You have a new version ready
mkdir versions/1.0.2
# (copy your new version files into versions/1.0.2/)

# 2. Generate patch files
patch-gen --versions-dir ./versions --new-version 1.0.2 --output ./patches
# Creates: patches/1.0.0-to-1.0.2.patch
#          patches/1.0.1-to-1.0.2.patch

# 3. Distribute patch files to users
# (upload to your website, CDN, etc.)

# 4. Users test the patch first (optional but recommended)
patch-apply --patch 1.0.0-to-1.0.2.patch --current-dir ./myapp --dry-run

# 5. Users apply the patch
patch-apply --patch 1.0.0-to-1.0.2.patch --current-dir ./myapp --verify
# Done! Their version 1.0.0 is now version 1.0.2
```

**More examples:** [CLI Examples](docs/CLI-EXAMPLES.md)

## What's Included

**Production-Ready CLI Tools:**
- **Patch Generator** (`patch-gen.exe` / `patch-gen`) - Create update files
- **Patch Applier** (`patch-apply.exe` / `patch-apply`) - Install updates
- Comprehensive verification and automatic rollback
- Tested with complex directory structures
- Handles files from 1KB to 5GB+
- Multiple compression formats (zstd, gzip)

**Experimental GUI (In Development):**
- Basic graphical interface available for patch generation
- Not yet recommended for production use
- CLI tools are the primary, supported interface

**Documentation:** [Full Documentation Index](docs/README.md)

## Testing

Want to verify everything works? Run the test suite:

```powershell
# Windows PowerShell
.\advanced-test.ps1
```

The test suite automatically validates:
- Patch generation and application
- Multiple compression formats
- Wrong version detection
- File corruption detection
- Backup and rollback systems
- Complex directory structures

**Testing documentation:**
- [Testing Guide](docs/testing-guide.md) - How to test your patches
- [Advanced Test Summary](docs/ADVANCED-TEST-SUMMARY.md) - Detailed test results

## Performance & Reliability

**Fast:**
- Generate patches for 5GB apps in under 5 minutes
- Apply patches in under 3 minutes
- Low memory usage (< 500MB)

**Small:**
- Typical patches are < 5% of full app size
- Smart compression reduces size by ~60%

**Safe:**
- Every file verified before and after patching
- Automatic backup and rollback
- Prevents applying wrong patches

**Technical details:**
- [Architecture](docs/architecture.md) - System design
- [Compression Guide](docs/compression-guide.md) - Compression options
- [Version Management](docs/version-management.md) - How versions are tracked

## Documentation

All documentation is in the [docs/](docs/) folder:

**Getting Started:**
- [Quick Start](docs/quick-start.md) - Get up and running fast
- [CLI Examples](docs/CLI-EXAMPLES.md) - Common usage patterns
- [Development Setup](docs/development-setup.md) - Developer guide

**Reference:**
- [Generator Guide](docs/generator-guide.md) - Creating patches
- [Applier Guide](docs/applier-guide.md) - Applying patches
- [Self-Contained Executables](docs/self-contained-executables.md) - Standalone patch distribution
- [CLI Reference](docs/cli-reference.md) - All commands and options

**Advanced:**
- [How It Works](docs/how-it-works.md) - Technical deep dive
- [Architecture](docs/architecture.md) - System design
- [Testing Guide](docs/testing-guide.md) - Testing your patches
- [Troubleshooting](docs/troubleshooting.md) - Common issues

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

## License

See [LICENSE](LICENSE) file for details.

## Author

Created by [CyberOfficial](https://github.com/cyberofficial)

---

**Like this project?** Give it a star on GitHub!

## Support the project

If you'd like to support continued development, consider sponsoring on GitHub: [Sponsor Me](https://github.com/sponsors/cyberofficial)