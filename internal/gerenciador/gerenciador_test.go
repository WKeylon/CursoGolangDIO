package gerenciador

import (
	"testing"
)

func TestGerenciador(t *testing.T) {
	g := NovoGerenciador()

	c := g.AdicionarCandidato("Alice", "Party A")
	if c.ID != 1 || c.Nome != "Alice" {
		t.Errorf("Expected candidate Alice with ID 1, got %v", c)
	}

	p := g.AdicionarPergunta("Is it sunny?")
	if p.ID != 1 || p.Texto != "Is it sunny?" {
		t.Errorf("Expected question Is it sunny? with ID 1, got %v", p)
	}

	g.RegistrarVotoCandidato(1)
	if g.ListarCandidatos()[0].Votos != 1 {
		t.Errorf("Expected 1 vote for candidate, got %d", g.ListarCandidatos()[0].Votos)
	}

	g.RegistrarRespostaPergunta(1, true)
	if g.ListarPerguntas()[0].Sim != 1 {
		t.Errorf("Expected 1 Sim for question, got %d", g.ListarPerguntas()[0].Sim)
	}
}
