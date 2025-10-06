package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// CreateFolderSelector creates a folder selection widget with label and browse button
func CreateFolderSelector(label string, window fyne.Window, onSelected func(string)) *fyne.Container {
	entry := widget.NewEntry()
	entry.SetPlaceHolder("Select folder...")

	browseBtn := widget.NewButton("Browse", func() {
		dialog.ShowFolderOpen(func(dir fyne.ListableURI, err error) {
			if err == nil && dir != nil {
				path := dir.Path()
				entry.SetText(path)
				if onSelected != nil {
					onSelected(path)
				}
			}
		}, window)
	})

	return container.NewBorder(
		nil, nil,
		widget.NewLabel(label),
		browseBtn,
		entry,
	)
}

// CreateLabeledEntry creates a labeled entry widget
func CreateLabeledEntry(label, placeholder string) (*widget.Entry, *fyne.Container) {
	entry := widget.NewEntry()
	entry.SetPlaceHolder(placeholder)

	return entry, container.NewBorder(
		nil, nil,
		widget.NewLabel(label),
		nil,
		entry,
	)
}

// CreateLabeledSelect creates a labeled select widget
func CreateLabeledSelect(label string, options []string, onChange func(string)) (*widget.Select, *fyne.Container) {
	sel := widget.NewSelect(options, onChange)

	return sel, container.NewBorder(
		nil, nil,
		widget.NewLabel(label),
		nil,
		sel,
	)
}

// CreateRadioGroup creates a radio group widget
func CreateRadioGroup(label string, options []string, onChange func(string)) *fyne.Container {
	radio := widget.NewRadioGroup(options, onChange)
	radio.Horizontal = true

	return container.NewVBox(
		widget.NewLabel(label),
		radio,
	)
}

// CreateProgressDialog creates a progress dialog using the new Fyne API
func CreateProgressDialog(title, message string, window fyne.Window) *dialog.CustomDialog {
	progressBar := widget.NewProgressBarInfinite()
	content := container.NewVBox(
		widget.NewLabel(message),
		progressBar,
	)

	progress := dialog.NewCustomWithoutButtons(title, content, window)
	return progress
}

// ShowInfoDialog shows an information dialog
func ShowInfoDialog(title, message string, window fyne.Window) {
	dialog.ShowInformation(title, message, window)
}

// ShowErrorDialog shows an error dialog
func ShowErrorDialog(title, message string, window fyne.Window) {
	dialog.ShowError(nil, window)
	dialog.ShowInformation(title, message, window)
}

// ShowConfirmDialog shows a confirmation dialog
func ShowConfirmDialog(title, message string, onConfirm func(bool), window fyne.Window) {
	dialog.ShowConfirm(title, message, onConfirm, window)
}
