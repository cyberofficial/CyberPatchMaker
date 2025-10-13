package gui

import (
	"bufio"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image/color"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
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
	versionsDir          string
	fromVersion          string
	toVersion            string
	fromDir              string // Custom path for from version (different drive support)
	toDir                string // Custom path for to version (different drive support)
	useCustomPaths       bool   // Toggle between legacy mode and custom paths mode
	outputDir            string
	compression          string
	compressionLevel     int
	verifyAfter          bool
	skipIdentical        bool
	batchMode            bool
	createExecutable     bool
	exeType              string // "gui" or "console" for self-contained executable type
	createReversePatches bool
	ignore1GB            bool
	simpleModeForUsers   bool   // Enable simplified UI for end users
	useScanCache         bool   // Enable scan caching
	forceRescan          bool   // Force rescan despite cache
	cacheDir             string // Custom cache directory
	workerThreads        int    // Number of worker threads for parallel operations
	fromKeyFile          string
	toKeyFile            string

	versionsDirEntry    *widget.Entry
	fromDirEntry        *widget.Entry // Custom from directory path
	toDirEntry          *widget.Entry // Custom to directory path
	customPathCheck     *widget.Check // Toggle for custom paths mode
	fromKeyFileEntry    *widget.Entry
	toKeyFileEntry      *widget.Entry
	fromVersionSelect   *widget.Select
	toVersionSelect     *widget.Select
	outputDirEntry      *widget.Entry
	compressionRadio    *widget.RadioGroup
	compressionSlider   *widget.Slider
	compressionLabel    *widget.Label
	verifyCheck         *widget.Check
	skipIdenticalCheck  *widget.Check
	batchModeCheck      *widget.Check
	createExeCheck      *widget.Check
	exeTypeRadio        *widget.RadioGroup
	crpCheck            *widget.Check
	ignore1GBCheck      *widget.Check
	simpleModeCheck     *widget.Check
	useScanCacheCheck   *widget.Check
	forceRescanCheck    *widget.Check
	cacheDirEntry       *widget.Entry
	workerThreadsSlider *widget.Slider
	workerThreadsLabel  *widget.Label
	generateBtn         *widget.Button
	statusLabel         *widget.Label
	logText             *widget.Entry

	// Data
	availableVersions []string
	manifestMgr       *manifest.Manager
}

