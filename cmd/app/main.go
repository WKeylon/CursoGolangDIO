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
	myWindow.Resize(fyne.NewSize(450, 750))

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

	// Aba Administração (Visível apenas se tiver permissão)
	var adminTab *container.TabItem
	if gerenciador.CheckPermission(user, modelos.AreaCandidatos, "editar") || gerenciador.CheckPermission(user, modelos.AreaPerguntas, "editar") {
		adminContent := createAdminContent(user)
		adminTab = container.NewTabItem("Administração", adminContent)
	}

	// Aba Usuários (Visível apenas se tiver permissão)
	var usersTab *container.TabItem
	if gerenciador.CheckPermission(user, modelos.AreaUsuarios, "ler") || gerenciador.CheckPermission(user, modelos.AreaUsuarios, "editar") {
		usersContent := createUsersContent(user)
		usersTab = container.NewTabItem("Usuários", usersContent)
	}

	// Aba Estatísticas
	var statsTab *container.TabItem
	if gerenciador.CheckPermission(user, modelos.AreaEstatisticas, "ler") {
		estatisticasContent := container.NewVBox()
		statsScroll := container.NewVScroll(estatisticasContent)
		atualizarEstatisticas(user, estatisticasContent, nil)

		statsTab = container.NewTabItem("Estatísticas", container.NewBorder(
			widget.NewButton("Atualizar Global", func() { atualizarEstatisticas(user, estatisticasContent, nil) }),
			nil, nil, nil,
			statsScroll,
		))
	}

	// Se pesquisador/votante, update screen
	updateVotingScreen(user, votacaoContent)

	// Configurar Abas
	tabs := container.NewAppTabs()

	// Adiciona abas conforme permissão
	// Votação: Se tiver permissão de editar (votar)
	if gerenciador.CheckPermission(user, modelos.AreaVotacao, "editar") {
		tabs.Append(container.NewTabItem("Votação", scrollVotacao))
	}

	if statsTab != nil {
		tabs.Append(statsTab)
	}
	if adminTab != nil {
		tabs.Append(adminTab)
	}
	if usersTab != nil {
		tabs.Append(usersTab)
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

	if !gerenciador.CheckPermission(user, modelos.AreaVotacao, "editar") {
		content.Add(widget.NewLabel("Você não tem permissão para votar."))
		content.Refresh()
		return
	}

	// Mapa de Cidades permitidas
	cityMap := make(map[string]uint)
	cityOptions := []string{}

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
		loadVotingForm(user, selectedCityID, content)
	})

	content.Add(selectLabel)
	content.Add(citySelect)

	// Se só tiver uma cidade, seleciona automaticamente
	if len(cityOptions) == 1 {
		citySelect.SetSelected(cityOptions[0])
	}
	content.Refresh()
}

