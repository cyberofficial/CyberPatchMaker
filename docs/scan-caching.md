# Scan Caching

CyberPatchMaker includes intelligent scan caching that dramatically speeds up subsequent patch generations. Instead of re-scanning large directories on every build, previously scanned versions can be loaded from cache in under a second.

## Overview

When generating patches, the most time-consuming operation is scanning directory trees and calculating SHA-256 hashes for every file. For large projects with tens of thousands of files, this can take 15+ minutes.

Scan caching eliminates this bottleneck by storing the complete scan results (including manifest) to disk. On subsequent patch generations, the cached data is loaded and validated instantly.

## Performance Impact

### Real-World Example (War Thunder)
- **Files**: 34,650 files
- **Size**: ~56 GB
- **First scan (without cache)**: ~15 minutes
- **Cached scan load**: <1 second
- **Speedup**: **900x faster**

### Benefits
- Instant patch generation for previously scanned versions
- Key file hash validation ensures cache integrity
- Location-based hashing prevents using wrong cached data
- Transparent operation - works with all generation modes

## How It Works

### Cache Storage

Cached scans are stored in a dedicated cache directory (default: `.data/`):

```
.data/
├── scan_1.0.0_a1b2c3d4e5f6g7h8.json
├── scan_1.0.1_i9j0k1l2m3n4o5p6.json
└── scan_1.0.2_q7r8s9t0u1v2w3x4.json
```

### Cache Filename Format

Each cache file is named: `scan_<version>_<locationHash>.json`

- `version`: The version number (e.g., "1.0.0")
- `locationHash`: First 16 characters of SHA-256 hash of the location path

The location hash ensures that cached scans from one directory cannot be mistakenly used for another directory, even if version numbers match.

### Cache File Structure

```json
{
  "version": "1.0.0",
  "location": "/path/to/versions/1.0.0",
  "key_file": {
    "Path": "game.exe",
    "Checksum": "abc123...",
    "Size": 15728640
  },
  "manifest": {
    "Version": "1.0.0",
    "KeyFile": {...},
    "Files": [...],
    "Directories": [...],
    "Timestamp": "2024-01-15T10:30:00Z",
    "TotalSize": 56987654321,
    "TotalFiles": 34650,
    "Checksum": "def456..."
  },
  "cached_at": "2024-01-15T10:30:05Z",
  "location_hash": "a1b2c3d4e5f6g7h8"
}
```

## Usage

### Enable Scan Caching

Use the `--savescans` flag when generating patches:

```bash
# First generation - creates cache
patch-gen --versions-dir ./versions --new-version 1.0.3 --output ./patches --savescans
```

### Subsequent Generations

The cache is automatically used on subsequent runs:

```bash
# Second generation - uses cache (instant!)
patch-gen --versions-dir ./versions --new-version 1.0.4 --output ./patches --savescans
```

### Custom Cache Location

Specify a custom cache directory with `--scandata`:

```bash
# Use shared cache for team collaboration
patch-gen --versions-dir ./versions --new-version 1.0.3 --output ./patches --savescans --scandata ./shared-cache
```

### Force Rescan

If files have changed and you need to invalidate cache:

```bash
# Force fresh scan
patch-gen --versions-dir ./versions --new-version 1.0.3 --output ./patches --savescans --rescan
```

## Cache Validation

### Key File Hash Verification

When loading from cache, CyberPatchMaker verifies:
1. Cached scan exists for the version
2. Location hash matches (prevents wrong directory usage)
3. Key file checksum matches the cached manifest

If the key file has changed, the cache is invalidated and a fresh scan is performed automatically.

### Location-Based Isolation

Each location gets its own cache entries:

```
# Project A at /builds/project-a/versions/1.0.0
.data/scan_1.0.0_a1b2c3d4e5f6g7h8.json

# Project B at /builds/project-b/versions/1.0.0
.data/scan_1.0.0_x9y0z1a2b3c4d5e6.json
```

Even with identical version numbers, projects maintain separate caches.

## Cache Management

### List Cached Scans

View all cached scans programmatically:

```go
cache := cache.NewScanCache(".data")
scans, _ := cache.ListCachedScans()

for _, scan := range scans {
    fmt.Printf("%s @ %s (%d files, %d bytes)\n",
        scan.Version, scan.Location, scan.TotalFiles, scan.TotalSize)
}
```

### Delete Specific Cache Entry

Remove a cached scan:

```go
cache.DeleteScan("1.0.0", "/path/to/versions/1.0.0")
```

### Clear All Cache

Remove all cached scans:

```go
cache.ClearCache()
```

## Internal API

### ScanCache Structure

```go
type ScanCache struct {
    cacheDir string  // Directory containing cache files
}
```

### Key Methods

| Method | Description |
|--------|-------------|
| `NewScanCache(cacheDir string)` | Create new cache manager (defaults to `.data/`) |
| `SaveScan(version *utils.Version)` | Save scan to cache |
| `LoadScan(versionNumber, location string)` | Load with validation |
| `HasCachedScan(versionNumber, location string)` | Check if cache exists |
| `DeleteScan(versionNumber, location string)` | Remove cached scan |
| `ClearCache()` | Remove all cached scans |
| `ListCachedScans()` | Get info about all cached scans |
| `GetCacheDir()` | Returns the cache directory path as a string |

## Best Practices

### For Development Workflows
1. **Enable caching by default** in CI/CD pipelines
2. **Use shared cache location** for team builds
3. **Version the cache directory** alongside source code
4. **Clean cache periodically** to remove old versions

### For Production Builds
1. **Cache on first build** of each version
2. **Reuse cache** for all patches from that version
3. **Validate cache integrity** with key file checksums
4. **Store cache separately** from build artifacts

### For Continuous Integration
```bash
# Mount cache volume in CI
docker run -v ./cache:/app/.data patch-gen \
  --versions-dir ./versions \
  --new-version 1.0.3 \
  --output ./patches \
  --savescans
```

## Troubleshooting

### Cache Not Working
- **Symptom**: Scans take full time despite `--savescans`
- **Cause**: Cache directory not writable or path incorrect
- **Solution**: Check `--scandata` path and permissions

### Stale Cache Data
- **Symptom**: Patches include old files
- **Cause**: Files changed without cache invalidation
- **Solution**: Use `--rescan` flag to force fresh scan

### Cache Location Mismatch
- **Symptom**: Cache not found for known version
- **Cause**: Location path changed (relative vs absolute)
- **Solution**: Use consistent paths or `--rescan`

## Version History

- **v1.0.10**: Initial scan caching implementation
  - `--savescans` flag for cache creation
  - `--scandata` flag for custom cache location
  - `--rescan` flag for cache invalidation
  - Location-based hashing for isolation
  - Key file validation for integrity

## Related Documentation

- [Version Management](version-management.md) - How versions are tracked
- [Generator Guide](generator-guide.md) - CLI flags for scan caching
- [Performance](performance.md) - Performance characteristics
