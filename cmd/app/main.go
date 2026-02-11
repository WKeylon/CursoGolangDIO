package main

import (
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"pesquisa-eleitoral/internal/database"
	"pesquisa-eleitoral/internal/gerenciador"
	"pesquisa-eleitoral/internal/modelos"
)

var myApp fyne.App
var myWindow fyne.Window

func main() {
	database.InitDB()
	myApp = app.New()
	myWindow = myApp.NewWindow("Sistema de Pesquisa Eleitoral")
	myWindow.Resize(fyne.NewSize(400, 700))

	showLogin()

	myWindow.ShowAndRun()
}

func showLogin() {
	userEntry := widget.NewEntry()
	userEntry.SetPlaceHolder("Usuário")
	passEntry := widget.NewPasswordEntry()
	passEntry.SetPlaceHolder("Senha")

	statusLabel := widget.NewLabel("")
	statusLabel.Alignment = fyne.TextAlignCenter

	loginBtn := widget.NewButton("Entrar", func() {
		user, err := gerenciador.Authenticate(userEntry.Text, passEntry.Text)
		if err != nil {
			statusLabel.SetText("Erro: " + err.Error())
			return
		}

		if user.PrimeiroAcesso {
			showChangePassword(user)
		} else {
			showMainApp(user)
		}
	})

	content := container.NewVBox(
		layout.NewSpacer(),
		widget.NewLabelWithStyle("Login", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		userEntry,
		passEntry,
		statusLabel,
		loginBtn,
		layout.NewSpacer(),
	)

	myWindow.SetContent(container.NewCenter(content))
}

func showChangePassword(user *modelos.Usuario) {
	passEntry := widget.NewPasswordEntry()
	passEntry.SetPlaceHolder("Nova Senha")
	confirmEntry := widget.NewPasswordEntry()
	confirmEntry.SetPlaceHolder("Confirmar Nova Senha")

	statusLabel := widget.NewLabel("")

	changeBtn := widget.NewButton("Alterar Senha", func() {
		if passEntry.Text != confirmEntry.Text {
			statusLabel.SetText("As senhas não conferem")
			return
		}
		if len(passEntry.Text) < 4 {
			statusLabel.SetText("A senha deve ter pelo menos 4 caracteres")
			return
		}

		err := gerenciador.ChangePassword(user.ID, passEntry.Text)
		if err != nil {
			statusLabel.SetText("Erro ao alterar senha: " + err.Error())
			return
		}

		dialog.ShowInformation("Sucesso", "Senha alterada com sucesso! Faça login novamente.", myWindow)
		showLogin()
	})

	content := container.NewVBox(
		widget.NewLabelWithStyle("Trocar Senha (Primeiro Acesso)", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		passEntry,
		confirmEntry,
		statusLabel,
		changeBtn,
	)

	myWindow.SetContent(container.NewCenter(content))
}

func showMainApp(user *modelos.Usuario) {
	// --- Componentes da UI ---

	// Aba Votação
	votacaoContent := container.NewVBox()
	scrollVotacao := container.NewVScroll(votacaoContent)

	// Aba Administração (Visível apenas para Admin e Gerente)
	var adminTab *container.TabItem
	if user.Role == modelos.RoleAdmin || user.Role == modelos.RoleGerente {
		adminContent := createAdminContent(user)
		adminTab = container.NewTabItem("Administração", adminContent)
	}

	// Aba Usuários (Visível apenas para Admin e Gerente)
	var usersTab *container.TabItem
	if user.Role == modelos.RoleAdmin || user.Role == modelos.RoleGerente {
		usersContent := createUsersContent(user)
		usersTab = container.NewTabItem("Usuários", usersContent)
	}

	// Aba Estatísticas
	estatisticasContent := container.NewVBox()
	statsScroll := container.NewVScroll(estatisticasContent)
	atualizarEstatisticas(estatisticasContent, nil) // Inicialmente Global

	// Se pesquisador, deve selecionar cidade antes de votar.
	updateVotingScreen(user, votacaoContent)

	// Configurar Abas
	var tabs *container.AppTabs

	statsTab := container.NewTabItem("Estatísticas", container.NewBorder(
		widget.NewButton("Atualizar Global", func() { atualizarEstatisticas(estatisticasContent, nil) }),
		nil, nil, nil,
		statsScroll,
	))

	votingTab := container.NewTabItem("Votação", scrollVotacao)

	if user.Role == modelos.RoleAdmin {
		tabs = container.NewAppTabs(votingTab, statsTab, adminTab, usersTab)
	} else if user.Role == modelos.RoleGerente {
		tabs = container.NewAppTabs(votingTab, statsTab, adminTab, usersTab)
	} else {
		// Pesquisador: Votação apenas
		tabs = container.NewAppTabs(votingTab)
	}

	logoutBtn := widget.NewButton("Sair", func() {
		showLogin()
	})

	topBar := container.NewHBox(
		widget.NewLabel("Usuário: "+user.Username+" ("+user.Role+")"),
		layout.NewSpacer(),
		logoutBtn,
	)

	mainContainer := container.NewBorder(
		topBar,
		nil, nil, nil,
		tabs,
	)

	myWindow.SetContent(mainContainer)
}

// --- Funções Auxiliares de UI ---

func updateVotingScreen(user *modelos.Usuario, content *fyne.Container) {
	content.Objects = nil

	// Mapa de Cidades permitidas
	cityMap := make(map[string]uint)
	cityOptions := []string{}

	// Se for Pesquisador, usa apenas as cidades atribuídas
	// Se for Admin/Gerente, pode ver todas?
	// O prompt diz: "pesquisador este tem acesso apenas as pesquisas destinadas a sua cidade"
	// Vamos assumir que Admin/Gerente podem selecionar qualquer cidade para TESTAR o voto, ou ver estatísticas.

	if user.Role == modelos.RoleAdmin || user.Role == modelos.RoleGerente {
		cities, _ := gerenciador.GetCities()
		for _, c := range cities {
			cityOptions = append(cityOptions, c.Nome)
			cityMap[c.Nome] = c.ID
		}
	} else {
		for _, c := range user.Cidades {
			cityOptions = append(cityOptions, c.Nome)
			cityMap[c.Nome] = c.ID
		}
	}

	if len(cityOptions) == 0 {
		content.Add(widget.NewLabel("Nenhuma cidade disponível para este usuário."))
		content.Refresh()
		return
	}

	selectLabel := widget.NewLabel("Selecione a Cidade de Pesquisa:")
	var selectedCityID uint

	citySelect := widget.NewSelect(cityOptions, func(s string) {
		selectedCityID = cityMap[s]
		loadVotingForm(selectedCityID, user.ID, content)
	})

	content.Add(selectLabel)
	content.Add(citySelect)

	// Se só tiver uma cidade, seleciona automaticamente
	if len(cityOptions) == 1 {
		citySelect.SetSelected(cityOptions[0])
	}
	content.Refresh()
}

func loadVotingForm(cityID, userID uint, content *fyne.Container) {
	// Preserva os dois primeiros widgets (Label e Select)
	// Mas como o Select chama essa função, se deletarmos tudo, o select some.
	// Melhor limpar do índice 2 em diante.

	// Hack: Recriar o container interno para o form
	// Mas o 'content' passado é o VBox principal da aba.
	// Vamos remover itens > 1 (Label=0, Select=1)

	// Remove objects safely
	if len(content.Objects) > 2 {
		content.Objects = content.Objects[:2]
	}

	content.Add(widget.NewSeparator())
	content.Add(widget.NewLabelWithStyle("Candidatos", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}))

	candidates, err := gerenciador.ListCandidates(cityID)
	if err != nil {
		log.Println("Erro ao listar candidatos:", err)
	}

	if len(candidates) == 0 {
		content.Add(widget.NewLabel("Nenhum candidato cadastrado para esta cidade."))
	} else {
		grupoCandidatos := widget.NewRadioGroup([]string{}, func(s string) {})
		candidatosMap := make(map[string]uint)
		opcoes := []string{}

		for _, c := range candidates {
			label := fmt.Sprintf("%s - %s", c.Nome, c.Partido)
			opcoes = append(opcoes, label)
			candidatosMap[label] = c.ID
		}
		grupoCandidatos.Options = opcoes
		content.Add(grupoCandidatos)

		content.Add(widget.NewSeparator())
		content.Add(widget.NewLabelWithStyle("Perguntas", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}))

		questions, _ := gerenciador.ListQuestions()
		respostas := make(map[uint]bool) // PerguntaID -> Resposta (Sim=true)

		for _, p := range questions {
			pID := p.ID
			label := widget.NewLabel(p.Texto)

			rg := widget.NewRadioGroup([]string{"Sim", "Não"}, func(s string) {
				respostas[pID] = (s == "Sim")
			})
			rg.Horizontal = true

			content.Add(label)
			content.Add(rg)
		}

		confirmarBtn := widget.NewButton("CONFIRMAR VOTO", func() {
			sel := grupoCandidatos.Selected
			if sel == "" {
				dialog.ShowInformation("Atenção", "Selecione um candidato.", myWindow)
				return
			}

			// Validação opcional: Todas as perguntas respondidas?
			if len(respostas) < len(questions) {
				dialog.ShowInformation("Atenção", "Responda todas as perguntas.", myWindow)
				return
			}

			candID := candidatosMap[sel]
			if err := gerenciador.RegisterVote(candID, cityID, userID); err != nil {
				dialog.ShowError(err, myWindow)
				return
			}

			for pID, resp := range respostas {
				gerenciador.RegisterAnswer(pID, cityID, userID, resp)
			}

			dialog.ShowInformation("Sucesso", "Voto registrado!", myWindow)

			// Resetar form (recarrega)
			loadVotingForm(cityID, userID, content)
		})

		content.Add(widget.NewSeparator())
		content.Add(confirmarBtn)
	}
	content.Refresh()
}

