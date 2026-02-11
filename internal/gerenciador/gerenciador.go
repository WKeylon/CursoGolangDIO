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
	err := database.DB.Preload("Cidades").Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, errors.New("usuário não encontrado")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.SenhaHash), []byte(password))
	if err != nil {
		return nil, errors.New("senha incorreta")
	}

	return &user, nil
}

// CreateUser creates a new user with role validation
func CreateUser(creatorRole string, username, password, role string, cityIDs []uint) error {
	// Simple validation rules
	if creatorRole == modelos.RolePesquisador {
		return errors.New("pesquisadores não podem criar usuários")
	}

	if creatorRole == modelos.RoleGerente && role != modelos.RolePesquisador {
		return errors.New("gerentes só podem criar pesquisadores")
	}

	return database.CreateUser(username, password, role, cityIDs)
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

func AddCandidate(name, party string, cityID *uint) error {
	candidate := modelos.Candidato{
		Nome:     name,
		Partido:  party,
		CidadeID: cityID,
	}
	return database.DB.Create(&candidate).Error
}

func AddQuestion(text string) error {
	question := modelos.Pergunta{
		Texto: text,
	}
	return database.DB.Create(&question).Error
}

func ListCandidates(cityID uint) ([]modelos.Candidato, error) {
	var candidates []modelos.Candidato
	// List candidates specific to the city OR global (null cityID)
	// Important: We need to filter by CidadeID = ? OR CidadeID IS NULL
	err := database.DB.Where("cidade_id = ? OR cidade_id IS NULL", cityID).Find(&candidates).Error
	return candidates, err
}

func ListQuestions() ([]modelos.Pergunta, error) {
	var questions []modelos.Pergunta
	err := database.DB.Find(&questions).Error
	return questions, err
}

func RegisterVote(candidateID, cityID, userID uint) error {
	vote := modelos.Voto{
		CandidatoID: candidateID,
		CidadeID:    cityID,
		UsuarioID:   userID,
	}
	return database.DB.Create(&vote).Error
}

func RegisterAnswer(questionID, cityID, userID uint, answer bool) error {
	resp := modelos.Resposta{
		PerguntaID: questionID,
		Sim:        answer,
		CidadeID:   cityID,
		UsuarioID:  userID,
	}
	return database.DB.Create(&resp).Error
}

// Statistics

func GetCandidateVotes(cityID *uint) (map[string]int64, error) {
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

func GetQuestionStats(cityID *uint) (map[string]map[string]int64, error) {
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
