package patcher

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/cyberofficial/cyberpatchmaker/pkg/utils"
)

// SplitPatchIntoParts splits a large patch into multiple parts based on size constraints
func (g *Generator) SplitPatchIntoParts(patch *utils.Patch, maxPartSize int64) ([]*utils.Patch, error) {
	// Calculate total size
	totalSize := g.CalculatePatchSize(patch)

	// If patch is smaller than max size, return as single-part
	if totalSize <= maxPartSize {
		return []*utils.Patch{patch}, nil
	}

	fmt.Printf("\nPatch size (%d bytes) exceeds limit (%d bytes), splitting into multiple parts...\n",
		totalSize, maxPartSize)

	// Sort operations by size (smallest first, but key file operations first)
	sortedOps := make([]utils.PatchOperation, len(patch.Operations))
	copy(sortedOps, patch.Operations)

	sort.Slice(sortedOps, func(i, j int) bool {
		// Prioritize key file operations (must be in part 1)
		iIsKeyFile := sortedOps[i].FilePath == patch.ToKeyFile.Path
		jIsKeyFile := sortedOps[j].FilePath == patch.ToKeyFile.Path

		if iIsKeyFile != jIsKeyFile {
			return iIsKeyFile // Key file comes first
		}

		// Then sort by size (smallest first)
		return sortedOps[i].Size < sortedOps[j].Size
	})

	// Split operations into parts
	var parts []*utils.Patch
	var currentPart *utils.Patch
	var currentSize int64

	for _, op := range sortedOps {
		opSize := op.Size
		if op.NewFile != nil {
			opSize = int64(len(op.NewFile))
		}

		// Check if we need a new part
		needNewPart := currentPart == nil ||
			(currentSize+opSize > maxPartSize && len(currentPart.Operations) > 0)

		if needNewPart {
			// If operation itself is larger than maxPartSize, it gets its own part
			if opSize > maxPartSize && currentPart != nil {
				parts = append(parts, currentPart)
				currentPart = nil
				currentSize = 0
			}

			// Create new part
			if currentPart != nil {
				parts = append(parts, currentPart)
			}

			currentPart = &utils.Patch{
				Header:        patch.Header,
				FromVersion:   patch.FromVersion,
				ToVersion:     patch.ToVersion,
				FromKeyFile:   patch.FromKeyFile,
				ToKeyFile:     patch.ToKeyFile,
				RequiredFiles: patch.RequiredFiles,
				Operations:    make([]utils.PatchOperation, 0),
				SimpleMode:    patch.SimpleMode,
			}
			currentSize = 0
		}

		// Add operation to current part
		currentPart.Operations = append(currentPart.Operations, op)
		currentSize += opSize
	}

	// Add last part
	if currentPart != nil && len(currentPart.Operations) > 0 {
		parts = append(parts, currentPart)
	}

	// Update multi-part metadata
	totalParts := len(parts)
	for i, part := range parts {
		part.MultiPart = &utils.MultiPartInfo{
			IsMultiPart: true,
			PartNumber:  i + 1,
			TotalParts:  totalParts,
			MaxPartSize: maxPartSize,
		}

		// Only part 1 will have PartHashes (filled after saving all parts)
		if i == 0 {
			part.MultiPart.PartHashes = make([]utils.PartHash, totalParts)
		}
	}

	fmt.Printf("Split patch into %d parts\n", totalParts)
	for i, part := range parts {
		partSize := g.CalculatePatchSize(part)
		fmt.Printf("  Part %d: %d operations, ~%d bytes\n", i+1, len(part.Operations), partSize)
	}

	return parts, nil
}

