# .cyberignore File Guide

## Overview

Place a `.cyberignore` file in your version directory to exclude files and directories from patches. The file is loaded automatically by the scanner during patch generation — each scanned version root (source and target) loads its own `.cyberignore`.

## Automatic Exclusions

The following are always excluded without needing to list them:
- `backup.cyberpatcher/` — patch backup directory
- `.cyberignore` — the ignore file itself

## Syntax

```
:: Lines starting with double-colon are comments (only at start of line)
:: Blank lines are ignored

*.key
logs/
config/secrets.json
E:\temp\build\*.log
```

- `*.key` — matches any `.key` file recursively
- `logs/` — trailing slash excludes entire directory tree
- `config/secrets.json` — exact nested path match
- `E:\temp\build\*.log` — absolute path with wildcard

## Pattern Matching

- **Exact filenames**: `secret.key` matches that file at root, plus any directory of the same name and its entire contents
- **Wildcards**: `*.log` matches any `.log` file at any depth
- **Directories**: `logs/` excludes the entire tree under `logs/`
- **Nested paths**: `config/secrets.json` matches only that specific path
- **Absolute paths**: `E:\temp\*.log` matches files by full filesystem path (added v1.0.14)
- **Case**: Exact path matching uses case-insensitive `strings.EqualFold`. Directory patterns with trailing slash (e.g., `logs/`) match case-insensitively via `strings.ToLower`. For patterns without a trailing slash (e.g., `logs`), the entry itself still matches case-insensitively via `EqualFold` (so a `LOGS` directory is ignored with its whole subtree), but matching files *inside* the directory by prefix is case-sensitive. Wildcard (`*`) matching via `filepath.Match` is case-sensitive on every platform.

## Examples

```
:: Sensitive files
*.key
*.pem
.env
config/secrets.json

:: Logs and temp
*.log
logs/
temp/

:: Dev tools
.vscode/
.idea/

:: User data (don't override)
saves/
user_config.json
```

## Verification

Generate a throwaway patch into a test output directory and check the reported file count to confirm patterns work (the generator has no dry-run mode):

```bash
patch-gen --from-dir ./versions/1.0.0 --to-dir ./versions/1.0.1 --output ./test
```
