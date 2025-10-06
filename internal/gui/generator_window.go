package gui

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image/color"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
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
	fromDir          string // Custom path for from version (different drive support)
	toDir            string // Custom path for to version (different drive support)
	useCustomPaths   bool   // Toggle between legacy mode and custom paths mode
	outputDir        string
	compression      string
	compressionLevel int
	verifyAfter      bool
	skipIdentical    bool
	batchMode        bool
	createExecutable bool
	fromKeyFile      string
	toKeyFile        string

	versionsDirEntry   *widget.Entry
	fromDirEntry       *widget.Entry // Custom from directory path
	toDirEntry         *widget.Entry // Custom to directory path
	customPathCheck    *widget.Check // Toggle for custom paths mode
	fromKeyFileSelect  *widget.Select
	toKeyFileSelect    *widget.Select
	fromVersionSelect  *widget.Select
	toVersionSelect    *widget.Select
	outputDirEntry     *widget.Entry
	compressionRadio   *widget.RadioGroup
	compressionSlider  *widget.Slider
	compressionLabel   *widget.Label
	verifyCheck        *widget.Check
	skipIdenticalCheck *widget.Check
	batchModeCheck     *widget.Check
	createExeCheck     *widget.Check
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

	// Create custom paths checkbox
	gw.customPathCheck = widget.NewCheck("Use Custom Paths (different drives/locations)", func(checked bool) {
		gw.useCustomPaths = checked
		if checked {
			// Enable custom path entries, disable legacy version selects and batch mode
			gw.fromDirEntry.Enable()
			gw.toDirEntry.Enable()
			gw.versionsDirEntry.Disable()
			gw.fromVersionSelect.Disable()
			gw.toVersionSelect.Disable()
			gw.batchModeCheck.Disable()
			gw.batchModeCheck.SetChecked(false)
			gw.batchMode = false
		} else {
			// Disable custom path entries, enable legacy version selects and batch mode
			gw.fromDirEntry.Disable()
			gw.toDirEntry.Disable()
			gw.versionsDirEntry.Enable()
			gw.fromVersionSelect.Enable()
			gw.toVersionSelect.Enable()
			gw.batchModeCheck.Enable()
		}
		gw.updateGenerateButton()
	})

	// Create custom from directory selector
	gw.fromDirEntry = widget.NewEntry()
	gw.fromDirEntry.SetPlaceHolder("Select source version directory...")
	gw.fromDirEntry.OnSubmitted = func(text string) {
		if text != "" {
			gw.fromDir = text
			gw.updateFromKeyFileOptionsCustom()
			gw.updateGenerateButton()
		}
	}
	gw.fromDirEntry.Disable() // Start disabled

	fromDirBrowse := widget.NewButton("Browse", func() {
		gw.selectFromDirectory()
	})

	fromDirContainer := container.NewBorder(
		nil, nil,
		widget.NewLabel("From Directory:"),
		fromDirBrowse,
		gw.fromDirEntry,
	)

	// Create custom to directory selector
	gw.toDirEntry = widget.NewEntry()
	gw.toDirEntry.SetPlaceHolder("Select target version directory...")
	gw.toDirEntry.OnSubmitted = func(text string) {
		if text != "" {
			gw.toDir = text
			gw.updateToKeyFileOptionsCustom()
			gw.updateGenerateButton()
		}
	}
	gw.toDirEntry.Disable() // Start disabled

	toDirBrowse := widget.NewButton("Browse", func() {
		gw.selectToDirectory()
	})

	toDirContainer := container.NewBorder(
		nil, nil,
		widget.NewLabel("To Directory:"),
		toDirBrowse,
		gw.toDirEntry,
	)

	// Create batch mode checkbox
	gw.batchModeCheck = widget.NewCheck("Batch Mode: Generate patches from ALL versions to target", func(checked bool) {
		gw.batchMode = checked
		if checked {
			// In batch mode, from version is not used
			gw.fromVersionSelect.Disable()
			gw.fromKeyFileSelect.Disable()
		} else {
			if !gw.useCustomPaths {
				gw.fromVersionSelect.Enable()
				gw.fromKeyFileSelect.Enable()
			}
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
		switch selected {
		case "zstd":
			gw.compressionSlider.Max = 4
			if gw.compressionLevel > 4 {
				gw.compressionLevel = 3
				gw.compressionSlider.Value = 3
			}
		case "gzip":
			gw.compressionSlider.Max = 9
		case "none":
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

	gw.createExeCheck = widget.NewCheck("Create self-contained executable", func(checked bool) {
		gw.createExecutable = checked
	})
	gw.createExeCheck.SetChecked(false)

	// Right column: Options
	rightColumn := container.NewVBox(
		widget.NewLabel("Compression:"),
		compressionContainer,
		widget.NewSeparator(),
		widget.NewLabel("Options:"),
		gw.verifyCheck,
		gw.skipIdenticalCheck,
		gw.createExeCheck,
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
		gw.customPathCheck,
		widget.NewSeparator(),
		// Legacy mode (single versions directory)
		versionsDirContainer,
		// Custom paths mode (different drives/locations)
		fromDirContainer,
		toDirContainer,
		widget.NewSeparator(),
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

// selectFromDirectory opens a folder dialog for selecting source version directory
func (gw *GeneratorWindow) selectFromDirectory() {
	if gw.window == nil {
		return
	}

	dialog.ShowFolderOpen(func(dir fyne.ListableURI, err error) {
		if err == nil && dir != nil {
			path := dir.Path()
			gw.fromDirEntry.SetText(path)
			gw.fromDir = path
			gw.updateFromKeyFileOptionsCustom()
			gw.updateGenerateButton()
		}
	}, gw.window)
}

// selectToDirectory opens a folder dialog for selecting target version directory
func (gw *GeneratorWindow) selectToDirectory() {
	if gw.window == nil {
		return
	}

	dialog.ShowFolderOpen(func(dir fyne.ListableURI, err error) {
		if err == nil && dir != nil {
			path := dir.Path()
			gw.toDirEntry.SetText(path)
			gw.toDir = path
			gw.updateToKeyFileOptionsCustom()
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
	var canGenerate bool

	if gw.useCustomPaths {
		// Custom paths mode: need fromDir, toDir, and outputDir
		canGenerate = gw.fromDir != "" && gw.toDir != "" && gw.outputDir != ""
	} else {
		// Legacy mode
		if gw.batchMode {
			// In batch mode, only need to version and output dir
			canGenerate = gw.toVersion != "" && gw.outputDir != ""
		} else {
			// In normal mode, need both from and to versions
			canGenerate = gw.fromVersion != "" && gw.toVersion != "" && gw.outputDir != ""
		}
	}

	if canGenerate {
		gw.generateBtn.Enable()
	} else {
		gw.generateBtn.Disable()
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

	// Auto-select if only one file found (works with any file type)
	if len(files) == 1 {
		gw.fromKeyFileSelect.SetSelected(files[0])
		gw.fromKeyFile = files[0]
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

	// Auto-select if only one file found (works with any file type)
	if len(files) == 1 {
		gw.toKeyFileSelect.SetSelected(files[0])
		gw.toKeyFile = files[0]
	}
}

// updateFromKeyFileOptionsCustom scans from directory and populates key file options (custom paths mode)
func (gw *GeneratorWindow) updateFromKeyFileOptionsCustom() {
	if gw.fromDir == "" {
		return
	}

	// Check if directory exists
	if _, err := os.Stat(gw.fromDir); os.IsNotExist(err) {
		return
	}

	files := gw.getFilesInDirectory(gw.fromDir)
	gw.fromKeyFileSelect.Options = files
	gw.fromKeyFileSelect.Refresh()

	// Auto-select if only one file found (works with any file type)
	if len(files) == 1 {
		gw.fromKeyFileSelect.SetSelected(files[0])
		gw.fromKeyFile = files[0]
	}
}

// updateToKeyFileOptionsCustom scans to directory and populates key file options (custom paths mode)
func (gw *GeneratorWindow) updateToKeyFileOptionsCustom() {
	if gw.toDir == "" {
		return
	}

	// Check if directory exists
	if _, err := os.Stat(gw.toDir); os.IsNotExist(err) {
		return
	}

	files := gw.getFilesInDirectory(gw.toDir)
	gw.toKeyFileSelect.Options = files
	gw.toKeyFileSelect.Refresh()

	// Auto-select if only one file found (works with any file type)
	if len(files) == 1 {
		gw.toKeyFileSelect.SetSelected(files[0])
		gw.toKeyFile = files[0]
	}
}

// getFilesInDirectory returns a list of all files in the directory tree (recursive)
// Returns relative paths from the root directory (e.g., "cmd/applier/main.go")
func (gw *GeneratorWindow) getFilesInDirectory(dirPath string) []string {
	files := []string{}

	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Skip directories/files that cause errors but continue walking
			return nil
		}

		// Skip directories, only include files
		if d.IsDir() {
			return nil
		}

		// Calculate relative path from root
		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			// If we can't get relative path, skip this file
			return nil
		}

		// Normalize path separators to forward slashes for consistency
		relPath = filepath.ToSlash(relPath)

		files = append(files, relPath)
		return nil
	})

	if err != nil {
		// If walk fails completely, return empty slice
		return []string{}
	}

	// Sort files alphabetically for better UX
	sort.Strings(files)

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

	// Determine paths and version numbers based on mode
	var fromPath, toPath, fromVersion, toVersion string

	if gw.useCustomPaths {
		// Custom paths mode: use full directory paths and extract version numbers
		fromPath = gw.fromDir
		toPath = gw.toDir
		fromVersion = filepath.Base(fromPath)
		toVersion = filepath.Base(toPath)
		gw.appendLog("=== CUSTOM PATHS MODE ===")
		gw.appendLog(fmt.Sprintf("From: %s (version: %s)", fromPath, fromVersion))
		gw.appendLog(fmt.Sprintf("To: %s (version: %s)", toPath, toVersion))
	} else {
		// Legacy mode: construct paths from versions directory and version names
		fromPath = filepath.Join(gw.versionsDir, gw.fromVersion)
		toPath = filepath.Join(gw.versionsDir, gw.toVersion)
		fromVersion = gw.fromVersion
		toVersion = gw.toVersion
		gw.appendLog("=== LEGACY MODE ===")
		gw.appendLog(fmt.Sprintf("From: %s", fromPath))
		gw.appendLog(fmt.Sprintf("To: %s", toPath))
	}

	// Validate selections
	if fromVersion == toVersion {
		gw.setStatus("Error: From and To versions must be different")
		gw.appendLog("ERROR: Cannot generate patch from same version to same version")
		gw.generateBtn.Enable()
		return
	}

	// Build output path using version numbers
	outputPath := filepath.Join(gw.outputDir, fmt.Sprintf("%s-to-%s.patch", fromVersion, toVersion))

	gw.appendLog(fmt.Sprintf("Output: %s", outputPath))
	gw.appendLog(fmt.Sprintf("Compression: %s", gw.compression))

	// Create version manager and manifest manager
	versionMgr := version.NewManager()
	gw.manifestMgr = manifest.NewManager()

	// Register and scan source version
	gw.appendLog("Scanning source version...")
	gw.appendLog(fmt.Sprintf("Using from key file: %s", gw.fromKeyFile))
	fromVer, err := versionMgr.RegisterVersion(fromVersion, fromPath, gw.fromKeyFile)
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
	fromVer.Manifest.Version = fromVersion
	gw.appendLog(fmt.Sprintf("Source version: %d files, %d directories", len(fromFiles), len(fromDirs)))

	// Register and scan target version
	gw.appendLog("Scanning target version...")
	gw.appendLog(fmt.Sprintf("Using to key file: %s", gw.toKeyFile))
	toVer, err := versionMgr.RegisterVersion(toVersion, toPath, gw.toKeyFile)
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
	toVer.Manifest.Version = toVersion
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

	// Create self-contained executable if requested
	if gw.createExecutable {
		exePath := strings.TrimSuffix(outputPath, ".patch") + ".exe"
		gw.appendLog("Creating self-contained executable...")
		if err := gw.createStandaloneExe(outputPath, exePath); err != nil {
			gw.appendLog("ERROR: Failed to create executable: " + err.Error())
		} else {
			info, err := os.Stat(exePath)
			if err == nil {
				sizeMB := float64(info.Size()) / (1024.0 * 1024.0)
				gw.appendLog(fmt.Sprintf("Executable size: %.2f MB", sizeMB))
			}
			gw.appendLog(fmt.Sprintf("✓ Created: %s", exePath))
		}
	}

	gw.generateBtn.Enable()

	// Show success dialog
	if gw.window != nil {
		successMsg := fmt.Sprintf("Patch generated successfully!\n\nOutput: %s", outputPath)
		if gw.createExecutable {
			exePath := strings.TrimSuffix(outputPath, ".patch") + ".exe"
			successMsg += fmt.Sprintf("\nExecutable: %s", exePath)
		}
		dialog.ShowInformation("Success", successMsg, gw.window)
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

// createStandaloneExe creates a self-contained executable by appending patch data to applier-gui.exe
func (gw *GeneratorWindow) createStandaloneExe(patchPath, exePath string) error {
	// Get path to patch-apply-gui.exe (same directory as patch-gen-gui.exe)
	genExe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	applierPath := filepath.Join(filepath.Dir(genExe), "patch-apply-gui.exe")

	// Read the applier GUI executable
	applierData, err := os.ReadFile(applierPath)
	if err != nil {
		return fmt.Errorf("failed to read applier executable: %w", err)
	}

	// Read the patch file
	patchData, err := os.ReadFile(patchPath)
	if err != nil {
		return fmt.Errorf("failed to read patch file: %w", err)
	}

	// Calculate checksum of patch data
	checksum := sha256.Sum256(patchData)

	// Create 128-byte header
	header := make([]byte, 128)

	// Magic bytes "CPMPATCH" (8 bytes)
	copy(header[0:8], []byte("CPMPATCH"))

	// Version (4 bytes, uint32)
	binary.LittleEndian.PutUint32(header[8:12], 1)

	// Stub size (8 bytes, uint64) - size of applier exe
	binary.LittleEndian.PutUint64(header[12:20], uint64(len(applierData)))

	// Data offset (8 bytes, uint64) - right after stub
	binary.LittleEndian.PutUint64(header[20:28], uint64(len(applierData)))

	// Data size (8 bytes, uint64)
	binary.LittleEndian.PutUint64(header[28:36], uint64(len(patchData)))

	// Compression type (16 bytes)
	compressionBytes := make([]byte, 16)
	copy(compressionBytes, []byte(gw.compression))
	copy(header[36:52], compressionBytes)

	// Checksum (32 bytes, SHA-256)
	copy(header[52:84], checksum[:])

	// Reserved (44 bytes) - already zeroed

	// Create output file
	outFile, err := os.Create(exePath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	// Write: applier.exe + patch data + header
	if _, err := outFile.Write(applierData); err != nil {
		return fmt.Errorf("failed to write applier data: %w", err)
	}

	if _, err := outFile.Write(patchData); err != nil {
		return fmt.Errorf("failed to write patch data: %w", err)
	}

	if _, err := outFile.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	return nil
}
