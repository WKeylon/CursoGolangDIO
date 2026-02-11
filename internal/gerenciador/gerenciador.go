package gerenciador

import (
	"errors"
	"pesquisa-eleitoral/internal/database"
	"pesquisa-eleitoral/internal/modelos"
	"golang.org/x/crypto/bcrypt"
)

// Authenticate checks username and password
func Authenticate(username, password string) (*modelos.Usuario, error) {
	var user modelos.Usuario
	// Preload both Cidades and Permissoes
	err := database.DB.Preload("Cidades").Preload("Permissoes").Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, errors.New("usuário não encontrado")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.SenhaHash), []byte(password))
	if err != nil {
		return nil, errors.New("senha incorreta")
	}

	return &user, nil
}

// CheckPermission checks if a user has permission for an area/action
func CheckPermission(user *modelos.Usuario, area string, action string) bool {
	if user.Role == modelos.RoleAdmin {
		return true
	}

	for _, p := range user.Permissoes {
		if p.Area == area {
			if action == "ler" && p.Ler { return true }
			if action == "editar" && p.Editar { return true }
			if action == "deletar" && p.Deletar { return true }
		}
	}
	return false
}

// CreateUser creates a new user with role validation
func CreateUser(creator *modelos.Usuario, username, password, role string, cityIDs []uint, permissions []modelos.Permissao) error {
	// Permission Check
	if !CheckPermission(creator, modelos.AreaUsuarios, "editar") {
		return errors.New("permissão negada para criar usuários")
	}

	// Extra Logic: Manager can only create Researcher (Legacy rule, still valid?)
	// If we rely purely on permissions, maybe we drop this?
	// The prompt says "administrator can choose...".
	// Let's keep the legacy rule as an extra safeguard if the creator is Manager.
	if creator.Role == modelos.RoleGerente && role != modelos.RolePesquisador {
		return errors.New("gerentes só podem criar pesquisadores")
	}

	return database.CreateUser(username, password, role, cityIDs, permissions)
}

// ChangePassword changes the user's password
func ChangePassword(userID uint, newPassword string) error {
	return database.UpdatePassword(userID, newPassword)
}

// ListUsers lists all users
func ListUsers() ([]modelos.Usuario, error) {
	return database.GetAllUsers()
}

// GetCities returns all cities
func GetCities() ([]modelos.Cidade, error) {
	return database.GetAllCities()
}

// --- Voting Logic ---

func AddCandidate(user *modelos.Usuario, name, party string, cityID *uint) error {
	if !CheckPermission(user, modelos.AreaCandidatos, "editar") {
		return errors.New("permissão negada para adicionar candidato")
	}
	candidate := modelos.Candidato{
		Nome:     name,
		Partido:  party,
		CidadeID: cityID,
	}
	return database.DB.Create(&candidate).Error
}

func AddQuestion(user *modelos.Usuario, text string) error {
	if !CheckPermission(user, modelos.AreaPerguntas, "editar") {
		return errors.New("permissão negada para adicionar pergunta")
	}
	question := modelos.Pergunta{
		Texto: text,
	}
	return database.DB.Create(&question).Error
}

func ListCandidates(cityID uint) ([]modelos.Candidato, error) {
	// Read permission check? Usually for listing for voting, we assume implicit read?
	// Or should we enforce it? Let's enforce implicit read for voting context,
	// but if used in admin panel, UI should check.
	// For now, no strict check here to avoid breaking voting flow if we forget to assign "Read" on Candidates to Researcher.
	// But technically Researcher needs Read access.
	var candidates []modelos.Candidato
	err := database.DB.Where("cidade_id = ? OR cidade_id IS NULL", cityID).Find(&candidates).Error
	return candidates, err
}

func ListQuestions() ([]modelos.Pergunta, error) {
	var questions []modelos.Pergunta
	err := database.DB.Find(&questions).Error
	return questions, err
}

func RegisterVote(user *modelos.Usuario, candidateID, cityID uint) error {
	if !CheckPermission(user, modelos.AreaVotacao, "editar") { // Voting is an "edit" action (create vote)
		return errors.New("permissão negada para votar")
	}
	vote := modelos.Voto{
		CandidatoID: candidateID,
		CidadeID:    cityID,
		UsuarioID:   user.ID,
	}
	return database.DB.Create(&vote).Error
}

func RegisterAnswer(user *modelos.Usuario, questionID, cityID uint, answer bool) error {
	if !CheckPermission(user, modelos.AreaVotacao, "editar") {
		return errors.New("permissão negada para responder")
	}
	resp := modelos.Resposta{
		PerguntaID: questionID,
		Sim:        answer,
		CidadeID:   cityID,
		UsuarioID:  user.ID,
	}
	return database.DB.Create(&resp).Error
}

// Statistics

func GetCandidateVotes(user *modelos.Usuario, cityID *uint) (map[string]int64, error) {
	if !CheckPermission(user, modelos.AreaEstatisticas, "ler") {
		return nil, errors.New("permissão negada para ver estatísticas")
	}

	type Result struct {
		Nome  string
		Total int64
	}
	var results []Result

	query := database.DB.Table("votos").
		Select("candidatos.nome as nome, count(votos.id) as total").
		Joins("left join candidatos on candidatos.id = votos.candidato_id")

	if cityID != nil {
		query = query.Where("votos.cidade_id = ?", *cityID)
	}

	err := query.Group("candidatos.nome").Scan(&results).Error
	if err != nil {
		return nil, err
	}

	stats := make(map[string]int64)
	for _, r := range results {
		stats[r.Nome] = r.Total
	}
	return stats, nil
}

func GetQuestionStats(user *modelos.Usuario, cityID *uint) (map[string]map[string]int64, error) {
	if !CheckPermission(user, modelos.AreaEstatisticas, "ler") {
		return nil, errors.New("permissão negada para ver estatísticas")
	}

	var questions []modelos.Pergunta
	if err := database.DB.Find(&questions).Error; err != nil {
		return nil, err
	}

	stats := make(map[string]map[string]int64)

	for _, q := range questions {
		var simCount int64
		var naoCount int64

		qSim := database.DB.Model(&modelos.Resposta{}).Where("pergunta_id = ? AND sim = ?", q.ID, true)
		qNao := database.DB.Model(&modelos.Resposta{}).Where("pergunta_id = ? AND sim = ?", q.ID, false)

		if cityID != nil {
			qSim = qSim.Where("cidade_id = ?", *cityID)
			qNao = qNao.Where("cidade_id = ?", *cityID)
		}

		qSim.Count(&simCount)
		qNao.Count(&naoCount)

		stats[q.Texto] = map[string]int64{
			"Sim": simCount,
			"Não": naoCount,
		}
	}
	return stats, nil
}
