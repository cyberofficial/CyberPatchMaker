package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// UI Spacing Constants
const (
	PaddingSmall  float32 = 5
	PaddingMedium float32 = 10
	PaddingLarge  float32 = 20

	ButtonMinWidth  float32 = 120
	ButtonMinHeight float32 = 40

	WindowMinWidth  float32 = 800
	WindowMinHeight float32 = 600
)

// UI Colors
var (
	ColorSuccess = color.RGBA{R: 76, G: 175, B: 80, A: 255}  // Green
	ColorError   = color.RGBA{R: 244, G: 67, B: 54, A: 255}  // Red
	ColorWarning = color.RGBA{R: 255, G: 152, B: 0, A: 255}  // Orange
	ColorInfo    = color.RGBA{R: 33, G: 150, B: 243, A: 255} // Blue
)

// GetTheme returns the application theme
func GetTheme() fyne.Theme {
	return theme.DefaultTheme()
}
