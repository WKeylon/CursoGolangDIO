package models

import (
	"time"
)

type OPM struct {
	ID        uint   `gorm:"primaryKey;column:id_opm" json:"id_opm"`
	Nome      string `gorm:"column:nome;not null" json:"nome"`
	Municipio string `gorm:"column:municipio;not null" json:"municipio"`
	Estado    string `gorm:"column:estado;not null;size:2" json:"estado"`
}

func (OPM) TableName() string {
	return "OPM"
}

type Usuario struct {
	ID             uint   `gorm:"primaryKey;column:id_usuario" json:"id_usuario"`
	Nome           string `gorm:"column:nome;not null" json:"nome"`
	Email          string `gorm:"column:email;unique;not null" json:"email"`
	Senha          string `gorm:"column:senha;not null" json:"-"` // Don't return password
	Matricula      string `gorm:"column:matricula;unique" json:"matricula"`
	PostoGraduacao string `gorm:"column:posto_graduacao" json:"posto_graduacao"`
	Funcao         string `gorm:"column:funcao" json:"funcao"`
	IDOPM          *uint  `gorm:"column:id_opm" json:"id_opm"`
	OPM            OPM    `gorm:"foreignKey:IDOPM" json:"opm,omitempty"`
}

func (Usuario) TableName() string {
	return "Usuarios"
}

type Pessoa struct {
	ID             uint       `gorm:"primaryKey;column:id_pessoa" json:"id_pessoa"`
	Nome           string     `gorm:"column:nome;not null" json:"nome"`
	CPF            string     `gorm:"column:cpf;size:14" json:"cpf"`
	Email          string     `gorm:"column:email" json:"email"`
	Telefone       string     `gorm:"column:telefone;size:20" json:"telefone"`
	DataNascimento *time.Time `gorm:"column:data_nascimento;type:date" json:"data_nascimento"`
	Funcao         string     `gorm:"column:funcao;size:150" json:"funcao"`
}

func (Pessoa) TableName() string {
	return "Pessoa"
}

type Fazenda struct {
	ID              uint   `gorm:"primaryKey;column:id_fazenda" json:"id_fazenda"`
	NomePropriedade string `gorm:"column:nome_propriedade;not null" json:"nome_propriedade"`
	MatriculaImovel string `gorm:"column:matricula_imovel;size:50" json:"matricula_imovel"`
	Atividade       string `gorm:"column:atividade;size:100" json:"atividade"`
	Municipio       string `gorm:"column:municipio;size:100" json:"municipio"`
	Estado          string `gorm:"column:estado;size:2" json:"estado"`
	// Coordenadas treated as string (WKT) for simplicity in this REST API
	Coordenadas         string `gorm:"column:coordenadas;type:POINT" json:"coordenadas"`
	Referencia          string `gorm:"column:referencia;type:text" json:"referencia"`
	NumPlacaPropriedade string `gorm:"column:num_placa_propriedade;size:7" json:"num_placa_propriedade"`
	PistaAviao          bool   `gorm:"column:pista_aviao;default:false" json:"pista_aviao"`
	InternetSSID        string `gorm:"column:internet_ssid;size:50" json:"internet_ssid"`
	InternetSenha       string `gorm:"column:internet_senha;size:50" json:"internet_senha"`
	Observacoes         string `gorm:"column:observacoes;type:text" json:"observacoes"`
	IDProprietario      *uint  `gorm:"column:id_proprietario" json:"id_proprietario"`
	Proprietario        Pessoa `gorm:"foreignKey:IDProprietario" json:"proprietario,omitempty"`
	DefensivoAgricola   bool   `gorm:"column:defensivo_agricola" json:"defensivo_agricola"`
	Descricao           string `gorm:"column:descricao;type:text" json:"descricao"`
}

func (Fazenda) TableName() string {
	return "Fazenda"
}

type Objeto struct {
	ID        uint    `gorm:"primaryKey;column:id_objeto" json:"id_objeto"`
	IDFazenda uint    `gorm:"column:id_fazenda;not null" json:"id_fazenda"`
	Descricao string  `gorm:"column:descricao;size:100" json:"descricao"`
	Placa     string  `gorm:"column:placa;size:20" json:"placa"`
	Chassi    string  `gorm:"column:chassi;size:50" json:"chassi"`
	Fazenda   Fazenda `gorm:"foreignKey:IDFazenda" json:"fazenda,omitempty"`
}

func (Objeto) TableName() string {
	return "Objetos"
}

type Visita struct {
	ID               uint      `gorm:"primaryKey;column:id_visita" json:"id_visita"`
	IDFazenda        uint      `gorm:"column:id_fazenda;not null" json:"id_fazenda"`
	IDUsuario        *uint     `gorm:"column:id_usuario" json:"id_usuario"`
	DataUltimaVisita time.Time `gorm:"column:data_ultima_visita;default:CURRENT_TIMESTAMP" json:"data_ultima_visita"`
	DadosGerais      string    `gorm:"column:dados_gerais;type:text" json:"dados_gerais"`
	Descricao        string    `gorm:"column:descricao;type:text" json:"descricao"`
	Fazenda          Fazenda   `gorm:"foreignKey:IDFazenda" json:"fazenda,omitempty"`
	Usuario          Usuario   `gorm:"foreignKey:IDUsuario" json:"usuario,omitempty"`
}

func (Visita) TableName() string {
	return "Visitas"
}

type Foto struct {
	ID         uint      `gorm:"primaryKey;column:id_foto" json:"id_foto"`
	URLCaminho string    `gorm:"column:url_caminho;not null" json:"url_caminho"`
	DataUpload time.Time `gorm:"column:data_upload;default:CURRENT_TIMESTAMP" json:"data_upload"`
	Descricao  string    `gorm:"column:descricao;size:150" json:"descricao"`
	IDUsuario  *uint     `gorm:"column:id_usuario" json:"id_usuario"`
	IDPessoa   *uint     `gorm:"column:id_pessoa" json:"id_pessoa"`
	IDFazenda  *uint     `gorm:"column:id_fazenda" json:"id_fazenda"`
	IDObjeto   *uint     `gorm:"column:id_objeto" json:"id_objeto"`
	IDVisita   *uint     `gorm:"column:id_visita" json:"id_visita"`
}

func (Foto) TableName() string {
	return "Fotos"
}
