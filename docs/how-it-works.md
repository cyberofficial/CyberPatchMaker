# How It Works

Understanding how CyberPatchMaker creates and applies efficient delta patches.

## The Problem

Traditional software updates require downloading the entire application again:
- **Version 1.0.0**: 5GB download
- **Version 1.0.1**: 5GB download (even if only 50MB changed!)

This wastes bandwidth, time, and storage.

## The Solution

CyberPatchMaker creates **delta patches** - small files containing only the changes:
- **Version 1.0.0**: 5GB (initial install)
- **Patch 1.0.0→1.0.1**: 10MB (only the changes!)

Users download and apply the 10MB patch instead of re-downloading 5GB.

---

## Core Concepts

### 1. Directory Tree Scanning

CyberPatchMaker scans **entire directory trees** recursively:

```
MyApp/
├── program.exe          ← Scanned & hashed
├── data/
│   ├── config.json      ← Scanned & hashed
│   └── assets/
│       ├── image1.png   ← Scanned & hashed
│       └── image2.png   ← Scanned & hashed
└── libs/
    ├── core.dll         ← Scanned & hashed
    └── dependencies/
        └── third.dll    ← Scanned & hashed
```

**Every file at every level** is discovered and hashed.

---

### 2. SHA-256 Hashing

For each file, calculate a SHA-256 hash (256-bit fingerprint):

```
program.exe → SHA-256 → 63573ff071ea5fa2b8c9d3e1a4f7c8e5...
```

This hash:
- **Uniquely identifies** the file contents
- **Detects any modification** (even 1 bit change = different hash)
- **Cannot be forged** (cryptographically secure)

---

### 3. Manifest Generation

Create a manifest (JSON file) describing the complete version:

```json
{
  "version": "1.0.0",
  "key_file": {
    "path": "program.exe",
    "checksum": "63573ff071ea5fa2...",
    "size": 52428800
  },
  "files": [
    {
      "path": "program.exe",
      "checksum": "63573ff071ea5fa2...",
      "size": 52428800
    },
    {
      "path": "data/config.json",
      "checksum": "a1b2c3d4e5f6...",
      "size": 1024
    },
    {
      "path": "libs/core.dll",
      "checksum": "f7g8h9i0j1k2...",
      "size": 2097152
    }
  ],
  "timestamp": "2025-10-04T10:00:00Z",
  "total_size": 5368709120,
  "checksum": "manifest_overall_hash..."
}
```

---

### 4. Manifest Comparison

Compare two manifests to find changes:

```
Version 1.0.0                      Version 1.0.1
program.exe: abc123...             program.exe: xyz789...  → MODIFIED
data/config.json: def456...        data/config.json: def456...  → UNCHANGED
libs/core.dll: ghi789...           libs/core.dll: uvw345...  → MODIFIED
libs/old.dll: jkl012...            (missing)  → DELETED
(missing)                          libs/new.dll: pqr678...  → ADDED
```

**Result:** List of operations needed to transform 1.0.0 into 1.0.1

---

### 5. Binary Diff Generation

For **modified files**, use bsdiff algorithm to create binary diffs:

```
Old file: program.exe (50MB)
New file: program.exe (52MB)
           ↓
      bsdiff algorithm
           ↓
Binary diff: (2MB) ← Much smaller than full file!
```

**bsdiff** is intelligent:
- Finds similar sections between files
- Stores only the differences
- Highly efficient for executable files
- Can handle insertions, deletions, and modifications

---

### 6. Patch Package Creation

Package all operations into a compressed patch file:

```
Patch File (1.0.0-to-1.0.1.patch):
┌─────────────────────────────────────┐
│ Header:                             │
│   Format Version: 1                 │
│   From Version: 1.0.0               │
│   To Version: 1.0.1                 │
│   Created: 2025-10-04               │
│   Compression: zstd                 │
│   Size: 10MB                        │
├─────────────────────────────────────┤
│ Key File Verification:              │
│   From: program.exe (abc123...)     │
│   To:   program.exe (xyz789...)     │
├─────────────────────────────────────┤
│ Required Files (must match):        │
│   data/config.json: def456...       │
│   libs/core.dll: ghi789...          │
│   ... (all source files)            │
├─────────────────────────────────────┤
│ Operations:                         │
│   1. MODIFY program.exe             │
│      Old hash: abc123...            │
│      New hash: xyz789...            │
│      Binary diff: [2MB data]        │
│   2. MODIFY libs/core.dll           │
│      Old hash: ghi789...            │
│      New hash: uvw345...            │
│      Binary diff: [500KB data]      │
│   3. DELETE libs/old.dll            │
│      Old hash: jkl012...            │
│   4. ADD libs/new.dll               │
│      New hash: pqr678...            │
│      Full file: [3MB data]          │
└─────────────────────────────────────┘
      ↓ zstd compression
Final size: 10MB (from ~5MB uncompressed)
```

---

## Patch Application Process

### Step-by-Step Application

```
1. LOAD PATCH
   ├─ Read patch file
   ├─ Decompress data
   └─ Display information

2. PRE-VERIFICATION ← Critical safety check!
   ├─ Verify target directory exists
   ├─ Verify key file exists and hash matches
   │  └─ If wrong: STOP - Wrong version!
   ├─ Verify ALL required files exist
   ├─ Verify ALL required file hashes match
   │  └─ If wrong: STOP - Modified installation!
   └─ SUCCESS: Confirmed clean version 1.0.0

3. CREATE BACKUP ← Only after verification passes!
   ├─ Create backup directory
   ├─ Recursively copy ALL files
   └─ Backup captures VERIFIED CLEAN STATE

4. APPLY OPERATIONS
   ├─ For each operation:
   │  ├─ DELETE: Remove file/directory
   │  ├─ ADD: Write new file
   │  └─ MODIFY: Apply binary diff
   └─ Report progress

5. POST-VERIFICATION ← Ensure operations worked!
   ├─ Verify ALL modified files have correct hashes
   ├─ Verify key file now matches target version
   │  └─ If wrong: RESTORE from backup!
   └─ SUCCESS: Confirmed clean version 1.0.1

6. CLEANUP
   ├─ If success: Remove backup
   └─ If failure: Keep backup for manual recovery
```

