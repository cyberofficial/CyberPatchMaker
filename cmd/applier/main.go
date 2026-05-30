package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cyberofficial/cyberpatchmaker/internal/core/patcher"
	"github.com/cyberofficial/cyberpatchmaker/internal/core/version"
	"github.com/cyberofficial/cyberpatchmaker/pkg/utils"
)

const (
	MAGIC_BYTES = "CPMPATCH"
	HEADER_SIZE = 128
)

type EmbeddedPatchHeader struct {
	Magic       [8]byte
	Version     uint32
	StubSize    uint64
	DataOffset  uint64
	DataSize    uint64
	Compression [16]byte
	Checksum    [32]byte
	Flags       byte
	Reserved    [43]byte
}

func main() {
	// Define flags
	patchFile := flag.String("patch", "", "Path to patch file")
	currentDir := flag.String("current-dir", "", "Directory containing current version")
	keyFile := flag.String("key-file", "", "Custom key file path (if renamed or moved)")
	dryRun := flag.Bool("dry-run", false, "Simulate patch without making changes")
	verify := flag.Bool("verify", true, "Verify file hashes before and after patching")
	backup := flag.Bool("backup", true, "Create backup before patching")
	ignore1GB := flag.Bool("ignore1gb", false, "Bypass 1GB patch size limit (use with caution)")
	silent := flag.Bool("silent", false, "Silent mode: apply patch automatically without prompts (for automation)")
	versionFlag := flag.Bool("version", false, "Show version information")
	help := flag.Bool("help", false, "Show help message")

	flag.Parse()

	// Show version if requested
	if *versionFlag {
		fmt.Printf("CyberPatchMaker Patch Applier v%s\n", version.GetVersion())
		return
	}

	if *help {
		printHelp()
		return
	}

	// Check if patch data is embedded in this executable
	patch, targetDir, isEmbedded, embeddedSilent := checkEmbeddedPatch(*ignore1GB)

	if isEmbedded && patch != nil {
		// Use embedded silent flag if set, otherwise check command-line flag
		if embeddedSilent || *silent {
			// Silent mode: apply patch automatically
			runSilentMode(patch, targetDir, *currentDir, *keyFile)
			return
		}
		// Interactive console mode for embedded patch
		fmt.Println("==============================================")
		fmt.Println("  CyberPatchMaker - Self-Contained Patch")
		fmt.Println("==============================================")
		runInteractiveMode(patch, targetDir, *ignore1GB)
		return
	}

	// Standard mode - require arguments
	if *patchFile == "" || *currentDir == "" {
		fmt.Println("Error: --patch and --current-dir are required")
		printHelp()
		os.Exit(1)
	}

	// Check if patch file exists
	if !utils.FileExists(*patchFile) {
		fmt.Printf("Error: patch file not found: %s\n", *patchFile)
		os.Exit(1)
	}

	// Check if current directory exists
	if !utils.FileExists(*currentDir) {
		fmt.Printf("Error: current directory not found: %s\n", *currentDir)
		os.Exit(1)
	}

	// Load patch
	patch, err := loadPatch(*patchFile)
	if err != nil {
		fmt.Printf("Error: failed to load patch: %v\n", err)
		os.Exit(1)
	}

	// Display patch information
	displayPatchInfo(patch)

	// Override key file path if custom one is provided
	if *keyFile != "" {
		fmt.Printf("\nUsing custom key file: %s\n", *keyFile)
		patch.FromKeyFile.Path = *keyFile
	}

	if *dryRun {
		fmt.Println("\n=== DRY RUN MODE ===")
		fmt.Println("No changes will be made")
		performDryRun(patch, *currentDir, *keyFile)
		return
	}

	applier := patcher.NewApplier()
	if err := applier.ApplyPatchWithPath(patch, *currentDir, *patchFile, *verify, *verify, *backup); err != nil {
		fmt.Printf("Error: patch application failed: %v\n", err)
		if *backup {
			fmt.Println("\nNote: If backup was created, automatic rollback may have been performed to restore original files.")
		}
		os.Exit(1)
	}

	fmt.Println("\n=== Patch Applied Successfully ===")
	fmt.Printf("Version updated from %s to %s\n", patch.FromVersion, patch.ToVersion)
}

func loadPatch(filename string) (*utils.Patch, error) {
	// Check if this is a multi-part patch (has .01.patch, .02.patch, etc. naming)
	if strings.HasSuffix(filename, ".01.patch") {
		// Load all parts of multi-part patch
		fmt.Println("Detected multi-part patch, loading all parts...")
		patch, err := patcher.LoadMultiPartPatch(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to load multi-part patch: %w", err)
		}
		return patch, nil
	}

	// Check if user provided a non-.01 part number
	// Look for pattern: .XX.patch where XX is 02-99
	for i := 2; i <= 99; i++ {
		partSuffix := fmt.Sprintf(".%02d.patch", i)
		if strings.HasSuffix(filename, partSuffix) {
			// User provided a non-first part, redirect to part 1
			part1File := strings.Replace(filename, partSuffix, ".01.patch", 1)
			if utils.FileExists(part1File) {
				fmt.Printf("Note: Part %d detected, loading from part 1: %s\n", i, part1File)
				return loadPatch(part1File)
			}
		}
	}

	// Single-part patch (legacy format)
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open patch file: %w", err)
	}
	defer file.Close()

	return parsePatchDataStreaming(file)
}

