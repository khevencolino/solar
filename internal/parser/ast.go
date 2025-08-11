package parser

import (
	"fmt"

	"github.com/khevencolino/Solar/internal/lexer"
)

// Expressao representa a interface base para todos os nós da AST
type Expressao interface {
	Aceitar(node Node) any
	String() string
}

// Variavel representa uma variavel na árvore
type Variavel struct {
	Nome  string
	Token lexer.Token
}

func (v *Variavel) Aceitar(node Node) interface{} {
	return node.Variavel(v)
}

func (v *Variavel) String() string {
	return v.Nome
}

// Atribuicao representa uma atribuicao na árvore
type Atribuicao struct {
	Nome  string
	Valor Expressao
	Token lexer.Token
}

func (a *Atribuicao) Aceitar(node Node) interface{} {
	return node.Atribuicao(a)
}

func (a *Atribuicao) String() string {
	return fmt.Sprintf("%s = %s", a.Nome, a.Valor.String())
}

// Constante representa um literal inteiro na árvore
type Constante struct {
	Valor int
	Token lexer.Token
}

// Aceitar implementa o padrão visitor para Constante
func (c *Constante) Aceitar(node Node) interface{} {
	return node.Constante(c)
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
func (o *OperacaoBinaria) Aceitar(node Node) interface{} {
	return node.OperacaoBinaria(o)
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
	DIVISAO
	POWER
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
	case POWER:
		return "**"
	default:
		return "?"
	}
}

// Node define a interface
type Node interface {
	Constante(constante *Constante) interface{}
	OperacaoBinaria(operacao *OperacaoBinaria) interface{}
	Variavel(variavel *Variavel) interface{}
	Atribuicao(atribuicao *Atribuicao) interface{}
}
