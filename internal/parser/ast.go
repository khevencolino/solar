package parser

import (
	"fmt"

	"github.com/khevencolino/Kite/internal/lexer"
)

// Expressao representa a interface base para todos os nós da AST
type Expressao interface {
	Aceitar(visitante Visitante) interface{}
	String() string
}

// Constante representa um literal inteiro na árvore
type Constante struct {
	Valor int
	Token lexer.Token
}

// Aceitar implementa o padrão visitor para Constante
func (c *Constante) Aceitar(visitante Visitante) interface{} {
	return visitante.VisitarConstante(c)
}

// String retorna representação em string da constante
func (c *Constante) String() string {
	return fmt.Sprintf("%d", c.Valor)
}

// OperacaoBinaria representa uma operação binária na árvore
type OperacaoBinaria struct {
	OperandoEsquerdo Expressao
	Operador         TipoOperador
	OperandoDireito  Expressao
	Token            lexer.Token
}

// Aceitar implementa o padrão visitor para OperacaoBinaria
func (o *OperacaoBinaria) Aceitar(visitante Visitante) interface{} {
	return visitante.VisitarOperacaoBinaria(o)
}

// String retorna representação em string da operação
func (o *OperacaoBinaria) String() string {
	return fmt.Sprintf("(%s %s %s)",
		o.OperandoEsquerdo.String(),
		o.Operador.String(),
		o.OperandoDireito.String())
}

// TipoOperador representa os tipos de operadores
type TipoOperador int

const (
	ADICAO TipoOperador = iota
	SUBTRACAO
	MULTIPLICACAO
	DIVISAO // FALTA POWER add dps
)

// String retorna representação em string do operador
func (t TipoOperador) String() string {
	switch t {
	case ADICAO:
		return "+"
	case SUBTRACAO:
		return "-"
	case MULTIPLICACAO:
		return "*"
	case DIVISAO:
		return "/"
	default:
		return "?"
	}
}

// Visitante define a interface para o padrão visitor
type Visitante interface {
	VisitarConstante(constante *Constante) interface{}
	VisitarOperacaoBinaria(operacao *OperacaoBinaria) interface{}
}