//// parsePatchData parses patch data, automatically detecting and decompressing if needed
//func parsePatchData(data []byte) (*utils.Patch, error) {
//	// Try to detect compression and decompress
//	// First try as JSON directly
//	var patch utils.Patch
//	if err := json.Unmarshal(data, &patch); err != nil {
//		// Try decompressing with zstd
//		decompressed, err := utils.DecompressData(data, "zstd")
//		if err != nil {
//			// Try gzip
//			decompressed, err = utils.DecompressData(data, "gzip")
//			if err != nil {
//				return nil, fmt.Errorf("failed to decompress or parse patch: %w", err)
//			}
//		}
//		data = decompressed
//
//		// Try parsing again
//		if err := json.Unmarshal(data, &patch); err != nil {
//			return nil, fmt.Errorf("failed to parse patch JSON: %w", err)
//		}
//	}
//
//	return &patch, nil
//}

// parsePatchDataStreaming parses patch data using streaming decompression with magic-byte detection.
func parsePatchDataStreaming(reader io.Reader) (*utils.Patch, error) {
	// Read first 4 bytes to detect compression format
	magic := make([]byte, 4)
	n, err := io.ReadFull(reader, magic)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return nil, fmt.Errorf("failed to read patch header: %w", err)
	}
	magic = magic[:n]

	// Detect compression from magic bytes
	algo := utils.DetectCompression(magic)

	// Reconstruct full reader
	fullReader := io.MultiReader(bytes.NewReader(magic), reader)

	var patchReader io.Reader
	switch algo {
	case "none":
		patchReader = fullReader
	case "zstd":
		pr, pw := io.Pipe()
		go func() {
			defer pw.Close()
			if err := utils.DecompressDataStreaming(fullReader, pw, "zstd"); err != nil {
				pw.CloseWithError(err)
			}
		}()
		patchReader = pr
	case "gzip":
		pr, pw := io.Pipe()
		go func() {
			defer pw.Close()
			if err := utils.DecompressDataStreaming(fullReader, pw, "gzip"); err != nil {
				pw.CloseWithError(err)
			}
		}()
		patchReader = pr
	default:
		return nil, fmt.Errorf("unsupported compression format")
	}

	// 64KB buffer is sufficient for streaming JSON decoding
	bufReader := bufio.NewReaderSize(patchReader, 64*1024)
	var patch utils.Patch
	decoder := json.NewDecoder(bufReader)
	decoder.UseNumber()
	if err := decoder.Decode(&patch); err != nil {
		return nil, fmt.Errorf("failed to parse patch: %w", err)
	}

	return &patch, nil
}

func displayPatchInfo(patch *utils.Patch) {
	fmt.Println("\n=== Patch Information ===")
	fmt.Printf("From Version:     %s\n", patch.FromVersion)
	fmt.Printf("To Version:       %s\n", patch.ToVersion)
	fmt.Printf("Key File:         %s\n", patch.FromKeyFile.Path)
	fmt.Printf("Required Hash:    %s\n", patch.FromKeyFile.Checksum)
	fmt.Printf("Patch Size:       %d bytes\n", patch.Header.PatchSize)
	fmt.Printf("Compression:      %s\n", patch.Header.Compression)
	fmt.Printf("Created:          %s\n", patch.Header.CreatedAt.Format("2006-01-02 15:04:05"))

	// Count operations
	addCount := 0
	modifyCount := 0
	deleteCount := 0
	addDirCount := 0
	deleteDirCount := 0

	for _, op := range patch.Operations {
		switch op.Type {
		case utils.OpAdd:
			addCount++
		case utils.OpModify:
			modifyCount++
		case utils.OpDelete:
			deleteCount++
		case utils.OpAddDir:
			addDirCount++
		case utils.OpDeleteDir:
			deleteDirCount++
		}
	}

	fmt.Printf("Files Added:      %d\n", addCount)
	fmt.Printf("Files Modified:   %d\n", modifyCount)
	fmt.Printf("Files Deleted:    %d\n", deleteCount)
	fmt.Printf("Dirs Added:       %d\n", addDirCount)
	fmt.Printf("Dirs Deleted:     %d\n", deleteDirCount)
	fmt.Printf("Required Files:   %d (must match exact hashes)\n", len(patch.RequiredFiles))
}

