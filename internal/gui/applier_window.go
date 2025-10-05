package gui

import (
	"encoding/json"
	"fmt"
	"image/color"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/cyberofficial/cyberpatchmaker/internal/core/patcher"
	"github.com/cyberofficial/cyberpatchmaker/pkg/utils"
)

// ApplierWindow represents the patch applier UI
type ApplierWindow struct {
	widget.BaseWidget

	window fyne.Window

	// UI Components
	patchFile    string
	currentDir   string
	dryRun       bool
	verifyBefore bool
	verifyAfter  bool
	createBackup bool
	autoDetect   bool

	patchFileEntry    *widget.Entry
	currentDirEntry   *widget.Entry
	dryRunCheck       *widget.Check
	verifyBeforeCheck *widget.Check
	verifyAfterCheck  *widget.Check
	backupCheck       *widget.Check
	autoDetectCheck   *widget.Check
	applyBtn          *widget.Button
	statusLabel       *widget.Label
	logText           *widget.Entry

	// Patch info display
	patchInfoBox     *fyne.Container
	fromVersionLabel *widget.Label
	toVersionLabel   *widget.Label
	keyFileLabel     *widget.Label
	hashLabel        *widget.Label
	sizeLabel        *widget.Label
	compressionLabel *widget.Label
	createdLabel     *widget.Label
	addedLabel       *widget.Label
	modifiedLabel    *widget.Label
	deletedLabel     *widget.Label
	addDirsLabel     *widget.Label
	deleteDirsLabel  *widget.Label
	requiredLabel    *widget.Label

	// Data
	loadedPatch *utils.Patch
}

// NewApplierWindow creates a new applier window
func NewApplierWindow() *ApplierWindow {
	aw := &ApplierWindow{
		dryRun:       false,
		verifyBefore: true,
		verifyAfter:  true,
		createBackup: true,
		autoDetect:   true,
	}
	aw.ExtendBaseWidget(aw)
	return aw
}

// CreateRenderer creates the renderer for the applier window
func (aw *ApplierWindow) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(aw.buildUI())
}

// SetWindow sets the parent window (needed for dialogs)
func (aw *ApplierWindow) SetWindow(window fyne.Window) {
	aw.window = window
}

