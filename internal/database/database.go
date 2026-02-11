package database

import (
	"log"
	"pesquisa-eleitoral/internal/modelos"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	var err error
	DB, err = gorm.Open(sqlite.Open("pesquisa.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database: ", err)
	}

	err = DB.AutoMigrate(
		&modelos.Usuario{},
		&modelos.Cidade{},
		&modelos.Candidato{},
		&modelos.Pergunta{},
		&modelos.Voto{},
		&modelos.Resposta{},
	)
	if err != nil {
		log.Fatal("failed to migrate database: ", err)
	}

	Seed()
}

func Seed() {
	// Seed Admin
	var admin modelos.Usuario
	result := DB.First(&admin, "username = ?", "admin")
	if result.Error != nil {
		hash, _ := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
		admin = modelos.Usuario{
			Username:       "admin",
			SenhaHash:      string(hash),
			Role:           modelos.RoleAdmin,
			PrimeiroAcesso: true,
		}
		DB.Create(&admin)
		log.Println("Admin user created with password 'admin'")
	}

	// Seed Cities
	cities := []string{
		"Palmas", "Araguaína", "Gurupi", "Porto Nacional", "Paraíso do Tocantins",
		"Araguatins", "Colinas do Tocantins", "Guaraí", "Tocantinópolis", "Dianópolis",
		"Augustinópolis", "Formoso do Araguaia", "Miracema do Tocantins", "Taguatinga", "Miranorte",
	}

	for _, name := range cities {
		var city modelos.Cidade
		if err := DB.FirstOrCreate(&city, modelos.Cidade{Nome: name}).Error; err != nil {
			log.Printf("Error creating city %s: %v", name, err)
		}
	}
}

// User Functions
func CreateUser(username, password, role string, cityIDs []uint) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := modelos.Usuario{
		Username:       username,
		SenhaHash:      string(hash),
		Role:           role,
		PrimeiroAcesso: true,
	}

	// Add cities
	if len(cityIDs) > 0 {
		var cities []modelos.Cidade
		DB.Find(&cities, cityIDs)
		user.Cidades = cities
	}

	return DB.Create(&user).Error
}

func GetUserByUsername(username string) (*modelos.Usuario, error) {
	var user modelos.Usuario
	err := DB.Preload("Cidades").First(&user, "username = ?", username).Error
	return &user, err
}

func UpdatePassword(userID uint, newPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return DB.Model(&modelos.Usuario{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"senha_hash":      string(hash),
		"primeiro_acesso": false,
	}).Error
}

func GetAllCities() ([]modelos.Cidade, error) {
	var cities []modelos.Cidade
	err := DB.Find(&cities).Error
	return cities, err
}

func GetAllUsers() ([]modelos.Usuario, error) {
	var users []modelos.Usuario
	err := DB.Preload("Cidades").Find(&users).Error
	return users, err
}
