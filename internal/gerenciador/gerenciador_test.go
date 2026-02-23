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
	db.AutoMigrate(
		&modelos.Usuario{},
		&modelos.Permissao{},
		&modelos.Cidade{},
		&modelos.Candidato{},
		&modelos.Voto{},
		&modelos.Pergunta{},
		&modelos.Resposta{},
	)
	database.DB = db

	database.DB.Create(&modelos.Cidade{Nome: "Palmas"})
	database.DB.Create(&modelos.Cidade{Nome: "Gurupi"})
}

func TestGranularPermissions(t *testing.T) {
	setupTestDB()

	permsAdmin := []modelos.Permissao{}
	database.CreateUser("admin", "admin", modelos.RoleAdmin, nil, permsAdmin)
	admin, _ := Authenticate("admin", "admin")

	// Create a "Reader" user who can Read Candidates but NOT Add
	permsReader := []modelos.Permissao{
		{Area: modelos.AreaCandidatos, Ler: true, Editar: false},
	}
	err := CreateUser(admin, "reader", "pass", modelos.RolePesquisador, nil, permsReader)
	if err != nil { t.Fatal(err) }

	reader, _ := Authenticate("reader", "pass")

	// Reader tries to add candidate -> Fail
	err = AddCandidate(reader, "Test", "T", nil)
	if err == nil {
		t.Error("Reader should NOT be able to add candidates")
	} else if err.Error() != "permissão negada para adicionar candidato" {
		t.Errorf("Unexpected error: %v", err)
	}

	// Create "Editor" user
	permsEditor := []modelos.Permissao{
		{Area: modelos.AreaCandidatos, Editar: true},
	}
	CreateUser(admin, "editor", "pass", modelos.RolePesquisador, nil, permsEditor)
	editor, _ := Authenticate("editor", "pass")

	// Editor tries to add candidate -> Success
	err = AddCandidate(editor, "Test2", "T2", nil)
	if err != nil {
		t.Errorf("Editor should be able to add candidate: %v", err)
	}
}
