package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"pesquisa-eleitoral/internal/gerenciador"
)

var atualizarVotacao func(g *gerenciador.Gerenciador, content *fyne.Container, win fyne.Window)

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Sistema de Pesquisa Eleitoral")
	myWindow.Resize(fyne.NewSize(400, 700))

	gerente := gerenciador.NovoGerenciador()

	// --- Componentes da UI ---

	// Aba Votação
	votacaoContent := container.NewVBox()
	scrollVotacao := container.NewVScroll(votacaoContent)

	// Aba Administração
	nomeEntry := widget.NewEntry()
	nomeEntry.SetPlaceHolder("Nome do Candidato")
	partidoEntry := widget.NewEntry()
	partidoEntry.SetPlaceHolder("Partido")

	perguntaEntry := widget.NewEntry()
	perguntaEntry.SetPlaceHolder("Nova Pergunta (Sim/Não)")

	statusLabel := widget.NewLabel("")

	addCandidatoBtn := widget.NewButton("Adicionar Candidato", func() {
		if nomeEntry.Text != "" && partidoEntry.Text != "" {
			gerente.AdicionarCandidato(nomeEntry.Text, partidoEntry.Text)
			statusLabel.SetText("Candidato adicionado: " + nomeEntry.Text)
			nomeEntry.SetText("")
			partidoEntry.SetText("")
			atualizarVotacao(gerente, votacaoContent, myWindow)
		} else {
			statusLabel.SetText("Erro: Preencha nome e partido")
		}
	})

	addPerguntaBtn := widget.NewButton("Adicionar Pergunta", func() {
		if perguntaEntry.Text != "" {
			gerente.AdicionarPergunta(perguntaEntry.Text)
			statusLabel.SetText("Pergunta adicionada")
			perguntaEntry.SetText("")
			atualizarVotacao(gerente, votacaoContent, myWindow)
		} else {
			statusLabel.SetText("Erro: Digite a pergunta")
		}
	})

	adminContent := container.NewVBox(
		widget.NewLabel("Adicionar Candidato"),
		nomeEntry,
		partidoEntry,
		addCandidatoBtn,
		widget.NewSeparator(),
		widget.NewLabel("Adicionar Pergunta"),
		perguntaEntry,
		addPerguntaBtn,
		widget.NewSeparator(),
		statusLabel,
	)

	// Aba Estatísticas
	estatisticasContent := container.NewVBox()
	atualizarEstatisticasBtn := widget.NewButton("Atualizar Estatísticas", func() {
		estatisticasContent.Objects = nil
		estatisticasContent.Add(widget.NewLabelWithStyle("Resultados dos Candidatos", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}))

		for _, c := range gerente.ListarCandidatos() {
			texto := fmt.Sprintf("%s (%s): %d votos", c.Nome, c.Partido, c.Votos)
			estatisticasContent.Add(widget.NewLabel(texto))
		}

		estatisticasContent.Add(widget.NewSeparator())
		estatisticasContent.Add(widget.NewLabelWithStyle("Resultados das Perguntas", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}))

		for _, p := range gerente.ListarPerguntas() {
			texto := fmt.Sprintf("%s\nSim: %d | Não: %d", p.Texto, p.Sim, p.Nao)
			estatisticasContent.Add(widget.NewLabel(texto))
		}
		estatisticasContent.Refresh()
	})

	estatisticasContainer := container.NewBorder(atualizarEstatisticasBtn, nil, nil, nil, container.NewVScroll(estatisticasContent))

	// Implementação da função de atualização
	atualizarVotacao = func(g *gerenciador.Gerenciador, content *fyne.Container, win fyne.Window) {
		content.Objects = nil // Limpar

		content.Add(widget.NewLabelWithStyle("Escolha um Candidato", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}))

		grupoCandidatos := widget.NewRadioGroup([]string{}, func(s string) {})
		candidatosMap := make(map[string]int) // Nome -> ID
		opcoes := []string{}

		for _, c := range g.ListarCandidatos() {
			label := fmt.Sprintf("%s - %s", c.Nome, c.Partido)
			opcoes = append(opcoes, label)
			candidatosMap[label] = c.ID
		}
		grupoCandidatos.Options = opcoes
		content.Add(grupoCandidatos)

		content.Add(widget.NewSeparator())
		content.Add(widget.NewLabelWithStyle("Responda as Perguntas", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}))

		// Mapa temporário para armazenar respostas da sessão atual de votação
		respostas := make(map[int]bool)

		for _, p := range g.ListarPerguntas() {
			pID := p.ID
			label := widget.NewLabel(p.Texto)

			rg := widget.NewRadioGroup([]string{"Sim", "Não"}, func(s string) {
				if s == "Sim" {
					respostas[pID] = true
				} else {
					respostas[pID] = false
				}
			})
			rg.Horizontal = true

			content.Add(label)
			content.Add(rg)
		}

		confirmarBtn := widget.NewButton("CONFIRMAR VOTO", func() {
			sel := grupoCandidatos.Selected
			if sel == "" {
				dialog.ShowInformation("Atenção", "Por favor, selecione um candidato.", win)
				return
			}

			// Verificar se todas as perguntas foram respondidas (opcional, mas recomendado)
			if len(respostas) < len(g.ListarPerguntas()) {
				dialog.ShowInformation("Atenção", "Por favor, responda todas as perguntas.", win)
				return
			}

			candID := candidatosMap[sel]
			g.RegistrarVotoCandidato(candID)

			for pID, resp := range respostas {
				g.RegistrarRespostaPergunta(pID, resp)
			}

			dialog.ShowInformation("Sucesso", "Voto registrado com sucesso!", win)

			// Resetar a tela para o próximo voto
			atualizarVotacao(g, content, win)
		})

		content.Add(widget.NewSeparator())
		content.Add(confirmarBtn)
		content.Refresh()
	}

	// Inicializar
	atualizarVotacao(gerente, votacaoContent, myWindow)

	tabs := container.NewAppTabs(
		container.NewTabItem("Votação", scrollVotacao),
		container.NewTabItem("Estatísticas", estatisticasContainer),
		container.NewTabItem("Administração", adminContent),
	)

	myWindow.SetContent(tabs)
	myWindow.ShowAndRun()
}
