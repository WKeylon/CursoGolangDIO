package modelos

type Candidato struct {
    ID      int
    Nome    string
    Partido string
    Votos   int
}

type Pergunta struct {
    ID      int
    Texto   string
    Sim     int
    Nao     int
}

type SistemaEleitoral struct {
    Candidatos []*Candidato
    Perguntas  []*Pergunta
}
