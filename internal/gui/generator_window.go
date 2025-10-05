package gui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/cyberofficial/cyberpatchmaker/internal/core/manifest"
	"github.com/cyberofficial/cyberpatchmaker/internal/core/patcher"
	"github.com/cyberofficial/cyberpatchmaker/internal/core/scanner"
	"github.com/cyberofficial/cyberpatchmaker/internal/core/version"
	"github.com/cyberofficial/cyberpatchmaker/pkg/utils"
)

// GeneratorWindow represents the patch generator UI
type GeneratorWindow struct {
	widget.BaseWidget

	window fyne.Window

	// UI Components
	versionsDir string
	fromVersion string
	toVersion   string
	outputDir   string
	compression string
	keyFile     string

	versionsDirEntry  *widget.Entry
	keyFileEntry      *widget.Entry
	fromVersionSelect *widget.Select
	toVersionSelect   *widget.Select
	outputDirEntry    *widget.Entry
	compressionRadio  *widget.RadioGroup
	generateBtn       *widget.Button
	statusLabel       *widget.Label
	logText           *widget.Entry

	// Data
	availableVersions []string
	manifestMgr       *manifest.Manager
}

// NewGeneratorWindow creates a new generator window
func NewGeneratorWindow() *GeneratorWindow {
	gw := &GeneratorWindow{
		compression: "zstd",
		keyFile:     "program.exe",
	}
	gw.ExtendBaseWidget(gw)
	return gw
}

// CreateRenderer creates the renderer for the generator window
func (gw *GeneratorWindow) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(gw.buildUI())
}

// SetWindow sets the parent window (needed for dialogs)
func (gw *GeneratorWindow) SetWindow(window fyne.Window) {
	gw.window = window
}

// buildUI builds the complete UI layout
func (gw *GeneratorWindow) buildUI() fyne.CanvasObject {
	// Create versions directory selector
	gw.versionsDirEntry = widget.NewEntry()
	gw.versionsDirEntry.SetPlaceHolder("Select versions directory...")

	versionsDirBrowse := widget.NewButton("Browse", func() {
		gw.selectVersionsDirectory()
	})

	versionsDirContainer := container.NewBorder(
		nil, nil,
		widget.NewLabel("Versions Directory:"),
		versionsDirBrowse,
		gw.versionsDirEntry,
	)

	// Create key file selector
	gw.keyFileEntry = widget.NewEntry()
	gw.keyFileEntry.SetText(gw.keyFile)
	gw.keyFileEntry.OnChanged = func(text string) {
		gw.keyFile = text
	}

	keyFileContainer := container.NewBorder(
		nil, nil,
		widget.NewLabel("Key File:"),
		nil,
		gw.keyFileEntry,
	)

	// Create from version selector
	gw.fromVersionSelect = widget.NewSelect([]string{}, func(selected string) {
		gw.fromVersion = selected
		gw.updateGenerateButton()
	})
	gw.fromVersionSelect.PlaceHolder = "Select source version..."

	fromVersionContainer := container.NewBorder(
		nil, nil,
		widget.NewLabel("From Version:"),
		nil,
		gw.fromVersionSelect,
	)

	// Create to version selector
	gw.toVersionSelect = widget.NewSelect([]string{}, func(selected string) {
		gw.toVersion = selected
		gw.updateGenerateButton()
	})
	gw.toVersionSelect.PlaceHolder = "Select target version..."

	toVersionContainer := container.NewBorder(
		nil, nil,
		widget.NewLabel("To Version:"),
		nil,
		gw.toVersionSelect,
	)

	// Create output directory selector
	gw.outputDirEntry = widget.NewEntry()
	gw.outputDirEntry.SetPlaceHolder("Select output directory...")

	outputDirBrowse := widget.NewButton("Browse", func() {
		gw.selectOutputDirectory()
	})

	outputDirContainer := container.NewBorder(
		nil, nil,
		widget.NewLabel("Output Directory:"),
		outputDirBrowse,
		gw.outputDirEntry,
	)

	// Create compression selector
	gw.compressionRadio = widget.NewRadioGroup([]string{"zstd", "gzip", "none"}, func(selected string) {
		gw.compression = selected
	})
	gw.compressionRadio.Horizontal = true
	gw.compressionRadio.SetSelected("zstd")

	compressionContainer := container.NewVBox(
		widget.NewLabel("Compression:"),
		gw.compressionRadio,
	)

	// Create generate button
	gw.generateBtn = widget.NewButton("Generate Patch", func() {
		gw.generatePatch()
	})
	gw.generateBtn.Disable()

	// Create status label
	gw.statusLabel = widget.NewLabel("Ready")

	// Create log output
	gw.logText = widget.NewMultiLineEntry()
	gw.logText.SetPlaceHolder("Log output will appear here...")
	gw.logText.Disable()

	logContainer := container.NewVScroll(gw.logText)
	logContainer.SetMinSize(fyne.NewSize(0, 200))

	// Assemble the UI
	return container.NewVBox(
		versionsDirContainer,
		keyFileContainer,
		widget.NewSeparator(),
		fromVersionContainer,
		toVersionContainer,
		widget.NewSeparator(),
		outputDirContainer,
		compressionContainer,
		widget.NewSeparator(),
		gw.generateBtn,
		widget.NewSeparator(),
		widget.NewLabel("Status:"),
		gw.statusLabel,
		widget.NewLabel("Log:"),
		logContainer,
	)
}