// NewGeneratorWindow creates a new generator window
func NewGeneratorWindow() *GeneratorWindow {
	gw := &GeneratorWindow{
		compression:          "zstd",
		compressionLevel:     3,
		verifyAfter:          true,
		skipIdentical:        true,
		batchMode:            false,
		exeType:              "gui",
		createReversePatches: false,
		useScanCache:         false,
		forceRescan:          false,
		cacheDir:             ".data",
		workerThreads:        0, // 0 = auto-detect
		fromKeyFile:          "",
		toKeyFile:            "",
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
	gw.versionsDirEntry.OnChanged = func(text string) {
		if text != "" {
			gw.versionsDir = text
		}
	}
	gw.versionsDirEntry.OnSubmitted = func(text string) {
		if text != "" {
			gw.versionsDir = text
			gw.scanVersions()
		}
	}

	versionsDirBrowse := widget.NewButton("Browse", func() {
		gw.selectVersionsDirectory()
	})

	versionsDirHint := widget.NewLabel("Hint: Press [Enter] to scan after pasting a path")
	versionsDirHint.TextStyle.Italic = true

	versionsDirContainer := container.NewBorder(
		nil, versionsDirHint,
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
	gw.fromDirEntry.OnChanged = func(text string) {
		gw.fromDir = text
		gw.updateGenerateButton()
	}
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
	gw.toDirEntry.OnChanged = func(text string) {
		gw.toDir = text
		gw.updateGenerateButton()
	}
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
			gw.fromKeyFileEntry.Disable()
		} else {
			if !gw.useCustomPaths {
				gw.fromVersionSelect.Enable()
				gw.fromKeyFileEntry.Enable()
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

	// Create from key file entry
	gw.fromKeyFileEntry = widget.NewEntry()
	gw.fromKeyFileEntry.SetPlaceHolder("Enter key file path (e.g., program.exe or cmd/main.go)...")
	gw.fromKeyFileEntry.SetText(gw.fromKeyFile)
	gw.fromKeyFileEntry.OnChanged = func(text string) {
		gw.fromKeyFile = text
	}

	// Create to version selector
	gw.toVersionSelect = widget.NewSelect([]string{}, func(selected string) {
		gw.toVersion = selected
		gw.updateToKeyFileOptions()
		gw.updateGenerateButton()
	})
	gw.toVersionSelect.PlaceHolder = "Select target version..."

	// Create to key file entry
	gw.toKeyFileEntry = widget.NewEntry()
	gw.toKeyFileEntry.SetPlaceHolder("Enter key file path (e.g., program.exe or cmd/main.go)...")
	gw.toKeyFileEntry.SetText(gw.toKeyFile)
	gw.toKeyFileEntry.OnChanged = func(text string) {
		gw.toKeyFile = text
	}

	// Left column: Version selection
	leftColumn := container.NewVBox(
		container.NewBorder(nil, nil, widget.NewLabel("From:"), nil, gw.fromVersionSelect),
		container.NewBorder(nil, nil, widget.NewLabel("Key:"), nil, gw.fromKeyFileEntry),
		container.NewBorder(nil, nil, widget.NewLabel("To:"), nil, gw.toVersionSelect),
		container.NewBorder(nil, nil, widget.NewLabel("Key:"), nil, gw.toKeyFileEntry),
	)

	// Create output directory selector
	gw.outputDirEntry = widget.NewEntry()
	gw.outputDirEntry.SetPlaceHolder("Select output directory...")
	gw.outputDirEntry.OnChanged = func(text string) {
		gw.outputDir = text
		gw.updateGenerateButton()
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
			gw.compressionSlider.Max = 3
			if gw.compressionLevel > 3 {
				gw.compressionLevel = 2
				gw.compressionSlider.Value = 2
			}
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
		container.NewBorder(nil, nil, widget.NewLabel("Lvl:"), gw.compressionLabel, gw.compressionSlider),
	)

	// Create advanced options
	gw.verifyCheck = widget.NewCheck("Verify after", func(checked bool) {
		gw.verifyAfter = checked
	})
	gw.verifyCheck.SetChecked(true)

	gw.skipIdenticalCheck = widget.NewCheck("Skip identical", func(checked bool) {
		gw.skipIdentical = checked
	})
	gw.skipIdenticalCheck.SetChecked(true)

	gw.createExeCheck = widget.NewCheck("Create .exe", func(checked bool) {
		gw.createExecutable = checked
		if checked {
			gw.exeTypeRadio.Enable()
		} else {
			gw.exeTypeRadio.Disable()
		}
	})
	gw.createExeCheck.SetChecked(false)

	// Executable type selector
	gw.exeTypeRadio = widget.NewRadioGroup([]string{"GUI", "Console"}, func(selected string) {
		if selected == "GUI" {
			gw.exeType = "gui"
		} else {
			gw.exeType = "console"
		}
	})
	gw.exeTypeRadio.Horizontal = true
	gw.exeTypeRadio.SetSelected("GUI")
	gw.exeTypeRadio.Disable() // Disabled until createExeCheck is enabled

	gw.crpCheck = widget.NewCheck("Reverse patch", func(checked bool) {
		gw.createReversePatches = checked
	})
	gw.crpCheck.SetChecked(false)

	gw.ignore1GBCheck = widget.NewCheck("Ignore 1GB limit", func(checked bool) {
		gw.ignore1GB = checked
	})
	gw.ignore1GBCheck.SetChecked(false)

	gw.simpleModeCheck = widget.NewCheck("Simple Mode for End Users", func(checked bool) {
		gw.simpleModeForUsers = checked
	})
	gw.simpleModeCheck.SetChecked(false)

	// Scan cache checkbox
	gw.useScanCacheCheck = widget.NewCheck("Scan cache", func(checked bool) {
		gw.useScanCache = checked
		if checked {
			gw.forceRescanCheck.Enable()
			gw.cacheDirEntry.Enable()
		} else {
			gw.forceRescanCheck.Disable()
			gw.cacheDirEntry.Disable()
		}
	})
	gw.useScanCacheCheck.SetChecked(false)

	// Force rescan checkbox
	gw.forceRescanCheck = widget.NewCheck("Force rescan", func(checked bool) {
		gw.forceRescan = checked
	})
	gw.forceRescanCheck.SetChecked(false)
	gw.forceRescanCheck.Disable() // Disabled until cache is enabled

	// Cache directory entry
	gw.cacheDirEntry = widget.NewEntry()
	gw.cacheDirEntry.SetPlaceHolder(".data")
	gw.cacheDirEntry.SetText(".data")
	gw.cacheDirEntry.OnChanged = func(text string) {
		if text != "" {
			gw.cacheDir = text
		} else {
			gw.cacheDir = ".data"
		}
	}
	gw.cacheDirEntry.Disable() // Disabled until cache is enabled

	// Worker threads slider - max is CPU core count
	cpuCores := runtime.NumCPU()
	gw.workerThreadsSlider = widget.NewSlider(0, float64(cpuCores))
	gw.workerThreadsSlider.Value = 0 // 0 = auto-detect (use all cores)
	gw.workerThreadsSlider.Step = 1
	gw.workerThreadsSlider.OnChanged = func(value float64) {
		gw.workerThreads = int(value)
		gw.updateWorkerThreadsLabel()
	}

	gw.workerThreadsLabel = widget.NewLabel(fmt.Sprintf("Auto (%d)", cpuCores))

	// Right column: Options
	rightColumn := container.NewVBox(
		compressionContainer,
		gw.verifyCheck,
		gw.skipIdenticalCheck,
		gw.createExeCheck,
		gw.exeTypeRadio,
		gw.crpCheck,
		gw.ignore1GBCheck,
		gw.simpleModeCheck,
	)

	// Third column: Cache options
	cacheColumn := container.NewVBox(
		gw.useScanCacheCheck,
		gw.forceRescanCheck,
		container.NewBorder(nil, nil, widget.NewLabel("Dir:"), nil, gw.cacheDirEntry),
		container.NewBorder(nil, nil, widget.NewLabel("Workers:"), gw.workerThreadsLabel, gw.workerThreadsSlider),
	)

	// Create three-column layout for version/options/cache
	threeColumnLayout := container.NewGridWithColumns(3,
		leftColumn,
		rightColumn,
		cacheColumn,
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
	logContainer.SetMinSize(fyne.NewSize(0, 80))

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
		threeColumnLayout,
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

	// Auto-set output directory to versionsDir/patchfiles if not already set
	if gw.outputDir == "" {
		patchFilesDir := filepath.Join(gw.versionsDir, "patchfiles")
		gw.outputDir = patchFilesDir
		gw.outputDirEntry.SetText(patchFilesDir)
		gw.appendLog("Auto-set output directory: " + patchFilesDir)
		// Note: Directory creation moved to generatePatch() to avoid creating folders before user clicks Generate
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

// updateWorkerThreadsLabel updates the worker threads label
func (gw *GeneratorWindow) updateWorkerThreadsLabel() {
	if gw.workerThreads == 0 {
		cpuCores := runtime.NumCPU()
		gw.workerThreadsLabel.SetText(fmt.Sprintf("Auto (%d)", cpuCores))
	} else {
		gw.workerThreadsLabel.SetText(fmt.Sprintf("%d", gw.workerThreads))
	}
}

// updateFromKeyFileOptions updates the from key file entry with auto-detected file if available
func (gw *GeneratorWindow) updateFromKeyFileOptions() {
	if gw.fromVersion == "" || gw.versionsDir == "" {
		return
	}

	versionPath := filepath.Join(gw.versionsDir, gw.fromVersion)
	files := gw.getFilesInDirectory(versionPath)

	// Auto-populate if only one file found (works with any file type)
	if len(files) == 1 {
		gw.fromKeyFileEntry.SetText(files[0])
		gw.fromKeyFile = files[0]
	}
}

// updateToKeyFileOptions updates the to key file entry with auto-detected file if available
func (gw *GeneratorWindow) updateToKeyFileOptions() {
	if gw.toVersion == "" || gw.versionsDir == "" {
		return
	}

	versionPath := filepath.Join(gw.versionsDir, gw.toVersion)
	files := gw.getFilesInDirectory(versionPath)

	// Auto-populate if only one file found (works with any file type)
	if len(files) == 1 {
		gw.toKeyFileEntry.SetText(files[0])
		gw.toKeyFile = files[0]
	}
}

// updateFromKeyFileOptionsCustom updates the from key file entry with auto-detected file if available (custom paths mode)
func (gw *GeneratorWindow) updateFromKeyFileOptionsCustom() {
	if gw.fromDir == "" {
		return
	}

	// Check if directory exists
	if _, err := os.Stat(gw.fromDir); os.IsNotExist(err) {
		return
	}

	files := gw.getFilesInDirectory(gw.fromDir)

	// Auto-populate if only one file found (works with any file type)
	if len(files) == 1 {
		gw.fromKeyFileEntry.SetText(files[0])
		gw.fromKeyFile = files[0]
	}
}

// updateToKeyFileOptionsCustom updates the to key file entry with auto-detected file if available (custom paths mode)
func (gw *GeneratorWindow) updateToKeyFileOptionsCustom() {
	if gw.toDir == "" {
		return
	}

	// Check if directory exists
	if _, err := os.Stat(gw.toDir); os.IsNotExist(err) {
		return
	}

	files := gw.getFilesInDirectory(gw.toDir)

	// Auto-populate if only one file found (works with any file type)
	if len(files) == 1 {
		gw.toKeyFileEntry.SetText(files[0])
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

	// Create output directory if it doesn't exist (only when user clicks Generate)
	if err := os.MkdirAll(gw.outputDir, 0755); err != nil {
		gw.setStatus("Error: Failed to create output directory")
		gw.appendLog("ERROR: Could not create output directory: " + err.Error())
		gw.generateBtn.Enable()
		return
	}

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

	// Set worker threads for parallel operations
	workerCount := gw.workerThreads
	if workerCount == 0 {
		// Auto-detect: use runtime.NumCPU()
		workerCount = 0 // Will be set by config in manager
	}
	if workerCount > 0 {
		versionMgr.SetWorkerThreads(workerCount)
		if workerCount > 1 {
			gw.appendLog(fmt.Sprintf("Using %d worker threads for parallel operations", workerCount))
		}
	}

	// Enable scan cache if requested
	if gw.useScanCache {
		versionMgr.EnableScanCache(gw.cacheDir, gw.forceRescan)
		gw.appendLog(fmt.Sprintf("✓ Scan caching enabled (cache dir: %s)", gw.cacheDir))
		if gw.forceRescan {
			gw.appendLog("  Force rescan: enabled")
		}
	}

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
	gw.appendLog("")

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
	gw.appendLog(fmt.Sprintf("Target version: %d files, %d directories", len(toFiles), len(toDirs)))
	gw.appendLog("")
	gw.appendLog("Preparing to compare versions...\nThis may take a while, program will build and gather the files to hash")
	gw.setStatus("Preparing comparison...")

	// Generate patch
	gw.appendLog("")
	gw.appendLog("=== Computing File Differences ===")
	gw.appendLog("Analyzing differences between versions...")
	gw.appendLog(fmt.Sprintf("Comparing %d source files with %d target files...", len(fromFiles), len(toFiles)))
	gw.appendLog("Please wait - this may take several minutes for large projects")
	gw.setStatus("Computing differences...")
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

	// Set simple mode if enabled
	patch.SimpleMode = gw.simpleModeForUsers

	// Report what was found
	gw.appendLog("\n=== Patch Operations Summary ===")
	addedCount := 0
	modifiedCount := 0
	deletedCount := 0
	for _, op := range patch.Operations {
		switch op.Type {
		case utils.OpAdd:
			addedCount++
		case utils.OpModify:
			modifiedCount++
		case utils.OpDelete:
			deletedCount++
		}
	}
	gw.appendLog(fmt.Sprintf("Files to add: %d", addedCount))
	gw.appendLog(fmt.Sprintf("Files to modify: %d", modifiedCount))
	gw.appendLog(fmt.Sprintf("Files to delete: %d", deletedCount))
	gw.appendLog(fmt.Sprintf("Total operations: %d", len(patch.Operations)))

	// Validate patch
	gw.appendLog("\nValidating patch structure...")
	gw.setStatus("Validating patch...")
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
		if gw.exeType == "console" {
			gw.appendLog("Creating self-contained Console Host executable...")
		} else {
			gw.appendLog("Creating self-contained GUI executable...")
		}
		if err := gw.createStandaloneExe(outputPath, exePath, gw.exeType); err != nil {
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

	// Create reverse patch if requested (for downgrades)
	if gw.createReversePatches {
		gw.appendLog("\n--- Creating reverse patch (for downgrades) ---")
		reversePatchPath := filepath.Join(gw.outputDir, fmt.Sprintf("%s-to-%s.patch", toVersion, fromVersion))
		gw.appendLog(fmt.Sprintf("Reverse patch: %s → %s", toVersion, fromVersion))

		// Generate reverse patch (swap from and to)
		reverseGenerator := patcher.NewGenerator()
		reversePatch, err := reverseGenerator.GeneratePatch(toVer, fromVer, options)
		if err != nil {
			gw.appendLog("ERROR: Failed to generate reverse patch: " + err.Error())
		} else {
			// Set simple mode if enabled
			reversePatch.SimpleMode = gw.simpleModeForUsers

			// Validate reverse patch
			if err := reverseGenerator.ValidatePatch(reversePatch); err != nil {
				gw.appendLog("ERROR: Reverse patch validation failed: " + err.Error())
			} else {
				// Save reverse patch
				if err := gw.savePatch(reversePatch, reversePatchPath, options); err != nil {
					gw.appendLog("ERROR: Failed to save reverse patch: " + err.Error())
				} else {
					// Get reverse patch file size
					info, err := os.Stat(reversePatchPath)
					if err == nil {
						sizeKB := float64(info.Size()) / 1024.0
						sizeMB := sizeKB / 1024.0
						if sizeMB >= 1.0 {
							gw.appendLog(fmt.Sprintf("Reverse patch size: %.2f MB", sizeMB))
						} else {
							gw.appendLog(fmt.Sprintf("Reverse patch size: %.2f KB", sizeKB))
						}
					}
					gw.appendLog(fmt.Sprintf("✓ Reverse patch saved: %s", reversePatchPath))

					// Create reverse exe if requested
					if gw.createExecutable {
						reverseExePath := strings.TrimSuffix(reversePatchPath, ".patch") + ".exe"
						gw.appendLog("Creating reverse executable...")
						if err := gw.createStandaloneExe(reversePatchPath, reverseExePath, gw.exeType); err != nil {
							gw.appendLog("ERROR: Failed to create reverse executable: " + err.Error())
						} else {
							info, err := os.Stat(reverseExePath)
							if err == nil {
								sizeMB := float64(info.Size()) / (1024.0 * 1024.0)
								gw.appendLog(fmt.Sprintf("Reverse executable size: %.2f MB", sizeMB))
							}
							gw.appendLog(fmt.Sprintf("✓ Created: %s", reverseExePath))
						}
					}
				}
			}
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
		if gw.createReversePatches {
			reversePatchPath := filepath.Join(gw.outputDir, fmt.Sprintf("%s-to-%s.patch", toVersion, fromVersion))
			successMsg += fmt.Sprintf("\nReverse patch: %s", reversePatchPath)
			if gw.createExecutable {
				reverseExePath := strings.TrimSuffix(reversePatchPath, ".patch") + ".exe"
				successMsg += fmt.Sprintf("\nReverse executable: %s", reverseExePath)
			}
		}
		dialog.ShowInformation("Success", successMsg, gw.window)
	}
}

// generateBatchPatches generates patches from all versions to target version
func (gw *GeneratorWindow) generateBatchPatches() {
	gw.appendLog("=== BATCH MODE: Generating patches from ALL versions ===")
	gw.appendLog(fmt.Sprintf("Target version: %s", gw.toVersion))
	gw.appendLog(fmt.Sprintf("Compression: %s (level %d)", gw.compression, gw.compressionLevel))

	// Create output directory if it doesn't exist (only when user clicks Generate)
	if err := os.MkdirAll(gw.outputDir, 0755); err != nil {
		gw.setStatus("Error: Failed to create output directory")
		gw.appendLog("ERROR: Could not create output directory: " + err.Error())
		gw.generateBtn.Enable()
		return
	}

	// Create version manager
	versionMgr := version.NewManager()
	gw.manifestMgr = manifest.NewManager()

	// Set worker threads for parallel operations
	workerCount := gw.workerThreads
	if workerCount == 0 {
		// Auto-detect: use runtime.NumCPU()
		workerCount = runtime.NumCPU()
	}
	if workerCount < 1 {
		workerCount = 1
	}
	versionMgr.SetWorkerThreads(workerCount)
	if workerCount > 1 {
		gw.appendLog(fmt.Sprintf("Using %d worker threads for parallel operations", workerCount))
	}

	// Enable scan cache if requested
	if gw.useScanCache {
		versionMgr.EnableScanCache(gw.cacheDir, gw.forceRescan)
		gw.appendLog(fmt.Sprintf("✓ Scan caching enabled (cache dir: %s)", gw.cacheDir))
		if gw.forceRescan {
			gw.appendLog("  Force rescan: enabled")
		}
	}

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
	gw.appendLog("")
	gw.appendLog("Target scan complete - ready for batch generation")
	gw.appendLog("Preparing to compare versions...\nThis may take a while, program will build and gather the files to hash")
	gw.setStatus("Preparing batch comparison...")

	// Process each source version
	patchCount := 0
	failCount := 0

	for _, fromVersion := range gw.availableVersions {
		if fromVersion == gw.toVersion {
			continue // Skip target version itself
		}

		gw.appendLog(fmt.Sprintf("\n--- Processing %s → %s ---", fromVersion, gw.toVersion))
		gw.setStatus(fmt.Sprintf("Processing %s → %s...", fromVersion, gw.toVersion))

		fromPath := filepath.Join(gw.versionsDir, fromVersion)

		// Auto-detect key file for this source version independently
		var fromKeyFile string
		if gw.fromKeyFile != "" {
			// Use custom key file if specified by user
			fromKeyFile = gw.fromKeyFile
		} else {
			// Auto-detect key file from common names
			keyFiles := []string{"program.exe", "game.exe", "app.exe", "main.exe"}
			for _, kf := range keyFiles {
				if utils.FileExists(filepath.Join(fromPath, kf)) {
					fromKeyFile = kf
					break
				}
			}
			if fromKeyFile == "" {
				// If auto-detection failed, try using the same key file as the target version
				if gw.toKeyFile != "" && utils.FileExists(filepath.Join(fromPath, gw.toKeyFile)) {
					fromKeyFile = gw.toKeyFile
					gw.appendLog(fmt.Sprintf("Using target key file for %s: %s", fromVersion, gw.toKeyFile))
				} else {
					gw.appendLog(fmt.Sprintf("WARNING: Skipping %s: could not find key file (tried: program.exe, game.exe, app.exe, main.exe)", fromVersion))
					failCount++
					continue
				}
			}
		}

		fromVer, err := versionMgr.RegisterVersion(fromVersion, fromPath, fromKeyFile)
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

		// Set simple mode if enabled
		patch.SimpleMode = gw.simpleModeForUsers

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

		// Create self-contained executable if requested
		if gw.createExecutable {
			exePath := strings.TrimSuffix(outputPath, ".patch") + ".exe"
			if gw.exeType == "console" {
				gw.appendLog("Creating self-contained Console Host executable...")
			} else {
				gw.appendLog("Creating self-contained GUI executable...")
			}
			if err := gw.createStandaloneExe(outputPath, exePath, gw.exeType); err != nil {
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

		patchCount++

		// Create reverse patch if requested (for downgrades)
		if gw.createReversePatches {
			gw.appendLog(fmt.Sprintf("Creating reverse patch: %s → %s", gw.toVersion, fromVersion))
			reversePatchPath := filepath.Join(gw.outputDir, fmt.Sprintf("%s-to-%s.patch", gw.toVersion, fromVersion))

			// Generate reverse patch (swap from and to)
			reverseGenerator := patcher.NewGenerator()
			reversePatch, err := reverseGenerator.GeneratePatch(toVer, fromVer, options)
			if err != nil {
				gw.appendLog(fmt.Sprintf("ERROR: Failed to generate reverse patch: %v", err))
			} else {
				// Set simple mode if enabled
				reversePatch.SimpleMode = gw.simpleModeForUsers

				// Validate reverse patch
				if err := reverseGenerator.ValidatePatch(reversePatch); err != nil {
					gw.appendLog(fmt.Sprintf("ERROR: Reverse patch validation failed: %v", err))
				} else {
					// Save reverse patch
					if err := gw.savePatch(reversePatch, reversePatchPath, options); err != nil {
						gw.appendLog(fmt.Sprintf("ERROR: Failed to save reverse patch: %v", err))
					} else {
						// Get reverse patch file size
						info, err := os.Stat(reversePatchPath)
						if err == nil {
							sizeKB := float64(info.Size()) / 1024.0
							sizeMB := sizeKB / 1024.0
							if sizeMB >= 1.0 {
								gw.appendLog(fmt.Sprintf("✓ Reverse patch saved: %.2f MB", sizeMB))
							} else {
								gw.appendLog(fmt.Sprintf("✓ Reverse patch saved: %.2f KB", sizeKB))
							}
						}

						// Create reverse exe if requested
						if gw.createExecutable {
							reverseExePath := strings.TrimSuffix(reversePatchPath, ".patch") + ".exe"
							gw.appendLog("Creating reverse executable...")
							if err := gw.createStandaloneExe(reversePatchPath, reverseExePath, gw.exeType); err != nil {
								gw.appendLog("ERROR: Failed to create reverse executable: " + err.Error())
							} else {
								info, err := os.Stat(reverseExePath)
								if err == nil {
									sizeMB := float64(info.Size()) / (1024.0 * 1024.0)
									gw.appendLog(fmt.Sprintf("Reverse executable size: %.2f MB", sizeMB))
								}
								gw.appendLog(fmt.Sprintf("✓ Created: %s", reverseExePath))
							}
						}

						patchCount++
					}
				}
			}
		}
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

// savePatch saves the patch to a file with compression and streaming to avoid memory exhaustion
func (gw *GeneratorWindow) savePatch(patch *utils.Patch, filename string, options *utils.PatchOptions) error {
	// Create output file
	outFile, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create patch file: %w", err)
	}
	defer outFile.Close()

	// Create a pipe for streaming JSON encoding
	jsonReader, jsonWriter := io.Pipe()
	defer jsonReader.Close()

	// Start custom streaming JSON encoding in a goroutine
	encodeErr := make(chan error, 1)
	go func() {
		defer jsonWriter.Close()
		encodeErr <- gw.encodePatchStreaming(patch, jsonWriter)
	}()

	// Set up compression if needed
	var finalReader io.Reader = jsonReader

	if options.Compression != "none" && options.Compression != "" {
		// Create a pipe for compression
		compressedReader, compressor := io.Pipe()

		// Start compression in a goroutine
		go func() {
			defer compressor.Close()
			err := utils.CompressDataStreaming(jsonReader, compressor, options.Compression, options.CompressionLevel)
			if err != nil {
				compressor.CloseWithError(err)
			}
		}()

		finalReader = compressedReader
	}

	// Create a hash writer to calculate checksum while writing
	hasher := sha256.New()
	multiWriter := io.MultiWriter(outFile, hasher)

	// Copy data through compression and hashing
	_, err = io.Copy(multiWriter, finalReader)
	if err != nil {
		return fmt.Errorf("failed to write patch data: %w", err)
	}

	// Close output file
	if err := outFile.Close(); err != nil {
		return fmt.Errorf("failed to close output file: %w", err)
	}

	// Check for encoding errors
	if err := <-encodeErr; err != nil {
		return fmt.Errorf("failed to encode patch: %w", err)
	}

	// Calculate checksum
	checksum := fmt.Sprintf("%x", hasher.Sum(nil))
	patch.Header.Checksum = checksum

	return nil
}

// encodePatchStreaming writes the patch as JSON in a streaming fashion to avoid memory exhaustion
func (gw *GeneratorWindow) encodePatchStreaming(patch *utils.Patch, writer io.Writer) error {
	// Create a buffered writer for better performance
	bufWriter := bufio.NewWriterSize(writer, 64*1024) // 64KB buffer
	defer bufWriter.Flush()

	// Write opening brace
	if _, err := bufWriter.WriteString("{\n"); err != nil {
		return err
	}

	// Encode header
	if err := gw.encodeField(bufWriter, "Header", patch.Header, true); err != nil {
		return err
	}

	// Encode simple fields
	if err := gw.encodeField(bufWriter, "FromVersion", patch.FromVersion, true); err != nil {
		return err
	}
	if err := gw.encodeField(bufWriter, "ToVersion", patch.ToVersion, true); err != nil {
		return err
	}
	if err := gw.encodeField(bufWriter, "FromKeyFile", patch.FromKeyFile, true); err != nil {
		return err
	}
	if err := gw.encodeField(bufWriter, "ToKeyFile", patch.ToKeyFile, true); err != nil {
		return err
	}
	if err := gw.encodeField(bufWriter, "RequiredFiles", patch.RequiredFiles, true); err != nil {
		return err
	}
	if err := gw.encodeField(bufWriter, "SimpleMode", patch.SimpleMode, true); err != nil {
		return err
	}

	// Encode operations array manually to stream large data
	if _, err := bufWriter.WriteString(`  "Operations": [`); err != nil {
		return err
	}

	for i, op := range patch.Operations {
		if i > 0 {
			if _, err := bufWriter.WriteString(",\n"); err != nil {
				return err
			}
		} else {
			if _, err := bufWriter.WriteString("\n"); err != nil {
				return err
			}
		}

		if err := gw.encodeOperation(bufWriter, op); err != nil {
			return err
		}
	}

	if _, err := bufWriter.WriteString("\n  ]\n"); err != nil {
		return err
	}

	// Write closing brace
	if _, err := bufWriter.WriteString("}\n"); err != nil {
		return err
	}

	return bufWriter.Flush()
}

// encodeField encodes a single field with proper JSON formatting
func (gw *GeneratorWindow) encodeField(writer io.Writer, name string, value interface{}, addComma bool) error {
	commaStr := ","
	if !addComma {
		commaStr = ""
	}

	// For simple types, use JSON encoding
	data, err := json.MarshalIndent(value, "  ", "  ")
	if err != nil {
		return err
	}

	// Write field name and value
	fieldStr := fmt.Sprintf("  \"%s\": ", name)
	if _, err := writer.Write([]byte(fieldStr)); err != nil {
		return err
	}

	// Write the JSON data, but indent it properly
	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		if i > 0 {
			if _, err := writer.Write([]byte("\n")); err != nil {
				return err
			}
		}
		if _, err := writer.Write([]byte(line)); err != nil {
			return err
		}
	}

	if _, err := writer.Write([]byte(commaStr + "\n")); err != nil {
		return err
	}

	return nil
}

// encodeOperation encodes a single patch operation with streaming for large binary data
func (gw *GeneratorWindow) encodeOperation(writer io.Writer, op utils.PatchOperation) error {
	// Write operation opening
	if _, err := writer.Write([]byte("    {\n")); err != nil {
		return err
	}

	// Encode simple fields
	if err := gw.encodeOperationField(writer, "Type", int(op.Type), true); err != nil {
		return err
	}
	if err := gw.encodeOperationField(writer, "FilePath", op.FilePath, true); err != nil {
		return err
	}
	if err := gw.encodeOperationField(writer, "BinaryDiff", op.BinaryDiff, true); err != nil {
		return err
	}

	// Encode NewFile data - this is the large binary data that needs streaming
	if err := gw.encodeOperationField(writer, "NewFile", op.NewFile, true); err != nil {
		return err
	}

	if err := gw.encodeOperationField(writer, "OldChecksum", op.OldChecksum, true); err != nil {
		return err
	}
	if err := gw.encodeOperationField(writer, "NewChecksum", op.NewChecksum, true); err != nil {
		return err
	}
	if err := gw.encodeOperationField(writer, "Size", op.Size, false); err != nil {
		return err
	}

	// Write operation closing
	if _, err := writer.Write([]byte("\n    }")); err != nil {
		return err
	}

	return nil
}

// encodeOperationField encodes a single field within an operation
func (gw *GeneratorWindow) encodeOperationField(writer io.Writer, name string, value interface{}, addComma bool) error {
	commaStr := ","
	if !addComma {
		commaStr = ""
	}

	fieldStr := fmt.Sprintf("      \"%s\": ", name)
	if _, err := writer.Write([]byte(fieldStr)); err != nil {
		return err
	}

	// For byte slices (binary data), encode as base64 using streaming encoder
	if byteData, ok := value.([]byte); ok {
		// Write opening quote
		if _, err := writer.Write([]byte("\"")); err != nil {
			return err
		}

		// Create base64 encoder that writes directly to output
		encoder := base64.NewEncoder(base64.StdEncoding, writer)

		// Write data in chunks to avoid memory exhaustion
		const chunkSize = 64 * 1024 // 64KB chunks
		for i := 0; i < len(byteData); i += chunkSize {
			end := i + chunkSize
			if end > len(byteData) {
				end = len(byteData)
			}
			if _, err := encoder.Write(byteData[i:end]); err != nil {
				encoder.Close()
				return err
			}
		}

		// Close encoder to flush any remaining data
		if err := encoder.Close(); err != nil {
			return err
		}

		// Write closing quote and comma
		if _, err := writer.Write([]byte(fmt.Sprintf("\"%s", commaStr))); err != nil {
			return err
		}
	} else {
		// For other types, use JSON encoding
		data, err := json.Marshal(value)
		if err != nil {
			return err
		}
		jsonStr := string(data) + commaStr
		if _, err := writer.Write([]byte(jsonStr)); err != nil {
			return err
		}
	}

	if _, err := writer.Write([]byte("\n")); err != nil {
		return err
	}

	return nil
}

// createStandaloneExe creates a self-contained executable by appending patch data to the chosen applier
func (gw *GeneratorWindow) createStandaloneExe(patchPath, exePath, exeType string) error {
	// Get path to the appropriate applier (same directory as patch-gen-gui.exe)
	genExe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Choose applier based on exe type
	var applierPath string
	if exeType == "console" {
		applierPath = filepath.Join(filepath.Dir(genExe), "patch-apply.exe")
	} else {
		applierPath = filepath.Join(filepath.Dir(genExe), "patch-apply-gui.exe")
	}

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
