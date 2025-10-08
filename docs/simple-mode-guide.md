# Simple Mode for End Users Guide

**NEW in v1.0.9**

A comprehensive guide to using Simple Mode to create user-friendly patches for non-technical end users.

## Overview

Simple Mode is a feature that patch creators can enable to provide end users with a simplified, streamlined interface. Instead of seeing all the technical options and settings, users get a clean, simple experience focused on the essentials.

**Important Distinction:**
- **Simple Mode** (SimpleMode field in patch): Simplified UI with user choices (Dry Run, Apply, Exit)
- **Silent Mode** (--silent flag): Fully automatic patching with zero user interaction for automation

This guide covers Simple Mode, which provides a user-friendly interface. For automation, see the Silent Mode section in the Applier Guide.

## Purpose

Simple Mode is designed for:
- **Client distributions** - Software vendors distributing patches to their clients
- **Non-technical users** - Users who don't need or want advanced options
- **Enterprise deployments** - IT departments deploying to end users
- **Reduced support burden** - Fewer questions about confusing options
- **Professional appearance** - Polished, user-friendly interface

## How It Works

### For Patch Creators

When generating a patch, enable "Simple Mode for End Users":

**GUI Method:**
1. Open Patch Generator GUI
2. Configure your patch (versions, compression, etc.)
3. Check **"Enable Simple Mode for End Users"** checkbox
4. Generate patch or create self-contained executable
5. Distribute to end users

**CLI Method:**
Currently, Simple Mode can only be enabled via the GUI. The CLI generator does not have a `--simple-mode` flag yet. Simple Mode is controlled by the `SimpleMode` field in the Patch struct, which the GUI sets when the checkbox is checked.

**Note:** Do not confuse this with `--silent` flag, which is for automation (fully automatic, no user interaction).

### For End Users

When end users run a patch created with Simple Mode enabled, they experience:

**GUI Mode:**
- Simple message showing version change
- Essential options only:
  - Create backup (checked by default)
  - Dry Run button
  - Apply Patch button
- Advanced options hidden

**CLI Mode:**
- Clean console interface
- Clear patch information
- Simple menu:
  1. Dry Run
  2. Apply Patch
  3. Exit
- Backup option before applying

## User Interface Comparison

### Standard Mode (Advanced)

**What Users See:**
```
=== Patch Information ===
From Version: 1.0.0
To Version: 1.0.1
Key File: program.exe
Hash: a3f8d9e...
Compression: zstd

Options:
☑ Verify before patching
☑ Verify after patching
☑ Create backup
☑ Auto-detect version
☐ Ignore 1GB limit
☐ Dry run mode

Custom key file: [_________________] [Browse]
Target directory: [_________________] [Browse]

[Advanced Options ▼]
```

### Simple Mode (Simplified)

**What Users See:**
```
==============================================
          Simple Patch Application
==============================================

You are about to patch from "1.0.0" to "1.0.1"

Create backup before patching? (Y/n): Y

==============================================
Options:
  1. Dry Run (test without making changes)
  2. Apply Patch
  3. Exit
==============================================
Select option [1-3]: 
```

## Technical Details

### What Gets Hidden/Disabled

In Simple Mode, the following options are automatically set and hidden from users:

**Always Enabled:**
- Pre-patch verification
- Post-patch verification
- Auto-detect version

**User Controllable:**
- Create backup (default: enabled)
- Dry run option

**Hidden/Disabled:**
- Custom key file selection
- Ignore 1GB limit
- Advanced compression settings
- Manual verification toggles

### Safety Features

Simple Mode **does not** compromise safety:
- All verification checks still run
- Backups are still created (user choice)
- Pre and post-verification mandatory
- Version auto-detection enabled
- Same reliability as standard mode

**Silent Mode vs Simple Mode:**
- **Silent Mode** (--silent flag): Fully automatic, zero interaction, applies patch immediately
- **Simple Mode** (patch field): Simplified menu with choices (Dry Run, Apply, Exit)

### Performance

Simple Mode has **no performance impact**:
- Same patch generation speed
- Same patch application speed
- Same verification process
- Only UI changes, not internals

## Use Cases

### Case 1: Software Vendor Distribution

**Scenario:**
A software company distributes patches to business clients who then deploy to their end users.

**Solution:**
```bash
# Generate patches for all versions
# Note: Simple Mode must be enabled via GUI checkbox
patch-gen --versions-dir ./releases \
          --new-version 2.0.0 \
          --output ./dist \
          --create-exe \
          --verify

# Distribute the .exe files to clients
# Clients' end users see simplified interface
```

**Benefits:**
- Professional appearance
- Reduced support calls
- User confidence
- Clear, simple process

### Case 2: Enterprise IT Deployment

**Scenario:**
IT department needs to deploy updates to 500 workstations with minimal user interaction.

**Solution:**
```bash
# Create GUI executable
# Note: Enable Simple Mode via GUI checkbox before generating
patch-gen --from-dir ./prod/v1.5 \
          --to-dir ./prod/v1.6 \
          --output ./deploy \
          --create-exe

# Users on workstations see simple interface:
# - Clear message about update
# - Backup option (default yes)
# - Dry run to test
# - Apply to execute
```

**Benefits:**
- Users can't accidentally disable critical safety
- Simple decision tree for users
- IT can confidently distribute
- Reduced help desk tickets

### Case 3: Game Update Distribution

**Scenario:**
Game developer needs to push updates to players worldwide with varying technical skills.

