package gui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// BatchScriptWindow represents the batch script generator UI
type BatchScriptWindow struct {
	widget.BaseWidget

	window fyne.Window

	// UI Components
	patchFileEntry    *widget.Entry
	targetDirEntry    *widget.Entry
	scriptNameEntry   *widget.Entry
	customKeyEntry    *widget.Entry
	dryRunCheck       *widget.Check
	silentModeCheck   *widget.Check
	verifyCheck       *widget.Check
	backupCheck       *widget.Check
	autoCloseCheck    *widget.Check
	showProgressCheck *widget.Check
	customInstrEntry  *widget.Entry
	previewText       *widget.Entry
	generateBtn       *widget.Button
	statusLabel       *widget.Label

	// Data
	patchFile     string
	targetDir     string
	scriptName    string
	customKeyFile string
	includeDryRun bool
	silentMode    bool
	disableVerify bool
	disableBackup bool
	autoClose     bool
	showProgress  bool
	customInstr   string
}

// NewBatchScriptWindow creates a new batch script generator window
func NewBatchScriptWindow() *BatchScriptWindow {
	bsw := &BatchScriptWindow{
		scriptName:    "apply_patch.bat",
		includeDryRun: false,
		silentMode:    false,
		disableVerify: false,
		disableBackup: false,
		autoClose:     true,
		showProgress:  true,
	}
	bsw.ExtendBaseWidget(bsw)
	return bsw
}

// CreateRenderer creates the renderer for the batch script window
func (bsw *BatchScriptWindow) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(bsw.buildUI())
}

// SetWindow sets the parent window (needed for dialogs)
func (bsw *BatchScriptWindow) SetWindow(window fyne.Window) {
	bsw.window = window
}

