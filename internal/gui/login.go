package gui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func ShowLoginScreen() {
	emailEntry := widget.NewEntry()
	emailEntry.PlaceHolder = "Email"
	emailEntry.SetText("antonio.garcia@policia.mg.gov.br")

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.PlaceHolder = "Senha"
	passwordEntry.SetText("123456")

	errorLabel := widget.NewLabel("")
	errorLabel.Hide()

	loginButton := widget.NewButton("Entrar", func() {
		if err := performLogin(emailEntry.Text, passwordEntry.Text); err != nil {
			errorLabel.SetText("Erro: " + err.Error())
			errorLabel.Show()
		} else {
			ShowDashboard()
		}
	})
	loginButton.Importance = widget.HighImportance

	form := container.NewVBox(
		widget.NewLabel("Email"),
		emailEntry,
		widget.NewLabel("Senha"),
		passwordEntry,
		errorLabel,
		loginButton,
	)

	content := container.NewCenter(
		container.NewVBox(
			widget.NewLabelWithStyle("Patrulha Rural", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			form,
		),
	)

	State.Window.SetContent(content)
}

func performLogin(email, password string) error {
	reqBody, _ := json.Marshal(map[string]string{
		"email": email,
		"senha": password,
	})

	resp, err := State.Client.Post(State.BaseURL+"/login", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed with status: %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	token, ok := result["token"].(string)
	if !ok {
		return fmt.Errorf("invalid token in response")
	}

	State.SetToken(token)
	return nil
}
