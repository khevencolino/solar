package parser

import (
	"fmt"

	"github.com/khevencolino/Solar/internal/lexer"
)

// Node define a interface
type Node interface {
	Constante(constante *Constante) interface{}
	OperacaoBinaria(operacao *OperacaoBinaria) interface{}
	Variavel(variavel *Variavel) interface{}
	Atribuicao(atribuicao *Atribuicao) interface{}
	ChamadaFuncao(chamada *ChamadaFuncao) interface{}
	ComandoSe(comando *ComandoSe) interface{}
	Bloco(bloco *Bloco) interface{}
}

// Expressao representa a interface base para todos os nós da AST
type Expressao interface {
	Aceitar(node Node) any
	String() string
}

// Constante representa um literal inteiro na árvore
type Constante struct {
	Valor int
	Token lexer.Token
}

// Aceitar implementa o padrão  para Constante
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

// Aceitar implementa o padrão para OperacaoBinaria
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
	// Operadores de comparação
	IGUALDADE
	DIFERENCA
	MENOR_QUE
	MAIOR_QUE
	MENOR_IGUAL
	MAIOR_IGUAL
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
	case IGUALDADE:
		return "=="
	case DIFERENCA:
		return "!="
	case MENOR_QUE:
		return "<"
	case MAIOR_QUE:
		return ">"
	case MENOR_IGUAL:
		return "<="
	case MAIOR_IGUAL:
		return ">="
	default:
		return "?"
	}
}

// Variavel representa uma variavel na árvore
type Variavel struct {
	Nome  string
	Token lexer.Token
}

func (v *Variavel) Aceitar(node Node) any {
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

// ChamadaFuncao representa uma chamada de função na árvore
type ChamadaFuncao struct {
	Nome       string
	Argumentos []Expressao
	Token      lexer.Token
}

func (c *ChamadaFuncao) Aceitar(node Node) interface{} {
	return node.ChamadaFuncao(c)
}

func (c *ChamadaFuncao) String() string {
	args := ""
	for i, arg := range c.Argumentos {
		if i > 0 {
			args += ", "
		}
		args += arg.String()
	}
	return fmt.Sprintf("%s(%s)", c.Nome, args)
}

// ComandoSe representa um comando if/else na árvore
type ComandoSe struct {
	Condicao   Expressao
	BlocoSe    *Bloco
	BlocoSenao *Bloco // pode ser nil se não há else
	Token      lexer.Token
}

func (c *ComandoSe) Aceitar(node Node) interface{} {
	return node.ComandoSe(c)
}

func (c *ComandoSe) String() string {
	str := fmt.Sprintf("se (%s) %s", c.Condicao.String(), c.BlocoSe.String())
	if c.BlocoSenao != nil {
		str += fmt.Sprintf(" senao %s", c.BlocoSenao.String())
	}
	return str
}

// Bloco representa um bloco de comandos na árvore
type Bloco struct {
	Comandos []Expressao
	Token    lexer.Token
}

func (b *Bloco) Aceitar(node Node) interface{} {
	return node.Bloco(b)
}

func (b *Bloco) String() string {
	comandosStr := ""
	for i, comando := range b.Comandos {
		if i > 0 {
			comandosStr += "; "
		}
		comandosStr += comando.String()
	}
	return fmt.Sprintf("{ %s }", comandosStr)
}