**Solution:**
```bash
# Create patches for all versions
# Note: Enable Simple Mode via GUI checkbox before generating
patch-gen --versions-dir ./builds \
          --new-version 3.2.0 \
          --output ./updates \
          --create-exe \
          --crp  # Include reverse patches

# Players download matching their version
# Simple interface: "Update game from X to Y?"
# One-click update process
```

**Benefits:**
- Professional game launcher feel
- Players can't break things
- Dry run lets them test safely
- Automatic version detection

## Best Practices

### When to Enable Silent Mode

✅ **DO use Silent Mode when:**
- Distributing to non-technical users
- Client or customer distributions
- End-user facing patches
- Support resources are limited
- Professional appearance matters
- Users should focus on essentials only

❌ **DON'T use Silent Mode when:**
- Internal development/testing
- Technical users need full control
- Advanced configuration required
- Debugging or troubleshooting
- Power users expected

### Patch Creator Checklist

Before distributing Silent Mode patches:

- [ ] Test the patch with Silent Mode enabled
- [ ] Verify the simplified interface works correctly
- [ ] Test both Dry Run and Apply Patch options
- [ ] Ensure backup option functions properly
- [ ] Test with both GUI and CLI executables
- [ ] Prepare simple user instructions
- [ ] Document the expected user experience
- [ ] Test on clean installation
- [ ] Verify error handling works correctly
- [ ] Prepare support documentation

### User Instructions Template

When distributing Silent Mode patches, provide clear instructions:

```
How to Update [Your App Name]

1. Download the patch file matching your current version
   Example: For version 1.0.0, download "1.0.0-to-1.0.1.exe"

2. Close [Your App Name] completely

3. Double-click the patch file

4. You will see: "You are about to patch from X to Y"

5. (Recommended) Click "Dry Run" to test without changes

6. When ready, click "Apply Patch"

7. Wait for the update to complete

8. Restart [Your App Name]

Note: A backup will be created automatically for safety.
```

## Troubleshooting

### Issue: Users Ask for Advanced Options

**Cause:** Silent Mode hides advanced options by design

**Solution:**
- Create two versions: one with Silent Mode, one without
- Distribute Silent Mode version to regular users
- Provide standard version on request for power users

### Issue: Users Can't Find Custom Key File Option

**Cause:** Custom key file is disabled in Silent Mode

**Solution:**
- For standard deployments, key file should be in standard location
- If custom key file needed, use standard mode patches
- Or rename key file to expected name before distributing app

### Issue: Users Want to See Technical Details

**Cause:** Silent Mode simplifies output for clarity

**Solution:**
- Provide standard mode patch for technical users
- Silent Mode targets non-technical users specifically
- Include version/build info in separate documentation

## Advanced Configuration

### Hybrid Approach

You can provide both modes:

```bash
# Generate standard patch
patch-gen --from-dir ./v1 --to-dir ./v2 \
          --output ./patches/standard \
          --create-exe

# Generate silent mode patch
patch-gen --from-dir ./v1 --to-dir ./v2 \
          --output ./patches/simple \
          --create-exe \
          --silent

# Distribute:
# - patches/simple/* to regular users
# - patches/standard/* to power users or support
```

### With Reverse Patches

Silent Mode works with reverse patches:

```bash
# Note: Enable Simple Mode via GUI checkbox
patch-gen --from-dir ./v2.0 --to-dir ./v2.1 \
          --output ./patches \
          --create-exe \
          --crp

# Creates:
# - 2.0-to-2.1.exe (upgrade, silent mode)
# - 2.1-to-2.0.exe (downgrade, silent mode)
```

Both upgrade and downgrade use Silent Mode interface.

## FAQ

**Q: Does Silent Mode make patches less safe?**  
A: No. All safety features (verification, backup, etc.) still operate. Only the UI is simplified.

**Q: Can users still see what's changing?**  
A: Users see high-level info (from version X to Y). Technical details are hidden for clarity.

**Q: What if users need the advanced options?**  
A: Provide standard mode patches for power users. Silent Mode targets non-technical users.

**Q: Can Simple Mode be disabled after patch creation?**  
A: No, it's embedded in the patch file (SimpleMode field). Create new patch without Simple Mode enabled if needed.

**Q: Does Simple Mode work with all patch types?**  
A: Yes - works with standard patches, self-contained exes, CLI and GUI modes.

**Q: Can users still do dry runs in Simple Mode?**  
A: Yes! Dry Run button is available and recommended before applying.

**Q: What happens if the patch fails?**  
A: Same as standard mode - automatic backup restoration, clear error message.

**Q: Is there a performance difference?**  
A: No performance impact. Only UI changes, internal processing identical.

## Related Documentation

- [Generator Guide](generator-guide.md) - Creating patches with Silent Mode
- [Applier Guide](applier-guide.md) - How simplified interface works
- [GUI Usage Guide](gui-usage.md) - Using GUI to enable Silent Mode
- [Self-Contained Executables](self-contained-executables.md) - Creating standalone patches

## Summary

Simple Mode provides:
- ✅ Simplified, user-friendly interface
- ✅ Professional appearance for distributions
- ✅ Reduced support burden
- ✅ Same safety and reliability
- ✅ Focus on essentials only
- ✅ Suitable for non-technical users

Use Simple Mode when distributing patches to end users who need simplicity and clarity without compromising safety or reliability.

**Remember:** Simple Mode (SimpleMode field) = Simplified UI with choices | Silent Mode (--silent flag) = Fully automatic for automation
