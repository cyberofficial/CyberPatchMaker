# Hash Verification

## Overview

Every file operation in CyberPatchMaker is protected by SHA-256 hashing at three verification points.

## Verification Points

### 1. Pre-Verification (Source Version)
Before any changes, the applier verifies:
- Key file exists and its SHA-256 matches the expected source hash
- All required files exist and match their expected hashes
- If ANY mismatch is found, patch is rejected, no backup created, no changes made

### 2. During Operations
- **Modify**: Old file checksum verified before applying replacement, new file checksum verified after write
- **Add**: Written file checksum verified against expected hash
- **Delete**: File verified to exist with correct checksum before deletion

### 3. Post-Verification (Target Version)
After all operations, verifies:
- Key file hash matches target version
- All modified and added files match expected target hashes
- If verification fails: automatic rollback from backup restores original state

Note: Delete, delete-directory, and add-directory operations are not re-verified post-patch (no meaningful checksum to verify).

## Implementation

```go
// pkg/utils/checksum.go
func CalculateFileChecksum(filePath string) (string, error)  // SHA-256 of file
func CalculateDataChecksum(data []byte) string              // SHA-256 of bytes
func VerifyFileChecksum(filePath, expectedChecksum string) (bool, error)  // compare

// Streaming: io.Copy into sha256.New() for memory-efficient hashing
```

Parallel checksum computation is handled by `scanner.ScanDirectoryParallel()` using a worker pool.

## Failure Example

```
Pre-verification failed: key file checksum mismatch
Expected: abc123...
Got:      xyz789...
This patch requires version 1.0.0
Your installation may be corrupted or modified
```

## Security Properties

- **Collision resistance**: Finding two files with same SHA-256 is computationally infeasible
- **Pre-image resistance**: Cannot create a file that matches a specific hash
- **Tamper evidence**: Any byte change produces a completely different hash
- **Checksum verification in the self-contained EXE header** prevents patch data tampering