// selectVersionsDirectory opens a folder dialog for selecting versions directory
func (gw *GeneratorWindow) selectVersionsDirectory() {
	if gw.window == nil {
		return
	}

	dialog.ShowFolderOpen(func(dir fyne.ListableURI, err error) {
		if err == nil && dir != nil {
			path := dir.Path()
			gw.versionsDirEntry.SetText(path)
			gw.versionsDir = path
			gw.scanVersions()
		}
	}, gw.window)
}

// selectOutputDirectory opens a folder dialog for selecting output directory
func (gw *GeneratorWindow) selectOutputDirectory() {
	if gw.window == nil {
		return
	}

	dialog.ShowFolderOpen(func(dir fyne.ListableURI, err error) {
		if err == nil && dir != nil {
			path := dir.Path()
			gw.outputDirEntry.SetText(path)
			gw.outputDir = path
			gw.updateGenerateButton()
		}
	}, gw.window)
}

// scanVersions scans the versions directory for available versions
func (gw *GeneratorWindow) scanVersions() {
	gw.appendLog("Scanning versions directory: " + gw.versionsDir)

	// Check if directory exists
	if _, err := os.Stat(gw.versionsDir); os.IsNotExist(err) {
		gw.setStatus("Error: Versions directory does not exist")
		gw.appendLog("ERROR: Directory does not exist")
		return
	}

	// Read directory contents
	entries, err := os.ReadDir(gw.versionsDir)
	if err != nil {
		gw.setStatus("Error: Could not read versions directory")
		gw.appendLog("ERROR: " + err.Error())
		return
	}

	// Find subdirectories that might be versions
	versions := []string{}
	for _, entry := range entries {
		if entry.IsDir() {
			versions = append(versions, entry.Name())
		}
	}

	if len(versions) == 0 {
		gw.setStatus("No versions found in directory")
		gw.appendLog("No version directories found")
		return
	}

	gw.availableVersions = versions
	gw.fromVersionSelect.Options = versions
	gw.toVersionSelect.Options = versions
	gw.fromVersionSelect.Refresh()
	gw.toVersionSelect.Refresh()

	gw.setStatus(fmt.Sprintf("Found %d versions", len(versions)))
	gw.appendLog(fmt.Sprintf("Found versions: %s", strings.Join(versions, ", ")))
}

// updateGenerateButton enables/disables generate button based on selections
func (gw *GeneratorWindow) updateGenerateButton() {
	if gw.fromVersion != "" && gw.toVersion != "" && gw.outputDir != "" {
		gw.generateBtn.Enable()
	} else {
		gw.generateBtn.Disable()
	}
}

