package parser

import (
	"fmt"
	"strconv"

	"github.com/khevencolino/Solar/internal/lexer"
	"github.com/khevencolino/Solar/internal/utils"
)

// Precedencia define a precedência dos operadores
type Precedencia int

const (
	PRECEDENCIA_NENHUMA       Precedencia = iota
	PRECEDENCIA_COMPARACAO                // == != < > <= >=
	PRECEDENCIA_SOMA                      // + -
	PRECEDENCIA_MULTIPLICACAO             // * /
	PRECEDENCIA_POTENCIA                  // **
)

// Parser representa o analisador sintático
type Parser struct {
	tokens       []lexer.Token
	posicaoAtual int
}

// obterPrecedencia retorna a precedência de um operador
func (p *Parser) obterPrecedencia(tokenType lexer.TokenType) Precedencia {
	switch tokenType {
	case lexer.EQUAL, lexer.NOT_EQUAL, lexer.LESS, lexer.GREATER, lexer.LESS_EQUAL, lexer.GREATER_EQUAL:
		return PRECEDENCIA_COMPARACAO
	case lexer.PLUS, lexer.MINUS:
		return PRECEDENCIA_SOMA
	case lexer.MULTIPLY, lexer.DIVIDE:
		return PRECEDENCIA_MULTIPLICACAO
	case lexer.POWER:
		return PRECEDENCIA_POTENCIA
	default:
		return PRECEDENCIA_NENHUMA
	}
}

// ehAssociativoADireita verifica se o operador é associativo à direita
func (p *Parser) ehAssociativoADireita(tokenType lexer.TokenType) bool {
	return tokenType == lexer.POWER // ** é associativo à direita
}

// NovoParser cria um novo analisador sintático
func NovoParser(tokens []lexer.Token) *Parser {
	return &Parser{
		tokens:       tokens,
		posicaoAtual: 0,
	}
}

// AnalisarPrograma analisa um programa
func (p *Parser) AnalisarPrograma() ([]Expressao, error) {
	var statements []Expressao

	for !p.chegouAoFim() {
		statement, err := p.analisarStatement()
		if err != nil {
			return nil, err
		}
		statements = append(statements, statement)
	}

	if len(statements) == 0 {
		return nil, utils.NovoErro("programa vazio", 0, 0, "")
	}

	return statements, nil
}

// analisarStatement implementa a análise descendente recursiva para expressões
func (p *Parser) analisarStatement() (Expressao, error) {
	token := p.tokenAtual()

	// Verifica se é um comando "se"
	if token.Type == lexer.SE {
		return p.analisarComandoSe()
	}

	// Verifica se é uma atribuição
	if token.Type == lexer.IDENTIFIER {
		p.proximoToken() // consome o identificador

		if p.tokenAtual().Type == lexer.ASSIGN {
			p.proximoToken() // consome o operador de atribuição
			valor, err := p.analisarExpressao(PRECEDENCIA_NENHUMA)
			if err != nil {
				return nil, err
			}
			return &Atribuicao{Nome: token.Value, Valor: valor, Token: token}, nil
		} else {
			// Se não é atribuição, volta um token e analisa como expressão
			p.posicaoAtual--
			return p.analisarExpressao(PRECEDENCIA_NENHUMA)
		}
	}

	// Caso contrário, analisa como expressão
	return p.analisarExpressao(PRECEDENCIA_NENHUMA)
}

// analisarExpressao implementa precedência de operadores usando o algoritmo Pratt
func (p *Parser) analisarExpressao(precedenciaMinima Precedencia) (Expressao, error) {
	// Analisa o lado esquerdo (prefixo)
	esquerda, err := p.analisarPrefixo()
	if err != nil {
		return nil, err
	}

	// Processa operadores binários com precedência adequada
	for {
		tokenAtual := p.tokenAtual()

		// Se chegou ao fim ou não é um operador binário, para
		if tokenAtual.Type == lexer.EOF || tokenAtual.Type == lexer.RPAREN || tokenAtual.Type == lexer.LBRACE {
			break
		}

		precedenciaAtual := p.obterPrecedencia(tokenAtual.Type)

		// Se não é um operador binário ou a precedência é menor que a mínima, para
		if precedenciaAtual == PRECEDENCIA_NENHUMA || precedenciaAtual < precedenciaMinima {
			break
		}

		// Se é associativo à direita, incrementa a precedência para a recursão
		proximaPrecedencia := precedenciaAtual
		if !p.ehAssociativoADireita(tokenAtual.Type) {
			proximaPrecedencia++
		}

		// Consome o operador
		operadorToken := p.proximoToken()
		operador, err := p.tokenParaOperador(operadorToken)
		if err != nil {
			return nil, err
		}

		// Analisa o lado direito com a precedência apropriada
		direita, err := p.analisarExpressao(proximaPrecedencia)
		if err != nil {
			return nil, err
		}

		// Cria a operação binária
		esquerda = &OperacaoBinaria{
			OperandoEsquerdo: esquerda,
			Operador:         operador,
			OperandoDireito:  direita,
			Token:            operadorToken,
		}
	}

	return esquerda, nil
}