func createAdminContent(user *modelos.Usuario) *fyne.Container {
	// Formulário para adicionar Candidato
	nomeEntry := widget.NewEntry()
	nomeEntry.SetPlaceHolder("Nome do Candidato")
	partidoEntry := widget.NewEntry()
	partidoEntry.SetPlaceHolder("Partido")

	// Seleção de Cidade para o Candidato
	cities, _ := gerenciador.GetCities()
	cityOptions := []string{"Global (Todas as Cidades)"}
	cityIDMap := make(map[string]*uint)
	cityIDMap["Global (Todas as Cidades)"] = nil

	for _, c := range cities {
		cityOptions = append(cityOptions, c.Nome)
		id := c.ID
		cityIDMap[c.Nome] = &id
	}

	citySelect := widget.NewSelect(cityOptions, nil)
	citySelect.PlaceHolder = "Cidade do Candidato"

	addCandidatoBtn := widget.NewButton("Salvar Candidato", func() {
		if nomeEntry.Text != "" && partidoEntry.Text != "" {
			selectedCity := citySelect.Selected
			cID := cityIDMap[selectedCity] // nil se vazio ou Global

			err := gerenciador.AddCandidate(nomeEntry.Text, partidoEntry.Text, cID)
			if err != nil {
				dialog.ShowError(err, myWindow)
			} else {
				dialog.ShowInformation("Sucesso", "Candidato adicionado", myWindow)
				nomeEntry.SetText("")
				partidoEntry.SetText("")
				citySelect.SetSelected("")
			}
		} else {
			dialog.ShowInformation("Erro", "Preencha Nome e Partido", myWindow)
		}
	})

	perguntaEntry := widget.NewEntry()
	perguntaEntry.SetPlaceHolder("Texto da Pergunta")
	addPerguntaBtn := widget.NewButton("Salvar Pergunta", func() {
		if perguntaEntry.Text != "" {
			gerenciador.AddQuestion(perguntaEntry.Text)
			dialog.ShowInformation("Sucesso", "Pergunta adicionada", myWindow)
			perguntaEntry.SetText("")
		}
	})

	return container.NewVBox(
		widget.NewLabelWithStyle("Adicionar Candidato", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		nomeEntry,
		partidoEntry,
		citySelect,
		addCandidatoBtn,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Adicionar Pergunta (Sim/Não)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		perguntaEntry,
		addPerguntaBtn,
	)
}

func createUsersContent(user *modelos.Usuario) *fyne.Container {
	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("Username")

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Senha Inicial")

	// Role Selection
	roleSelect := widget.NewSelect([]string{}, nil)
	if user.Role == modelos.RoleAdmin {
		roleSelect.Options = []string{modelos.RoleAdmin, modelos.RoleGerente, modelos.RolePesquisador}
	} else {
		// Gerente só cria pesquisador
		roleSelect.Options = []string{modelos.RolePesquisador}
		roleSelect.SetSelected(modelos.RolePesquisador)
	}

	// City Multi-Selection
	// Fyne doesn't have a multi-select widget natively easily, so we use Checkboxes in a scroll container
	cities, _ := gerenciador.GetCities()
	selectedCities := make(map[uint]bool)

	cityChecks := container.NewVBox()
	for _, c := range cities {
		cID := c.ID
		check := widget.NewCheck(c.Nome, func(b bool) {
			selectedCities[cID] = b
		})
		cityChecks.Add(check)
	}
	cityScroll := container.NewVScroll(cityChecks)
	cityScroll.SetMinSize(fyne.NewSize(0, 150))

	addUserBtn := widget.NewButton("Criar Usuário", func() {
		if usernameEntry.Text == "" || passwordEntry.Text == "" || roleSelect.Selected == "" {
			dialog.ShowInformation("Erro", "Preencha todos os campos", myWindow)
			return
		}

		// Coletar IDs selecionados
		var cityIDs []uint
		for id, selected := range selectedCities {
			if selected {
				cityIDs = append(cityIDs, id)
			}
		}

		err := gerenciador.CreateUser(user.Role, usernameEntry.Text, passwordEntry.Text, roleSelect.Selected, cityIDs)
		if err != nil {
			dialog.ShowError(err, myWindow)
		} else {
			dialog.ShowInformation("Sucesso", "Usuário criado com sucesso!", myWindow)
			usernameEntry.SetText("")
			passwordEntry.SetText("")
			// Limpar checkboxes (manual refresh needed usually)
		}
	})

	// Lista de usuários
	listContainer := container.NewVBox()

	refreshBtn := widget.NewButton("Atualizar Lista", func() {
		listContainer.Objects = nil
		users, _ := gerenciador.ListUsers()
		for _, u := range users {
			cidadesStr := ""
			if len(u.Cidades) > 0 {
				cidadesStr = fmt.Sprintf("%d cidades", len(u.Cidades))
			} else {
				cidadesStr = "Nenhuma"
			}
			label := widget.NewLabel(fmt.Sprintf("%s [%s] - %s", u.Username, u.Role, cidadesStr))
			listContainer.Add(label)
		}
		listContainer.Refresh()
	})

	// Auto-load list
	// We can't auto-click the button here easily without goroutines or lifecycle, let's just show the button.

	return container.NewVBox(
		widget.NewLabelWithStyle("Novo Usuário", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		usernameEntry,
		passwordEntry,
		widget.NewLabel("Função:"),
		roleSelect,
		widget.NewLabel("Cidades Permitidas:"),
		cityScroll,
		addUserBtn,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Usuários Existentes", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		refreshBtn,
		container.NewVScroll(listContainer),
	)
}

func atualizarEstatisticas(content *fyne.Container, cityID *uint) {
	content.Objects = nil
	content.Add(widget.NewLabelWithStyle("Resultados dos Candidatos", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}))

	votes, _ := gerenciador.GetCandidateVotes(cityID)
	for nome, count := range votes {
		texto := fmt.Sprintf("%s: %d votos", nome, count)
		content.Add(widget.NewLabel(texto))
	}

	content.Add(widget.NewSeparator())
	content.Add(widget.NewLabelWithStyle("Resultados das Perguntas", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}))

	qStats, _ := gerenciador.GetQuestionStats(cityID)
	for texto, stats := range qStats {
		res := fmt.Sprintf("%s\nSim: %d | Não: %d", texto, stats["Sim"], stats["Não"])
		content.Add(widget.NewLabel(res))
	}
	content.Refresh()
}