// generatePatch generates the patch file
func (gw *GeneratorWindow) generatePatch() {
	gw.setStatus("Generating patch...")
	gw.generateBtn.Disable()

	// Validate selections
	if gw.fromVersion == gw.toVersion {
		gw.setStatus("Error: From and To versions must be different")
		gw.appendLog("ERROR: Cannot generate patch from same version to same version")
		gw.generateBtn.Enable()
		return
	}

	// Build paths
	fromPath := filepath.Join(gw.versionsDir, gw.fromVersion)
	toPath := filepath.Join(gw.versionsDir, gw.toVersion)
	outputPath := filepath.Join(gw.outputDir, fmt.Sprintf("%s-to-%s.patch", gw.fromVersion, gw.toVersion))

	gw.appendLog(fmt.Sprintf("From: %s", fromPath))
	gw.appendLog(fmt.Sprintf("To: %s", toPath))
	gw.appendLog(fmt.Sprintf("Output: %s", outputPath))
	gw.appendLog(fmt.Sprintf("Compression: %s", gw.compression))

	// Create version manager and manifest manager
	versionMgr := version.NewManager()
	gw.manifestMgr = manifest.NewManager()

	// Register and scan source version
	gw.appendLog("Scanning source version...")
	gw.appendLog(fmt.Sprintf("Using key file: %s", gw.keyFile))
	fromVer, err := versionMgr.RegisterVersion(gw.fromVersion, fromPath, gw.keyFile)
	if err != nil {
		gw.setStatus("Error: Failed to register source version")
		gw.appendLog("ERROR: " + err.Error())
		gw.generateBtn.Enable()
		return
	}

	// Scan source directory
	fromScanner := scanner.NewScanner(fromPath)
	fromFiles, fromDirs, err := fromScanner.ScanDirectory()
	if err != nil {
		gw.setStatus("Error: Failed to scan source version")
		gw.appendLog("ERROR: " + err.Error())
		gw.generateBtn.Enable()
		return
	}

	// Create manifest for source version
	if fromVer.Manifest == nil {
		fromVer.Manifest = &utils.Manifest{}
	}
	fromVer.Manifest.Files = fromFiles
	fromVer.Manifest.Directories = fromDirs
	fromVer.Manifest.Version = gw.fromVersion
	gw.appendLog(fmt.Sprintf("Source version: %d files, %d directories", len(fromFiles), len(fromDirs)))

	// Register and scan target version
	gw.appendLog("Scanning target version...")
	toVer, err := versionMgr.RegisterVersion(gw.toVersion, toPath, gw.keyFile)
	if err != nil {
		gw.setStatus("Error: Failed to register target version")
		gw.appendLog("ERROR: " + err.Error())
		gw.generateBtn.Enable()
		return
	}

	// Scan target directory
	toScanner := scanner.NewScanner(toPath)
	toFiles, toDirs, err := toScanner.ScanDirectory()
	if err != nil {
		gw.setStatus("Error: Failed to scan target version")
		gw.appendLog("ERROR: " + err.Error())
		gw.generateBtn.Enable()
		return
	}

	// Create manifest for target version
	if toVer.Manifest == nil {
		toVer.Manifest = &utils.Manifest{}
	}
	toVer.Manifest.Files = toFiles
	toVer.Manifest.Directories = toDirs
	toVer.Manifest.Version = gw.toVersion
	gw.appendLog(fmt.Sprintf("Target version: %d files, %d directories", len(toFiles), len(toDirs))) // Generate patch
	gw.appendLog("Generating patch operations...")
	compressionStr := "zstd"
	switch gw.compression {
	case "gzip":
		compressionStr = "gzip"
	case "none":
		compressionStr = "none"
	}

	options := &utils.PatchOptions{
		Compression:      compressionStr,
		CompressionLevel: 3,
		VerifyAfter:      false,
		DiffThresholdKB:  1,
		SkipIdentical:    true,
	}

	generator := patcher.NewGenerator()
	patch, err := generator.GeneratePatch(fromVer, toVer, options)
	if err != nil {
		gw.setStatus("Error: Failed to generate patch")
		gw.appendLog("ERROR: " + err.Error())
		gw.generateBtn.Enable()
		return
	}

	// Validate patch
	if err := generator.ValidatePatch(patch); err != nil {
		gw.setStatus("Error: Patch validation failed")
		gw.appendLog("ERROR: " + err.Error())
		gw.generateBtn.Enable()
		return
	}

	// Save patch to file
	gw.appendLog("Saving patch file...")
	if err := gw.savePatch(patch, outputPath, options); err != nil {
		gw.setStatus("Error: Failed to save patch")
		gw.appendLog("ERROR: " + err.Error())
		gw.generateBtn.Enable()
		return
	}

	// Get patch file size
	info, err := os.Stat(outputPath)
	if err == nil {
		sizeKB := float64(info.Size()) / 1024.0
		sizeMB := sizeKB / 1024.0
		if sizeMB >= 1.0 {
			gw.appendLog(fmt.Sprintf("Patch size: %.2f MB", sizeMB))
		} else {
			gw.appendLog(fmt.Sprintf("Patch size: %.2f KB", sizeKB))
		}
	}

	gw.setStatus("Patch generated successfully!")
	gw.appendLog("SUCCESS: Patch generated successfully")
	gw.generateBtn.Enable()

	// Show success dialog
	if gw.window != nil {
		dialog.ShowInformation("Success",
			fmt.Sprintf("Patch generated successfully!\n\nOutput: %s", outputPath),
			gw.window)
	}
}

// setStatus updates the status label
func (gw *GeneratorWindow) setStatus(status string) {
	gw.statusLabel.SetText(status)
}

// appendLog appends a message to the log
func (gw *GeneratorWindow) appendLog(message string) {
	current := gw.logText.Text
	if current != "" {
		current += "\n"
	}
	current += message
	gw.logText.SetText(current)

	// Auto-scroll to bottom (approximate by setting text again)
	gw.logText.Refresh()
}

// savePatch saves the patch to a file with compression
func (gw *GeneratorWindow) savePatch(patch *utils.Patch, filename string, options *utils.PatchOptions) error {
	// Marshal patch to JSON
	data, err := json.MarshalIndent(patch, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal patch: %w", err)
	}

	// Compress if needed
	if options.Compression != "none" && options.Compression != "" {
		compressedData, err := utils.CompressData(data, options.Compression, options.CompressionLevel)
		if err != nil {
			return fmt.Errorf("failed to compress patch: %w", err)
		}
		data = compressedData
	}

	// Calculate checksum
	checksum := utils.CalculateDataChecksum(data)
	patch.Header.Checksum = checksum

	// Write to file
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write patch file: %w", err)
	}

	return nil
}
