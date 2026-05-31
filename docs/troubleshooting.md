# Troubleshooting

## Quick Diagnostics

```bash
go build -o patch-gen.exe ./cmd/generator
go build -o patch-apply.exe ./cmd/applier
.\advanced-test.ps1
```

## Generator Issues

**"versions directory not found"**: Check the path exists and contains version folders.

**"key file not found"**: No `program.exe`, `game.exe`, `app.exe`, or `main.exe` found. Use `--key-file <name>` if your executable has a custom name, or rename it to one of the auto-detected names.

**"failed to create patch file"**: Check output directory exists, has write permissions, and has sufficient disk space.

**Slow performance**: Large files take time. Use SSD storage, enable scan caching with `--savescans`, use `--jobs` for parallel processing. Compression level 1 is faster than level 4.

## Applier Issues

**"key file checksum mismatch"**: You're trying to apply the wrong patch or your installation is modified/corrupted. Check the patch's expected source version and verify you have that exact version installed.

**"required file missing or checksum mismatch"**: A file in your installation was modified or deleted. Re-install the clean source version and try again.

**"permission denied"**: Close the application before patching. On Windows, run as administrator for Program Files. On Linux, check ownership and permissions.

**"insufficient disk space"**: Free up space. The selective backup only needs space for changed files (typically much less than the full installation).

**"post-verification failed"**: Patch produced wrong results. The system will attempt automatic rollback. Check disk for errors, re-download the patch file, and retry.

## Backup Issues

**Need to manually rollback**: Copy files from `backup.cyberpatcher/` to their original locations using the mirror structure. Delete the backup folder when confirmed.

**Automatic rollback fails**: Check disk space and file permissions. Manually restore from `backup.cyberpatcher/` if needed.

## Silent Mode / Self-Contained Exes

**1GB limit warning**: Run with `--ignore1gb` flag if you have sufficient RAM. Or use the standalone `.patch` file + `patch-apply.exe` instead.

**Silent mode exits with code 1**: Check `log_<timestamp>.txt` for the full error. Common causes: wrong directory, key file mismatch, insufficient permissions.

**Self-contained exe won't run**: Windows may block downloaded exes — right-click → Properties → Unblock. Verify checksum if re-downloading.

## Getting Help

When reporting issues, include: platform, Go version, full command used, complete error output, and relevant file sizes.