// resolveKeyFilePath resolves the actual key file path, using custom path if provided
func resolveKeyFilePath(patch *utils.Patch, currentDir string, customKeyFile string) string {
	if customKeyFile != "" {
		// Use custom key file path (can be absolute or relative)
		if strings.Contains(customKeyFile, string(os.PathSeparator)) || strings.Contains(customKeyFile, "/") {
			// If it contains path separators, treat as-is
			return customKeyFile
		}
		// Otherwise, treat as relative to currentDir
		return currentDir + string(os.PathSeparator) + customKeyFile
	}
	// Use default key file from patch
	return currentDir + string(os.PathSeparator) + patch.FromKeyFile.Path
}

func performDryRun(patch *utils.Patch, currentDir string, customKeyFile string) {
	fmt.Println("\nSimulating patch application...")

	// Verify key file
	if customKeyFile != "" {
		fmt.Printf("\nVerifying custom key file: %s\n", customKeyFile)
	} else {
		fmt.Printf("\nVerifying key file: %s\n", patch.FromKeyFile.Path)
	}
	keyFilePath := resolveKeyFilePath(patch, currentDir, customKeyFile)
	if !utils.FileExists(keyFilePath) {
		fmt.Printf("✗ Key file not found: %s\n", keyFilePath)
		return
	}

	checksum, err := utils.CalculateFileChecksum(keyFilePath)
	if err != nil {
		fmt.Printf("✗ Failed to calculate key file checksum: %v\n", err)
		return
	}

	if checksum != patch.FromKeyFile.Checksum {
		fmt.Printf("✗ Key file hash mismatch\n")
		fmt.Printf("  Expected: %s\n", patch.FromKeyFile.Checksum[:16]+"...")
		fmt.Printf("  Got:      %s\n", checksum[:16]+"...")
		return
	}
	fmt.Println("✓ Key file verified")

	// Verify required files
	fmt.Printf("\nVerifying %d required files...\n", len(patch.RequiredFiles))
	mismatches := 0
	for i, req := range patch.RequiredFiles {
		if i < 5 || mismatches > 0 { // Show first 5 or any mismatches
			filePath := currentDir + string(os.PathSeparator) + req.Path
			if !utils.FileExists(filePath) {
				fmt.Printf("✗ Required file missing: %s\n", req.Path)
				mismatches++
				continue
			}

			checksum, err := utils.CalculateFileChecksum(filePath)
			if err != nil {
				fmt.Printf("✗ Failed to verify: %s\n", req.Path)
				mismatches++
				continue
			}

			if checksum != req.Checksum {
				fmt.Printf("✗ Hash mismatch: %s\n", req.Path)
				mismatches++
			}
		}
	}

	if mismatches > 0 {
		fmt.Printf("\n✗ %d file(s) have mismatches - patch cannot be applied\n", mismatches)
		return
	}

	fmt.Println("✓ All required files verified")

	// Show operations that would be performed
	fmt.Println("\nOperations that would be performed:")
	for _, op := range patch.Operations {
		switch op.Type {
		case utils.OpAdd:
			fmt.Printf("  ADD: %s\n", op.FilePath)
		case utils.OpModify:
			fmt.Printf("  MODIFY: %s\n", op.FilePath)
		case utils.OpDelete:
			fmt.Printf("  DELETE: %s\n", op.FilePath)
		case utils.OpAddDir:
			fmt.Printf("  ADD DIR: %s\n", op.FilePath)
		case utils.OpDeleteDir:
			fmt.Printf("  DELETE DIR: %s\n", op.FilePath)
		}
	}

	fmt.Println("\n✓ Dry run completed - patch can be applied safely")
}

