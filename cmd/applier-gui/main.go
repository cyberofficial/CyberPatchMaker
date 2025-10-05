package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/cyberofficial/cyberpatchmaker/internal/gui"
)

func main() {
	// Create new Fyne application
	myApp := app.New()
	myWindow := myApp.NewWindow("CyberPatchMaker - Patch Applier")

	// Set window size
	myWindow.Resize(fyne.NewSize(800, 600))

	// Create applier UI
	applierUI := gui.NewApplierWindow()
	applierUI.SetWindow(myWindow)

	// Set up the main content
	content := container.NewVBox(
		widget.NewLabel("CyberPatchMaker - Patch Applier"),
		widget.NewSeparator(),
		applierUI,
	)

	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}
