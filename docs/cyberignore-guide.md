# .cyberignore File Guide

## Overview

The `.cyberignore` file allows you to exclude specific files and directories from patch generation. This is useful for:

- **Excluding sensitive data**: API keys, certificates, passwords
- **Ignoring temporary files**: Logs, cache files, build artifacts
- **Skipping development files**: IDE configurations, test data
- **Reducing patch size**: Exclude large files that shouldn't be distributed

## How It Works

When generating a patch, CyberPatchMaker automatically looks for a `.cyberignore` file in the **source version directory** (the `--from-dir` or version directory being scanned). If found, patterns listed in the file will be excluded from scanning.

### Automatic Exclusions

The following are **always excluded** automatically:
- `backup.cyberpatcher/` - Backup directory created by patch application
- `.cyberignore` - The ignore file itself

## File Format

### Basic Syntax

```
:: This is a comment (lines starting with ::)
pattern_to_ignore
another_pattern
folder_to_ignore/
```

### Comments

Lines starting with `::` are treated as comments and ignored:

```
:: Ignore all secret keys
*.key

:: Ignore configuration files
config/secrets.json
```

### Empty Lines

Empty lines are ignored and can be used to improve readability.

## Pattern Types

### 1. Exact File Names

Ignore specific files by name:

```
secret.key
passwords.txt
api_keys.json
```

Matches:
- `secret.key` in root directory
- Does NOT match `other_secret.key` or `folder/secret.key`

### 2. Wildcard Patterns

Use `*` to match any characters:

```
*.key
*.log
*.tmp
*.bak
```

Matches:
- `app.key`, `database.key`, `any.key`
- `debug.log`, `error.log`, `application.log`
- Works with any file extension

### 3. Directory Paths

Ignore entire directories and their contents:

```
logs/
temp/
test_data/
.vscode/
```

**Important**: 
- Include trailing `/` for directories
- All files within the directory are excluded
- Subdirectories are also excluded

### 4. Nested Paths

Ignore specific files in nested locations:

```
config/secrets.json
data/temp/cache.db
build/intermediate/
```

**Path Format**:
- Use forward slashes `/` (preferred)
- Backslashes `\` are automatically converted to `/`
- Both `config/secrets.json` and `config\secrets.json` work

### 5. Absolute Path Patterns

Ignore files using full absolute paths (Windows-style with drive letters):

```
E:\projects\myapp\*.key
C:\temp\logs\*.log
D:\shared\secrets\*
```

**Absolute Path Features**:
- Must include drive letter (e.g., `C:`, `D:`, `E:`)
- Use backslashes `\` or forward slashes `/` (both work)
- Supports wildcards (`*`) for flexible matching
- Case-insensitive matching on Windows
- Useful for excluding files outside the project directory

**Examples**:
```
:: Exclude all .key files on E: drive
E:\projects\myapp\*.key

:: Exclude temp logs from specific directory
C:\temp\build\logs\*.log

:: Exclude entire temp directory on D: drive
D:\temp\*
```

## Common Use Cases

### Example 1: Web Application

```
:: .cyberignore for web application

:: Sensitive files
*.key
*.pem
*.crt
.env
.env.local
config/secrets.json

:: Logs and temporary files
*.log
*.tmp
logs/
temp/

:: Development files
node_modules/
.vscode/
.idea/
```

### Example 2: Game Application

```
:: .cyberignore for game