// checkEmbeddedPatch checks if this executable contains an embedded patch
// Returns: patch, targetDir, isEmbedded, embeddedSilent
func checkEmbeddedPatch(ignore1GB bool) (*utils.Patch, string, bool, bool) {
	// Get path to this executable
	exePath, err := os.Executable()
	if err != nil {
		return nil, "", false, false
	}

	// Open executable for reading
	file, err := os.Open(exePath)
	if err != nil {
		return nil, "", false, false
	}
	defer file.Close()

	// Get file size
	stat, err := file.Stat()
	if err != nil {
		return nil, "", false, false
	}
	fileSize := stat.Size()

	// Check if file is large enough for header
	if fileSize < HEADER_SIZE {
		return nil, "", false, false
	}

	// Read header from end of file
	headerOffset := fileSize - HEADER_SIZE
	if _, err := file.Seek(headerOffset, io.SeekStart); err != nil {
		return nil, "", false, false
	}

	headerBytes := make([]byte, HEADER_SIZE)
	if _, err := io.ReadFull(file, headerBytes); err != nil {
		return nil, "", false, false
	}

	// Parse header
	var header EmbeddedPatchHeader
	buf := bytes.NewReader(headerBytes)
	if err := binary.Read(buf, binary.LittleEndian, &header); err != nil {
		return nil, "", false, false
	}

	// Validate magic bytes
	magic := string(bytes.TrimRight(header.Magic[:], "\x00"))
	if magic != MAGIC_BYTES {
		return nil, "", false, false
	}

	// Validate version
	if header.Version != 1 {
		return nil, "", false, false
	}

	// Validate structure: data must start immediately after stub
	if header.DataOffset != header.StubSize {
		return nil, "", false, false
	}

	// Allow optional sidecar blob between patch data and header (we may embed chunk JSONs)
	minExpectedSize := header.StubSize + header.DataSize + HEADER_SIZE
	if uint64(fileSize) < minExpectedSize {
		return nil, "", false, false
	}

	// Extract silent flag from Flags byte (bit 0)
	embeddedSilent := (header.Flags & 0x01) != 0

	// Check size limit (1GB)
	const maxPatchSize = 1 << 30 // 1 GB
	if !ignore1GB && header.DataSize > maxPatchSize {
		fmt.Printf("Warning: Patch size (%d bytes) exceeds 1GB limit\n", header.DataSize)
		fmt.Println("Use --ignore1gb flag if you want to proceed anyway")
		return nil, "", false, false
	}


	// Read patch data
	if _, err := file.Seek(int64(header.DataOffset), io.SeekStart); err != nil {
		return nil, "", false, false
	}

	patchData := make([]byte, header.DataSize)
	if _, err := io.ReadFull(file, patchData); err != nil {
		return nil, "", false, false
	}

	// If there are extra bytes between patch data and header, treat them as sidecar blob
	extraBytes := int64(fileSize) - int64(minExpectedSize)
	var writtenSidecars []string
	if extraBytes > 0 {
		// Read sidecar blob
		sidecarOffset := int64(header.DataOffset + header.DataSize)
		if _, err := file.Seek(sidecarOffset, io.SeekStart); err == nil {
			sidecarData := make([]byte, extraBytes)
			if _, err := io.ReadFull(file, sidecarData); err == nil {
				// Parse sidecar format: uint32 count, then for each: uint16 nameLen, name bytes, uint64 dataLen, data bytes
				r := bytes.NewReader(sidecarData)
				var count uint32
				if err := binary.Read(r, binary.LittleEndian, &count); err == nil {
					for i := uint32(0); i < count; i++ {
						var nameLen uint16
						if err := binary.Read(r, binary.LittleEndian, &nameLen); err != nil {
							break
						}
						nameBytes := make([]byte, nameLen)
						if _, err := io.ReadFull(r, nameBytes); err != nil {
							break
						}
						var dataLen uint64
						if err := binary.Read(r, binary.LittleEndian, &dataLen); err != nil {
							break
						}
						dataBytes := make([]byte, dataLen)
						if _, err := io.ReadFull(r, dataBytes); err != nil {
							break
						}
						// Write sidecar file into exe directory
						exeDir := filepath.Dir(exePath)
						sidePath := filepath.Join(exeDir, string(nameBytes))
						if err := os.WriteFile(sidePath, dataBytes, 0644); err == nil {
							writtenSidecars = append(writtenSidecars, sidePath)
						}
					}
				}
			}
		}
	}

	// Verify checksum
	actualChecksum := sha256.Sum256(patchData)
	if !bytes.Equal(actualChecksum[:], header.Checksum[:]) {
		return nil, "", false, false
	}

	// The embedded patch data is the raw .patch file content (part 01 if multi-part)
	// Check if there are additional parts (.02, .03, etc.) in the same directory as the exe
	exeDir := filepath.Dir(exePath)
	exeBaseName := strings.TrimSuffix(filepath.Base(exePath), ".exe")

	// Check if part 02 exists to determine if this is multi-part
	part02Path := filepath.Join(exeDir, exeBaseName+".02.patch")
	isMultiPart := utils.FileExists(part02Path)

	var patch *utils.Patch
	if isMultiPart {
		// Multi-part patch: Save embedded part 01 temporarily and load all parts
		// Save in exe directory with the correct base name so LoadMultiPartPatch finds the parts
		tempPart01 := filepath.Join(exeDir, exeBaseName+".01.patch")

		// Before writing part 01 temp, verify that any extracted sidecars reference existing chunk files.
		// If any listed chunk file is missing, fail early to avoid silently falling back to a full part.
		exeDir := filepath.Dir(exePath)
		for _, sc := range writtenSidecars {
			// Read sidecar JSON
			sb, err := os.ReadFile(sc)
			if err != nil {
				// cleanup and fail
				for _, p := range writtenSidecars {
					_ = os.Remove(p)
				}
				return nil, "", false, false
			}
			var parsed struct {
				PartNumber int                 `json:"part_number"`
				Chunks     []utils.PartChunk   `json:"chunks"`
			}
			if err := json.Unmarshal(sb, &parsed); err != nil {
				for _, p := range writtenSidecars {
					_ = os.Remove(p)
				}
				return nil, "", false, false
			}
			// Verify each chunk file exists
			for _, ch := range parsed.Chunks {
				chunkPath := filepath.Join(exeDir, ch.FileName)
				if !utils.FileExists(chunkPath) {
					// cleanup and fail with clear message
					for _, p := range writtenSidecars {
						_ = os.Remove(p)
					}
					fmt.Fprintf(os.Stderr, "Error: missing chunk file required by embedded sidecar: %s\n", chunkPath)
					return nil, "", false, false
				}
			}
		}

		// Write part 01 data to temporary file
		if err := os.WriteFile(tempPart01, patchData, 0644); err != nil {
			// cleanup any written sidecars
			for _, p := range writtenSidecars {
				_ = os.Remove(p)
			}
			return nil, "", false, false
		}
		// Clean up temporary file and any written sidecars after loading
		defer func() {
			_ = os.Remove(tempPart01)
			for _, p := range writtenSidecars {
				_ = os.Remove(p)
			}
		}()

		// Load all parts using multi-part loader
		patch, err = patcher.LoadMultiPartPatch(tempPart01)
		if err != nil {
			return nil, "", false, false
		}

		fmt.Printf("✓ Loaded multi-part patch from embedded part 01 + external parts\n")
	} else {
		// Single-part patch: Parse embedded data directly
		patch, err = parsePatchDataStreaming(bytes.NewReader(patchData))
		if err != nil {
			// cleanup any written sidecars
			for _, p := range writtenSidecars {
				_ = os.Remove(p)
			}
			return nil, "", false, false
		}
	} // Get current directory as default target
	targetDir, _ := os.Getwd()

	return patch, targetDir, true, embeddedSilent
}

