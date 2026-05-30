# Downgrade Guide

## Overview

CyberPatchMaker fully supports generating patches in either direction. A downgrade patch is identical in format to an upgrade patch — simply swap the source and target versions.

## Generating

### Manual (swap --from/--to)

```bash
# Upgrade: 1.0.0 -> 1.0.1
patch-gen --from 1.0.0 --to 1.0.1 --versions-dir ./versions --output ./patches

# Downgrade: 1.0.1 -> 1.0.0 (just swap)
patch-gen --from 1.0.1 --to 1.0.0 --versions-dir ./versions --output ./patches/downgrade
```

### Automatic (--crp flag)

Use `--crp` to generate both forward and reverse patches in one invocation:

```bash
# Creates both 1.0.0-to-1.0.1.patch AND 1.0.1-to-1.0.0.patch
patch-gen --from-dir ./v1.0.0 --to-dir ./v1.0.1 --output ./patches --crp

# With self-contained executables for both directions
patch-gen --from-dir ./v1.0.0 --to-dir ./v1.0.1 --output ./patches --crp --create-exe
```

## Applying

Downgrade patches are applied identically to upgrade patches:

```bash
patch-apply --patch downgrade/1.0.3-to-1.0.2.patch --current-dir ./app --verify
```

## Recommended Directory Layout

```
patches/
├── upgrade/
│   ├── 1.0.0-to-1.0.1.patch
│   └── 1.0.1-to-1.0.2.patch
└── downgrade/
    ├── 1.0.2-to-1.0.1.patch
    └── 1.0.1-to-1.0.0.patch
```

## Safety Notes

- Downgrade patches include the same pre/post verification and backup protections as upgrade patches
- If the newer version changed data formats (databases, config schemas), those are NOT rolled back by file patching — handle data migrations separately
- Always test downgrade on a copy first before production use
