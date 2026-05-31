package patcher

import (
	"crypto/sha256"
	"encoding/json"
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
func (g *Generator) SaveMultiPartPatch(parts []*utils.Patch, basePath string, compression string, chunkSize int64, level int) error {
	if len(parts) == 0 {
		return fmt.Errorf("no parts to save")
	}

	// Extract base filename and directory
	baseDir := filepath.Dir(basePath)
	baseFile := filepath.Base(basePath)
	baseFile = strings.TrimSuffix(baseFile, ".patch")

	// First, save all parts to their final locations (without PartHashes yet)
	partPaths := make([]string, len(parts))
	for i, part := range parts {
		partFile := fmt.Sprintf("%s.%02d.patch", baseFile, i+1)
		partPath := filepath.Join(baseDir, partFile)
		partPaths[i] = partPath

		// For parts 2+, don't include PartHashes in metadata
		if i > 0 && part.MultiPart != nil {
			part.MultiPart.PartHashes = nil
		}

		if err := utils.SavePatch(part, partPath, compression, level); err != nil {
			return fmt.Errorf("failed to save part %d: %w", i+1, err)
		}
	}

	// Now calculate hashes of all saved part files
	partHashes := make([]utils.PartHash, len(parts))

	// Prepare PartChunks map if needed (only stored in part 1)
	if parts[0].MultiPart != nil {
		if parts[0].MultiPart.PartHashes == nil {
			parts[0].MultiPart.PartHashes = make([]utils.PartHash, len(parts))
		}
	}

	// Track which parts were chunked so we can write a small stub for part 1 if needed
	chunked := make([]bool, len(parts))

	// Iterate and compute hashes; also perform chunking if requested
	for i, partPath := range partPaths {
		data, err := os.ReadFile(partPath)
		if err != nil {
			return fmt.Errorf("failed to read saved part %d for hashing: %w", i+1, err)
		}

		// Compute overall checksum for the part
		fullHash := sha256.Sum256(data)
		partHashes[i] = utils.PartHash{
			PartNumber: i + 1,
			Checksum:   fmt.Sprintf("%x", fullHash),
			Size:       int64(len(data)),
		}

		// If chunking requested and part exceeds chunkSize, split into chunks
		if chunkSize > 0 && int64(len(data)) > chunkSize {
			// Ensure PartChunks map exists
			if parts[0].MultiPart != nil {
				if parts[0].MultiPart.PartHashes == nil {
					parts[0].MultiPart.PartHashes = make([]utils.PartHash, len(parts))
				}
			}

			// Create chunk entries
			var chunks []utils.PartChunk
			totalLen := len(data)
			chunkIndex := 0
			for offset := 0; offset < totalLen; offset += int(chunkSize) {
				end := offset + int(chunkSize)
				if end > totalLen {
					end = totalLen
				}
				chunkIndex++
				chunkData := data[offset:end]

				// Chunk filename: <baseFile>.part<partNum>.<chunkIdx>.patch
				chunkFileName := fmt.Sprintf("%s.part%d.%d.patch", baseFile, i+1, chunkIndex)
				chunkPath := filepath.Join(baseDir, chunkFileName)

				if err := os.WriteFile(chunkPath, chunkData, 0644); err != nil {
					return fmt.Errorf("failed to write chunk file %s: %w", chunkPath, err)
				}

				// Compute chunk checksum
				chash := sha256.Sum256(chunkData)
				chunks = append(chunks, utils.PartChunk{
					PartNumber:  i + 1,
					ChunkNumber: chunkIndex,
					FileName:    chunkFileName,
					Checksum:    fmt.Sprintf("%x", chash),
					Size:        int64(len(chunkData)),
				})
			}

			// Remove original large part file to avoid confusion (we will reconstruct when loading)
			if err := os.Remove(partPath); err != nil {
				return fmt.Errorf("failed to remove original part after chunking: %w", err)
			}

			// Mark this part as chunked
			chunked[i] = true

			// Save chunk sidecar JSON file for reconstruction during load
			// Sidecar filename: <baseFile>.part<partNum>.chunks.json
			sidecarName := fmt.Sprintf("%s.part%d.chunks.json", baseFile, i+1)
			sidecarPath := filepath.Join(baseDir, sidecarName)
			// Build minimal JSON for chunks
			type chunkOut struct {
				PartNumber int                 `json:"part_number"`
				Chunks     []utils.PartChunk   `json:"chunks"`
			}
			out := chunkOut{PartNumber: i + 1, Chunks: chunks}
			jb, err := json.MarshalIndent(out, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal chunk sidecar: %w", err)
			}
			if err := os.WriteFile(sidecarPath, jb, 0644); err != nil {
				return fmt.Errorf("failed to write chunk sidecar: %w", err)
			}

			// Also record chunk sidecar filename in a simple header area: we will set a PartHash.Size to the combined size (already set) and rely on sidecar for reconstruction
			// The part file on disk was removed; reconstruction will use sidecar + chunk files
		}
	}

	// Attach final PartHashes to part 1 and save it again.
	// If part 1 was chunked, write a small stub part 01 that contains only header/multi-part metadata
	parts[0].MultiPart.PartHashes = partHashes

	if chunked[0] {
		// Create a shallow copy and strip large binary blobs so the saved part 01 is small
		stub := *parts[0]
		stubOps := make([]utils.PatchOperation, len(stub.Operations))
		for i, op := range stub.Operations {
			// Copy operation but remove large binary fields
			stubOps[i] = utils.PatchOperation{
				Type:        op.Type,
				FilePath:    op.FilePath,
				BinaryDiff:  nil,
				NewFile:     nil,
				OldChecksum: op.OldChecksum,
				NewChecksum: op.NewChecksum,
				Size:        0,
			}
		}
		stub.Operations = stubOps

		// Save stub part 01
		if err := utils.SavePatch(&stub, partPaths[0], compression, level); err != nil {
			return fmt.Errorf("failed to save stubbed part 1: %w", err)
		}
		// Update reported size for part 1 to the stub size
		finalData, err := os.ReadFile(partPaths[0])
		if err == nil {
			partHashes[0].Size = int64(len(finalData))
		}
	} else {
		// No stub needed; save full part 01 (with PartHashes filled)
		if err := utils.SavePatch(parts[0], partPaths[0], compression, level); err != nil {
			return fmt.Errorf("failed to save updated part 1: %w", err)
		}

		// Update sizes if part1 file changed
		finalData, err := os.ReadFile(partPaths[0])
		if err == nil {
			partHashes[0].Size = int64(len(finalData))
		}
	}

	fmt.Printf("✓ Multi-part patch saved successfully:\n")
	fmt.Printf("  Base name: %s\n", baseFile)
	fmt.Printf("  Total parts: %d\n", len(parts))
	fmt.Printf("  Compression: %s\n", compression)
	for i, ph := range partHashes {
		fmt.Printf("  Part %d: %d bytes\n", i+1, ph.Size)
	}

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

		// Check for chunk sidecar for this part: <baseFile>.part<partNum>.chunks.json
		sidecarName := fmt.Sprintf("%s.part%d.chunks.json", baseFile, i)
		sidecarPath := filepath.Join(baseDir, sidecarName)

		var toLoadPath string

		if utils.FileExists(sidecarPath) {
			// Reconstruct full part from chunks listed in sidecar
			sideData, err := os.ReadFile(sidecarPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read chunk sidecar for part %d: %w", i, err)
			}

			var parsed struct {
				PartNumber int               `json:"part_number"`
				Chunks     []utils.PartChunk `json:"chunks"`
			}
			if err := json.Unmarshal(sideData, &parsed); err != nil {
				return nil, fmt.Errorf("failed to parse chunk sidecar for part %d: %w", i, err)
			}

			// Create temporary file to reassemble
			tmp, err := os.CreateTemp("", "cpm_part_reconstruct_*.patch")
			if err != nil {
				return nil, fmt.Errorf("failed to create temp file for part %d reconstruction: %w", i, err)
			}
			defer func() { tmp.Close(); os.Remove(tmp.Name()) }()

			// Write chunks in order
			for _, chunk := range parsed.Chunks {
				chunkPath := filepath.Join(baseDir, chunk.FileName)
				cdata, err := os.ReadFile(chunkPath)
				if err != nil {
					return nil, fmt.Errorf("failed to read chunk %s for part %d: %w", chunk.FileName, i, err)
				}
				// verify chunk checksum
				ch := sha256.Sum256(cdata)
				if fmt.Sprintf("%x", ch) != chunk.Checksum {
					return nil, fmt.Errorf("chunk %s checksum mismatch for part %d", chunk.FileName, i)
				}
				if _, err := tmp.Write(cdata); err != nil {
					return nil, fmt.Errorf("failed to write chunk data to temp for part %d: %w", i, err)
				}
			}

			// Ensure flush
			if err := tmp.Sync(); err != nil {
				return nil, fmt.Errorf("failed to sync temp file for part %d: %w", i, err)
			}

			// Compute overall hash of reconstructed file and compare to expected
			if _, err := tmp.Seek(0, 0); err != nil {
				return nil, fmt.Errorf("failed to rewind temp file for part %d: %w", i, err)
			}
			reconData, err := os.ReadFile(tmp.Name())
			if err != nil {
				return nil, fmt.Errorf("failed to read reconstructed temp for part %d: %w", i, err)
			}
			rhash := sha256.Sum256(reconData)
			expectedHash := part1.MultiPart.PartHashes[i-1].Checksum
			if fmt.Sprintf("%x", rhash) != expectedHash {
				return nil, fmt.Errorf("reconstructed part %d hash mismatch: expected %s, got %s", i, expectedHash[:16]+"...", fmt.Sprintf("%x", rhash)[:16]+"...")
			}

			fmt.Printf("  ✓ Reconstructed Part %d hash verified\n", i)

			toLoadPath = tmp.Name()
		} else {
			// No sidecar; load the part file directly
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

			toLoadPath = partPath
		}

		// Load part (either reconstructed temp or the file on disk)
		part, err := utils.LoadPatch(toLoadPath)
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
