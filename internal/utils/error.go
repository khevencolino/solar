package utils

import (
	"fmt"
	"strings"
)

// CompilerError representa um erro do compilador com informações de posição
type CompilerError struct {
	Mensagem string // Mensagem de erro
	Linha    int    // Linha onde ocorreu o erro
	Coluna   int    // Coluna onde ocorreu o erro
	Detalhes string // Detalhes adicionais do erro
}

// Error implementa a interface error de forma otimizada
func (e *CompilerError) Error() string {
	if e.Linha > 0 && e.Coluna > 0 {
		var builder strings.Builder
		builder.WriteString(e.Mensagem)
		builder.WriteString(" em linha ")
		builder.WriteString(fmt.Sprintf("%d", e.Linha))
		builder.WriteString(", coluna ")
		builder.WriteString(fmt.Sprintf("%d", e.Coluna))
		if e.Detalhes != "" {
			builder.WriteString(" (")
			builder.WriteString(e.Detalhes)
			builder.WriteString(")")
		}
		return builder.String()
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
