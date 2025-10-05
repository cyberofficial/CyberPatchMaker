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
	myWindow := myApp.NewWindow("CyberPatchMaker - Patch Generator")

	// Set window size
	myWindow.Resize(fyne.NewSize(800, 600))

	// Create generator UI
	generatorUI := gui.NewGeneratorWindow()
	generatorUI.SetWindow(myWindow)

	// Set up the main content
	content := container.NewVBox(
		widget.NewLabel("CyberPatchMaker - Patch Generator"),
		widget.NewSeparator(),
		generatorUI,
	)

	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}
