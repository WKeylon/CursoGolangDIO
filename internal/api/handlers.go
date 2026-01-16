package api

import (
	"net/http"
	"patrulha_rural/internal/auth"
	"patrulha_rural/internal/database"
	"patrulha_rural/internal/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Login Handler
func Login(c *gin.Context) {
	var input struct {
		Email string `json:"email"`
		Senha string `json:"senha"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.Usuario
	// Checking DB connection
	if database.DB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database not connected"})
		return
	}

	if err := database.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Simplification: Plain text password comparison as per SQL insert example
	if user.Senha != input.Senha {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token, err := auth.GenerateToken(user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token, "user": user})
}

// --- CRUD Fazenda ---

func GetFazendas(c *gin.Context) {
	var fazendas []models.Fazenda
	database.DB.Preload("Proprietario").Find(&fazendas)
	c.JSON(http.StatusOK, fazendas)
}

func GetFazenda(c *gin.Context) {
	id := c.Param("id")
	var fazenda models.Fazenda
	if err := database.DB.Preload("Proprietario").First(&fazenda, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Fazenda not found"})
		return
	}
	c.JSON(http.StatusOK, fazenda)
}

func CreateFazenda(c *gin.Context) {
	var fazenda models.Fazenda
	if err := c.ShouldBindJSON(&fazenda); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Note: Coordenadas is a string, expected to be in correct format or handled by caller if raw SQL is used.
	// GORM will try to insert the string into the POINT column.
	// If the string is "POINT(X Y)", MySQL might complain if we bind it as string literal to a geometry column without ST_GeomFromText.
	// But let's assume standard GORM behavior for now. If this fails, we'd need a BeforeCreate hook.

	if err := database.DB.Create(&fazenda).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, fazenda)
}

func UpdateFazenda(c *gin.Context) {
	id := c.Param("id")
	var fazenda models.Fazenda
	if err := database.DB.First(&fazenda, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Fazenda not found"})
		return
	}

	var input models.Fazenda
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := database.DB.Model(&fazenda).Updates(input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, fazenda)
}

func DeleteFazenda(c *gin.Context) {
	id := c.Param("id")
	if err := database.DB.Delete(&models.Fazenda{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Fazenda deleted"})
}

// --- CRUD Pessoa ---

func GetPessoas(c *gin.Context) {
	var pessoas []models.Pessoa
	database.DB.Find(&pessoas)
	c.JSON(http.StatusOK, pessoas)
}

func CreatePessoa(c *gin.Context) {
	var pessoa models.Pessoa
	if err := c.ShouldBindJSON(&pessoa); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := database.DB.Create(&pessoa).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, pessoa)
}

// --- CRUD Visitas ---

func GetVisitas(c *gin.Context) {
	var visitas []models.Visita
	// Handling query param for filtering by Fazenda
	fazendaID := c.Query("id_fazenda")
	query := database.DB.Preload("Fazenda").Preload("Usuario")
	if fazendaID != "" {
		query = query.Where("id_fazenda = ?", fazendaID)
	}
	query.Find(&visitas)
	c.JSON(http.StatusOK, visitas)
}

func CreateVisita(c *gin.Context) {
	var visita models.Visita
	if err := c.ShouldBindJSON(&visita); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Auto set IDUsuario from token
	userID, exists := c.Get("userID")
	if exists {
		uid := userID.(uint)
		visita.IDUsuario = &uid
	}

	if err := database.DB.Create(&visita).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, visita)
}

// --- CRUD OPM ---
func GetOPMs(c *gin.Context) {
	var opms []models.OPM
	database.DB.Find(&opms)
	c.JSON(http.StatusOK, opms)
}

// --- Dashboard Stats ---
func GetStats(c *gin.Context) {
	var countFazendas int64
	var countVisitas int64
	var countPessoas int64

	database.DB.Model(&models.Fazenda{}).Count(&countFazendas)
	database.DB.Model(&models.Visita{}).Count(&countVisitas)
	database.DB.Model(&models.Pessoa{}).Count(&countPessoas)

	c.JSON(http.StatusOK, gin.H{
		"fazendas": countFazendas,
		"visitas":  countVisitas,
		"pessoas":  countPessoas,
	})
}

// RegisterRoutes registers all routes
func RegisterRoutes(r *gin.Engine) {
	r.POST("/login", Login)

	api := r.Group("/api")
	api.Use(auth.AuthMiddleware())
	{
		api.GET("/fazendas", GetFazendas)
		api.GET("/fazendas/:id", GetFazenda)
		api.POST("/fazendas", CreateFazenda)
		api.PUT("/fazendas/:id", UpdateFazenda)
		api.DELETE("/fazendas/:id", DeleteFazenda)

		api.GET("/pessoas", GetPessoas)
		api.POST("/pessoas", CreatePessoa)

		api.GET("/visitas", GetVisitas)
		api.POST("/visitas", CreateVisita)

		api.GET("/opms", GetOPMs)

		api.GET("/stats", GetStats)
	}

	// Setup static file serving for photos if needed, or upload handler
	api.POST("/upload", func(c *gin.Context) {
		file, _ := c.FormFile("file")
		// logic to save file
		dst := "./uploads/" + file.Filename
		// c.SaveUploadedFile(file, dst)
		// save to DB...
		c.JSON(http.StatusOK, gin.H{"url": dst})
	})
}

// Helper for string to int
func atoi(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