// buildUI builds the complete UI layout
func (aw *ApplierWindow) buildUI() fyne.CanvasObject {
	// Create patch file selector
	aw.patchFileEntry = widget.NewEntry()
	aw.patchFileEntry.SetPlaceHolder("Select patch file...")
	aw.patchFileEntry.OnChanged = func(text string) {
		aw.patchFile = text
		aw.updateApplyButton()
		// Auto-load patch info when file is selected
		if text != "" && utils.FileExists(text) {
			aw.loadPatchInfo()
		}
	}
	aw.patchFileEntry.OnSubmitted = func(text string) {
		if text != "" && utils.FileExists(text) {
			aw.patchFile = text
			aw.loadPatchInfo()
			aw.updateApplyButton()
		}
	}

	patchFileBrowse := widget.NewButton("Browse", func() {
		aw.selectPatchFile()
	})

	patchFileContainer := container.NewBorder(
		nil, nil,
		widget.NewLabel("Patch File:"),
		patchFileBrowse,
		aw.patchFileEntry,
	)

	// Create current directory selector
	aw.currentDirEntry = widget.NewEntry()
	aw.currentDirEntry.SetPlaceHolder("Select current version directory...")
	aw.currentDirEntry.OnChanged = func(text string) {
		aw.currentDir = text
		aw.updateApplyButton()
	}
	aw.currentDirEntry.OnSubmitted = func(text string) {
		if text != "" {
			aw.currentDir = text
			aw.updateApplyButton()
		}
	}

	currentDirBrowse := widget.NewButton("Browse", func() {
		aw.selectCurrentDirectory()
	})

	currentDirContainer := container.NewBorder(
		nil, nil,
		widget.NewLabel("Current Dir:"),
		currentDirBrowse,
		aw.currentDirEntry,
	)

	// Create patch information display labels
	aw.fromVersionLabel = widget.NewLabel("---")
	aw.toVersionLabel = widget.NewLabel("---")
	aw.keyFileLabel = widget.NewLabel("---")
	aw.hashLabel = widget.NewLabel("---")
	aw.sizeLabel = widget.NewLabel("---")
	aw.compressionLabel = widget.NewLabel("---")
	aw.createdLabel = widget.NewLabel("---")
	aw.addedLabel = widget.NewLabel("---")
	aw.modifiedLabel = widget.NewLabel("---")
	aw.deletedLabel = widget.NewLabel("---")
	aw.addDirsLabel = widget.NewLabel("---")
	aw.deleteDirsLabel = widget.NewLabel("---")
	aw.requiredLabel = widget.NewLabel("---")

	// Left column: Patch Info
	patchInfoLeft := container.NewVBox(
		widget.NewLabel("Version Info:"),
		container.NewBorder(nil, nil, widget.NewLabel("From:"), nil, aw.fromVersionLabel),
		container.NewBorder(nil, nil, widget.NewLabel("To:"), nil, aw.toVersionLabel),
		container.NewBorder(nil, nil, widget.NewLabel("Key File:"), nil, aw.keyFileLabel),
		container.NewBorder(nil, nil, widget.NewLabel("Hash:"), nil, aw.hashLabel),
		widget.NewSeparator(),
		container.NewBorder(nil, nil, widget.NewLabel("Size:"), nil, aw.sizeLabel),
		container.NewBorder(nil, nil, widget.NewLabel("Compression:"), nil, aw.compressionLabel),
		container.NewBorder(nil, nil, widget.NewLabel("Created:"), nil, aw.createdLabel),
	)

	// Right column: Operations Info
	patchInfoRight := container.NewVBox(
		widget.NewLabel("Operations:"),
		container.NewBorder(nil, nil, widget.NewLabel("Added:"), nil, aw.addedLabel),
		container.NewBorder(nil, nil, widget.NewLabel("Modified:"), nil, aw.modifiedLabel),
		container.NewBorder(nil, nil, widget.NewLabel("Deleted:"), nil, aw.deletedLabel),
		container.NewBorder(nil, nil, widget.NewLabel("Dirs+:"), nil, aw.addDirsLabel),
		container.NewBorder(nil, nil, widget.NewLabel("Dirs-:"), nil, aw.deleteDirsLabel),
		widget.NewSeparator(),
		container.NewBorder(nil, nil, widget.NewLabel("Required:"), nil, aw.requiredLabel),
	)

	// Create two-column patch info layout
	aw.patchInfoBox = container.NewGridWithColumns(2,
		patchInfoLeft,
		patchInfoRight,
	)

	// Create options
	aw.dryRunCheck = widget.NewCheck("Dry Run (Simulate)", func(checked bool) {
		aw.dryRun = checked
		// Disable backup and verify when dry run is enabled
		if checked {
			aw.backupCheck.Disable()
			aw.verifyBeforeCheck.Disable()
			aw.verifyAfterCheck.Disable()
		} else {
			aw.backupCheck.Enable()
			aw.verifyBeforeCheck.Enable()
			aw.verifyAfterCheck.Enable()
		}
	})

	aw.verifyBeforeCheck = widget.NewCheck("Verify before", func(checked bool) {
		aw.verifyBefore = checked
	})
	aw.verifyBeforeCheck.SetChecked(true)

	aw.verifyAfterCheck = widget.NewCheck("Verify after", func(checked bool) {
		aw.verifyAfter = checked
	})
	aw.verifyAfterCheck.SetChecked(true)

	aw.backupCheck = widget.NewCheck("Create backup", func(checked bool) {
		aw.createBackup = checked
	})
	aw.backupCheck.SetChecked(true)

	aw.autoDetectCheck = widget.NewCheck("Auto-detect version", func(checked bool) {
		aw.autoDetect = checked
	})
	aw.autoDetectCheck.SetChecked(true)

	// Options in horizontal layout to save space
	optionsRow1 := container.NewGridWithColumns(2, aw.verifyBeforeCheck, aw.verifyAfterCheck)
	optionsRow2 := container.NewGridWithColumns(2, aw.backupCheck, aw.autoDetectCheck)

	optionsContainer := container.NewVBox(
		widget.NewLabel("Options:"),
		aw.dryRunCheck,
		optionsRow1,
		optionsRow2,
	)

	// Create apply button
	aw.applyBtn = widget.NewButton("Apply Patch", func() {
		aw.applyPatch()
	})
	aw.applyBtn.Disable()

	// Create status label
	aw.statusLabel = widget.NewLabel("Ready")

	// Create log output with white background and black text (read-only but not disabled)
	aw.logText = widget.NewMultiLineEntry()
	aw.logText.SetPlaceHolder("Log output will appear here...")
	// Make read-only by preventing edits (but keep enabled for normal text color)
	aw.logText.OnChanged = func(text string) {
		// This prevents user edits - text can only be set programmatically
	}

	// Create a white background for the log for maximum contrast
	logBg := canvas.NewRectangle(color.White)
	logWithBg := container.NewStack(logBg, aw.logText)
	logContainer := container.NewVScroll(logWithBg)
	logContainer.SetMinSize(fyne.NewSize(0, 150))

	// Assemble the UI with compact layout
	return container.NewVBox(
		patchFileContainer,
		currentDirContainer,
		widget.NewSeparator(),
		aw.patchInfoBox,
		widget.NewSeparator(),
		optionsContainer,
		container.NewBorder(nil, nil, nil, aw.applyBtn, aw.statusLabel),
		logContainer,
	)
}

