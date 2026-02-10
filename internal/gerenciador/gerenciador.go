package gerenciador

import (
	"pesquisa-eleitoral/internal/modelos"
)

type Gerenciador struct {
	sistema    *modelos.SistemaEleitoral
	nextCandID int
	nextPergID int
}

func NovoGerenciador() *Gerenciador {
	return &Gerenciador{
		sistema: &modelos.SistemaEleitoral{
			Candidatos: []*modelos.Candidato{},
			Perguntas:  []*modelos.Pergunta{},
		},
		nextCandID: 1,
		nextPergID: 1,
	}
}

func (g *Gerenciador) AdicionarCandidato(nome, partido string) *modelos.Candidato {
	c := &modelos.Candidato{
		ID:      g.nextCandID,
		Nome:    nome,
		Partido: partido,
		Votos:   0,
	}
	g.sistema.Candidatos = append(g.sistema.Candidatos, c)
	g.nextCandID++
	return c
}

func (g *Gerenciador) AdicionarPergunta(texto string) *modelos.Pergunta {
	p := &modelos.Pergunta{
		ID:    g.nextPergID,
		Texto: texto,
		Sim:   0,
		Nao:   0,
	}
	g.sistema.Perguntas = append(g.sistema.Perguntas, p)
	g.nextPergID++
	return p
}

func (g *Gerenciador) RegistrarVotoCandidato(id int) {
	for _, c := range g.sistema.Candidatos {
		if c.ID == id {
			c.Votos++
			break
		}
	}
}

func (g *Gerenciador) RegistrarRespostaPergunta(id int, respostaSim bool) {
	for _, p := range g.sistema.Perguntas {
		if p.ID == id {
			if respostaSim {
				p.Sim++
			} else {
				p.Nao++
			}
			break
		}
	}
}

func (g *Gerenciador) ListarCandidatos() []*modelos.Candidato {
	return g.sistema.Candidatos
}

func (g *Gerenciador) ListarPerguntas() []*modelos.Pergunta {
	return g.sistema.Perguntas
}
