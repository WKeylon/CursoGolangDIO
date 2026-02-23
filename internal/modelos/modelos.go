package modelos

import (
	"gorm.io/gorm"
)

// Roles
const (
	RoleAdmin       = "admin"
	RoleGerente     = "gerente"
	RolePesquisador = "pesquisador"
)

// Áreas do Sistema
const (
	AreaCandidatos   = "Candidatos"
	AreaPerguntas    = "Perguntas"
	AreaUsuarios     = "Usuarios"
	AreaEstatisticas = "Estatisticas"
	AreaVotacao      = "Votacao"
)

type Cidade struct {
	ID   uint   `gorm:"primaryKey"`
	Nome string `gorm:"uniqueIndex"`
}

type Permissao struct {
	gorm.Model
	UsuarioID uint
	Area      string // Ex: "Candidatos"
	Ler       bool
	Editar    bool // Criar e Atualizar
	Deletar   bool
}

type Usuario struct {
	gorm.Model
	Username       string `gorm:"uniqueIndex"`
	SenhaHash      string
	Role           string
	PrimeiroAcesso bool
	Cidades        []Cidade    `gorm:"many2many:usuario_cidades;"` // Cidades permitidas para este usuário
	Permissoes     []Permissao `gorm:"foreignKey:UsuarioID"`
}

type Candidato struct {
	gorm.Model
	Nome     string
	Partido  string
	CidadeID *uint // Opcional: Se nulo, pode ser estadual/federal
	Cidade   *Cidade
}

type Pergunta struct {
	gorm.Model
	Texto string
}

type Voto struct {
	gorm.Model
	CandidatoID uint
	Candidato   Candidato
	CidadeID    uint
	Cidade      Cidade
	UsuarioID   uint
	Usuario     Usuario
}

type Resposta struct {
	gorm.Model
	PerguntaID uint
	Pergunta   Pergunta
	Sim        bool // true = Sim, false = Não
	CidadeID   uint
	Cidade     Cidade
	UsuarioID  uint
	Usuario    Usuario
}
