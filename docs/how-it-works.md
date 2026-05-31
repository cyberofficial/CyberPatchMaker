# How It Works

## Overview

CyberPatchMaker creates delta patches by comparing directory trees and storing full file replacements for changed files.

## Process

### 1. Directory Scanning
Every file in the version directory is recursively discovered and hashed with SHA-256. The scanner respects `.cyberignore` patterns and excludes `backup.cyberpatcher/`.

### 2. Manifest Creation
A JSON manifest records every file's relative path, size, checksum, and the key file used for version identification.

### 3. Manifest Comparison
Source and target manifests are compared to identify added, modified, and deleted files and directories.

### 4. Patch Packaging
For each modified or added file, the full file content is read and stored in the patch. All operations (add/modify/delete files, add/delete directories) are serialized into a compressed JSON file.

### 5. Compression
The patch JSON is stream-encoded and compressed (zstd by default, configurable levels 1-4) for distribution.

## Patch Application

1. **Load and decompress** patch file
2. **Pre-verify**: check key file and all required file hashes match expected source version
3. **Create selective backup** of files being modified/deleted to `backup.cyberpatcher/`
4. **Apply operations**: add new files/directories, replace modified files, delete removed files/directories
5. **Post-verify**: check key file and modified files match expected target version
6. **On failure**: automatic rollback from backup restores original state

## Safety Guarantees

- **Wrong version detection**: key file hash mismatch prevents applying patch to wrong version
- **Corruption detection**: any modified file in source is caught before changes begin
- **Atomic operations**: all-or-nothing with automatic rollback on any failure
- **Selective backup**: only changed files backed up, preserved after success for manual rollback

## Performance

- Scanning: O(n) where n = number of files
- Typical patch size: 2-20MB for a 5GB application with small changes
- Compression reduces size by ~60% on average
