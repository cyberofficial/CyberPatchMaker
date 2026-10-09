package patcher

import (
	"bytes"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cyberofficial/cyberpatchmaker/internal/core/version"
	"github.com/cyberofficial/cyberpatchmaker/pkg/utils"
)

// deltaTestKeyFile is the key file every test version directory contains.
const deltaTestKeyFile = "program.exe"

// deterministicBytes returns n pseudo-random bytes generated from a fixed
// seed so tests are fully reproducible.
func deterministicBytes(seed int64, n int) []byte {
	rng := rand.New(rand.NewSource(seed))
	data := make([]byte, n)
	rng.Read(data)
	return data
}

// makeVersionDir writes fileContents into path (always including the
// program.exe key file) and registers the directory as a version.
func makeVersionDir(t *testing.T, path string, fileContents map[string][]byte) *utils.Version {
	t.Helper()

	if _, ok := fileContents[deltaTestKeyFile]; !ok {
		fileContents[deltaTestKeyFile] = []byte("CyberPatchMaker delta test key file\n")
	}
	for name, data := range fileContents {
		full := filepath.Join(path, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatalf("failed to create directory for %s: %v", name, err)
		}
		if err := os.WriteFile(full, data, 0644); err != nil {
			t.Fatalf("failed to write %s: %v", name, err)
		}
	}

	mgr := version.NewManager()
	ver, err := mgr.RegisterVersion(filepath.Base(path), path, deltaTestKeyFile)
	if err != nil {
		t.Fatalf("failed to register version at %s: %v", path, err)
	}
	return ver
}

// findModifyOp returns the modify operation for the given slash-separated path.
func findModifyOp(t *testing.T, patch *utils.Patch, path string) utils.PatchOperation {
	t.Helper()
	for _, op := range patch.Operations {
		if op.Type == utils.OpModify && filepath.ToSlash(op.FilePath) == path {
			return op
		}
	}
	t.Fatalf("no modify operation found for %s in patch", path)
	return utils.PatchOperation{}
}

// deltaFixture is the shared two-version setup: only data/big.bin changes
// (2MB of pseudo-random data with two flipped bytes) while program.exe is
// identical in both versions.
type deltaFixture struct {
	baseDir string
	v1Path  string
	v2Path  string
	bigOld  []byte
	bigNew  []byte
	v1      *utils.Version
	v2      *utils.Version
}

func newDeltaFixture(t *testing.T) *deltaFixture {
	t.Helper()

	f := &deltaFixture{baseDir: t.TempDir()}
	f.v1Path = filepath.Join(f.baseDir, "v1")
	f.v2Path = filepath.Join(f.baseDir, "v2")

	f.bigOld = deterministicBytes(1, 2*1024*1024)
	f.bigNew = append([]byte(nil), f.bigOld...)
	f.bigNew[100] ^= 0xFF
	f.bigNew[len(f.bigNew)-1] ^= 0xFF

	key := []byte("CyberPatchMaker delta test key file\n")
	f.v1 = makeVersionDir(t, f.v1Path, map[string][]byte{
		deltaTestKeyFile: key,
		"data/big.bin":   f.bigOld,
	})
	f.v2 = makeVersionDir(t, f.v2Path, map[string][]byte{
		deltaTestKeyFile: key,
		"data/big.bin":   f.bigNew,
	})
	return f
}