// buildUI builds the complete UI layout
func (bsw *BatchScriptWindow) buildUI() fyne.CanvasObject {
	// Instructions label (compact)
	instructionsLabel := widget.NewLabel(
		"Generate batch scripts for end users. Choose options including silent mode for automated deployment.",
	)
	instructionsLabel.Wrapping = fyne.TextWrapWord

	// Top row: Patch file and Target directory side by side
	bsw.patchFileEntry = widget.NewEntry()
	bsw.patchFileEntry.SetPlaceHolder("Select patch file...")
	bsw.patchFileEntry.OnChanged = func(text string) {
		bsw.patchFile = text
		bsw.updatePreview()
		bsw.updateGenerateButton()
	}

	patchFileBrowse := widget.NewButton("Browse", func() {
		bsw.selectPatchFile()
	})

	patchFileContainer := container.NewBorder(
		nil, nil,
		widget.NewLabel("Patch File:"),
		patchFileBrowse,
		bsw.patchFileEntry,
	)

	bsw.targetDirEntry = widget.NewEntry()
	bsw.targetDirEntry.SetPlaceHolder("Leave empty for current directory")
	bsw.targetDirEntry.OnChanged = func(text string) {
		bsw.targetDir = text
		bsw.updatePreview()
	}

	targetDirBrowse := widget.NewButton("Browse", func() {
		bsw.selectTargetDirectory()
	})

	targetDirContainer := container.NewBorder(
		nil, nil,
		widget.NewLabel("Target Dir:"),
		targetDirBrowse,
		bsw.targetDirEntry,
	)

	topRow := container.NewGridWithColumns(2, patchFileContainer, targetDirContainer)

	// Second row: Script name and Custom key file side by side
	bsw.scriptNameEntry = widget.NewEntry()
	bsw.scriptNameEntry.SetText("apply_patch.bat")
	bsw.scriptNameEntry.OnChanged = func(text string) {
		bsw.scriptName = text
	}

	scriptNameContainer := container.NewBorder(
		nil, nil,
		widget.NewLabel("Script Name:"),
		nil,
		bsw.scriptNameEntry,
	)

	bsw.customKeyEntry = widget.NewEntry()
	bsw.customKeyEntry.SetPlaceHolder("Optional: Custom key file (e.g., MyApp.exe)")
	bsw.customKeyEntry.OnChanged = func(text string) {
		bsw.customKeyFile = text
		bsw.updatePreview()
	}

	customKeyContainer := container.NewBorder(
		nil, nil,
		widget.NewLabel("Key File:"),
		nil,
		bsw.customKeyEntry,
	)

	secondRow := container.NewGridWithColumns(2, scriptNameContainer, customKeyContainer)

	// Options in a more compact grid layout (2 columns)
	bsw.dryRunCheck = widget.NewCheck("Dry-run mode", func(checked bool) {
		bsw.includeDryRun = checked
		bsw.updatePreview()
	})

	bsw.silentModeCheck = widget.NewCheck("Silent mode", func(checked bool) {
		bsw.silentMode = checked
		bsw.updatePreview()
	})

	bsw.verifyCheck = widget.NewCheck("Disable verify", func(checked bool) {
		bsw.disableVerify = checked
		bsw.updatePreview()
	})

	bsw.backupCheck = widget.NewCheck("Disable backup", func(checked bool) {
		bsw.disableBackup = checked
		bsw.updatePreview()
	})

	bsw.autoCloseCheck = widget.NewCheck("Auto-close success", func(checked bool) {
		bsw.autoClose = checked
		bsw.updatePreview()
	})

	bsw.showProgressCheck = widget.NewCheck("Show progress", func(checked bool) {
		bsw.showProgress = checked
		bsw.updatePreview()
	})

	// Group options in 2 columns
	leftOptions := container.NewVBox(
		bsw.dryRunCheck,
		bsw.silentModeCheck,
		bsw.verifyCheck,
	)
	rightOptions := container.NewVBox(
		bsw.backupCheck,
		bsw.autoCloseCheck,
		bsw.showProgressCheck,
	)
	optionsRow := container.NewGridWithColumns(2, leftOptions, rightOptions)

	// Custom instructions (compact)
	bsw.customInstrEntry = widget.NewMultiLineEntry()
	bsw.customInstrEntry.SetPlaceHolder("Optional: Custom instructions for end users...")
	bsw.customInstrEntry.OnChanged = func(text string) {
		bsw.customInstr = text
		bsw.updatePreview()
	}

	customInstrContainer := container.NewBorder(
		widget.NewLabel("Instructions:"),
		nil, nil, nil,
		container.NewVScroll(bsw.customInstrEntry),
	)

	// Preview text area (smaller, more compact)
	bsw.previewText = widget.NewMultiLineEntry()
	bsw.previewText.SetPlaceHolder("Batch script preview...")
	bsw.previewText.Wrapping = fyne.TextWrapOff

	previewScroll := container.NewVScroll(bsw.previewText)
	previewScroll.SetMinSize(fyne.NewSize(0, 200)) // Reduced from 300 to 200

	previewContainer := container.NewBorder(
		widget.NewLabel("Preview:"),
		nil, nil, nil,
		previewScroll,
	)

	// Set default values after all UI components are created
	bsw.autoCloseCheck.SetChecked(true)
	bsw.showProgressCheck.SetChecked(true)

	// Bottom row: Status and generate button
	bsw.generateBtn = widget.NewButton("Generate Script", func() {
		bsw.generateBatchScript()
	})
	bsw.generateBtn.Importance = widget.HighImportance
	bsw.generateBtn.Disable()

	bsw.statusLabel = widget.NewLabel("Ready")
	bsw.statusLabel.Alignment = fyne.TextAlignLeading

	bottomRow := container.NewBorder(
		nil, nil,
		bsw.statusLabel,
		bsw.generateBtn,
		nil,
	)

	// Assemble the UI with better spacing
	return container.NewVBox(
		instructionsLabel,
		widget.NewSeparator(),
		topRow,
		secondRow,
		widget.NewSeparator(),
		container.NewBorder(widget.NewLabel("Options:"), nil, nil, nil, optionsRow),
		widget.NewSeparator(),
		customInstrContainer,
		widget.NewSeparator(),
		previewContainer,
		bottomRow,
	)
}