// runSilentMode applies the patch automatically without user interaction (for automation)
func runSilentMode(patch *utils.Patch, defaultTargetDir string, customTargetDir string, customKeyFile string) {
	// Use custom target directory if provided, otherwise use default (current directory)
	targetDir := defaultTargetDir
	if customTargetDir != "" {
		targetDir = customTargetDir
	}

	// Create log file with current epoch timestamp
	logFileName := fmt.Sprintf("log_%d.txt", time.Now().Unix())
	logFile, err := os.Create(logFileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to create log file: %v\n", err)
		logFile = nil
	}
	defer func() {
		if logFile != nil {
			logFile.Close()
		}
	}()

	// Helper function to write to both console and log
	logOutput := func(format string, args ...interface{}) {
		msg := fmt.Sprintf(format, args...)
		fmt.Print(msg)
		if logFile != nil {
			logFile.WriteString(msg)
		}
	}

	// Log header with timestamp
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logOutput("========================================\n")
	logOutput("CyberPatchMaker Silent Mode Log\n")
	logOutput("Started: %s\n", timestamp)
	logOutput("========================================\n\n")

	// Check if directory exists
	if !utils.FileExists(targetDir) {
		logOutput("Error: Target directory not found: %s\n", targetDir)
		logOutput("\n========================================\n")
		logOutput("Status: FAILED\n")
		logOutput("Completed: %s\n", time.Now().Format("2006-01-02 15:04:05"))
		logOutput("========================================\n")
		if logFile != nil {
			logOutput("\nLog saved to: %s\n", logFileName)
		}
		os.Exit(1)
	}

	// Override key file path if custom one is provided
	if customKeyFile != "" {
		patch.FromKeyFile.Path = customKeyFile
		logOutput("Using custom key file: %s\n", customKeyFile)
	}

	// Log patch details
	logOutput("Patch Information:\n")
	logOutput("  From Version: %s\n", patch.FromVersion)
	logOutput("  To Version:   %s\n", patch.ToVersion)
	logOutput("  Key File:     %s\n", patch.FromKeyFile.Path)
	logOutput("  Target Dir:   %s\n", targetDir)
	logOutput("  Compression:  %s\n", patch.Header.Compression)
	logOutput("\n")

	// Display simple startup message
	logOutput("Applying patch...\n\n")

	// Apply patch with default settings (verify=true, backup=true)
	applier := patcher.NewApplier()
	if err := applier.ApplyPatchWithPath(patch, targetDir, "", true, true, true); err != nil {
		logOutput("\nError: Patch application failed: %v\n", err)
		logOutput("\n========================================\n")
		logOutput("Status: FAILED\n")
		logOutput("Completed: %s\n", time.Now().Format("2006-01-02 15:04:05"))
		logOutput("========================================\n")
		if logFile != nil {
			logOutput("\nLog saved to: %s\n", logFileName)
		}
		os.Exit(1)
	}

	// Success - output minimal message
	logOutput("\n")
	logOutput("Patch applied successfully: %s → %s\n", patch.FromVersion, patch.ToVersion)
	logOutput("\n========================================\n")
	logOutput("Status: SUCCESS\n")
	logOutput("Completed: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	logOutput("========================================\n")
	if logFile != nil {
		logOutput("\nLog saved to: %s\n", logFileName)
	}
	os.Exit(0)
}

// runSimpleMode runs a fully automated interface for end users when patch creator enabled simple mode
// This mode automatically:
// - Uses current directory as target
// - Enables backup (yes by default)
// - Runs dry-run first to verify
// - Applies patch if dry-run succeeds
// - Logs everything to <patchname>_<utctime>_log.txt
func runSimpleMode(patch *utils.Patch, defaultTargetDir string) {
	// Use current directory as target
	targetDir := defaultTargetDir

	// Create log file with patch name and UTC timestamp
	exePath, _ := os.Executable()
	exeBaseName := strings.TrimSuffix(exePath, ".exe")
	if exeBaseName == exePath {
		exeBaseName = "patch"
	} else {
		// Extract just the filename without directory
		parts := strings.Split(exeBaseName, string(os.PathSeparator))
		if len(parts) > 0 {
			exeBaseName = parts[len(parts)-1]
		}
	}
	logFileName := fmt.Sprintf("%s_%d_log.txt", exeBaseName, time.Now().UTC().Unix())
	logFile, err := os.Create(logFileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to create log file: %v\n", err)
		logFile = nil
	}
	defer func() {
		if logFile != nil {
			logFile.Close()
		}
	}()

	// Helper function to write to both console and log
	logOutput := func(format string, args ...interface{}) {
		msg := fmt.Sprintf(format, args...)
		fmt.Print(msg)
		if logFile != nil {
			logFile.WriteString(msg)
		}
	}

	// Header
	logOutput("\n")
	logOutput("==============================================\n")
	logOutput("  CyberPatchMaker - Self-Contained Patch\n")
	logOutput("==============================================\n")
	logOutput("\n")
	logOutput("==============================================\n")
	logOutput("          Simple Patch Application\n")
	logOutput("==============================================\n")
	logOutput("\n")
	logOutput("Automated patching from \"%s\" to \"%s\"\n", patch.FromVersion, patch.ToVersion)
	logOutput("\n")

	// Log details
	timestamp := time.Now().UTC().Format("2006-01-02 15:04:05 UTC")
	logOutput("Patch Information:\n")
	logOutput("  Started:      %s\n", timestamp)
	logOutput("  From Version: %s\n", patch.FromVersion)
	logOutput("  To Version:   %s\n", patch.ToVersion)
	logOutput("  Key File:     %s\n", patch.FromKeyFile.Path)
	logOutput("  Target Dir:   %s\n", targetDir)
	logOutput("  Backup:       Enabled\n")
	logOutput("  Compression:  %s\n", patch.Header.Compression)
	logOutput("\n")

	// Check if directory exists
	if !utils.FileExists(targetDir) {
		logOutput("Error: Directory not found: %s\n", targetDir)
		logOutput("\n========================================\n")
		logOutput("Status: FAILED\n")
		logOutput("Completed: %s\n", time.Now().UTC().Format("2006-01-02 15:04:05 UTC"))
		logOutput("========================================\n")
		if logFile != nil {
			logOutput("\nLog saved to: %s\n", logFileName)
		}
		os.Exit(1)
	}

	// Step 1: Dry run
	logOutput("==============================================\n")
	logOutput("Step 1: Dry Run (Validation)\n")
	logOutput("==============================================\n")
	logOutput("\n")
	logOutput("Testing patch application without making changes...\n")
	logOutput("\n")

	// Perform dry run validation
	dryRunSuccess := true

	// Verify key file
	logOutput("Verifying key file: %s\n", patch.FromKeyFile.Path)
	keyFilePath := targetDir + string(os.PathSeparator) + patch.FromKeyFile.Path
	if !utils.FileExists(keyFilePath) {
		logOutput("✗ Key file not found: %s\n", keyFilePath)
		dryRunSuccess = false
	} else {
		checksum, err := utils.CalculateFileChecksum(keyFilePath)
		if err != nil {
			logOutput("✗ Failed to calculate key file checksum: %v\n", err)
			dryRunSuccess = false
		} else if checksum != patch.FromKeyFile.Checksum {
			logOutput("✗ Key file hash mismatch\n")
			logOutput("  Expected: %s\n", patch.FromKeyFile.Checksum[:16]+"...")
			logOutput("  Got:      %s\n", checksum[:16]+"...")
			dryRunSuccess = false
		} else {
			logOutput("✓ Key file verified\n")
		}
	}

	// Verify required files
	if dryRunSuccess {
		logOutput("\nVerifying %d required files...\n", len(patch.RequiredFiles))
		mismatches := 0
		for _, req := range patch.RequiredFiles {
			filePath := targetDir + string(os.PathSeparator) + req.Path
			if !utils.FileExists(filePath) {
				logOutput("✗ Required file missing: %s\n", req.Path)
				mismatches++
				dryRunSuccess = false
				continue
			}

			checksum, err := utils.CalculateFileChecksum(filePath)
			if err != nil {
				logOutput("✗ Failed to verify: %s\n", req.Path)
				mismatches++
				dryRunSuccess = false
				continue
			}

			if checksum != req.Checksum {
				logOutput("✗ Hash mismatch: %s\n", req.Path)
				mismatches++
				dryRunSuccess = false
			}
		}

		if mismatches == 0 {
			logOutput("✓ All required files verified\n")
		}
	}

	if !dryRunSuccess {
		logOutput("\n✗ Dry run validation failed - patch cannot be applied\n")
		logOutput("\n========================================\n")
		logOutput("Status: FAILED\n")
		logOutput("Completed: %s\n", time.Now().UTC().Format("2006-01-02 15:04:05 UTC"))
		logOutput("========================================\n")
		if logFile != nil {
			logOutput("\nLog saved to: %s\n", logFileName)
		}
		os.Exit(1)
	}

	logOutput("\n✓ Dry run completed successfully\n")
	logOutput("\n")

	// Step 2: Apply patch
	logOutput("==============================================\n")
	logOutput("Step 2: Applying Patch\n")
	logOutput("==============================================\n")
	logOutput("\n")
	logOutput("Applying patch with backup enabled...\n")
	logOutput("\n")

	applier := patcher.NewApplier()
	if err := applier.ApplyPatchWithPath(patch, targetDir, "", true, true, true); err != nil {
		logOutput("\nError: Patch application failed: %v\n", err)
		logOutput("\nNote: Automatic rollback may have been performed to restore original files.\n")
		logOutput("\n========================================\n")
		logOutput("Status: FAILED\n")
		logOutput("Completed: %s\n", time.Now().UTC().Format("2006-01-02 15:04:05 UTC"))
		logOutput("========================================\n")
		if logFile != nil {
			logOutput("\nLog saved to: %s\n", logFileName)
		}
		os.Exit(1)
	}

	// Success
	logOutput("\n")
	logOutput("==============================================\n")
	logOutput("          Patch Applied Successfully\n")
	logOutput("==============================================\n")
	logOutput("\n")
	logOutput("Version updated: %s → %s\n", patch.FromVersion, patch.ToVersion)
	logOutput("\n========================================\n")
	logOutput("Status: SUCCESS\n")
	logOutput("Completed: %s\n", time.Now().UTC().Format("2006-01-02 15:04:05 UTC"))
	logOutput("========================================\n")
	if logFile != nil {
		logOutput("\nLog saved to: %s\n", logFileName)
	}

	os.Exit(0)
}

// runInteractiveMode runs the interactive console interface for embedded patches
func runInteractiveMode(patch *utils.Patch, defaultTargetDir string, ignore1GB bool) {
	reader := bufio.NewReader(os.Stdin)
	customKeyFile := "" // Track custom key file path

	// Check if patch creator enabled simple mode for end users
	if patch.SimpleMode {
		runSimpleMode(patch, defaultTargetDir)
		return
	}

	// Display patch information
	fmt.Println()
	displayPatchInfo(patch)

	// Ask for target directory
	fmt.Printf("\nTarget directory [%s]: ", defaultTargetDir)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	targetDir := defaultTargetDir
	if input != "" {
		targetDir = input
	}

	// Check if directory exists
	if !utils.FileExists(targetDir) {
		fmt.Printf("Error: Directory not found: %s\n", targetDir)
		fmt.Println("\nPress Enter to exit...")
		reader.ReadString('\n')
		os.Exit(1)
	}

	// Show menu
	for {
		fmt.Println("\n==============================================")
		fmt.Println("Options:")
		fmt.Println("  1. Dry Run (simulate without changes)")
		fmt.Println("  2. Apply Patch")
		fmt.Println("  3. Toggle 1GB Bypass Mode (currently: " + formatBoolState(ignore1GB) + ")")
		fmt.Println("  4. Change Target Directory")
		fmt.Println("  5. Specify Custom Key File")
		if customKeyFile != "" {
			fmt.Printf("     (Currently: %s)\n", customKeyFile)
		} else {
			fmt.Printf("     (Currently: %s - default)\n", patch.FromKeyFile.Path)
		}
		fmt.Println("  6. Exit")
		fmt.Println("==============================================")
		fmt.Print("Select option [1-6]: ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		switch input {
		case "1":
			// Dry run
			fmt.Println("\n=== DRY RUN MODE ===")
			fmt.Println("Simulating patch application...")
			performDryRun(patch, targetDir, customKeyFile)
			fmt.Println("\nPress Enter to continue...")
			reader.ReadString('\n')

		case "2":
			// Apply patch
			fmt.Println("\n=== APPLYING PATCH ===")
			fmt.Printf("Target: %s\n", targetDir)

			// Override key file path if custom one is provided
			if customKeyFile != "" {
				fmt.Printf("Using custom key file: %s\n", customKeyFile)
				patch.FromKeyFile.Path = customKeyFile
			}

			fmt.Print("Are you sure you want to apply this patch? (yes/no): ")
			confirm, _ := reader.ReadString('\n')
			confirm = strings.TrimSpace(strings.ToLower(confirm))

			if confirm == "yes" || confirm == "y" {
				fmt.Println("\nApplying patch...")
				applier := patcher.NewApplier()
				if err := applier.ApplyPatchWithPath(patch, targetDir, "", true, true, true); err != nil {
					fmt.Printf("\nError: Patch application failed: %v\n", err)
					fmt.Println("\nNote: Automatic rollback may have been performed to restore original files.")
					fmt.Println("\nPress Enter to exit...")
					reader.ReadString('\n')
					os.Exit(1)
				}

				fmt.Println("\n=== SUCCESS ===")
				fmt.Printf("Patch applied successfully!\n")
				fmt.Printf("Version updated from %s to %s\n", patch.FromVersion, patch.ToVersion)
				fmt.Println("\nPress Enter to exit...")
				reader.ReadString('\n')
				return
			} else {
				fmt.Println("Patch application cancelled")
			}

		case "3":
			// Toggle 1GB bypass
			ignore1GB = !ignore1GB
			fmt.Printf("\n1GB Bypass Mode: %s\n", formatBoolState(ignore1GB))
			if ignore1GB {
				fmt.Println("Warning: Large patches may consume significant memory!")
			}

		case "4":
			// Change target directory
			fmt.Print("\nEnter new target directory: ")
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)
			if input != "" {
				if !utils.FileExists(input) {
					fmt.Printf("Error: Directory not found: %s\n", input)
				} else {
					targetDir = input
					fmt.Printf("Target directory changed to: %s\n", targetDir)
				}
			}

		case "5":
			// Specify custom key file
			fmt.Print("\nEnter custom key file path (or press Enter to use default): ")
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)
			if input == "" {
				customKeyFile = ""
				fmt.Printf("Using default key file: %s\n", patch.FromKeyFile.Path)
			} else {
				// Validate the custom key file exists
				testPath := input
				if !strings.Contains(input, string(os.PathSeparator)) && !strings.Contains(input, "/") {
					// Relative path - check in target directory
					testPath = targetDir + string(os.PathSeparator) + input
				}
				if !utils.FileExists(testPath) {
					fmt.Printf("Warning: File not found: %s\n", testPath)
					fmt.Println("You can still proceed - verification will happen during patch application.")
				}
				customKeyFile = input
				fmt.Printf("Custom key file set to: %s\n", customKeyFile)
			}

		case "6":
			// Exit
			fmt.Println("\nExiting...")
			return

		default:
			fmt.Println("Invalid option. Please select 1-6.")
		}
	}
}