// TestGeneratePatchStoresDelta verifies that a two-byte change to a 2MB
// file is stored as a compact delta operation with consistent metadata.
func TestGeneratePatchStoresDelta(t *testing.T) {
	f := newDeltaFixture(t)

	options := &utils.PatchOptions{Compression: "zstd", CompressionLevel: 3, SkipIdentical: true}
	g := NewGenerator()
	patch, err := g.GeneratePatch(f.v1, f.v2, options)
	if err != nil {
		t.Fatalf("GeneratePatch returned error: %v", err)
	}

	op := findModifyOp(t, patch, "data/big.bin")
	if len(op.BinaryDiff) == 0 {
		t.Fatal("expected a non-empty BinaryDiff (delta) for data/big.bin")
	}
	if len(op.NewFile) != 0 {
		t.Errorf("delta op must not carry a full NewFile, got %d bytes", len(op.NewFile))
	}
	if op.Size != int64(len(op.BinaryDiff)) {
		t.Errorf("op.Size = %d, want %d (len(BinaryDiff))", op.Size, len(op.BinaryDiff))
	}
	if want := utils.CalculateDataChecksum(f.bigOld); op.OldChecksum != want {
		t.Errorf("op.OldChecksum = %s, want %s (v1 big.bin hash)", op.OldChecksum, want)
	}
	if want := utils.CalculateDataChecksum(f.bigNew); op.NewChecksum != want {
		t.Errorf("op.NewChecksum = %s, want %s (v2 big.bin hash)", op.NewChecksum, want)
	}
	if len(op.BinaryDiff) >= 1024 {
		t.Errorf("delta for a two-byte change is %d bytes, want < 1024", len(op.BinaryDiff))
	}
	if err := g.ValidatePatch(patch); err != nil {
		t.Errorf("ValidatePatch returned error: %v", err)
	}
	if patch.Header.FormatVersion != 2 {
		t.Errorf("patch.Header.FormatVersion = %d, want 2", patch.Header.FormatVersion)
	}
}

// TestGeneratePatchFallbackWhenDeltaNotSmaller verifies that a modification
// with completely unrelated content falls back to full file replacement
// because a delta against unrelated random data is never strictly smaller.
func TestGeneratePatchFallbackWhenDeltaNotSmaller(t *testing.T) {
	baseDir := t.TempDir()
	v1Path := filepath.Join(baseDir, "v1")
	v2Path := filepath.Join(baseDir, "v2")

	oldData := deterministicBytes(11, 512*1024)
	newData := deterministicBytes(22, 512*1024)

	key := []byte("CyberPatchMaker delta test key file\n")
	v1 := makeVersionDir(t, v1Path, map[string][]byte{
		deltaTestKeyFile: key,
		"data/big.bin":   oldData,
	})
	v2 := makeVersionDir(t, v2Path, map[string][]byte{
		deltaTestKeyFile: key,
		"data/big.bin":   newData,
	})

	options := &utils.PatchOptions{Compression: "zstd", CompressionLevel: 3, SkipIdentical: true}
	g := NewGenerator()
	patch, err := g.GeneratePatch(v1, v2, options)
	if err != nil {
		t.Fatalf("GeneratePatch returned error: %v", err)
	}

	op := findModifyOp(t, patch, "data/big.bin")
	if len(op.NewFile) == 0 {
		t.Fatal("expected full replacement (non-empty NewFile) for unrelated random data")
	}
	if op.BinaryDiff != nil {
		t.Errorf("BinaryDiff must be nil for full replacement, got %d bytes", len(op.BinaryDiff))
	}
	if !bytes.Equal(op.NewFile, newData) {
		t.Error("NewFile must equal the new file content")
	}
	if op.Size != int64(len(newData)) {
		t.Errorf("op.Size = %d, want %d (new file size)", op.Size, len(newData))
	}
	if err := g.ValidatePatch(patch); err != nil {
		t.Errorf("ValidatePatch returned error: %v", err)
	}
}

// TestGeneratePatchFallbackWhenBaseTooLarge exercises the base size guard
// directly: an oversized base must be rejected before any file is read.
func TestGeneratePatchFallbackWhenBaseTooLarge(t *testing.T) {
	g := NewGenerator()
	op := &utils.PatchOperation{
		Type:        utils.OpModify,
		FilePath:    "data/big.bin",
		OldChecksum: "original-base-checksum",
		NewChecksum: "new-file-checksum",
	}
	sourceFile := &utils.FileEntry{
		Path:     "data/big.bin",
		Size:     utils.MaxDeltaBaseSize + 1, // base above the delta limit
		Checksum: "original-base-checksum",
	}

	// fromDir intentionally does not contain the file: the size guard must
	// reject the delta attempt before any file access happens.
	missingDir := filepath.Join(t.TempDir(), "no", "such", "dir")
	if g.tryDelta(op, missingDir, sourceFile, []byte("new file content"), 3) {
		t.Error("tryDelta must return false when the base exceeds MaxDeltaBaseSize")
	}
	if op.BinaryDiff != nil {
		t.Error("tryDelta must not set BinaryDiff when it rejects the base size")
	}
}

