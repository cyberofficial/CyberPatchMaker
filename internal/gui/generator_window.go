package gui

import (
	"encoding/json"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
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
	versionsDir      string
	fromVersion      string
	toVersion        string
	outputDir        string
	compression      string
	compressionLevel int
	verifyAfter      bool
	diffThresholdKB  int
	skipIdentical    bool
	batchMode        bool
	fromKeyFile      string
	toKeyFile        string

	versionsDirEntry   *widget.Entry
	fromKeyFileSelect  *widget.Select
	toKeyFileSelect    *widget.Select
	fromVersionSelect  *widget.Select
	toVersionSelect    *widget.Select
	outputDirEntry     *widget.Entry
	compressionRadio   *widget.RadioGroup
	compressionSlider  *widget.Slider
	compressionLabel   *widget.Label
	verifyCheck        *widget.Check
	diffThresholdEntry *widget.Entry
	skipIdenticalCheck *widget.Check
	batchModeCheck     *widget.Check
	generateBtn        *widget.Button
	statusLabel        *widget.Label
	logText            *widget.Entry

	// Data
	availableVersions []string
	manifestMgr       *manifest.Manager
}

// NewGeneratorWindow creates a new generator window
func NewGeneratorWindow() *GeneratorWindow {
	gw := &GeneratorWindow{
		compression:      "zstd",
		compressionLevel: 3,
		verifyAfter:      true,
		diffThresholdKB:  1,
		skipIdentical:    true,
		batchMode:        false,
		fromKeyFile:      "program.exe",
		toKeyFile:        "program.exe",
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
	gw.versionsDirEntry.OnSubmitted = func(text string) {
		if text != "" {
			gw.versionsDir = text
			gw.scanVersions()
		}
	}

	versionsDirBrowse := widget.NewButton("Browse", func() {
		gw.selectVersionsDirectory()
	})

	versionsDirContainer := container.NewBorder(
		nil, nil,
		widget.NewLabel("Versions Directory:"),
		versionsDirBrowse,
		gw.versionsDirEntry,
	)

	// Create batch mode checkbox
	gw.batchModeCheck = widget.NewCheck("Batch Mode: Generate patches from ALL versions to target", func(checked bool) {
		gw.batchMode = checked
		if checked {
			// In batch mode, from version is not used
			gw.fromVersionSelect.Disable()
			gw.fromKeyFileSelect.Disable()
		} else {
			gw.fromVersionSelect.Enable()
			gw.fromKeyFileSelect.Enable()
		}
		gw.updateGenerateButton()
	})

	// Create from version selector
	gw.fromVersionSelect = widget.NewSelect([]string{}, func(selected string) {
		gw.fromVersion = selected
		gw.updateFromKeyFileOptions()
		gw.updateGenerateButton()
	})
	gw.fromVersionSelect.PlaceHolder = "Select source version..."

	// Create from key file selector
	gw.fromKeyFileSelect = widget.NewSelect([]string{}, func(selected string) {
		gw.fromKeyFile = selected
	})
	gw.fromKeyFileSelect.PlaceHolder = "Select key file..."
	gw.fromKeyFileSelect.SetSelected(gw.fromKeyFile)

	// Create to version selector
	gw.toVersionSelect = widget.NewSelect([]string{}, func(selected string) {
		gw.toVersion = selected
		gw.updateToKeyFileOptions()
		gw.updateGenerateButton()
	})
	gw.toVersionSelect.PlaceHolder = "Select target version..."

	// Create to key file selector
	gw.toKeyFileSelect = widget.NewSelect([]string{}, func(selected string) {
		gw.toKeyFile = selected
	})
	gw.toKeyFileSelect.PlaceHolder = "Select key file..."
	gw.toKeyFileSelect.SetSelected(gw.toKeyFile)

	// Left column: Version selection
	leftColumn := container.NewVBox(
		widget.NewLabel("Version Selection:"),
		container.NewBorder(nil, nil, widget.NewLabel("From:"), nil, gw.fromVersionSelect),
		container.NewBorder(nil, nil, widget.NewLabel("Key:"), nil, gw.fromKeyFileSelect),
		widget.NewSeparator(),
		container.NewBorder(nil, nil, widget.NewLabel("To:"), nil, gw.toVersionSelect),
		container.NewBorder(nil, nil, widget.NewLabel("Key:"), nil, gw.toKeyFileSelect),
	)

	// Create output directory selector
	gw.outputDirEntry = widget.NewEntry()
	gw.outputDirEntry.SetPlaceHolder("Select output directory...")
	gw.outputDirEntry.OnSubmitted = func(text string) {
		if text != "" {
			gw.outputDir = text
			gw.updateGenerateButton()
		}
	}

	outputDirBrowse := widget.NewButton("Browse", func() {
		gw.selectOutputDirectory()
	})

	outputDirContainer := container.NewBorder(
		nil, nil,
		widget.NewLabel("Output:"),
		outputDirBrowse,
		gw.outputDirEntry,
	)

	// Create compression level slider FIRST (before radio group that references it)
	gw.compressionSlider = widget.NewSlider(1, 4)
	gw.compressionSlider.Value = 3
	gw.compressionSlider.Step = 1
	gw.compressionSlider.OnChanged = func(value float64) {
		gw.compressionLevel = int(value)
		gw.updateCompressionLabel()
	}

	gw.compressionLabel = widget.NewLabel("Level: 3")

	// Create compression selector (references slider, so create it after)
	gw.compressionRadio = widget.NewRadioGroup([]string{"zstd", "gzip", "none"}, func(selected string) {
		gw.compression = selected
		// Update slider range based on compression type
		if selected == "zstd" {
			gw.compressionSlider.Max = 4
			if gw.compressionLevel > 4 {
				gw.compressionLevel = 3
				gw.compressionSlider.Value = 3
			}
		} else if selected == "gzip" {
			gw.compressionSlider.Max = 9
		} else {
			gw.compressionSlider.Disable()
			return
		}
		gw.compressionSlider.Enable()
		gw.compressionSlider.Refresh()
		gw.updateCompressionLabel()
	})
	gw.compressionRadio.Horizontal = true
	gw.compressionRadio.SetSelected("zstd")

	compressionContainer := container.NewVBox(
		gw.compressionRadio,
		container.NewBorder(nil, nil, widget.NewLabel("Level:"), gw.compressionLabel, gw.compressionSlider),
	)

	// Create advanced options
	gw.verifyCheck = widget.NewCheck("Verify after creation", func(checked bool) {
		gw.verifyAfter = checked
	})
	gw.verifyCheck.SetChecked(true)

	gw.skipIdenticalCheck = widget.NewCheck("Skip identical files", func(checked bool) {
		gw.skipIdentical = checked
	})
	gw.skipIdenticalCheck.SetChecked(true)

	gw.diffThresholdEntry = widget.NewEntry()
	gw.diffThresholdEntry.SetText("1")
	gw.diffThresholdEntry.OnChanged = func(text string) {
		var threshold int
		fmt.Sscanf(text, "%d", &threshold)
		if threshold > 0 {
			gw.diffThresholdKB = threshold
		}
	}

	// Right column: Options
	rightColumn := container.NewVBox(
		widget.NewLabel("Compression:"),
		compressionContainer,
		widget.NewSeparator(),
		widget.NewLabel("Options:"),
		gw.verifyCheck,
		gw.skipIdenticalCheck,
		container.NewBorder(nil, nil, widget.NewLabel("Diff Threshold (KB):"), nil, gw.diffThresholdEntry),
	)

	// Create two-column layout for version/options
	twoColumnLayout := container.NewGridWithColumns(2,
		leftColumn,
		rightColumn,
	)

	// Create generate button
	gw.generateBtn = widget.NewButton("Generate Patch", func() {
		gw.generatePatch()
	})
	gw.generateBtn.Disable()

	// Create status label
	gw.statusLabel = widget.NewLabel("Ready")

	// Create log output with white background and black text (read-only but not disabled)
	gw.logText = widget.NewMultiLineEntry()
	gw.logText.SetPlaceHolder("Log output will appear here...")
	// Make read-only by preventing edits (but keep enabled for normal text color)
	gw.logText.OnChanged = func(text string) {
		// This prevents user edits - text can only be set programmatically
	}

	// Create a white background for the log for maximum contrast
	logBg := canvas.NewRectangle(color.White)
	logWithBg := container.NewStack(logBg, gw.logText)
	logContainer := container.NewVScroll(logWithBg)
	logContainer.SetMinSize(fyne.NewSize(0, 150))

	// Assemble the UI with compact layout
	return container.NewVBox(
		versionsDirContainer,
		gw.batchModeCheck,
		widget.NewSeparator(),
		twoColumnLayout,
		widget.NewSeparator(),
		outputDirContainer,
		container.NewBorder(nil, nil, nil, gw.generateBtn, gw.statusLabel),
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
	if gw.batchMode {
		// In batch mode, only need to version and output dir
		if gw.toVersion != "" && gw.outputDir != "" {
			gw.generateBtn.Enable()
		} else {
			gw.generateBtn.Disable()
		}
	} else {
		// In normal mode, need both from and to versions
		if gw.fromVersion != "" && gw.toVersion != "" && gw.outputDir != "" {
			gw.generateBtn.Enable()
		} else {
			gw.generateBtn.Disable()
		}
	}
}

// updateCompressionLabel updates the compression level label
func (gw *GeneratorWindow) updateCompressionLabel() {
	gw.compressionLabel.SetText(fmt.Sprintf("Level: %d", gw.compressionLevel))
}

// updateFromKeyFileOptions scans from version and populates key file options
func (gw *GeneratorWindow) updateFromKeyFileOptions() {
	if gw.fromVersion == "" || gw.versionsDir == "" {
		return
	}

	versionPath := filepath.Join(gw.versionsDir, gw.fromVersion)
	files := gw.getFilesInDirectory(versionPath)
	gw.fromKeyFileSelect.Options = files
	gw.fromKeyFileSelect.Refresh()

	// Auto-select if only one executable found
	if len(files) > 0 {
		for _, file := range files {
			if strings.HasSuffix(strings.ToLower(file), ".exe") {
				gw.fromKeyFileSelect.SetSelected(file)
				gw.fromKeyFile = file
				break
			}
		}
	}
}

// updateToKeyFileOptions scans to version and populates key file options
func (gw *GeneratorWindow) updateToKeyFileOptions() {
	if gw.toVersion == "" || gw.versionsDir == "" {
		return
	}

	versionPath := filepath.Join(gw.versionsDir, gw.toVersion)
	files := gw.getFilesInDirectory(versionPath)
	gw.toKeyFileSelect.Options = files
	gw.toKeyFileSelect.Refresh()

	// Auto-select if only one executable found
	if len(files) > 0 {
		for _, file := range files {
			if strings.HasSuffix(strings.ToLower(file), ".exe") {
				gw.toKeyFileSelect.SetSelected(file)
				gw.toKeyFile = file
				break
			}
		}
	}
}

// getFilesInDirectory returns a list of all files in the directory (not recursive)
func (gw *GeneratorWindow) getFilesInDirectory(dirPath string) []string {
	files := []string{}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return files
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}

	return files
}

// generatePatch generates the patch file
func (gw *GeneratorWindow) generatePatch() {
	gw.setStatus("Generating patch...")
	gw.generateBtn.Disable()

	if gw.batchMode {
		gw.generateBatchPatches()
		return
	}

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
	gw.appendLog(fmt.Sprintf("Using from key file: %s", gw.fromKeyFile))
	fromVer, err := versionMgr.RegisterVersion(gw.fromVersion, fromPath, gw.fromKeyFile)
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
	gw.appendLog(fmt.Sprintf("Using to key file: %s", gw.toKeyFile))
	toVer, err := versionMgr.RegisterVersion(gw.toVersion, toPath, gw.toKeyFile)
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
		CompressionLevel: gw.compressionLevel,
		VerifyAfter:      gw.verifyAfter,
		DiffThresholdKB:  gw.diffThresholdKB,
		SkipIdentical:    gw.skipIdentical,
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

// generateBatchPatches generates patches from all versions to target version
func (gw *GeneratorWindow) generateBatchPatches() {
	gw.appendLog("=== BATCH MODE: Generating patches from ALL versions ===")
	gw.appendLog(fmt.Sprintf("Target version: %s", gw.toVersion))
	gw.appendLog(fmt.Sprintf("Compression: %s (level %d)", gw.compression, gw.compressionLevel))

	// Create version manager
	versionMgr := version.NewManager()
	gw.manifestMgr = manifest.NewManager()

	// Register target version
	toPath := filepath.Join(gw.versionsDir, gw.toVersion)
	gw.appendLog("Scanning target version...")
	gw.appendLog(fmt.Sprintf("Using to key file: %s", gw.toKeyFile))

	toVer, err := versionMgr.RegisterVersion(gw.toVersion, toPath, gw.toKeyFile)
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

	if toVer.Manifest == nil {
		toVer.Manifest = &utils.Manifest{}
	}
	toVer.Manifest.Files = toFiles
	toVer.Manifest.Directories = toDirs
	toVer.Manifest.Version = gw.toVersion
	gw.appendLog(fmt.Sprintf("Target version: %d files, %d directories", len(toFiles), len(toDirs)))

	// Process each source version
	patchCount := 0
	failCount := 0

	for _, fromVersion := range gw.availableVersions {
		if fromVersion == gw.toVersion {
			continue // Skip target version itself
		}

		gw.appendLog(fmt.Sprintf("\n--- Processing %s → %s ---", fromVersion, gw.toVersion))

		fromPath := filepath.Join(gw.versionsDir, fromVersion)

		// Register source version (use from key file or fall back to to key file)
		keyFile := gw.fromKeyFile
		if keyFile == "" {
			keyFile = gw.toKeyFile
		}

		fromVer, err := versionMgr.RegisterVersion(fromVersion, fromPath, keyFile)
		if err != nil {
			gw.appendLog(fmt.Sprintf("WARNING: Skipping %s: %v", fromVersion, err))
			failCount++
			continue
		}

		// Scan source directory
		fromScanner := scanner.NewScanner(fromPath)
		fromFiles, fromDirs, err := fromScanner.ScanDirectory()
		if err != nil {
			gw.appendLog(fmt.Sprintf("WARNING: Failed to scan %s: %v", fromVersion, err))
			failCount++
			continue
		}

		if fromVer.Manifest == nil {
			fromVer.Manifest = &utils.Manifest{}
		}
		fromVer.Manifest.Files = fromFiles
		fromVer.Manifest.Directories = fromDirs
		fromVer.Manifest.Version = fromVersion

		// Generate patch
		outputPath := filepath.Join(gw.outputDir, fmt.Sprintf("%s-to-%s.patch", fromVersion, gw.toVersion))

		compressionStr := "zstd"
		switch gw.compression {
		case "gzip":
			compressionStr = "gzip"
		case "none":
			compressionStr = "none"
		}

		options := &utils.PatchOptions{
			Compression:      compressionStr,
			CompressionLevel: gw.compressionLevel,
			VerifyAfter:      gw.verifyAfter,
			DiffThresholdKB:  gw.diffThresholdKB,
			SkipIdentical:    gw.skipIdentical,
		}

		generator := patcher.NewGenerator()
		patch, err := generator.GeneratePatch(fromVer, toVer, options)
		if err != nil {
			gw.appendLog(fmt.Sprintf("ERROR: Failed to generate patch: %v", err))
			failCount++
			continue
		}

		// Validate patch
		if err := generator.ValidatePatch(patch); err != nil {
			gw.appendLog(fmt.Sprintf("ERROR: Patch validation failed: %v", err))
			failCount++
			continue
		}

		// Save patch
		if err := gw.savePatch(patch, outputPath, options); err != nil {
			gw.appendLog(fmt.Sprintf("ERROR: Failed to save patch: %v", err))
			failCount++
			continue
		}

		// Get patch file size
		info, err := os.Stat(outputPath)
		if err == nil {
			sizeKB := float64(info.Size()) / 1024.0
			sizeMB := sizeKB / 1024.0
			if sizeMB >= 1.0 {
				gw.appendLog(fmt.Sprintf("✓ Patch saved: %.2f MB", sizeMB))
			} else {
				gw.appendLog(fmt.Sprintf("✓ Patch saved: %.2f KB", sizeKB))
			}
		}

		patchCount++
	}

	gw.appendLog("\n=== BATCH COMPLETE ===")
	gw.appendLog(fmt.Sprintf("Generated: %d patches", patchCount))
	if failCount > 0 {
		gw.appendLog(fmt.Sprintf("Failed: %d patches", failCount))
	}

	if patchCount > 0 {
		gw.setStatus(fmt.Sprintf("Generated %d patches", patchCount))
		if gw.window != nil {
			dialog.ShowInformation("Batch Complete",
				fmt.Sprintf("Successfully generated %d patches\n\nOutput: %s", patchCount, gw.outputDir),
				gw.window)
		}
	} else {
		gw.setStatus("Error: No patches generated")
	}

	gw.generateBtn.Enable()
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