// analisarPrefixo analisa expressões prefixas (números, variáveis, expressões parentizadas)
func (p *Parser) analisarPrefixo() (Expressao, error) {
	token := p.proximoToken()

	switch token.Type {
	case lexer.NUMBER:
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

	case lexer.IDENTIFIER:
		return &Variavel{Nome: token.Value, Token: token}, nil

	case lexer.FUNCTION:
		// Chamada de função
		return p.analisarChamadaFuncao(token)

	case lexer.LPAREN:
		// Expressão parentizada
		expressao, err := p.analisarExpressao(PRECEDENCIA_NENHUMA)
		if err != nil {
			return nil, err
		}

		// Verifica parêntese fechando
		if err := p.verificarProximoToken(lexer.RPAREN); err != nil {
			return nil, err
		}

		return expressao, nil

	default:
		return nil, utils.NovoErro(
			"expressão inválida",
			token.Position.Line,
			token.Position.Column,
			fmt.Sprintf("esperado número, variável ou '(', encontrado '%s'", token.Value),
		)
	}
}

// tokenParaOperador converte um token em um TipoOperador
func (p *Parser) tokenParaOperador(token lexer.Token) (TipoOperador, error) {
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
	case lexer.EQUAL:
		return IGUALDADE, nil
	case lexer.NOT_EQUAL:
		return DIFERENCA, nil
	case lexer.LESS:
		return MENOR_QUE, nil
	case lexer.GREATER:
		return MAIOR_QUE, nil
	case lexer.LESS_EQUAL:
		return MENOR_IGUAL, nil
	case lexer.GREATER_EQUAL:
		return MAIOR_IGUAL, nil
	default:
		return 0, utils.NovoErro(
			"operador inválido",
			token.Position.Line,
			token.Position.Column,
			fmt.Sprintf("esperado operador (+, -, *, /, **, ==, !=, <, >, <=, >=), encontrado '%s'", token.Value),
		)
	}
}

// analisarChamadaFuncao analisa uma chamada de função
func (p *Parser) analisarChamadaFuncao(tokenFuncao lexer.Token) (Expressao, error) {
	// Espera parêntese de abertura
	if err := p.verificarProximoToken(lexer.LPAREN); err != nil {
		return nil, err
	}

	var argumentos []Expressao

	// Se não é um parêntese de fechamento, analisa argumentos
	if p.tokenAtual().Type != lexer.RPAREN {
		for {
			argumento, err := p.analisarExpressao(PRECEDENCIA_NENHUMA)
			if err != nil {
				return nil, err
			}
			argumentos = append(argumentos, argumento)

			// Se o próximo token é uma vírgula, consome e continua
			if p.tokenAtual().Type == lexer.COMMA {
				p.proximoToken()
			} else {
				break
			}
		}
	}

	// Espera parêntese de fechamento
	if err := p.verificarProximoToken(lexer.RPAREN); err != nil {
		return nil, err
	}

	return &ChamadaFuncao{
		Nome:       tokenFuncao.Value,
		Argumentos: argumentos,
		Token:      tokenFuncao,
	}, nil
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
		msg := fmt.Sprintf("esperado %s, encontrado %s", tipoEsperado, token.Type)
		if token.Type == lexer.EOF {
			msg += " — possível parêntese não fechado"
		}
		return utils.NovoErro("token inesperado", token.Position.Line, token.Position.Column, msg)
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

// analisarComandoSe analisa um comando if/else
func (p *Parser) analisarComandoSe() (Expressao, error) {
	tokenSe := p.proximoToken() // consome "se"

	// Analisa a condição
	condicao, err := p.analisarExpressao(PRECEDENCIA_NENHUMA)
	if err != nil {
		return nil, fmt.Errorf("erro ao analisar condição do 'se': %v", err)
	}

	// Espera '{'
	if err := p.verificarProximoToken(lexer.LBRACE); err != nil {
		return nil, fmt.Errorf("esperado '{' após condição do 'se': %v", err)
	}

	// Analisa o bloco do "se"
	blocoSe, err := p.analisarBloco()
	if err != nil {
		return nil, fmt.Errorf("erro ao analisar bloco 'se': %v", err)
	}

	// Verifica se há "senao"
	var blocoSenao *Bloco
	if p.tokenAtual().Type == lexer.SENAO {
		p.proximoToken() // consome "senao"

		// Espera '{'
		if err := p.verificarProximoToken(lexer.LBRACE); err != nil {
			return nil, err
		}

		blocoSenao, err = p.analisarBloco()
		if err != nil {
			return nil, err
		}
	}

	return &ComandoSe{
		Condicao:   condicao,
		BlocoSe:    blocoSe,
		BlocoSenao: blocoSenao,
		Token:      tokenSe,
	}, nil
}

// analisarBloco analisa um bloco de comandos
func (p *Parser) analisarBloco() (*Bloco, error) {
	var comandos []Expressao
	tokenInicio := p.tokenAtual()

	// Processa comandos até encontrar '}'
	for !p.chegouAoFim() && p.tokenAtual().Type != lexer.RBRACE {
		comando, err := p.analisarStatement()
		if err != nil {
			return nil, err
		}
		comandos = append(comandos, comando)
	}

	// Espera '}'
	if err := p.verificarProximoToken(lexer.RBRACE); err != nil {
		return nil, err
	}

	return &Bloco{
		Comandos: comandos,
		Token:    tokenInicio,
	}, nil
}