// SaveMultiPartPatch saves a multi-part patch to disk
func (g *Generator) SaveMultiPartPatch(parts []*utils.Patch, basePath string, compression string) error {
	if len(parts) == 0 {
		return fmt.Errorf("no parts to save")
	}

	// Extract base filename and directory
	baseDir := filepath.Dir(basePath)
	baseFile := filepath.Base(basePath)
	baseFile = strings.TrimSuffix(baseFile, ".patch")

	// Pre-calculate part hashes by saving parts to temporary files first
	partHashes := make([]utils.PartHash, len(parts))
	tempFiles := make([]string, len(parts))

	for i, part := range parts {
		// Generate temporary filename
		tempFile := fmt.Sprintf("%s.%02d.patch.tmp", baseFile, i+1)
		tempPath := filepath.Join(baseDir, tempFile)

		// Save part to temporary file
		if err := utils.SavePatch(part, tempPath, compression); err != nil {
			return fmt.Errorf("failed to save part %d: %w", i+1, err)
		}

		// Calculate hash of the temporary file
		data, err := os.ReadFile(tempPath)
		if err != nil {
			os.Remove(tempPath) // Clean up on error
			return fmt.Errorf("failed to read part %d for hashing: %w", i+1, err)
		}

		hash := sha256.Sum256(data)
		partHashes[i] = utils.PartHash{
			PartNumber: i + 1,
			Checksum:   fmt.Sprintf("%x", hash),
			Size:       int64(len(data)),
		}

		tempFiles[i] = tempPath
	}

	// Update all parts with complete hash information
	for _, part := range parts {
		if part.MultiPart != nil {
			part.MultiPart.PartHashes = partHashes
		}
	}

	// Now save all parts with correct hash information
	for i, part := range parts {
		// Generate final filename
		partFile := fmt.Sprintf("%s.%02d.patch", baseFile, i+1)
		partPath := filepath.Join(baseDir, partFile)

		// Remove temp file and save final version
		os.Remove(tempFiles[i])
		if err := utils.SavePatch(part, partPath, compression); err != nil {
			return fmt.Errorf("failed to save final part %d: %w", i+1, err)
		}

		fmt.Printf("Saved part %d: %s (%d bytes)\n", i+1, partPath, partHashes[i].Size)
	}

	fmt.Printf("\n✓ Multi-part patch saved successfully:\n")
	fmt.Printf("  Base name: %s\n", baseFile)
	fmt.Printf("  Total parts: %d\n", len(parts))
	fmt.Printf("  Compression: %s\n", compression)

	return nil
}

// LoadMultiPartPatch loads all parts of a multi-part patch
func LoadMultiPartPatch(part1Path string) (*utils.Patch, error) {
	// Load part 1
	part1, err := utils.LoadPatch(part1Path)
	if err != nil {
		return nil, fmt.Errorf("failed to load part 1: %w", err)
	}

	// Check if it's multi-part
	if part1.MultiPart == nil || !part1.MultiPart.IsMultiPart {
		// Single-part patch
		return part1, nil
	}

	fmt.Printf("Detected multi-part patch: %d parts\n", part1.MultiPart.TotalParts)

	// Verify part 1 has hash information
	if len(part1.MultiPart.PartHashes) != part1.MultiPart.TotalParts {
		return nil, fmt.Errorf("part 1 missing hash information for all parts")
	}

	// Extract base path
	baseDir := filepath.Dir(part1Path)
	baseFile := filepath.Base(part1Path)
	baseFile = strings.TrimSuffix(baseFile, ".01.patch")

	// Load and verify all parts
	var allOperations []utils.PatchOperation
	allOperations = append(allOperations, part1.Operations...)

	for i := 2; i <= part1.MultiPart.TotalParts; i++ {
		partFile := fmt.Sprintf("%s.%02d.patch", baseFile, i)
		partPath := filepath.Join(baseDir, partFile)

		fmt.Printf("Loading part %d: %s\n", i, partFile)

		// Read part file
		partData, err := os.ReadFile(partPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read part %d: %w", i, err)
		}

		// Verify hash
		hash := sha256.Sum256(partData)
		expectedHash := part1.MultiPart.PartHashes[i-1].Checksum
		actualHash := fmt.Sprintf("%x", hash)

		if actualHash != expectedHash {
			return nil, fmt.Errorf("part %d hash mismatch: expected %s, got %s",
				i, expectedHash[:16]+"...", actualHash[:16]+"...")
		}

		fmt.Printf("  ✓ Part %d hash verified\n", i)

		// Load part
		part, err := utils.LoadPatch(partPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load part %d: %w", i, err)
		}

		// Merge operations
		allOperations = append(allOperations, part.Operations...)
	}

	// Create combined patch
	combinedPatch := &utils.Patch{
		Header:        part1.Header,
		FromVersion:   part1.FromVersion,
		ToVersion:     part1.ToVersion,
		FromKeyFile:   part1.FromKeyFile,
		ToKeyFile:     part1.ToKeyFile,
		RequiredFiles: part1.RequiredFiles,
		Operations:    allOperations,
		SimpleMode:    part1.SimpleMode,
		MultiPart:     part1.MultiPart, // Keep multi-part info for reference
	}

	fmt.Printf("✓ Loaded %d total operations from %d parts\n",
		len(allOperations), part1.MultiPart.TotalParts)

	return combinedPatch, nil
}
