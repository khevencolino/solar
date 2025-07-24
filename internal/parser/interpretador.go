package parser

import "github.com/khevencolino/Kite/internal/utils"

// Interpretador executa a árvore sintática
type Interpretador struct{}

// NovoInterpretador cria um novo interpretador
func NovoInterpretador() *Interpretador {
	return &Interpretador{}
}

// Interpretar executa uma expressão e retorna o resultado
func (i *Interpretador) Interpretar(expressao Expressao) (int, error) {
	resultado := expressao.Aceitar(i)
	if erro, ok := resultado.(error); ok {
		return 0, erro
	}
	return resultado.(int), nil
}

// VisitarConstante implementa visitor para constantes
func (i *Interpretador) VisitarConstante(constante *Constante) interface{} {
	return constante.Valor
}

// VisitarOperacaoBinaria implementa visitor para operações binárias
func (i *Interpretador) VisitarOperacaoBinaria(operacao *OperacaoBinaria) interface{} {
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
	case ADICAO:
		return esquerdo + direito
	case SUBTRACAO:
		return esquerdo - direito
	case MULTIPLICACAO:
		return esquerdo * direito
	case DIVISAO:
		if direito == 0 {
			return utils.NovoErro(
				"divisão por zero",
				operacao.Token.Position.Line,
				operacao.Token.Position.Column,
				"",
			)
		}
		return esquerdo / direito
	default:
		return utils.NovoErro(
			"operador desconhecido",
			operacao.Token.Position.Line,
			operacao.Token.Position.Column,
			"",
		)
	}
}