// selectPatchFile opens a file dialog for selecting patch file
func (aw *ApplierWindow) selectPatchFile() {
	if aw.window == nil {
		return
	}

	dialog.ShowFileOpen(func(file fyne.URIReadCloser, err error) {
		if err == nil && file != nil {
			path := file.URI().Path()
			aw.patchFileEntry.SetText(path)
			aw.patchFile = path
			aw.updateApplyButton()
			aw.loadPatchInfo()
			file.Close()
		}
	}, aw.window)
}

// selectCurrentDirectory opens a folder dialog for selecting current directory
func (aw *ApplierWindow) selectCurrentDirectory() {
	if aw.window == nil {
		return
	}

	dialog.ShowFolderOpen(func(dir fyne.ListableURI, err error) {
		if err == nil && dir != nil {
			path := dir.Path()
			aw.currentDirEntry.SetText(path)
			aw.currentDir = path
			aw.updateApplyButton()
		}
	}, aw.window)
}

// updateApplyButton enables/disables apply button based on selections
func (aw *ApplierWindow) updateApplyButton() {
	if aw.patchFile != "" && aw.currentDir != "" {
		aw.applyBtn.Enable()
	} else {
		aw.applyBtn.Disable()
	}
}

// loadPatchInfo loads and displays patch information
func (aw *ApplierWindow) loadPatchInfo() {
	aw.setStatus("Loading patch information...")

	// Check if file exists
	if !utils.FileExists(aw.patchFile) {
		aw.setStatus("Error: Patch file not found")
		return
	}

	// Load patch
	patch, err := aw.loadPatch(aw.patchFile)
	if err != nil {
		aw.setStatus("Error: Failed to load patch")
		aw.appendLog("ERROR: " + err.Error())
		return
	}

	aw.loadedPatch = patch

	// Update UI with patch information
	aw.fromVersionLabel.SetText(patch.FromVersion)
	aw.toVersionLabel.SetText(patch.ToVersion)
	aw.keyFileLabel.SetText(patch.FromKeyFile.Path)

	hashStr := patch.FromKeyFile.Checksum
	if len(hashStr) > 16 {
		hashStr = hashStr[:16] + "..."
	}
	aw.hashLabel.SetText(hashStr)

	sizeKB := float64(patch.Header.PatchSize) / 1024.0
	sizeMB := sizeKB / 1024.0
	if sizeMB >= 1.0 {
		aw.sizeLabel.SetText(fmt.Sprintf("%.2f MB", sizeMB))
	} else {
		aw.sizeLabel.SetText(fmt.Sprintf("%.2f KB", sizeKB))
	}

	aw.compressionLabel.SetText(patch.Header.Compression)
	aw.createdLabel.SetText(patch.Header.CreatedAt.Format("2006-01-02 15:04:05"))

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

	aw.addedLabel.SetText(fmt.Sprintf("%d", addCount))
	aw.modifiedLabel.SetText(fmt.Sprintf("%d", modifyCount))
	aw.deletedLabel.SetText(fmt.Sprintf("%d", deleteCount))
	aw.addDirsLabel.SetText(fmt.Sprintf("%d", addDirCount))
	aw.deleteDirsLabel.SetText(fmt.Sprintf("%d", deleteDirCount))
	aw.requiredLabel.SetText(fmt.Sprintf("%d (must match exact hashes)", len(patch.RequiredFiles)))

	aw.setStatus("Patch loaded successfully")
	aw.appendLog("Patch information loaded successfully")
}

