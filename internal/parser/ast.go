package parser

import (
	"fmt"

	"github.com/khevencolino/Solar/internal/lexer"
)

// Node define a interface
type Node interface {
	Constante(constante *Constante) interface{}
	Booleano(boolLit *Booleano) interface{}
	LiteralTexto(literal *LiteralTexto) interface{}
	LiteralDecimal(literal *LiteralDecimal) interface{}
	OperacaoBinaria(operacao *OperacaoBinaria) interface{}
	Variavel(variavel *Variavel) interface{}
	Atribuicao(atribuicao *Atribuicao) interface{}
	ChamadaFuncao(chamada *ChamadaFuncao) interface{}
	ComandoSe(comando *ComandoSe) interface{}
	ComandoEnquanto(cmd *ComandoEnquanto) interface{}
	ComandoPara(cmd *ComandoPara) interface{}
	Bloco(bloco *Bloco) interface{}
	FuncaoDeclaracao(fn *FuncaoDeclaracao) interface{}
	Retorno(ret *Retorno) interface{}
	Importacao(imp *Importacao) interface{}
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

// Booleano representa um literal booleano na árvore
type Booleano struct {
	Valor bool
	Token lexer.Token
}

func (b *Booleano) Aceitar(node Node) interface{} { return node.Booleano(b) }
func (b *Booleano) String() string {
	if b.Valor {
		return "verdadeiro"
	}
	return "falso"
}

// LiteralTexto representa um literal de string na árvore
type LiteralTexto struct {
	Valor string
	Token lexer.Token
}

func (lt *LiteralTexto) Aceitar(node Node) interface{} { return node.LiteralTexto(lt) }
func (lt *LiteralTexto) String() string {
	return fmt.Sprintf("\"%s\"", lt.Valor)
}

// LiteralDecimal representa um literal decimal na árvore
type LiteralDecimal struct {
	Valor float64
	Token lexer.Token
}

func (ld *LiteralDecimal) Aceitar(node Node) interface{} { return node.LiteralDecimal(ld) }
func (ld *LiteralDecimal) String() string {
	return fmt.Sprintf("%g", ld.Valor)
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
	Nome        string
	Valor       Expressao
	Token       lexer.Token
	TipoAnotado *Tipo
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

// Enquanto (while)
type ComandoEnquanto struct {
	Condicao Expressao
	Corpo    *Bloco
	Token    lexer.Token
}

func (e *ComandoEnquanto) Aceitar(node Node) interface{} { return node.ComandoEnquanto(e) }
func (e *ComandoEnquanto) String() string {
	return fmt.Sprintf("enquanto (%s) %s", e.Condicao.String(), e.Corpo.String())
}

// Para (for) com estilo C: init; cond; pos
type ComandoPara struct {
	Inicializacao Expressao // pode ser nil
	Condicao      Expressao // pode ser nil (trata como verdadeiro)
	PosIteracao   Expressao // pode ser nil
	Corpo         *Bloco
	Token         lexer.Token
}

func (p *ComandoPara) Aceitar(node Node) interface{} { return node.ComandoPara(p) }
func (p *ComandoPara) String() string {
	return fmt.Sprintf("para (%s; %s; %s) %s", strOr(p.Inicializacao), strOr(p.Condicao), strOr(p.PosIteracao), p.Corpo.String())
}

func strOr(e Expressao) string {
	if e == nil {
		return ""
	}
	return e.String()
}

// Tipagem simples
type Tipo int

const (
	TipoVazio    Tipo = iota // semelhante a void/null
	TipoInteiro              // inteiros (i64)
	TipoDecimal              // ponto flutuante (double)
	TipoTexto                // strings
	TipoBooleano             // booleano
)

func (t Tipo) String() string {
	switch t {
	case TipoVazio:
		return "vazio"
	case TipoInteiro:
		return "inteiro"
	case TipoDecimal:
		return "decimal"
	case TipoTexto:
		return "texto"
	case TipoBooleano:
		return "booleano"
	default:
		return "?"
	}
}

// ParametroFuncao representa um parâmetro de função com nome e tipo
type ParametroFuncao struct {
	Nome string
	Tipo Tipo
}

// FuncaoDeclaracao representa a declaração de uma função do usuário
type FuncaoDeclaracao struct {
	Nome       string
	Parametros []ParametroFuncao // Parâmetros com nome e tipo explícito
	Retorno    Tipo              // Tipo de retorno (default: TipoInteiro)
	Corpo      *Bloco
	Token      lexer.Token
}

func (f *FuncaoDeclaracao) Aceitar(node Node) any { return node.FuncaoDeclaracao(f) }

func (f *FuncaoDeclaracao) String() string {
	params := ""
	for i, param := range f.Parametros {
		if i > 0 {
			params += ", "
		}
		params += fmt.Sprintf("%s: %s", param.Nome, param.Tipo.String())
	}
	return fmt.Sprintf("definir %s(%s): %s %s", f.Nome, params, f.Retorno.String(), f.Corpo.String())
}

// Retorno representa um comando de retorno na árvore
type Retorno struct {
	Valor Expressao // pode ser nil para retorno vazio
	Token lexer.Token
}

func (r *Retorno) Aceitar(node Node) any { return node.Retorno(r) }

func (r *Retorno) String() string {
	if r.Valor == nil {
		return "retornar"
	}
	return fmt.Sprintf("retornar %s", r.Valor.String())
}

// Importacao representa uma declaração de importação na árvore
type Importacao struct {
	Simbolos []string // Lista de símbolos a importar (função/variável)
	Modulo   string   // Caminho do módulo/arquivo
	Token    lexer.Token
}

func (i *Importacao) Aceitar(node Node) any { return node.Importacao(i) }

func (i *Importacao) String() string {
	if len(i.Simbolos) == 1 {
		return fmt.Sprintf("importar %s de %s", i.Simbolos[0], i.Modulo)
	}
	return fmt.Sprintf("importar %v de %s", i.Simbolos, i.Modulo)
}