func loadVotingForm(user *modelos.Usuario, cityID uint, content *fyne.Container) {
	// Remove objects safely (Keep first 2: Label + Select)
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
		respostas := make(map[uint]bool)

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

			if len(respostas) < len(questions) {
				dialog.ShowInformation("Atenção", "Responda todas as perguntas.", myWindow)
				return
			}

			candID := candidatosMap[sel]
			if err := gerenciador.RegisterVote(user, candID, cityID); err != nil {
				dialog.ShowError(err, myWindow)
				return
			}

			for pID, resp := range respostas {
				gerenciador.RegisterAnswer(user, pID, cityID, resp)
			}

			dialog.ShowInformation("Sucesso", "Voto registrado!", myWindow)
			loadVotingForm(user, cityID, content)
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

			err := gerenciador.AddCandidate(user, nomeEntry.Text, partidoEntry.Text, cID)
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

	if !gerenciador.CheckPermission(user, modelos.AreaCandidatos, "editar") {
		addCandidatoBtn.Disable()
	}

	perguntaEntry := widget.NewEntry()
	perguntaEntry.SetPlaceHolder("Texto da Pergunta")
	addPerguntaBtn := widget.NewButton("Salvar Pergunta", func() {
		if perguntaEntry.Text != "" {
			err := gerenciador.AddQuestion(user, perguntaEntry.Text)
			if err != nil {
				dialog.ShowError(err, myWindow)
			} else {
				dialog.ShowInformation("Sucesso", "Pergunta adicionada", myWindow)
				perguntaEntry.SetText("")
			}
		}
	})

	if !gerenciador.CheckPermission(user, modelos.AreaPerguntas, "editar") {
		addPerguntaBtn.Disable()
	}

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

	roleSelect := widget.NewSelect([]string{}, nil)
	if user.Role == modelos.RoleAdmin {
		roleSelect.Options = []string{modelos.RoleAdmin, modelos.RoleGerente, modelos.RolePesquisador}
	} else {
		roleSelect.Options = []string{modelos.RolePesquisador}
		roleSelect.SetSelected(modelos.RolePesquisador)
	}

	// City Multi-Selection
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
	cityScroll.SetMinSize(fyne.NewSize(0, 100))

	// Permission Matrix
	areas := []string{modelos.AreaCandidatos, modelos.AreaPerguntas, modelos.AreaUsuarios, modelos.AreaEstatisticas, modelos.AreaVotacao}
	permMap := make(map[string]map[string]bool) // Area -> Action -> bool

	permContainer := container.NewVBox(widget.NewLabelWithStyle("Permissões", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))

	for _, area := range areas {
		area := area // Capture loop variable
		permMap[area] = make(map[string]bool)

		lbl := widget.NewLabel(area)
		chkLer := widget.NewCheck("Ler", func(b bool) { permMap[area]["ler"] = b })
		chkEdit := widget.NewCheck("Editar", func(b bool) { permMap[area]["editar"] = b })
		chkDel := widget.NewCheck("Deletar", func(b bool) { permMap[area]["deletar"] = b })

		// Default checks for convenience?
		if area == modelos.AreaVotacao {
			chkEdit.SetChecked(true)
			permMap[area]["editar"] = true
		} // Default voter

		row := container.NewHBox(lbl, layout.NewSpacer(), chkLer, chkEdit, chkDel)
		permContainer.Add(row)
	}

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

		// Collect Permissions
		var permissions []modelos.Permissao
		for area, actions := range permMap {
			if actions["ler"] || actions["editar"] || actions["deletar"] {
				permissions = append(permissions, modelos.Permissao{
					Area:    area,
					Ler:     actions["ler"],
					Editar:  actions["editar"],
					Deletar: actions["deletar"],
				})
			}
		}

		err := gerenciador.CreateUser(user, usernameEntry.Text, passwordEntry.Text, roleSelect.Selected, cityIDs, permissions)
		if err != nil {
			dialog.ShowError(err, myWindow)
		} else {
			dialog.ShowInformation("Sucesso", "Usuário criado com sucesso!", myWindow)
			usernameEntry.SetText("")
			passwordEntry.SetText("")
			// Limpar checkboxes (TODO: Reset manual)
		}
	})

	if !gerenciador.CheckPermission(user, modelos.AreaUsuarios, "editar") {
		addUserBtn.Disable()
	}

	listContainer := container.NewVBox()
	refreshBtn := widget.NewButton("Atualizar Lista", func() {
		listContainer.Objects = nil
		users, _ := gerenciador.ListUsers()
		for _, u := range users {
			cidadesStr := fmt.Sprintf("%d cidades", len(u.Cidades))
			label := widget.NewLabel(fmt.Sprintf("%s [%s] - %s", u.Username, u.Role, cidadesStr))
			listContainer.Add(label)
		}
		listContainer.Refresh()
	})

	return container.NewVBox(
		widget.NewLabelWithStyle("Novo Usuário", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		usernameEntry,
		passwordEntry,
		widget.NewLabel("Função:"),
		roleSelect,
		widget.NewLabel("Cidades Permitidas:"),
		cityScroll,
		permContainer,
		addUserBtn,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Usuários Existentes", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		refreshBtn,
		container.NewVScroll(listContainer),
	)
}

func atualizarEstatisticas(user *modelos.Usuario, content *fyne.Container, cityID *uint) {
	content.Objects = nil
	content.Add(widget.NewLabelWithStyle("Resultados dos Candidatos", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}))

	votes, err := gerenciador.GetCandidateVotes(user, cityID)
	if err != nil {
		content.Add(widget.NewLabel("Erro ao carregar estatísticas: " + err.Error()))
		content.Refresh()
		return
	}

	for nome, count := range votes {
		texto := fmt.Sprintf("%s: %d votos", nome, count)
		content.Add(widget.NewLabel(texto))
	}

	content.Add(widget.NewSeparator())
	content.Add(widget.NewLabelWithStyle("Resultados das Perguntas", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}))

	qStats, _ := gerenciador.GetQuestionStats(user, cityID)
	for texto, stats := range qStats {
		res := fmt.Sprintf("%s\nSim: %d | Não: %d", texto, stats["Sim"], stats["Não"])
		content.Add(widget.NewLabel(res))
	}
	content.Refresh()
}