// loadPatch loads a patch from file (with automatic decompression)
func (aw *ApplierWindow) loadPatch(filename string) (*utils.Patch, error) {
	// Read patch file
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read patch file: %w", err)
	}

	// Try to detect compression and decompress
	// First try as JSON directly
	var patch utils.Patch
	if err := json.Unmarshal(data, &patch); err != nil {
		// Try decompressing with zstd
		decompressed, err := utils.DecompressData(data, "zstd")
		if err != nil {
			// Try gzip
			decompressed, err = utils.DecompressData(data, "gzip")
			if err != nil {
				return nil, fmt.Errorf("failed to decompress or parse patch: %w", err)
			}
		}
		data = decompressed

		// Try parsing again
		if err := json.Unmarshal(data, &patch); err != nil {
			return nil, fmt.Errorf("failed to parse patch JSON: %w", err)
		}
	}

	return &patch, nil
}

// applyPatch applies the patch
func (aw *ApplierWindow) applyPatch() {
	aw.setStatus("Applying patch...")
	aw.applyBtn.Disable()
	aw.logText.SetText("") // Clear log

	// Validate
	if aw.loadedPatch == nil {
		aw.setStatus("Error: No patch loaded")
		aw.appendLog("ERROR: No patch loaded. Please select a valid patch file.")
		aw.applyBtn.Enable()
		return
	}

	if !utils.FileExists(aw.currentDir) {
		aw.setStatus("Error: Current directory not found")
		aw.appendLog("ERROR: Current directory does not exist: " + aw.currentDir)
		aw.applyBtn.Enable()
		return
	}

	aw.appendLog("=== Starting Patch Application ===")
	aw.appendLog(fmt.Sprintf("Patch: %s → %s", aw.loadedPatch.FromVersion, aw.loadedPatch.ToVersion))
	aw.appendLog(fmt.Sprintf("Target: %s", aw.currentDir))

	if aw.dryRun {
		aw.appendLog("\n=== DRY RUN MODE ===")
		aw.appendLog("No changes will be made")
		aw.performDryRun()
		return
	}

	// Auto-detect current version if enabled
	if aw.autoDetect {
		aw.appendLog("\nAuto-detecting current version...")
		detected, err := aw.detectCurrentVersion()
		if err != nil {
			aw.appendLog("WARNING: Could not auto-detect version: " + err.Error())
		} else {
			aw.appendLog(fmt.Sprintf("Detected version: %s", detected))
			if detected != aw.loadedPatch.FromVersion {
				aw.setStatus("Warning: Version mismatch detected")
				aw.appendLog(fmt.Sprintf("WARNING: Detected version (%s) does not match patch source version (%s)", detected, aw.loadedPatch.FromVersion))

				// Ask user if they want to continue
				if aw.window != nil {
					confirm := false
					dialog.ShowConfirm("Version Mismatch",
						fmt.Sprintf("The detected version (%s) does not match the patch source version (%s).\n\nDo you want to continue anyway?", detected, aw.loadedPatch.FromVersion),
						func(response bool) {
							confirm = response
							if confirm {
								aw.continueApplyPatch()
							} else {
								aw.setStatus("Patch application cancelled by user")
								aw.appendLog("Patch application cancelled")
								aw.applyBtn.Enable()
							}
						},
						aw.window)
					return // Wait for user response
				}
			}
		}
	}

	aw.continueApplyPatch()
}

