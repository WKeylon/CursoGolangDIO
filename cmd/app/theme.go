package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type myTheme struct{}

var _ fyne.Theme = (*myTheme)(nil)

func (m *myTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		if variant == theme.VariantLight {
			return color.NRGBA{R: 0xF0, G: 0xF2, B: 0xF5, A: 0xFF} // Light Gray
		}
		return theme.DefaultTheme().Color(name, variant)
	case theme.ColorNameForeground:
		if variant == theme.VariantLight {
			return color.NRGBA{R: 0x33, G: 0x33, B: 0x33, A: 0xFF} // Dark Gray
		}
		return theme.DefaultTheme().Color(name, variant)
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 0x00, G: 0x7B, B: 0xFF, A: 0xFF} // Blue
	case theme.ColorNameInputBackground:
		if variant == theme.VariantLight {
			return color.White
		}
		return theme.DefaultTheme().Color(name, variant)
	case theme.ColorNamePlaceHolder:
		return theme.DefaultTheme().Color(name, variant)
	case theme.ColorNameButton:
		return color.NRGBA{R: 0xE0, G: 0xE0, B: 0xE0, A: 0xFF} // Light Gray for secondary buttons
	case theme.ColorNameDisabled:
		return color.NRGBA{R: 0xCC, G: 0xCC, B: 0xCC, A: 0xFF}
	case theme.ColorNameScrollBar:
		return theme.DefaultTheme().Color(name, variant)
	case theme.ColorNameShadow:
		return color.NRGBA{A: 0x33}
	}
	return theme.DefaultTheme().Color(name, variant)
}

func (m *myTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (m *myTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (m *myTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		return 16
	case theme.SizeNamePadding:
		return 8
	case theme.SizeNameInlineIcon:
		return 24
	}
	return theme.DefaultTheme().Size(name)
}
