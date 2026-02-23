package main

import (
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"image/color"
	"pesquisa-eleitoral/internal/database"
	"pesquisa-eleitoral/internal/gerenciador"
	"pesquisa-eleitoral/internal/modelos"
)

var myApp fyne.App
var myWindow fyne.Window

func main() {
	database.InitDB()
	myApp = app.New()
	myApp.Settings().SetTheme(&myTheme{}) // Apply Custom Theme

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
	statusLabel.TextStyle = fyne.TextStyle{Italic: true}
	statusLabel.Refresh()

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
	loginBtn.Importance = widget.HighImportance // Primary color

	// Logo Placeholder (Icon)
	logo := widget.NewIcon(theme.AccountIcon())
	logo.SetMinSize(fyne.NewSize(64, 64))

	cardContent := container.NewVBox(
		container.NewCenter(logo),
		widget.NewLabelWithStyle("Bem-vindo", fyne.TextAlignCenter, fyne.TextStyle{Bold: true, Monospace: false}),
		widget.NewSeparator(),
		userEntry,
		passEntry,
		layout.NewSpacer(),
		loginBtn,
		statusLabel,
	)

	// A Card-like container for Login
	loginCard := widget.NewCard("", "", cardContent)

	centeredContent := container.NewCenter(container.New(layout.NewGridWrapLayout(fyne.NewSize(300, 350)), loginCard))

	myWindow.SetContent(centeredContent)
}

