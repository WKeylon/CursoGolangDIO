package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"patrulha_rural/internal/gui"
)

func main() {
	a := app.New()
	w := a.NewWindow("Patrulha Rural")

	gui.State.App = a
	gui.State.Window = w

	w.Resize(fyne.NewSize(800, 600))

	gui.ShowLoginScreen()

	w.ShowAndRun()
}
