package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// UI Spacing Constants
const (
	PaddingSmall  float32 = 2
	PaddingMedium float32 = 4
	PaddingLarge  float32 = 8

	ButtonMinWidth  float32 = 90
	ButtonMinHeight float32 = 28

	WindowMinWidth  float32 = 650
	WindowMinHeight float32 = 450
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
