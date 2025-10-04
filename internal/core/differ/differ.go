package differ

import (
	"fmt"
	"io"
	"os"

	"github.com/gabstv/go-bsdiff/pkg/bsdiff"
	"github.com/gabstv/go-bsdiff/pkg/bspatch"
)

// Differ handles binary diff operations
type Differ struct{}

// NewDiffer creates a new differ
func NewDiffer() *Differ {
	return &Differ{}
}

// GenerateDiff generates a binary diff between two files
func (d *Differ) GenerateDiff(oldFilePath, newFilePath string) ([]byte, error) {
	// Read old file
	oldData, err := os.ReadFile(oldFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read old file: %w", err)
	}

	// Read new file
	newData, err := os.ReadFile(newFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read new file: %w", err)
	}

	return d.GenerateDiffFromData(oldData, newData)
}

// GenerateDiffFromData generates a binary diff from byte slices
func (d *Differ) GenerateDiffFromData(oldData, newData []byte) ([]byte, error) {
	if len(oldData) == 0 {
		return nil, fmt.Errorf("old data is empty")
	}
	if len(newData) == 0 {
		return nil, fmt.Errorf("new data is empty")
	}

	// Generate diff using bsdiff (returns the patch bytes directly)
	patchData, err := bsdiff.Bytes(oldData, newData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate diff: %w", err)
	}

	return patchData, nil
}

// GenerateDiffStreaming generates a binary diff for large files using streaming
func (d *Differ) GenerateDiffStreaming(oldFilePath, newFilePath string, output io.Writer) error {
	// Open old file
	oldFile, err := os.Open(oldFilePath)
	if err != nil {
		return fmt.Errorf("failed to open old file: %w", err)
	}
	defer oldFile.Close()

	// Open new file
	newFile, err := os.Open(newFilePath)
	if err != nil {
		return fmt.Errorf("failed to open new file: %w", err)
	}
	defer newFile.Close()

	// Read both files into memory (bsdiff requires full data)
	// For very large files, this could be optimized with chunking
	oldData, err := io.ReadAll(oldFile)
	if err != nil {
		return fmt.Errorf("failed to read old file: %w", err)
	}

	newData, err := io.ReadAll(newFile)
	if err != nil {
		return fmt.Errorf("failed to read new file: %w", err)
	}

	// Generate diff
	patchData, err := bsdiff.Bytes(oldData, newData)
	if err != nil {
		return fmt.Errorf("failed to generate diff: %w", err)
	}

	// Write patch to output
	if _, err := output.Write(patchData); err != nil {
		return fmt.Errorf("failed to write patch data: %w", err)
	}

	return nil
}

// ApplyPatch applies a binary patch to a file
func (d *Differ) ApplyPatch(oldFilePath string, patchData []byte) ([]byte, error) {
	// Read old file
	oldData, err := os.ReadFile(oldFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read old file: %w", err)
	}

	return d.ApplyPatchToData(oldData, patchData)
}

// ApplyPatchToData applies a binary patch to byte data
func (d *Differ) ApplyPatchToData(oldData, patchData []byte) ([]byte, error) {
	if len(oldData) == 0 {
		return nil, fmt.Errorf("old data is empty")
	}
	if len(patchData) == 0 {
		return nil, fmt.Errorf("patch data is empty")
	}

	// Apply patch using bspatch (returns the new data directly)
	newData, err := bspatch.Bytes(oldData, patchData)
	if err != nil {
		return nil, fmt.Errorf("failed to apply patch: %w", err)
	}

	return newData, nil
}

// ApplyPatchToFile applies a binary patch and writes the result to a file
func (d *Differ) ApplyPatchToFile(oldFilePath, outputPath string, patchData []byte) error {
	// Apply patch
	newData, err := d.ApplyPatch(oldFilePath, patchData)
	if err != nil {
		return err
	}

	// Write to output file
	if err := os.WriteFile(outputPath, newData, 0644); err != nil {
		return fmt.Errorf("failed to write patched file: %w", err)
	}

	return nil
}

// CompareSizes returns the size difference between old and new data
func (d *Differ) CompareSizes(oldSize, newSize int64) (delta int64, percent float64) {
	delta = newSize - oldSize
	if oldSize > 0 {
		percent = (float64(delta) / float64(oldSize)) * 100
	}
	return delta, percent
}

// EstimatePatchSize estimates the patch size based on file sizes
// This is a rough estimate and actual size may vary
func (d *Differ) EstimatePatchSize(oldSize, newSize int64, similarity float64) int64 {
	// Very rough estimation:
	// If files are similar, patch size is roughly proportional to changes
	// If files are very different, patch size approaches newSize

	if similarity >= 0.9 {
		// Highly similar: patch is small
		return int64(float64(newSize) * (1 - similarity) * 1.2)
	} else if similarity >= 0.5 {
		// Moderately similar
		return int64(float64(newSize) * (1 - similarity) * 1.5)
	} else {
		// Very different: patch size approaches new file size
		return int64(float64(newSize) * 0.8)
	}
}

// ValidatePatch validates that patch data appears to be valid bsdiff format
func (d *Differ) ValidatePatch(patchData []byte) error {
	if len(patchData) < 32 {
		return fmt.Errorf("patch data too small to be valid")
	}

	// Check for bsdiff header magic bytes
	// bsdiff patches start with "BSDIFF40" or "BSDIFF4\x00"
	header := patchData[:8]
	if string(header[:7]) != "BSDIFF4" {
		return fmt.Errorf("invalid patch header: missing bsdiff magic bytes")
	}

	return nil
}
