package parser

import (
	"fmt"
	"strconv"

	"github.com/khevencolino/Kite/internal/lexer"
	"github.com/khevencolino/Kite/internal/utils"
)

// Parser representa o analisador sintático
type Parser struct {
	tokens       []lexer.Token
	posicaoAtual int
}

// NovoParser cria um novo analisador sintático
func NovoParser(tokens []lexer.Token) *Parser {
	return &Parser{
		tokens:       tokens,
		posicaoAtual: 0,
	}
}

// AnalisarPrograma analisa um programa completo
func (p *Parser) AnalisarPrograma() (Expressao, error) {
	expressao, err := p.analisarExpressao()
	if err != nil {
		return nil, err
	}

	// Verifica se chegou ao fim da entrada
	if !p.chegouAoFim() {
		token := p.tokenAtual()
		return nil, utils.NovoErro(
			"tokens inesperados após expressão",
			token.Position.Line,
			token.Position.Column,
			fmt.Sprintf("token: %s", token.Value),
		)
	}

	return expressao, nil
}

// analisarExpressao implementa a análise descendente recursiva para expressões
func (p *Parser) analisarExpressao() (Expressao, error) {
	token := p.proximoToken()

	if token.Type == lexer.NUMBER {
		// Caso: <expressao> ::= <literal-inteiro>
		valor, err := strconv.Atoi(token.Value)
		if err != nil {
			return nil, utils.NovoErro(
				"erro ao converter número",
				token.Position.Line,
				token.Position.Column,
				err.Error(),
			)
		}
		return &Constante{Valor: valor, Token: token}, nil

	} else if token.Type == lexer.LPAREN {
		// Caso: <expressao> ::= (<expressao> <operador> <expressao>)
		operandoEsquerdo, err := p.analisarExpressao()
		if err != nil {
			return nil, err
		}

		operador, err := p.analisarOperador()
		if err != nil {
			return nil, err
		}

		operandoDireito, err := p.analisarExpressao()
		if err != nil {
			return nil, err
		}

		// Verifica parêntese fechando
		if err := p.verificarProximoToken(lexer.RPAREN); err != nil {
			return nil, err
		}

		return &OperacaoBinaria{
			OperandoEsquerdo: operandoEsquerdo,
			Operador:         operador,
			OperandoDireito:  operandoDireito,
			Token:            token,
		}, nil

	} else {
		return nil, utils.NovoErro(
			"expressão inválida",
			token.Position.Line,
			token.Position.Column,
			fmt.Sprintf("esperado número ou '(', encontrado '%s'", token.Value),
		)
	}
}

// analisarOperador analisa operadores
func (p *Parser) analisarOperador() (TipoOperador, error) {
	token := p.proximoToken()

	switch token.Type {
	case lexer.PLUS:
		return ADICAO, nil
	case lexer.MINUS:
		return SUBTRACAO, nil
	case lexer.MULTIPLY:
		return MULTIPLICACAO, nil
	case lexer.POWER:
		return POWER, nil
	case lexer.DIVIDE:
		return DIVISAO, nil
	default:
		return 0, utils.NovoErro(
			"operador inválido",
			token.Position.Line,
			token.Position.Column,
			fmt.Sprintf("esperado operador (+, -, *, /), encontrado '%s'", token.Value),
		)
	}
}

// proximoToken retorna o próximo token e avança a posição
func (p *Parser) proximoToken() lexer.Token {
	if p.chegouAoFim() {
		// Retorna EOF se não há mais tokens
		return lexer.NovoToken(lexer.EOF, "", lexer.NovaPosicao(0, 0, 0))
	}

	token := p.tokens[p.posicaoAtual]
	p.posicaoAtual++
	return token
}

// verificarProximoToken verifica se o próximo token é do tipo esperado
func (p *Parser) verificarProximoToken(tipoEsperado lexer.TokenType) error {
	token := p.proximoToken()
	if token.Type != tipoEsperado {
		return utils.NovoErro(
			"token inesperado",
			token.Position.Line,
			token.Position.Column,
			fmt.Sprintf("esperado %s, encontrado %s", tipoEsperado, token.Type),
		)
	}
	return nil
}

// tokenAtual retorna o token atual sem avançar
func (p *Parser) tokenAtual() lexer.Token {
	if p.chegouAoFim() {
		return lexer.NovoToken(lexer.EOF, "", lexer.NovaPosicao(0, 0, 0))
	}
	return p.tokens[p.posicaoAtual]
}

// chegouAoFim verifica se chegou ao fim dos tokens
func (p *Parser) chegouAoFim() bool {
	return p.posicaoAtual >= len(p.tokens) ||
		(p.posicaoAtual < len(p.tokens) && p.tokens[p.posicaoAtual].Type == lexer.EOF)
}