// TestApplyDeltaPatch applies a delta patch end to end and verifies the
// patched file matches the v2 content.
func TestApplyDeltaPatch(t *testing.T) {
	f := newDeltaFixture(t)

	options := &utils.PatchOptions{Compression: "zstd", CompressionLevel: 3, SkipIdentical: true}
	g := NewGenerator()
	patch, err := g.GeneratePatch(f.v1, f.v2, options)
	if err != nil {
		t.Fatalf("GeneratePatch returned error: %v", err)
	}
	if op := findModifyOp(t, patch, "data/big.bin"); len(op.BinaryDiff) == 0 {
		t.Fatal("expected a delta op for data/big.bin before applying")
	}

	target := filepath.Join(f.baseDir, "target")
	if err := utils.CopyDir(f.v1Path, target); err != nil {
		t.Fatalf("failed to copy v1 to target: %v", err)
	}

	a := NewApplier()
	if err := a.ApplyPatchWithPath(patch, target, "", true, true, true); err != nil {
		t.Fatalf("ApplyPatchWithPath returned error: %v", err)
	}

	got, err := utils.CalculateFileChecksum(filepath.Join(target, "data", "big.bin"))
	if err != nil {
		t.Fatalf("failed to hash patched data/big.bin: %v", err)
	}
	if want := utils.CalculateDataChecksum(f.bigNew); got != want {
		t.Errorf("patched data/big.bin checksum = %s, want %s (v2 content)", got, want)
	}
}

// TestApplyDeltaWrongBaseRejected verifies that applying a delta patch
// against a target whose base file was tampered with fails on the per-op
// old checksum check and leaves the target untouched.
func TestApplyDeltaWrongBaseRejected(t *testing.T) {
	f := newDeltaFixture(t)

	options := &utils.PatchOptions{Compression: "zstd", CompressionLevel: 3, SkipIdentical: true}
	g := NewGenerator()
	patch, err := g.GeneratePatch(f.v1, f.v2, options)
	if err != nil {
		t.Fatalf("GeneratePatch returned error: %v", err)
	}
	findModifyOp(t, patch, "data/big.bin") // sanity: the op exists

	target := filepath.Join(f.baseDir, "target")
	if err := utils.CopyDir(f.v1Path, target); err != nil {
		t.Fatalf("failed to copy v1 to target: %v", err)
	}

	// Tamper the on-disk base so it no longer matches op.OldChecksum.
	tampered := append([]byte(nil), f.bigOld...)
	tampered[2048] ^= 0xFF
	tamperedPath := filepath.Join(target, "data", "big.bin")
	if err := os.WriteFile(tamperedPath, tampered, 0644); err != nil {
		t.Fatalf("failed to tamper target base file: %v", err)
	}

	// Pre-verification is off so the per-operation old checksum check fires.
	a := NewApplier()
	err = a.ApplyPatchWithPath(patch, target, "", false, false, false)
	if err == nil {
		t.Fatal("ApplyPatchWithPath must fail when the on-disk base does not match OldChecksum")
	}
	if !strings.Contains(err.Error(), "old file checksum mismatch") {
		t.Errorf("error %q must contain 'old file checksum mismatch'", err)
	}

	got, err := os.ReadFile(tamperedPath)
	if err != nil {
		t.Fatalf("failed to read target data/big.bin: %v", err)
	}
	if !bytes.Equal(got, tampered) {
		t.Error("target data/big.bin must still contain the tampered bytes after the failed apply")
	}
}

