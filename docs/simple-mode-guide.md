# Simple Mode Guide

## Overview

Simple Mode is controlled by the `SimpleMode` boolean field in the `Patch` struct. When set to `true`, the applier's `runSimpleMode()` function provides a fully automated patching experience.

**Current status:** The `SimpleMode` field and `runSimpleMode()` function exist in the codebase, but no generator code path currently sets `SimpleMode = true`. The field is reserved for future use.

**Distinction:**
- **Simple Mode** (`Patch.SimpleMode`): Fully automated — runs dry-run validation then automatically applies the patch with verbose progress output
- **Silent Mode** (`--silent` flag): Works with any self-contained executable, minimal output, designed for scripting and CI/CD

## Behavior (When Enabled)

When Simple Mode is enabled in the patch data, the applier:

1. Reads patch metadata from the embedded self-contained executable
2. Uses current directory as target (no prompt)
3. Automatically runs a dry-run validation (key file + required files)
4. If validation passes, automatically applies the patch with verification and backup
5. Logs all output to `<patchname>_<utctime>_log.txt`
6. Exits with code 0 on success, 1 on failure

**No menu, no prompts, no user interaction required.**

## Safety

All verification and backup features operate normally. Simple Mode only changes the UI — it does not disable any safety checks.