// selectPatchFile opens a file dialog for selecting patch file
func (bsw *BatchScriptWindow) selectPatchFile() {
	if bsw.window == nil {
		return
	}

	dialog.ShowFileOpen(func(file fyne.URIReadCloser, err error) {
		if err == nil && file != nil {
			path := file.URI().Path()
			bsw.patchFileEntry.SetText(path)
			bsw.patchFile = path
			bsw.updatePreview()
			bsw.updateGenerateButton()
		}
	}, bsw.window)
}

// selectTargetDirectory opens a folder dialog for selecting target directory
func (bsw *BatchScriptWindow) selectTargetDirectory() {
	if bsw.window == nil {
		return
	}

	dialog.ShowFolderOpen(func(dir fyne.ListableURI, err error) {
		if err == nil && dir != nil {
			path := dir.Path()
			bsw.targetDirEntry.SetText(path)
			bsw.targetDir = path
			bsw.updatePreview()
		}
	}, bsw.window)
}

// updatePreview updates the batch script preview
func (bsw *BatchScriptWindow) updatePreview() {
	if bsw.previewText == nil {
		return // UI not fully initialized yet
	}

	if bsw.patchFile == "" {
		bsw.previewText.SetText("")
		return
	}

	script := bsw.generateBatchScriptContent()
	bsw.previewText.SetText(script)
}

// updateGenerateButton enables/disables generate button
func (bsw *BatchScriptWindow) updateGenerateButton() {
	if bsw.patchFile != "" && bsw.scriptName != "" {
		bsw.generateBtn.Enable()
	} else {
		bsw.generateBtn.Disable()
	}
}

// generateBatchScriptContent generates the batch script content
func (bsw *BatchScriptWindow) generateBatchScriptContent() string {
	var sb strings.Builder

	// Header
	sb.WriteString("@echo off\n")
	sb.WriteString("REM CyberPatchMaker Patch Application Script\n")
	sb.WriteString("REM Generated by CyberPatchMaker GUI\n")
	sb.WriteString("REM\n\n")

	// Custom instructions
	if bsw.customInstr != "" {
		lines := strings.Split(bsw.customInstr, "\n")
		for _, line := range lines {
			sb.WriteString(fmt.Sprintf("REM %s\n", line))
		}
		sb.WriteString("REM\n\n")
	}

	// Title
	sb.WriteString("title CyberPatchMaker - Applying Patch\n\n")

	// Instructions
	sb.WriteString("echo ========================================\n")
	sb.WriteString("echo CyberPatchMaker - Patch Application\n")
	sb.WriteString("echo ========================================\n")
	sb.WriteString("echo.\n")

	if bsw.customInstr != "" {
		lines := strings.Split(bsw.customInstr, "\n")
		for _, line := range lines {
			if line != "" {
				sb.WriteString(fmt.Sprintf("echo %s\n", line))
			}
		}
		sb.WriteString("echo.\n")
	}

	if bsw.silentMode {
		sb.WriteString("echo SILENT MODE: This will apply the patch automatically without prompts.\n")
		sb.WriteString("echo.\n")
	} else if bsw.includeDryRun {
		sb.WriteString("echo DRY RUN MODE: This will test the patch without making changes.\n")
		sb.WriteString("echo.\n")
	} else {
		sb.WriteString("echo This script will apply the patch to your installation.\n")
		sb.WriteString("echo.\n")
	}

	if !bsw.silentMode {
		sb.WriteString("pause\n\n")
	}

	// Get patch filename
	patchFileName := filepath.Base(bsw.patchFile)

	// Build applier command
	sb.WriteString("echo.\n")
	if bsw.showProgress {
		sb.WriteString("echo Applying patch...\n")
		sb.WriteString("echo.\n\n")
	}

	// Determine target directory
	targetDir := ""
	if bsw.targetDir != "" {
		targetDir = bsw.targetDir
	} else {
		targetDir = "%~dp0" // Use script directory
	}

	// Build command with options
	command := fmt.Sprintf("applier.exe --patch \"%s\" --current-dir \"%s\"", patchFileName, targetDir)

	if bsw.customKeyFile != "" {
		command += fmt.Sprintf(" --key-file \"%s\"", bsw.customKeyFile)
	}

	if bsw.includeDryRun {
		command += " --dry-run"
	}

	if bsw.silentMode {
		command += " --silent"
	}

	if bsw.disableVerify {
		command += " --verify=false"
	}

	if bsw.disableBackup {
		command += " --backup=false"
	}

	sb.WriteString(command + "\n\n")

	// Check result
	sb.WriteString("if %ERRORLEVEL% EQU 0 (\n")

	if bsw.includeDryRun {
		sb.WriteString("    echo.\n")
		sb.WriteString("    echo ========================================\n")
		sb.WriteString("    echo DRY RUN SUCCESSFUL!\n")
		sb.WriteString("    echo ========================================\n")
		sb.WriteString("    echo.\n")
		sb.WriteString("    echo The patch can be applied safely.\n")
		sb.WriteString("    echo Remove --dry-run flag to apply for real.\n")
	} else {
		sb.WriteString("    echo.\n")
		sb.WriteString("    echo ========================================\n")
		sb.WriteString("    echo PATCH APPLIED SUCCESSFULLY!\n")
		sb.WriteString("    echo ========================================\n")
		sb.WriteString("    echo.\n")
		sb.WriteString("    echo Your installation has been updated.\n")
	}

	if bsw.showProgress {
		sb.WriteString("    echo.\n")
		sb.WriteString("    echo Press any key to continue...\n")
	}

	sb.WriteString("    echo.\n")
	sb.WriteString(") else (\n")
	sb.WriteString("    echo.\n")
	sb.WriteString("    echo ========================================\n")
	sb.WriteString("    echo PATCH APPLICATION FAILED!\n")
	sb.WriteString("    echo ========================================\n")
	sb.WriteString("    echo.\n")
	sb.WriteString("    echo Error code: %ERRORLEVEL%\n")
	sb.WriteString("    echo.\n")
	sb.WriteString("    echo Please check the error messages above.\n")

	if !bsw.disableBackup {
		sb.WriteString("    echo If backup was enabled, you can restore from the .backup folder.\n")
	}

	sb.WriteString("    echo.\n")
	sb.WriteString(")\n\n")

	if bsw.autoClose && !bsw.silentMode {
		sb.WriteString("echo.\n")
		sb.WriteString("pause\n")
	} else if !bsw.silentMode {
		sb.WriteString("echo.\n")
		sb.WriteString("pause\n")
	}

	return sb.String()
}

