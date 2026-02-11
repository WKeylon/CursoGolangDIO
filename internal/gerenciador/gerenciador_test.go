package gerenciador

import (
	"pesquisa-eleitoral/internal/database"
	"pesquisa-eleitoral/internal/modelos"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&modelos.Usuario{}, &modelos.Cidade{}, &modelos.Candidato{}, &modelos.Voto{}, &modelos.Pergunta{}, &modelos.Resposta{})
	database.DB = db

	database.DB.Create(&modelos.Cidade{Nome: "Palmas"})
	database.DB.Create(&modelos.Cidade{Nome: "Gurupi"})
}

func TestVotingFlow(t *testing.T) {
	setupTestDB()

	// 1. Create Admin
	CreateUser("system", "admin", "admin", modelos.RoleAdmin, nil)
	admin, _ := Authenticate("admin", "admin")

	// 2. Admin Creates Manager
	err := CreateUser(admin.Role, "manager", "pass", modelos.RoleGerente, nil)
	if err != nil {
		t.Errorf("Admin should be able to create Manager: %v", err)
	}

	// 3. Manager Creates Researcher for Palmas (ID 1)
	manager, _ := Authenticate("manager", "pass")
	err = CreateUser(manager.Role, "researcher", "pass", modelos.RolePesquisador, []uint{1})
	if err != nil {
		t.Errorf("Manager should be able to create Researcher: %v", err)
	}

	// 4. Researcher tries to create user (Should Fail)
	researcher, _ := Authenticate("researcher", "pass")
	err = CreateUser(researcher.Role, "hacker", "pass", modelos.RoleAdmin, nil)
	if err == nil {
		t.Error("Researcher should NOT be able to create users")
	}

	// 5. Add Candidate
	AddCandidate("Candidato 1", "P1", nil) // Global

	// 6. Researcher Votes in Palmas
	err = RegisterVote(1, 1, researcher.ID)
	if err != nil {
		t.Errorf("Vote failed: %v", err)
	}

	// 7. Stats
	votes, _ := GetCandidateVotes(nil) // Global
	if votes["Candidato 1"] != 1 {
		t.Errorf("Expected 1 vote, got %d", votes["Candidato 1"])
	}
}
