package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/cyberofficial/cyberpatchmaker/internal/gui"
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
	Reserved    [44]byte
}

func main() {
	// Check if patch data is embedded in this executable
	patch, targetDir, isEmbedded := checkEmbeddedPatch()

	// Create Fyne application with unique ID
	myApp := app.NewWithID("com.cyberofficial.cyberpatchmaker.applier")

	if isEmbedded && patch != nil {
		// Show embedded patch applier
		myWindow := myApp.NewWindow(fmt.Sprintf("Apply Patch: %s → %s", patch.FromVersion, patch.ToVersion))
		myWindow.Resize(fyne.NewSize(700, 600))

		applierUI := gui.NewApplierWindow()
		applierUI.SetWindow(myWindow)

		content := container.NewVBox(
			widget.NewLabel(fmt.Sprintf("Self-Contained Patch: %s → %s", patch.FromVersion, patch.ToVersion)),
			widget.NewSeparator(),
			applierUI,
		)

		myWindow.SetContent(content)

		// Load embedded patch AFTER UI is created
		applierUI.LoadEmbeddedPatch(patch, targetDir)

		myWindow.ShowAndRun()
	} else {
		// Normal mode - show file browser UI
		myWindow := myApp.NewWindow("CyberPatchMaker - Patch Applier")
		myWindow.Resize(fyne.NewSize(800, 600))

		applierUI := gui.NewApplierWindow()
		applierUI.SetWindow(myWindow)

		content := container.NewVBox(
			widget.NewLabel("CyberPatchMaker - Patch Applier"),
			widget.NewSeparator(),
			applierUI,
		)

		myWindow.SetContent(content)
		myWindow.ShowAndRun()
	}
}

func checkEmbeddedPatch() (*utils.Patch, string, bool) {
	// Get path to this executable
	exePath, err := os.Executable()
	if err != nil {
		return nil, "", false
	}

	// Open executable for reading
	file, err := os.Open(exePath)
	if err != nil {
		return nil, "", false
	}
	defer file.Close()

	// Get file size
	stat, err := file.Stat()
	if err != nil {
		return nil, "", false
	}
	fileSize := stat.Size()

	// Check if file is large enough for header
	if fileSize < HEADER_SIZE {
		return nil, "", false
	}

	// Read header from end of file
	headerOffset := fileSize - HEADER_SIZE
	if _, err := file.Seek(headerOffset, io.SeekStart); err != nil {
		return nil, "", false
	}

	headerBytes := make([]byte, HEADER_SIZE)
	if _, err := io.ReadFull(file, headerBytes); err != nil {
		return nil, "", false
	}

	// Parse header
	var header EmbeddedPatchHeader
	buf := bytes.NewReader(headerBytes)
	if err := binary.Read(buf, binary.LittleEndian, &header); err != nil {
		return nil, "", false
	}

	// Validate magic bytes
	magic := string(bytes.TrimRight(header.Magic[:], "\x00"))
	if magic != MAGIC_BYTES {
		return nil, "", false
	}

	// Read patch data
	if _, err := file.Seek(int64(header.DataOffset), io.SeekStart); err != nil {
		return nil, "", false
	}

	patchData := make([]byte, header.DataSize)
	if _, err := io.ReadFull(file, patchData); err != nil {
		return nil, "", false
	}

	// Verify checksum
	actualChecksum := sha256.Sum256(patchData)
	if !bytes.Equal(actualChecksum[:], header.Checksum[:]) {
		return nil, "", false
	}

	// Decompress if needed
	compression := string(bytes.TrimRight(header.Compression[:], "\x00"))
	if compression != "none" && compression != "" {
		decompressed, err := utils.DecompressData(patchData, compression)
		if err != nil {
			return nil, "", false
		}
		patchData = decompressed
	}

	// Parse JSON patch
	var patch utils.Patch
	if err := json.Unmarshal(patchData, &patch); err != nil {
		return nil, "", false
	}

	// Get current directory as default target
	targetDir, _ := os.Getwd()

	return &patch, targetDir, true
}