// generateBatchScript generates and saves the batch script
func (bsw *BatchScriptWindow) generateBatchScript() {
	bsw.setStatus("Generating batch script...")
	bsw.generateBtn.Disable()

	// Validate patch file exists
	if _, err := os.Stat(bsw.patchFile); os.IsNotExist(err) {
		bsw.setStatus("Error: Patch file does not exist")
		if bsw.window != nil {
			dialog.ShowError(fmt.Errorf("patch file does not exist: %s", bsw.patchFile), bsw.window)
		}
		bsw.generateBtn.Enable()
		return
	}

	// Determine output path (same directory as patch file)
	patchDir := filepath.Dir(bsw.patchFile)
	outputPath := filepath.Join(patchDir, bsw.scriptName)

	// Generate script content
	scriptContent := bsw.generateBatchScriptContent()

	// Write to file
	err := os.WriteFile(outputPath, []byte(scriptContent), 0644)
	if err != nil {
		bsw.setStatus("Error: Failed to save batch script")
		if bsw.window != nil {
			dialog.ShowError(fmt.Errorf("failed to save batch script: %w", err), bsw.window)
		}
		bsw.generateBtn.Enable()
		return
	}

	bsw.setStatus("Batch script generated successfully!")
	bsw.generateBtn.Enable()

	// Show success dialog
	if bsw.window != nil {
		dialog.ShowInformation("Success",
			fmt.Sprintf("Batch script generated successfully!\n\nSaved to: %s\n\nEnd users can double-click this file to apply the patch.", outputPath),
			bsw.window)
	}
}

// setStatus updates the status label
func (bsw *BatchScriptWindow) setStatus(status string) {
	bsw.statusLabel.SetText(status)
}