:: User data (don't override user settings)
saves/
config/user_settings.ini
player_data.db

:: Logs and cache
*.log
cache/
screenshots/temp/

:: Development assets
*.psd
*.blend
dev_assets/
```

### Example 3: Desktop Application

```
:: .cyberignore for desktop app

:: Sensitive data
license_keys.txt
*.p12
*.pfx

:: User configurations
user_config.xml
recent_files.json

:: Temporary files
*.tmp
*.bak
~*

:: Build artifacts
*.obj
*.pdb
debug_symbols/
```

### Example 4: Application with External Dependencies

```
:: .cyberignore for app with external files

:: Local project files (relative paths)
*.log
temp/
config/secrets.json

:: External temp directories (absolute paths)
C:\temp\build\*
E:\shared\cache\logs\*.log

:: Network drive exclusions
\\server\share\temp\*
Z:\network\backups\*
```

## Behavior Details

### Pattern Matching

1. **Exact Match**: `secret.key` matches only `secret.key` in root
2. **Wildcard in Root**: `*.log` matches any `.log` file in root
3. **Wildcard Recursive**: `*.log` also matches `.log` files in subdirectories
4. **Directory Match**: `logs/` matches entire `logs/` directory tree
5. **Nested Path**: `config/secrets.json` matches only that specific file
6. **Absolute Path Match**: `E:\temp\*.log` matches `.log` files in `E:\temp\` directory
7. **Absolute Wildcard**: `C:\shared\*` matches all files in `C:\shared\` directory

### Path Normalization

- Paths are normalized to forward slashes `/` internally
- Both `folder/file.txt` and `folder\file.txt` in `.cyberignore` work identically
- Windows and Unix path styles are supported
- Absolute paths are converted to consistent format for matching

### Case Sensitivity

- Pattern matching is **case-sensitive** on case-sensitive file systems (Linux/macOS)
- Pattern matching is **case-insensitive** on case-insensitive file systems (Windows)
- This applies to both relative and absolute path patterns

## Best Practices

### 1. Place in Source Directory

Put `.cyberignore` in your **source version directory** (not the target):

```
versions/
  1.0.0/
    .cyberignore    ← Place here
    app.exe
    data.txt
    secret.key
  1.0.1/
    app.exe
    data.txt
```

### 2. Document Your Patterns

Use comments to explain why files are ignored:

```
:: User-specific files that should never be overwritten
user_settings.json
recent_files.db

:: Temporary files that regenerate on launch
*.tmp
cache/
```

### 3. Test Your Patterns

Generate a test patch and check the output:

```powershell
.\patch-gen.exe --from-dir .\versions\1.0.0 --to-dir .\versions\1.0.1 --output .\patches
```

Look for the file count in the output:
```
Version 1.0.0 registered: 4 files, 1 directories
```

If the count is higher than expected, check your `.cyberignore` patterns.

### 4. Version-Specific Ignores

Each version can have its own `.cyberignore` file:

```
versions/
  1.0.0/
    .cyberignore    ← Ignores for 1.0.0
  1.0.1/
    .cyberignore    ← Different ignores for 1.0.1
```

This is useful when file structures change between versions.

### 5. Keep It Simple

Start with basic patterns and add more as needed:

```
:: Start simple
*.log
*.tmp
*.bak

:: Add specific files as you discover them
user_data.db
config/private.json
```

## Verification

### Check What's Ignored

To verify your `.cyberignore` is working:

1. Run patch generation with `--compression none` for faster testing
2. Check the "registered" file count in the output
3. Compare with actual file count in the directory

**Example**:

```powershell
# Total files in directory
Get-ChildItem .\versions\1.0.0 -Recurse -File | Measure-Object

# Run patch generator
.\patch-gen.exe --from-dir .\versions\1.0.0 --to-dir .\versions\1.0.1 --output .\patches

# Output shows: "Version 1.0.0 registered: 4 files"
```

If directory has 10 files but only 4 are registered, 6 files were ignored (including `.cyberignore` itself).

## Troubleshooting

### Pattern Not Working

1. **Check path format**: Use `/` not `\` in patterns
2. **Check for typos**: Pattern matching is exact
3. **Check wildcards**: `*.log` matches `app.log`, not `app.log.txt`
4. **Check directory slash**: `logs/` needs trailing `/` to match directory

### Too Many Files Ignored

1. Review patterns for wildcards that are too broad
2. Check if directory patterns are excluding more than intended
3. Use more specific paths: `temp/cache.db` instead of `*.db`

### File Still Included

1. Verify `.cyberignore` is in the **source directory** (--from-dir)
2. Check pattern syntax and spelling
3. Ensure no syntax errors (empty lines are OK, but malformed patterns are not)

## Technical Details

### Pattern Matching Algorithm

1. Exact path match: `path == pattern`
2. Directory prefix match: `path` starts with `pattern/`
3. Wildcard match using filepath.Match (Go standard library)
4. Extension match for `*.ext` patterns
5. **Absolute path match**: Full absolute paths are compared case-insensitively on Windows
6. **Absolute wildcard match**: Absolute paths with wildcards use enhanced pattern matching

### Performance Impact

- `.cyberignore` is loaded once at scan start
- Minimal performance overhead (simple string matching)
- Directory exclusions improve performance (skips entire trees)
- Absolute path patterns add slight overhead for path normalization

### Compatibility

- Feature added in CyberPatchMaker v1.0.2
- Works with all patch modes: legacy `--versions-dir` and custom `--from-dir/--to-dir`
- Compatible with all compression modes (zstd, gzip, none)
- Absolute path support added in v1.0.12

## Example Workflow

### Step 1: Create .cyberignore

```powershell
# In your source version directory
cd versions/1.0.0

# Create .cyberignore
@"
:: Ignore sensitive files
*.key
*.crt
config/secrets.json

:: Ignore logs
*.log
logs/
"@ | Out-File .cyberignore -Encoding UTF8
```

### Step 2: Generate Patch

```powershell
# Generate patch (ignored files excluded automatically)
.\patch-gen.exe --from-dir .\versions\1.0.0 --to-dir .\versions\1.0.1 --output .\patches
```

### Step 3: Verify Results

Check the output:
```
Version 1.0.0 registered: 4 files, 1 directories
```

Files like `secret.key`, `*.log`, and `config/secrets.json` won't be in the patch.

## Related Documentation

- [Quick Start Guide](quick-start.md) - Getting started with CyberPatchMaker
- [Generator Guide](generator-guide.md) - Complete patch generation reference
- [Backup System](backup-system.md) - Understanding the backup.cyberpatcher directory
- [CLI Reference](cli-reference.md) - Command-line options

## Summary

The `.cyberignore` file provides fine-grained control over which files are included in patches:

- **Simple Syntax**: Comment lines start with `::`
- **Flexible Patterns**: Exact names, wildcards, directories, nested paths, **absolute paths**
- **Automatic**: Just place `.cyberignore` in source directory
- **Safe**: `.cyberignore` and `backup.cyberpatcher/` always excluded
- **Performance**: Directory exclusions skip entire trees efficiently
- **Cross-Platform**: Works with relative and absolute paths on Windows and Unix systems

Use `.cyberignore` to keep sensitive data out of patches and reduce patch size!
