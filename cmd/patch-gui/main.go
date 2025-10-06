package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/cyberofficial/cyberpatchmaker/internal/core/version"
	"github.com/cyberofficial/cyberpatchmaker/internal/gui"
)

func main() {
	// Create new Fyne application with unique ID
	myApp := app.NewWithID("com.cyberofficial.cyberpatchmaker.generator")
	myWindow := myApp.NewWindow("CyberPatchMaker v" + version.GetVersion())

	// Set window size (increased for better tab visibility)
	myWindow.Resize(fyne.NewSize(900, 700))

	// Create generator UI
	generatorUI := gui.NewGeneratorWindow()
	generatorUI.SetWindow(myWindow)

	// Create batch script generator UI
	batchScriptUI := gui.NewBatchScriptWindow()
	batchScriptUI.SetWindow(myWindow)

	// Create tabs
	tabs := container.NewAppTabs(
		container.NewTabItem("Patch Generator", generatorUI),
		container.NewTabItem("Batch Script Generator", batchScriptUI),
	)

	// Set up the main content with title and tabs
	content := container.NewBorder(
		container.NewVBox(
			widget.NewLabel("CyberPatchMaker"),
			widget.NewSeparator(),
		),
		nil, nil, nil,
		tabs,
	)

	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}
