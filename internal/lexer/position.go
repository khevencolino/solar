package lexer

import "fmt"

// Position representa uma posição no código fonte
type Position struct {
	Line   int // Linha no código
	Column int // Coluna no código
	Offset int // Posição absoluta no arquivo
}

// String retorna uma representação em string da posição
func (p Position) String() string {
	return fmt.Sprintf("linha %d, coluna %d", p.Line, p.Column)
}

// NovaPosicao cria uma nova posição
func NovaPosicao(linha, coluna, offset int) Position {
	return Position{
		Line:   linha,
		Column: coluna,
		Offset: offset,
	}
}
