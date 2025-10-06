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
	patchFileEntry   *widget.Entry
	targetDirEntry   *widget.Entry
	scriptNameEntry  *widget.Entry
	dryRunCheck      *widget.Check
	verifyCheck      *widget.Check
	backupCheck      *widget.Check
	customInstrEntry *widget.Entry
	previewText      *widget.Entry
	generateBtn      *widget.Button
	statusLabel      *widget.Label

	// Data
	patchFile     string
	targetDir     string
	scriptName    string
	includeDryRun bool
	disableVerify bool
	disableBackup bool
	customInstr   string
}

// NewBatchScriptWindow creates a new batch script generator window
func NewBatchScriptWindow() *BatchScriptWindow {
	bsw := &BatchScriptWindow{
		scriptName:    "apply_patch.bat",
		includeDryRun: false,
		disableVerify: false,
		disableBackup: false,
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
	// Patch file selector
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

	// Target directory selector
	bsw.targetDirEntry = widget.NewEntry()
	bsw.targetDirEntry.SetPlaceHolder("Target directory (leave empty for current directory)")
	bsw.targetDirEntry.OnChanged = func(text string) {
		bsw.targetDir = text
		bsw.updatePreview()
	}

	targetDirBrowse := widget.NewButton("Browse", func() {
		bsw.selectTargetDirectory()
	})

	targetDirContainer := container.NewBorder(
		nil, nil,
		widget.NewLabel("Target Directory:"),
		targetDirBrowse,
		bsw.targetDirEntry,
	)

	// Script name entry
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

	// Options checkboxes
	bsw.dryRunCheck = widget.NewCheck("Include dry-run option (safe test run)", func(checked bool) {
		bsw.includeDryRun = checked
		bsw.updatePreview()
	})

	bsw.verifyCheck = widget.NewCheck("Disable verification (--verify=false)", func(checked bool) {
		bsw.disableVerify = checked
		bsw.updatePreview()
	})

	bsw.backupCheck = widget.NewCheck("Disable backup (--backup=false)", func(checked bool) {
		bsw.disableBackup = checked
		bsw.updatePreview()
	})

	// Custom instructions
	bsw.customInstrEntry = widget.NewMultiLineEntry()
	bsw.customInstrEntry.SetPlaceHolder("Optional: Add custom instructions for end users...")
	bsw.customInstrEntry.OnChanged = func(text string) {
		bsw.customInstr = text
		bsw.updatePreview()
	}

	customInstrContainer := container.NewBorder(
		widget.NewLabel("Custom Instructions:"),
		nil, nil, nil,
		bsw.customInstrEntry,
	)

	// Preview text area
	bsw.previewText = widget.NewMultiLineEntry()
	bsw.previewText.SetPlaceHolder("Batch script preview will appear here...")
	bsw.previewText.Wrapping = fyne.TextWrapOff
	// Keep it visible but read-only by preventing text changes

	previewScroll := container.NewVScroll(bsw.previewText)
	previewScroll.SetMinSize(fyne.NewSize(0, 300)) // Set minimum height for better visibility

	previewContainer := container.NewBorder(
		widget.NewLabel("Preview:"),
		nil, nil, nil,
		previewScroll,
	)

	// Generate button
	bsw.generateBtn = widget.NewButton("Generate Batch Script", func() {
		bsw.generateBatchScript()
	})
	bsw.generateBtn.Importance = widget.HighImportance
	bsw.generateBtn.Disable()

	// Status label
	bsw.statusLabel = widget.NewLabel("Ready")
	bsw.statusLabel.Alignment = fyne.TextAlignLeading

	// Instructions label
	instructionsLabel := widget.NewLabel(
		"This tool generates a Windows batch script (.bat) that end users can double-click to apply the patch.\n" +
			"The script will be saved in the same directory as the patch file.",
	)
	instructionsLabel.Wrapping = fyne.TextWrapWord

	// Assemble the UI
	return container.NewVBox(
		instructionsLabel,
		widget.NewSeparator(),
		patchFileContainer,
		targetDirContainer,
		scriptNameContainer,
		widget.NewSeparator(),
		widget.NewLabel("Options:"),
		bsw.dryRunCheck,
		bsw.verifyCheck,
		bsw.backupCheck,
		widget.NewSeparator(),
		customInstrContainer,
		widget.NewSeparator(),
		previewContainer,
		container.NewHBox(
			bsw.statusLabel,
			widget.NewSeparator(),
			bsw.generateBtn,
		),
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
	sb.WriteString("echo This script will apply the patch to your installation.\n")
	if bsw.includeDryRun {
		sb.WriteString("echo.\n")
		sb.WriteString("echo DRY RUN MODE: The script will test the patch without making changes.\n")
	}
	sb.WriteString("echo.\n")
	sb.WriteString("pause\n\n")

	// Get patch filename
	patchFileName := filepath.Base(bsw.patchFile)

	// Build applier command
	sb.WriteString("echo.\n")
	sb.WriteString("echo Applying patch...\n")
	sb.WriteString("echo.\n\n")

	// Determine target directory
	targetDir := ""
	if bsw.targetDir != "" {
		targetDir = bsw.targetDir
	} else {
		targetDir = "%~dp0" // Use script directory
	}

	// Build command with options
	command := fmt.Sprintf("applier.exe --patch \"%s\" --current-dir \"%s\"", patchFileName, targetDir)

	if bsw.includeDryRun {
		command += " --dry-run"
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
	sb.WriteString("    echo If verification failed, your installation may have been modified.\n")
	sb.WriteString("    echo If backup was enabled, you can restore from the .backup folder.\n")
	sb.WriteString("    echo.\n")
	sb.WriteString(")\n\n")

	sb.WriteString("echo.\n")
	sb.WriteString("pause\n")

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
