package gui

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func ShowDashboard() {
	content := container.NewMax()

	menu := widget.NewList(
		func() int { return 4 },
		func() fyne.CanvasObject { return widget.NewLabel("Template") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			switch i {
			case 0:
				o.(*widget.Label).SetText("Home")
			case 1:
				o.(*widget.Label).SetText("Fazendas")
			case 2:
				o.(*widget.Label).SetText("Visitas")
			case 3:
				o.(*widget.Label).SetText("Sair")
			}
		},
	)

	menu.OnSelected = func(id widget.ListItemID) {
		switch id {
		case 0:
			content.Objects = []fyne.CanvasObject{MakeHomeTab()}
		case 1:
			content.Objects = []fyne.CanvasObject{MakeFazendasTab()}
		case 2:
			content.Objects = []fyne.CanvasObject{widget.NewLabel("Visitas - Em construção")}
		case 3:
			State.SetToken("")
			ShowLoginScreen()
		}
		content.Refresh()
	}

	split := container.NewHSplit(menu, content)
	split.Offset = 0.2
	State.Window.SetContent(split)

	// Default selection
	menu.Select(0)
}

func MakeHomeTab() fyne.CanvasObject {
	stats, err := fetchStats()
	if err != nil {
		return widget.NewLabel("Erro ao carregar stats: " + err.Error())
	}

	return container.NewVBox(
		widget.NewLabelWithStyle("Resumo", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewCard("Fazendas", "", widget.NewLabel(fmt.Sprintf("%v", stats["fazendas"]))),
		widget.NewCard("Visitas", "", widget.NewLabel(fmt.Sprintf("%v", stats["visitas"]))),
		widget.NewCard("Pessoas", "", widget.NewLabel(fmt.Sprintf("%v", stats["pessoas"]))),
	)
}

func fetchStats() (map[string]interface{}, error) {
	req, _ := http.NewRequest("GET", State.BaseURL+"/api/stats", nil)
	req.Header.Set("Authorization", "Bearer "+State.Token)

	resp, err := State.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code %d", resp.StatusCode)
	}

	var stats map[string]interface{}
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &stats)
	return stats, nil
}