func showChangePassword(user *modelos.Usuario) {
	passEntry := widget.NewPasswordEntry()
	passEntry.SetPlaceHolder("Nova Senha")
	confirmEntry := widget.NewPasswordEntry()
	confirmEntry.SetPlaceHolder("Confirmar Nova Senha")

	statusLabel := widget.NewLabel("")
	statusLabel.Alignment = fyne.TextAlignCenter

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
	changeBtn.Importance = widget.HighImportance

	content := container.NewVBox(
		widget.NewLabelWithStyle("Trocar Senha (Primeiro Acesso)", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		passEntry,
		confirmEntry,
		statusLabel,
		changeBtn,
	)

	myWindow.SetContent(container.NewCenter(container.New(layout.NewGridWrapLayout(fyne.NewSize(300, 300)), widget.NewCard("", "", content))))
}

func showMainApp(user *modelos.Usuario) {
	// --- Custom Header ---
	headerText := canvas.NewText("Pesquisa Eleitoral", color.White)
	headerText.TextSize = 18
	headerText.TextStyle = fyne.TextStyle{Bold: true}

	userText := canvas.NewText(user.Username, color.White)
	userText.TextSize = 14
	userText.Alignment = fyne.TextAlignTrailing

	logoutIcon := widget.NewButtonWithIcon("", theme.LogoutIcon(), func() {
		showLogin()
	})
	logoutIcon.Importance = widget.LowImportance // Transparent-ish

	headerBar := container.NewBorder(nil, nil,
		container.NewHBox(widget.NewIcon(theme.HomeIcon()), container.NewCenter(headerText)),
		container.NewHBox(container.NewCenter(userText), logoutIcon),
	)

	// Blue Background for Header
	bgRect := canvas.NewRectangle(color.NRGBA{R: 0x00, G: 0x7B, B: 0xFF, A: 0xFF})
	headerStack := container.NewStack(bgRect, container.NewPadded(headerBar))


	// --- Content Areas ---

	// Aba Votação
	votacaoContent := container.NewVBox()
	scrollVotacao := container.NewVScroll(votacaoContent)

	// Role-Based Access Control + Granular Permissions
	isAdminOrManager := user.Role == modelos.RoleAdmin || user.Role == modelos.RoleGerente

	// Aba Administração
	var adminTab *container.TabItem
	if isAdminOrManager || gerenciador.CheckPermission(user, modelos.AreaCandidatos, "editar") || gerenciador.CheckPermission(user, modelos.AreaPerguntas, "editar") {
		adminContent := createAdminContent(user)
		adminTab = container.NewTabItem("Administração", adminContent)
		adminTab.Icon = theme.SettingsIcon()
	}

	// Aba Usuários
	var usersTab *container.TabItem
	if isAdminOrManager || gerenciador.CheckPermission(user, modelos.AreaUsuarios, "ler") || gerenciador.CheckPermission(user, modelos.AreaUsuarios, "editar") {
		usersContent := createUsersContent(user)
		usersTab = container.NewTabItem("Usuários", usersContent)
		usersTab.Icon = theme.AccountIcon()
	}

	// Aba Estatísticas
	var statsTab *container.TabItem
	if isAdminOrManager || gerenciador.CheckPermission(user, modelos.AreaEstatisticas, "ler") {
		estatisticasContent := container.NewVBox()
		statsScroll := container.NewVScroll(estatisticasContent)
		atualizarEstatisticas(user, estatisticasContent, nil)

		statsTab = container.NewTabItem("Estatísticas", container.NewBorder(
			widget.NewButton("Atualizar Global", func() { atualizarEstatisticas(user, estatisticasContent, nil) }),
			nil, nil, nil,
			statsScroll,
		))
		statsTab.Icon = theme.InfoIcon()
	}

	updateVotingScreen(user, votacaoContent)

	// Configurar Abas
	tabs := container.NewAppTabs()
	tabs.SetTabLocation(container.TabLocationBottom) // Bottom Navigation for Mobile feel

	if gerenciador.CheckPermission(user, modelos.AreaVotacao, "editar") || user.Role == modelos.RolePesquisador {
		vTab := container.NewTabItem("Votação", scrollVotacao)
		vTab.Icon = theme.ConfirmIcon()
		tabs.Append(vTab)
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

	// Main Layout
	mainContainer := container.NewBorder(headerStack, nil, nil, nil, tabs)
	myWindow.SetContent(mainContainer)
}

// --- UI Helpers ---

func updateVotingScreen(user *modelos.Usuario, content *fyne.Container) {
	content.Objects = nil

	if !gerenciador.CheckPermission(user, modelos.AreaVotacao, "editar") {
		content.Add(widget.NewLabel("Você não tem permissão para votar."))
		content.Refresh()
		return
	}

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
		content.Add(widget.NewCard("Aviso", "Nenhuma cidade disponível.", nil))
		content.Refresh()
		return
	}

	selectLabel := widget.NewLabel("Cidade de Pesquisa:")
	var selectedCityID uint

	citySelect := widget.NewSelect(cityOptions, func(s string) {
		selectedCityID = cityMap[s]
		loadVotingForm(user, selectedCityID, content)
	})

	// Card for City Selection
	cityCard := widget.NewCard("", "", container.NewVBox(selectLabel, citySelect))
	content.Add(cityCard)

	if len(cityOptions) == 1 {
		citySelect.SetSelected(cityOptions[0])
	}
	content.Refresh()
}

func loadVotingForm(user *modelos.Usuario, cityID uint, content *fyne.Container) {
	// Preserve City Selection Card (Index 0)
	if len(content.Objects) > 1 {
		content.Objects = content.Objects[:1]
	}

	candidates, err := gerenciador.ListCandidates(cityID)
	if err != nil {
		log.Println("Erro ao listar candidatos:", err)
	}

	formContainer := container.NewVBox()

	if len(candidates) == 0 {
		formContainer.Add(widget.NewLabel("Nenhum candidato cadastrado."))
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

		candCard := widget.NewCard("Candidatos", "Selecione um candidato", grupoCandidatos)
		formContainer.Add(candCard)

		questions, _ := gerenciador.ListQuestions()
		respostas := make(map[uint]bool)

		questContainer := container.NewVBox()
		for _, p := range questions {
			pID := p.ID
			rg := widget.NewRadioGroup([]string{"Sim", "Não"}, func(s string) {
				respostas[pID] = (s == "Sim")
			})
			rg.Horizontal = true

			qCard := widget.NewCard(p.Texto, "", rg)
			questContainer.Add(qCard)
		}

		formContainer.Add(widget.NewCard("Perguntas", "Responda abaixo", questContainer))

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
		confirmarBtn.Importance = widget.HighImportance

		formContainer.Add(container.NewPadded(confirmarBtn))
	}

	content.Add(formContainer)
	content.Refresh()
}

func createAdminContent(user *modelos.Usuario) *fyne.Container {
	nomeEntry := widget.NewEntry()
	nomeEntry.SetPlaceHolder("Nome")
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
			cID := cityIDMap[selectedCity]

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
	addCandidatoBtn.Importance = widget.HighImportance

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
	addPerguntaBtn.Importance = widget.HighImportance

	if !gerenciador.CheckPermission(user, modelos.AreaPerguntas, "editar") {
		addPerguntaBtn.Disable()
	}

	candCard := widget.NewCard("Adicionar Candidato", "", container.NewVBox(nomeEntry, partidoEntry, citySelect, addCandidatoBtn))
	pergCard := widget.NewCard("Adicionar Pergunta", "", container.NewVBox(perguntaEntry, addPerguntaBtn))

	return container.NewVBox(candCard, pergCard)
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

	areas := []string{modelos.AreaCandidatos, modelos.AreaPerguntas, modelos.AreaUsuarios, modelos.AreaEstatisticas, modelos.AreaVotacao}
	permMap := make(map[string]map[string]bool)

	permContainer := container.NewVBox(widget.NewLabelWithStyle("Permissões", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))

	for _, area := range areas {
		area := area
		permMap[area] = make(map[string]bool)

		lbl := widget.NewLabel(area)
		chkLer := widget.NewCheck("Ler", func(b bool) { permMap[area]["ler"] = b })
		chkEdit := widget.NewCheck("Editar", func(b bool) { permMap[area]["editar"] = b })
		chkDel := widget.NewCheck("Deletar", func(b bool) { permMap[area]["deletar"] = b })

		if area == modelos.AreaVotacao {
			chkEdit.SetChecked(true)
			permMap[area]["editar"] = true
		}

		row := container.NewHBox(lbl, layout.NewSpacer(), chkLer, chkEdit, chkDel)
		permContainer.Add(row)
	}

	addUserBtn := widget.NewButton("Criar Usuário", func() {
		if usernameEntry.Text == "" || passwordEntry.Text == "" || roleSelect.Selected == "" {
			dialog.ShowInformation("Erro", "Preencha todos os campos", myWindow)
			return
		}

		var cityIDs []uint
		for id, selected := range selectedCities {
			if selected {
				cityIDs = append(cityIDs, id)
			}
		}

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
			dialog.ShowInformation("Sucesso", "Usuário criado!", myWindow)
			usernameEntry.SetText("")
			passwordEntry.SetText("")
		}
	})
	addUserBtn.Importance = widget.HighImportance

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

	newUserCard := widget.NewCard("Novo Usuário", "", container.NewVBox(
		usernameEntry, passwordEntry,
		widget.NewLabel("Função:"), roleSelect,
		widget.NewLabel("Cidades:"), cityScroll,
		permContainer, addUserBtn,
	))

	listCard := widget.NewCard("Usuários Existentes", "", container.NewVBox(refreshBtn, container.NewVScroll(listContainer)))

	return container.NewVBox(newUserCard, listCard)
}

func atualizarEstatisticas(user *modelos.Usuario, content *fyne.Container, cityID *uint) {
	content.Objects = nil

	votes, err := gerenciador.GetCandidateVotes(user, cityID)
	if err != nil {
		content.Add(widget.NewLabel("Erro ao carregar estatísticas: " + err.Error()))
		content.Refresh()
		return
	}

	cContainer := container.NewVBox()
	for nome, count := range votes {
		texto := fmt.Sprintf("%s: %d votos", nome, count)
		cContainer.Add(widget.NewLabel(texto))
	}
	content.Add(widget.NewCard("Resultados: Candidatos", "", cContainer))

	qStats, _ := gerenciador.GetQuestionStats(user, cityID)
	qContainer := container.NewVBox()
	for texto, stats := range qStats {
		res := fmt.Sprintf("%s\nSim: %d | Não: %d", texto, stats["Sim"], stats["Não"])
		qContainer.Add(widget.NewLabel(res))
	}
	content.Add(widget.NewCard("Resultados: Perguntas", "", qContainer))

	content.Refresh()
}
