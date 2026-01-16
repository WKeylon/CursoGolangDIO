package gui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"patrulha_rural/internal/models"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func MakeFazendasTab() fyne.CanvasObject {
	listContainer := container.NewStack()

	var refreshList func()
	refreshList = func() {
		fazendas, err := fetchFazendas()
		if err != nil {
			listContainer.Objects = []fyne.CanvasObject{widget.NewLabel("Erro: " + err.Error())}
			listContainer.Refresh()
			return
		}

		list := widget.NewList(
			func() int { return len(fazendas) },
			func() fyne.CanvasObject { return widget.NewLabel("Template") },
			func(i widget.ListItemID, o fyne.CanvasObject) {
				o.(*widget.Label).SetText(fazendas[i].NomePropriedade + " - " + fazendas[i].Municipio)
			},
		)

		listContainer.Objects = []fyne.CanvasObject{list}
		listContainer.Refresh()
	}

	// Initial Load
	refreshList()

	addButton := widget.NewButton("Nova Fazenda", func() {
		showAddFazendaDialog(refreshList)
	})

	topBar := container.NewHBox(
		widget.NewLabelWithStyle("Fazendas", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		layout.NewSpacer(),
		addButton,
	)

	return container.NewBorder(topBar, nil, nil, nil, listContainer)
}

func fetchFazendas() ([]models.Fazenda, error) {
	req, _ := http.NewRequest("GET", State.BaseURL+"/api/fazendas", nil)
	req.Header.Set("Authorization", "Bearer "+State.Token)

	resp, err := State.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code %d", resp.StatusCode)
	}

	var fazendas []models.Fazenda
	body, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(body, &fazendas)
	return fazendas, err
}

func showAddFazendaDialog(onSuccess func()) {
	w := State.App.NewWindow("Nova Fazenda")

	nomeEntry := widget.NewEntry()
	municipioEntry := widget.NewEntry()
	estadoEntry := widget.NewEntry()
	estadoEntry.SetText("TO") // default

	submitFunc := func() {
		fazenda := models.Fazenda{
			NomePropriedade: nomeEntry.Text,
			Municipio:       municipioEntry.Text,
			Estado:          estadoEntry.Text,
		}
		if err := createFazenda(fazenda); err != nil {
			dialog := widget.NewModalPopUp(widget.NewLabel("Erro: "+err.Error()), w.Canvas())
			dialog.Show()
		} else {
			onSuccess()
			w.Close()
		}
	}

	form := widget.NewForm(
		widget.NewFormItem("Nome Propriedade", nomeEntry),
		widget.NewFormItem("Município", municipioEntry),
		widget.NewFormItem("Estado", estadoEntry),
	)

	form.OnSubmit = submitFunc

	w.SetContent(container.NewVBox(form))
	w.Resize(fyne.NewSize(400, 300))
	w.Show()
}

func createFazenda(fazenda models.Fazenda) error {
	data, _ := json.Marshal(fazenda)
	req, _ := http.NewRequest("POST", State.BaseURL+"/api/fazendas", bytes.NewBuffer(data))
	req.Header.Set("Authorization", "Bearer "+State.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := State.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed: %s", string(body))
	}
	return nil
}
