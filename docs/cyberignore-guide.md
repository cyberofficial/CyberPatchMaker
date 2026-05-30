# .cyberignore File Guide

## Overview

Place a `.cyberignore` file in your source version directory to exclude files and directories from patches. The file is loaded automatically by the scanner during patch generation.

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

- **Exact filenames**: `secret.key` matches only that file at root
- **Wildcards**: `*.log` matches any `.log` file at any depth
- **Directories**: `logs/` excludes the entire tree under `logs/`
- **Nested paths**: `config/secrets.json` matches only that specific path
- **Absolute paths**: `E:\temp\*.log` matches files by full filesystem path (added v1.0.12)
- **Case**: Exact path matching uses case-insensitive `strings.EqualFold`. Directory patterns with trailing slash (e.g., `logs/`) match case-insensitively via `strings.ToLower`. Directory patterns without trailing slash (e.g., `logs`) are case-sensitive. Wildcard (`*`) matching via `filepath.Match` is platform-dependent (case-insensitive on Windows, case-sensitive on Linux/macOS).

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

Run a dry patch generation and check the file count in the output to confirm patterns work:

```bash
patch-gen --from-dir ./versions/1.0.0 --to-dir ./versions/1.0.1 --output ./test
```