// formatBoolState formats a boolean as "Enabled" or "Disabled"
func formatBoolState(enabled bool) string {
	if enabled {
		return "Enabled"
	}
	return "Disabled"
}

func printHelp() {
	fmt.Printf("CyberPatchMaker - Patch Applier v%s\n", version.GetVersion())
	fmt.Println("\nUsage:")
	fmt.Println("  patch-apply --patch <file> --current-dir <directory>")
	fmt.Println("\nOptions:")
	fmt.Println("  --patch         Path to patch file (required)")
	fmt.Println("  --current-dir   Directory containing current version (required)")
	fmt.Println("  --key-file      Custom key file path (if renamed or moved)")
	fmt.Println("  --dry-run       Simulate patch without making changes")
	fmt.Println("  --verify        Verify file hashes before and after patching (default: true)")
	fmt.Println("  --backup        Create backup before patching (default: true)")
	fmt.Println("  --ignore1gb     Bypass 1GB patch size limit (use with caution)")
	fmt.Println("  --silent        Silent mode: apply patch automatically without prompts")
	fmt.Println("  --version       Show version information")
	fmt.Println("  --help          Show this help message")
	fmt.Println("\nSelf-Contained Executable Mode:")
	fmt.Println("  When run as a self-contained executable, an interactive console")
	fmt.Println("  interface will guide you through the patch application process.")
	fmt.Println("  Use --silent flag for automated patching without user interaction.")
	fmt.Println("\nExamples:")
	fmt.Println("  # Apply patch")
	fmt.Println("  patch-apply --patch 1.0.0-to-1.0.3.patch --current-dir C:\\MyApp")
	fmt.Println("\n  # Dry run (simulate only)")
	fmt.Println("  patch-apply --patch 1.0.0-to-1.0.3.patch --current-dir C:\\MyApp --dry-run")
	fmt.Println("\n  # Run self-contained executable with 1GB bypass")
	fmt.Println("  1.0.0-to-1.0.1.exe --ignore1gb")
	fmt.Println("\n  # Run self-contained executable in silent mode (automation)")
	fmt.Println("  1.2.4-to-1.2.5.exe --silent")
	fmt.Println("\n  # Silent mode with custom target directory")
	fmt.Println("  1.2.4-to-1.2.5.exe --silent --current-dir C:\\MyApp")
}