// TestReconstructDeltaWrongBase verifies reconstructDelta rejects a base
// whose content does not match the op's OldChecksum.
func TestReconstructDeltaWrongBase(t *testing.T) {
	dir := t.TempDir()

	base := deterministicBytes(3, 64*1024)
	updated := append([]byte(nil), base...)
	updated[42] ^= 0xFF

	delta, err := utils.EncodeDelta(base, updated, 3)
	if err != nil {
		t.Fatalf("EncodeDelta returned error: %v", err)
	}

	// The file on disk differs from the base the delta was encoded against.
	tampered := append([]byte(nil), base...)
	tampered[7] ^= 0xFF
	targetPath := filepath.Join(dir, "big.bin")
	if err := os.WriteFile(targetPath, tampered, 0644); err != nil {
		t.Fatalf("failed to write tampered base file: %v", err)
	}

	op := utils.PatchOperation{
		Type:        utils.OpModify,
		FilePath:    "data/big.bin",
		OldChecksum: utils.CalculateDataChecksum(base),
		NewChecksum: utils.CalculateDataChecksum(updated),
		BinaryDiff:  delta,
	}

	a := NewApplier()
	_, err = a.reconstructDelta(targetPath, op)
	if err == nil {
		t.Fatal("reconstructDelta must reject a wrong base")
	}
	if !strings.Contains(err.Error(), "delta base checksum mismatch") {
		t.Errorf("error %q must contain 'delta base checksum mismatch'", err)
	}
}

// TestReconstructDeltaCorruptDeltaRejected verifies reconstructDelta detects
// a reconstruction that does not match the op's NewChecksum.
func TestReconstructDeltaCorruptDeltaRejected(t *testing.T) {
	dir := t.TempDir()

	base := deterministicBytes(4, 64*1024)
	updated := append([]byte(nil), base...)
	updated[1000] ^= 0xFF

	delta, err := utils.EncodeDelta(base, updated, 3)
	if err != nil {
		t.Fatalf("EncodeDelta returned error: %v", err)
	}

	targetPath := filepath.Join(dir, "big.bin")
	if err := os.WriteFile(targetPath, base, 0644); err != nil {
		t.Fatalf("failed to write base file: %v", err)
	}

	op := utils.PatchOperation{
		Type:        utils.OpModify,
		FilePath:    "data/big.bin",
		OldChecksum: utils.CalculateDataChecksum(base),
		// NewChecksum deliberately names different data: the
		// reconstruction must be detected as corrupt before writing.
		NewChecksum: utils.CalculateDataChecksum([]byte("completely different content")),
		BinaryDiff:  delta,
	}

	a := NewApplier()
	_, err = a.reconstructDelta(targetPath, op)
	if err == nil {
		t.Fatal("reconstructDelta must reject a reconstruction that does not match NewChecksum")
	}
	if !strings.Contains(err.Error(), "delta reconstruction checksum mismatch") {
		t.Errorf("error %q must contain 'delta reconstruction checksum mismatch'", err)
	}
}

// TestApplyDeltaRollbackRestoresFiles verifies that when one delta op fails
// mid-apply, the automatic rollback restores the already-patched file from
// the backup while the failing file keeps its (tampered) content.
func TestApplyDeltaRollbackRestoresFiles(t *testing.T) {
	baseDir := t.TempDir()
	v1Path := filepath.Join(baseDir, "v1")
	v2Path := filepath.Join(baseDir, "v2")

	firstOld := deterministicBytes(101, 1024*1024)
	secondOld := deterministicBytes(202, 1024*1024)
	firstNew := append([]byte(nil), firstOld...)
	firstNew[512] ^= 0xFF
	secondNew := append([]byte(nil), secondOld...)
	secondNew[999] ^= 0xFF

	key := []byte("CyberPatchMaker delta test key file\n")
	v1 := makeVersionDir(t, v1Path, map[string][]byte{
		deltaTestKeyFile:  key,
		"data/first.bin":  firstOld,
		"data/second.bin": secondOld,
	})
	v2 := makeVersionDir(t, v2Path, map[string][]byte{
		deltaTestKeyFile:  key,
		"data/first.bin":  firstNew,
		"data/second.bin": secondNew,
	})

	options := &utils.PatchOptions{Compression: "zstd", CompressionLevel: 3, SkipIdentical: true}
	g := NewGenerator()
	patch, err := g.GeneratePatch(v1, v2, options)
	if err != nil {
		t.Fatalf("GeneratePatch returned error: %v", err)
	}

	firstOp := findModifyOp(t, patch, "data/first.bin")
	secondOp := findModifyOp(t, patch, "data/second.bin")
	if len(firstOp.BinaryDiff) == 0 || len(secondOp.BinaryDiff) == 0 {
		t.Fatal("expected both modified files to be stored as deltas")
	}

	target := filepath.Join(baseDir, "target")
	if err := utils.CopyDir(v1Path, target); err != nil {
		t.Fatalf("failed to copy v1 to target: %v", err)
	}

	// Tamper second.bin so its op fails on the old checksum check.
	tamperedSecond := append([]byte(nil), secondOld...)
	tamperedSecond[12345] ^= 0xFF
	secondPath := filepath.Join(target, "data", "second.bin")
	if err := os.WriteFile(secondPath, tamperedSecond, 0644); err != nil {
		t.Fatalf("failed to tamper target second.bin: %v", err)
	}

	// Backup on, pre-verification off: the failure happens during apply and
	// must trigger the automatic rollback.
	a := NewApplier()
	err = a.ApplyPatchWithPath(patch, target, "", false, false, true)
	if err == nil {
		t.Fatal("ApplyPatchWithPath must fail when second.bin does not match its old checksum")
	}

	// first.bin must be back to its pre-patch (v1) content.
	gotFirst, err := os.ReadFile(filepath.Join(target, "data", "first.bin"))
	if err != nil {
		t.Fatalf("failed to read target data/first.bin: %v", err)
	}
	if !bytes.Equal(gotFirst, firstOld) {
		t.Error("target data/first.bin must be restored to the v1 content after rollback")
	}

	// second.bin must still hold the tampered bytes.
	gotSecond, err := os.ReadFile(secondPath)
	if err != nil {
		t.Fatalf("failed to read target data/second.bin: %v", err)
	}
	if !bytes.Equal(gotSecond, tamperedSecond) {
		t.Error("target data/second.bin must still contain the tampered bytes after rollback")
	}

	// The backup directory must exist and contain both modified files.
	backupDir := filepath.Join(target, "backup.cyberpatcher")
	for _, rel := range []string{"data/first.bin", "data/second.bin"} {
		if !utils.FileExists(filepath.Join(backupDir, filepath.FromSlash(rel))) {
			t.Errorf("backup directory must contain %s", rel)
		}
	}
}