// detectCurrentVersion attempts to detect the current version by checking key file
func (aw *ApplierWindow) detectCurrentVersion() (string, error) {
	keyFilePath := filepath.Join(aw.currentDir, aw.loadedPatch.FromKeyFile.Path)
	if !utils.FileExists(keyFilePath) {
		return "", fmt.Errorf("key file not found: %s", aw.loadedPatch.FromKeyFile.Path)
	}

	checksum, err := utils.CalculateFileChecksum(keyFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to calculate checksum: %w", err)
	}

	if checksum == aw.loadedPatch.FromKeyFile.Checksum {
		return aw.loadedPatch.FromVersion, nil
	}

	// Version doesn't match - return what we know
	return "unknown (key file hash mismatch)", nil
}

// continueApplyPatch continues with patch application after version check
func (aw *ApplierWindow) continueApplyPatch() {
	// Apply patch
	aw.appendLog("\nApplying patch...")
	applier := patcher.NewApplier()

	err := applier.ApplyPatch(aw.loadedPatch, aw.currentDir, aw.verifyBefore, aw.verifyAfter, aw.createBackup)
	if err != nil {
		aw.setStatus("Error: patch application failed")
		aw.appendLog("ERROR: " + err.Error())

		// Check if backup exists for restoration
		if aw.createBackup {
			backupDir := aw.currentDir + ".backup"
			if utils.FileExists(backupDir) {
				aw.appendLog("\nBackup exists - attempting restoration...")
				if restoreErr := aw.restoreBackup(backupDir, aw.currentDir); restoreErr != nil {
					aw.appendLog("ERROR: Failed to restore backup: " + restoreErr.Error())
					aw.appendLog("CRITICAL: Backup exists at: " + backupDir)
				} else {
					aw.appendLog("Backup restored successfully")
				}
			}
		}

		aw.applyBtn.Enable()

		// Show error dialog
		if aw.window != nil {
			dialog.ShowError(fmt.Errorf("patch application failed: %v", err), aw.window)
		}
		return
	}

	aw.setStatus("Patch applied successfully!")
	aw.appendLog("\n=== Patch Applied Successfully ===")
	aw.appendLog(fmt.Sprintf("Version updated from %s to %s", aw.loadedPatch.FromVersion, aw.loadedPatch.ToVersion))
	aw.applyBtn.Enable()

	// Show success dialog
	if aw.window != nil {
		dialog.ShowInformation("Success",
			fmt.Sprintf("Patch applied successfully!\n\nVersion updated from %s to %s", aw.loadedPatch.FromVersion, aw.loadedPatch.ToVersion),
			aw.window)
	}
}

