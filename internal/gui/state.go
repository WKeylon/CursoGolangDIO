package gui

import (
	"net/http"
	"time"

	"fyne.io/fyne/v2"
)

type AppState struct {
	Token   string
	BaseURL string
	Client  *http.Client
	Window  fyne.Window
	App     fyne.App
}

var State = &AppState{
	BaseURL: "http://localhost:8080",
	Client:  &http.Client{Timeout: 10 * time.Second},
}

func (s *AppState) SetToken(token string) {
	s.Token = token
}
