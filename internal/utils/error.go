package utils

import "fmt"

// CompilerError representa um erro do compilador com informações de posição
type CompilerError struct {
	Mensagem string // Mensagem de erro
	Linha    int    // Linha onde ocorreu o erro
	Coluna   int    // Coluna onde ocorreu o erro
	Detalhes string // Detalhes adicionais do erro
}

// Error implementa a interface error
func (e *CompilerError) Error() string {
	if e.Linha > 0 && e.Coluna > 0 {
		return fmt.Sprintf("%s em linha %d, coluna %d", e.Mensagem, e.Linha, e.Coluna)
	}
	return e.Mensagem
}

// NovoErro cria um novo erro do compilador
func NovoErro(mensagem string, linha, coluna int, detalhes string) *CompilerError {
	return &CompilerError{
		Mensagem: mensagem,
		Linha:    linha,
		Coluna:   coluna,
		Detalhes: detalhes,
	}
}