// TestGenerateReversePatchDelta verifies the downgrade direction: v2 -> v1
// also stores a delta, and applying it to a v2 copy yields the v1 content.
func TestGenerateReversePatchDelta(t *testing.T) {
	f := newDeltaFixture(t)

	options := &utils.PatchOptions{Compression: "zstd", CompressionLevel: 3, SkipIdentical: true}
	g := NewGenerator()
	patch, err := g.GeneratePatch(f.v2, f.v1, options)
	if err != nil {
		t.Fatalf("GeneratePatch (reverse direction) returned error: %v", err)
	}

	op := findModifyOp(t, patch, "data/big.bin")
	if len(op.BinaryDiff) == 0 {
		t.Fatal("expected a delta op for data/big.bin in the reverse patch")
	}
	if len(op.NewFile) != 0 {
		t.Errorf("reverse delta op must not carry a full NewFile, got %d bytes", len(op.NewFile))
	}
	if want := utils.CalculateDataChecksum(f.bigNew); op.OldChecksum != want {
		t.Errorf("reverse op.OldChecksum = %s, want %s (v2 big.bin hash)", op.OldChecksum, want)
	}
	if want := utils.CalculateDataChecksum(f.bigOld); op.NewChecksum != want {
		t.Errorf("reverse op.NewChecksum = %s, want %s (v1 big.bin hash)", op.NewChecksum, want)
	}
	if err := g.ValidatePatch(patch); err != nil {
		t.Errorf("ValidatePatch returned error: %v", err)
	}

	target := filepath.Join(f.baseDir, "reverse-target")
	if err := utils.CopyDir(f.v2Path, target); err != nil {
		t.Fatalf("failed to copy v2 to target: %v", err)
	}

	a := NewApplier()
	if err := a.ApplyPatchWithPath(patch, target, "", true, true, true); err != nil {
		t.Fatalf("ApplyPatchWithPath returned error: %v", err)
	}

	got, err := utils.CalculateFileChecksum(filepath.Join(target, "data", "big.bin"))
	if err != nil {
		t.Fatalf("failed to hash patched data/big.bin: %v", err)
	}
	if want := utils.CalculateDataChecksum(f.bigOld); got != want {
		t.Errorf("patched data/big.bin checksum = %s, want %s (v1 content)", got, want)
	}
}