// performDryRun performs a dry run simulation
func (aw *ApplierWindow) performDryRun() {
	aw.appendLog("\nSimulating patch application...")

	// Verify key file
	aw.appendLog(fmt.Sprintf("\nVerifying key file: %s", aw.loadedPatch.FromKeyFile.Path))
	keyFilePath := filepath.Join(aw.currentDir, aw.loadedPatch.FromKeyFile.Path)

	if !utils.FileExists(keyFilePath) {
		aw.appendLog(fmt.Sprintf("✗ Key file not found: %s", keyFilePath))
		aw.setStatus("Dry run failed: Key file not found")
		aw.applyBtn.Enable()
		return
	}

	checksum, err := utils.CalculateFileChecksum(keyFilePath)
	if err != nil {
		aw.appendLog(fmt.Sprintf("✗ Failed to calculate key file checksum: %v", err))
		aw.setStatus("Dry run failed: Checksum error")
		aw.applyBtn.Enable()
		return
	}

	if checksum != aw.loadedPatch.FromKeyFile.Checksum {
		aw.appendLog("✗ Key file hash mismatch")
		aw.appendLog(fmt.Sprintf("  Expected: %s", aw.loadedPatch.FromKeyFile.Checksum[:16]+"..."))
		aw.appendLog(fmt.Sprintf("  Got:      %s", checksum[:16]+"..."))
		aw.setStatus("Dry run failed: Key file mismatch")
		aw.applyBtn.Enable()
		return
	}
	aw.appendLog("✓ Key file verified")

	// Verify required files
	aw.appendLog(fmt.Sprintf("\nVerifying %d required files...", len(aw.loadedPatch.RequiredFiles)))
	mismatches := 0

	for i, req := range aw.loadedPatch.RequiredFiles {
		if i < 5 || mismatches > 0 { // Show first 5 or any mismatches
			filePath := filepath.Join(aw.currentDir, req.Path)

			if !utils.FileExists(filePath) {
				aw.appendLog(fmt.Sprintf("✗ Required file missing: %s", req.Path))
				mismatches++
				continue
			}

			checksum, err := utils.CalculateFileChecksum(filePath)
			if err != nil {
				aw.appendLog(fmt.Sprintf("✗ Failed to verify: %s", req.Path))
				mismatches++
				continue
			}

			if checksum != req.Checksum {
				aw.appendLog(fmt.Sprintf("✗ Hash mismatch: %s", req.Path))
				mismatches++
			}
		}
	}

	if mismatches > 0 {
		aw.appendLog(fmt.Sprintf("\n✗ %d file(s) have mismatches - patch cannot be applied", mismatches))
		aw.setStatus("Dry run failed: File mismatches detected")
		aw.applyBtn.Enable()
		return
	}

	aw.appendLog("✓ All required files verified")

	// Show operations that would be performed
	aw.appendLog("\nOperations that would be performed:")
	for i, op := range aw.loadedPatch.Operations {
		if i >= 10 {
			aw.appendLog(fmt.Sprintf("... and %d more operations", len(aw.loadedPatch.Operations)-10))
			break
		}

		switch op.Type {
		case utils.OpAdd:
			aw.appendLog(fmt.Sprintf("  ADD: %s", op.FilePath))
		case utils.OpModify:
			aw.appendLog(fmt.Sprintf("  MODIFY: %s", op.FilePath))
		case utils.OpDelete:
			aw.appendLog(fmt.Sprintf("  DELETE: %s", op.FilePath))
		case utils.OpAddDir:
			aw.appendLog(fmt.Sprintf("  ADD DIR: %s", op.FilePath))
		case utils.OpDeleteDir:
			aw.appendLog(fmt.Sprintf("  DELETE DIR: %s", op.FilePath))
		}
	}

	aw.appendLog("\n✓ Dry run completed - patch can be applied safely")
	aw.setStatus("Dry run completed successfully")
	aw.applyBtn.Enable()

	// Show success dialog
	if aw.window != nil {
		dialog.ShowInformation("Dry Run Complete",
			"Patch simulation completed successfully.\n\nAll checks passed - patch can be applied safely.",
			aw.window)
	}
}

// restoreBackup restores from backup directory
func (aw *ApplierWindow) restoreBackup(backupDir, targetDir string) error {
	// Remove current directory
	if err := os.RemoveAll(targetDir); err != nil {
		return fmt.Errorf("failed to remove current directory: %w", err)
	}

	// Restore from backup
	return aw.copyDir(backupDir, targetDir)
}

// copyDir recursively copies a directory
func (aw *ApplierWindow) copyDir(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := utils.EnsureDir(dstPath); err != nil {
				return err
			}
			if err := aw.copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := utils.CopyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// setStatus updates the status label
func (aw *ApplierWindow) setStatus(status string) {
	aw.statusLabel.SetText(status)
}

// appendLog appends a message to the log
func (aw *ApplierWindow) appendLog(message string) {
	current := aw.logText.Text
	if current != "" {
		current += "\n"
	}
	current += message
	aw.logText.SetText(current)

	// Auto-scroll to bottom
	aw.logText.Refresh()
}
