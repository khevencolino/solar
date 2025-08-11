package interpreter

import (
	"fmt"
	"math"

	"github.com/khevencolino/Solar/internal/parser"
	"github.com/khevencolino/Solar/internal/utils"
)

type InterpreterBackend struct {
	variaveis map[string]int
}

func NewInterpreterBackend() *InterpreterBackend {
	return &InterpreterBackend{
		variaveis: make(map[string]int),
	}
}

func (i *InterpreterBackend) GetName() string      { return "Interpretador AST" }
func (i *InterpreterBackend) GetExtension() string { return "" }

func (i *InterpreterBackend) Compile(statements []parser.Expressao) error {
	fmt.Printf("🔍 Interpretando diretamente da AST...\n")

	var ultimoResultado interface{}

	for idx, stmt := range statements {
		fmt.Printf("--- Statement %d ---\n", idx+1)

		// Imprime a árvore (opcional - pode ser configurável)
		visualizador := parser.NovoVisualizador()
		visualizador.ImprimirArvore(stmt)

		// Interpreta
		resultado, err := i.interpretar(stmt)
		if err != nil {
			return err
		}

		fmt.Printf("Resultado: %d\n", resultado)
		ultimoResultado = resultado
	}

	fmt.Printf("\n✅ Interpretação concluída! Resultado final: %d\n", ultimoResultado)
	return nil
}

// Interpretar executa uma expressão e retorna o resultado
func (i *InterpreterBackend) interpretar(expressao parser.Expressao) (interface{}, error) {
	resultado := expressao.Aceitar(i)
	if erro, ok := resultado.(error); ok {
		return 0, erro
	}
	return resultado, nil
}

// Implementa interface Node (visitor pattern)
func (i *InterpreterBackend) Constante(constante *parser.Constante) interface{} {
	return constante.Valor
}

func (i *InterpreterBackend) Variavel(variavel *parser.Variavel) interface{} {
	valor, existe := i.variaveis[variavel.Nome]
	if !existe {
		return utils.NovoErro(
			fmt.Sprintf("variável '%s' não definida", variavel.Nome),
			variavel.Token.Position.Line,
			variavel.Token.Position.Column,
			"",
		)
	}
	return valor
}

func (i *InterpreterBackend) Atribuicao(atribuicao *parser.Atribuicao) interface{} {
	// Avalia o valor da expressão
	valorInterface := atribuicao.Valor.Aceitar(i)
	if erro, ok := valorInterface.(error); ok {
		return erro
	}
	valor := valorInterface.(int)

	// Armazena na tabela de símbolos
	i.variaveis[atribuicao.Nome] = valor
	return valor
}

func (i *InterpreterBackend) OperacaoBinaria(operacao *parser.OperacaoBinaria) interface{} {
	// Interpreta operando esquerdo
	esquerdoInterface := operacao.OperandoEsquerdo.Aceitar(i)
	if erro, ok := esquerdoInterface.(error); ok {
		return erro
	}
	esquerdo := esquerdoInterface.(int)

	// Interpreta operando direito
	direitoInterface := operacao.OperandoDireito.Aceitar(i)
	if erro, ok := direitoInterface.(error); ok {
		return erro
	}
	direito := direitoInterface.(int)

	// Executa operação
	switch operacao.Operador {
	case parser.ADICAO:
		return esquerdo + direito
	case parser.SUBTRACAO:
		return esquerdo - direito
	case parser.MULTIPLICACAO:
		return esquerdo * direito
	case parser.DIVISAO:
		if direito == 0 {
			return utils.NovoErro(
				"divisão por zero",
				operacao.Token.Position.Line,
				operacao.Token.Position.Column,
				"",
			)
		}
		return esquerdo / direito
	case parser.POWER:
		return int(math.Pow(float64(esquerdo), float64(direito)))
	default:
		return utils.NovoErro(
			"operador desconhecido",
			operacao.Token.Position.Line,
			operacao.Token.Position.Column,
			"",
		)
	}
}