---

## Safety Features

### 1. Key File System

Every version has a **key file** (main executable):
- **Unique identifier** for the version
- **Prevents wrong patches** (1.0.0→1.0.3 patch won't apply to 1.0.1)
- **Detects corruption** (modified key file = reject patch)

**Example:**
```
Patch requires: program.exe with hash abc123...
User has:       program.exe with hash xyz789...
Result:         REJECTED - Wrong version or corrupted!
```

---

### 2. Complete File Verification

**Pre-verification** checks **EVERY file**:
```
✓ Key file: program.exe (63573ff...)
✓ Required: data/config.json (a1b2c3d4...)
✓ Required: libs/core.dll (f7g8h9i0...)
✓ Required: libs/dependencies/third.dll (m4n5o6p7...)
... (ALL files verified)
```

If **ANY file is wrong** → Patch rejected, no changes made

---

### 3. Atomic Operations

**All-or-nothing approach:**
- Either **all operations succeed** → Update complete
- Or **any operation fails** → Restore from backup

**Never leaves system in broken state!**

---

### 4. Backup & Rollback

**Backup created after verification:**
- Captures **verified clean state**
- Enables **safe restoration** on failure
- **Never backs up corrupted state**

**Restoration triggers:**
- Operation failure (disk full, permission error)
- Post-verification failure (wrong result)
- User interruption (Ctrl+C handling planned)

---

## Why This Design?

### Compared to Other Update Systems

**Full Downloads (Traditional):**
- ❌ Wastes bandwidth (download everything)
- ❌ Wastes time (5GB vs 10MB)
- ❌ Wastes storage (keep multiple full versions)
- ✅ Simple (just overwrite)

**Incremental Patches (Our System):**
- ✅ Minimal bandwidth (10MB patch vs 5GB download)
- ✅ Fast updates (seconds vs hours)
- ✅ Efficient storage (small patches)
- ✅ Safe (verification + rollback)
- ⚠️ More complex (but worth it!)

---

**rsync-style sync:**
- ✅ Good for file synchronization
- ❌ Not designed for versioned software
- ❌ No version verification
- ❌ No atomic operations
- ❌ Can break on interruption

**Git-style version control:**
- ✅ Excellent for source code
- ❌ Not optimized for binary files
- ❌ Overhead of full repository
- ❌ Complex for end users

**Our System (Delta Patches):**
- ✅ Optimized for binary applications
- ✅ Versioned and verified
- ✅ Atomic and safe
- ✅ Simple user experience
- ✅ Production-ready

---

## Performance Characteristics

### Patch Generation

**Time:** O(n) where n = number of files
- Scan source version: ~30 seconds for 5GB
- Scan target version: ~30 seconds for 5GB
- Compare manifests: <1 second
- Generate diffs: ~2 minutes for 50MB of changes
- Compress patch: ~10 seconds
- **Total: ~3-5 minutes**

---

### Patch Application

**Time:** O(k) where k = number of operations
- Load patch: <1 second
- Pre-verification: ~10 seconds for 5GB installation
- Create backup: ~30 seconds for 5GB
- Apply operations: ~30 seconds for 50MB of changes
- Post-verification: ~10 seconds
- **Total: ~1-2 minutes**

---

### Patch Sizes

**Typical scenarios:**

| Change Amount | Full Download | Patch Size | Savings |
|--------------|---------------|------------|---------|
| Bug fixes (10MB) | 5GB | 2-5MB | 99.9% |
| Feature update (50MB) | 5GB | 10-20MB | 99.6% |
| Major overhaul (500MB) | 5GB | 100-200MB | 96% |

---

## Technical Deep Dive

### bsdiff Algorithm

**How it works:**
1. **Scan old file** for similar blocks
2. **Find matching blocks** in new file
3. **Compute differences** between blocks
4. **Generate instructions**:
   - Copy block from old file
   - Add new bytes
   - Skip bytes
5. **Compress instructions**

**Why it's efficient:**
- Exploits similarity between versions
- Executables often have large unchanged sections
- Very small diffs for minor code changes

**Example:**
```
Old program.exe: 50MB
New program.exe: 52MB (added one function)
Binary diff: 2MB (only the changes + control data)
```

---

### zstd Compression

**Why zstd?**
- **Modern algorithm** (released 2016 by Facebook)
- **Fast compression** (~500 MB/s)
- **Fast decompression** (~1500 MB/s)
- **Excellent ratio** (better than gzip)
- **Tunable levels** (1=fast, 4=best compression)

**Example:**
```
Uncompressed patch: 5MB
zstd level 3: 1.2MB (76% reduction)
zstd level 4: 1.0MB (80% reduction)
```

---

## Related Documentation

- [Architecture](architecture.md) - System design
- [Backup Lifecycle](backup-lifecycle.md) - Backup timing details
- [Hash Verification](hash-verification.md) - Verification explained
- [Generator Guide](generator-guide.md) - Creating patches
- [Applier Guide](applier-guide.md) - Applying patches
