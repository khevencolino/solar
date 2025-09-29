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

		// Estruturas de controle (se, enquanto, para) e declarações de função não precisam de semicolon
		if _, ok := statement.(*ComandoSe); ok {
			continue
		}
		if _, ok := statement.(*ComandoEnquanto); ok {
			continue
		}
		if _, ok := statement.(*ComandoPara); ok {
			continue
		}
		if _, ehFuncDecl := statement.(*FuncaoDeclaracao); ehFuncDecl {
			continue
		}
		if err := p.esperarSemicolon(); err != nil {
			return nil, err
		}
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

	// enquanto (cond) { bloco }
	if token.Type == lexer.ENQUANTO {
		return p.analisarComandoEnquanto()
	}

	// para (init; cond; pos) { bloco }
	if token.Type == lexer.PARA {
		return p.analisarComandoPara()
	}

	// Verifica se é uma declaração de função: definir nome(params) { bloco }
	if token.Type == lexer.DEFINIR {
		return p.analisarDeclaracaoFuncao()
	}

	// Verifica se é um retorno
	if token.Type == lexer.RETORNAR {
		return p.analisarRetorno()
	}

	// Verifica se é uma importação
	if token.Type == lexer.IMPORTAR {
		return p.analisarImportacao()
	}

	// Verifica se é uma atribuição (com ou sem anotação de tipo)
	if token.Type == lexer.IDENTIFIER {
		p.proximoToken() // consome o identificador

		// Suporta anotação de tipo: IDENT ':' tipo '~>' expr
		var tipoAnnot *Tipo
		if p.tokenAtual().Type == lexer.COLON {
			p.proximoToken()
			tTok := p.proximoToken()
			if tTok.Type != lexer.IDENTIFIER {
				return nil, utils.NovoErro("tipo inválido", tTok.Position.Line, tTok.Position.Column, "esperado identificador de tipo")
			}
			tp, err := p.parseTipoPorNome(tTok.Value)
			if err != nil {
				return nil, utils.NovoErro("tipo inválido", tTok.Position.Line, tTok.Position.Column, err.Error())
			}
			tipoAnnot = &tp
		}

		if p.tokenAtual().Type == lexer.ASSIGN {
			p.proximoToken() // consome o operador de atribuição
			valor, err := p.analisarExpressao(PRECEDENCIA_NENHUMA)
			if err != nil {
				return nil, err
			}
			return &Atribuicao{Nome: token.Value, Valor: valor, Token: token, TipoAnotado: tipoAnnot}, nil
		} else if p.tokenAtual().Type == lexer.LPAREN {
			return p.analisarChamadaFuncao(lexer.NovoToken(lexer.FUNCTION, token.Value, token.Position))
		} else {
			// Se não é atribuição nem chamada, volta um token e analisa como expressão
			p.posicaoAtual--
			return p.analisarExpressao(PRECEDENCIA_NENHUMA)
		}
	}

	// Caso contrário, analisa como expressão
	return p.analisarExpressao(PRECEDENCIA_NENHUMA)
}

// analisarRetorno: 'retornar' expressao? ';'
func (p *Parser) analisarRetorno() (Expressao, error) {
	tok := p.proximoToken() // consome 'retornar'
	var expr Expressao
	// retorno pode ser vazio: se próximo é ';' ou '}'
	if t := p.tokenAtual(); t.Type != lexer.SEMICOLON && t.Type != lexer.RBRACE {
		e, err := p.analisarExpressao(PRECEDENCIA_NENHUMA)
		if err != nil {
			return nil, err
		}
		expr = e
	}
	return &Retorno{Valor: expr, Token: tok}, nil
}

// analisarImportacao: 'importar' simbolos 'de' modulo ';'
// Suporta: importar imprime de io;
//
//	importar soma, mul de math;
func (p *Parser) analisarImportacao() (Expressao, error) {
	tok := p.proximoToken() // consome 'importar'

	// Parse símbolos a importar
	var simbolos []string

	// Primeiro símbolo é obrigatório
	if p.tokenAtual().Type != lexer.IDENTIFIER {
		return nil, fmt.Errorf("esperado identificador após 'importar' em %s", p.tokenAtual().Position)
	}
	simbolos = append(simbolos, p.tokenAtual().Value)
	p.proximoToken() // consome o identificador

	// Símbolos adicionais separados por vírgula
	for p.tokenAtual().Type == lexer.COMMA {
		p.proximoToken() // consome vírgula
		if p.tokenAtual().Type != lexer.IDENTIFIER {
			return nil, fmt.Errorf("esperado identificador após ',' em %s", p.tokenAtual().Position)
		}
		simbolos = append(simbolos, p.tokenAtual().Value)
		p.proximoToken() // consome identificador
	}

	// Espera 'de'
	if p.tokenAtual().Type != lexer.DE {
		return nil, fmt.Errorf("esperado 'de' após símbolos em %s", p.tokenAtual().Position)
	}
	p.proximoToken() // consome 'de'

	// Nome do módulo/arquivo
	if p.tokenAtual().Type != lexer.IDENTIFIER {
		return nil, fmt.Errorf("esperado nome do módulo após 'de' em %s", p.tokenAtual().Position)
	}
	modulo := p.tokenAtual().Value
	p.proximoToken() // consome módulo

	return &Importacao{
		Simbolos: simbolos,
		Modulo:   modulo,
		Token:    tok,
	}, nil
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
		if tokenAtual.Type == lexer.EOF || tokenAtual.Type == lexer.RPAREN || tokenAtual.Type == lexer.LBRACE || tokenAtual.Type == lexer.SEMICOLON {
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

	case lexer.FLOAT:
		valor, err := strconv.ParseFloat(token.Value, 64)
		if err != nil {
			return nil, utils.NovoErro(
				"erro ao converter número decimal",
				token.Position.Line,
				token.Position.Column,
				err.Error(),
			)
		}
		return &LiteralDecimal{Valor: valor, Token: token}, nil

	case lexer.STRING:
		// Remove as aspas do início e fim
		valor := token.Value[1 : len(token.Value)-1]
		return &LiteralTexto{Valor: valor, Token: token}, nil

	case lexer.VERDADEIRO:
		return &Booleano{Valor: true, Token: token}, nil
	case lexer.FALSO:
		return &Booleano{Valor: false, Token: token}, nil

	case lexer.MINUS:
		// Operador unário negativo
		proximo := p.proximoToken()
		switch proximo.Type {
		case lexer.NUMBER:
			valor, err := strconv.Atoi(proximo.Value)
			if err != nil {
				return nil, utils.NovoErro(
					"erro ao converter número negativo",
					proximo.Position.Line,
					proximo.Position.Column,
					err.Error(),
				)
			}
			return &Constante{Valor: -valor, Token: token}, nil
		case lexer.FLOAT:
			valor, err := strconv.ParseFloat(proximo.Value, 64)
			if err != nil {
				return nil, utils.NovoErro(
					"erro ao converter número decimal negativo",
					proximo.Position.Line,
					proximo.Position.Column,
					err.Error(),
				)
			}
			return &LiteralDecimal{Valor: -valor, Token: token}, nil
		default:
			// Se não for número, retorna erro
			return nil, utils.NovoErro(
				"operador negativo deve ser seguido de número",
				proximo.Position.Line,
				proximo.Position.Column,
				fmt.Sprintf("encontrado %s", proximo.Type.String()),
			)
		}

	case lexer.IDENTIFIER:
		// Pode ser variável ou início de chamada de função do usuário
		if p.tokenAtual().Type == lexer.LPAREN {
			return p.analisarChamadaFuncao(lexer.NovoToken(lexer.FUNCTION, token.Value, token.Position))
		}
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

// analisarDeclaracaoFuncao: 'definir' IDENT '(' params? ')' '{' bloco '}'
func (p *Parser) analisarDeclaracaoFuncao() (Expressao, error) {
	tokDef := p.proximoToken() // consumir 'definir'

	// nome da função
	nomeTok := p.proximoToken()
	if nomeTok.Type != lexer.IDENTIFIER {
		return nil, utils.NovoErro("nome de função inválido", nomeTok.Position.Line, nomeTok.Position.Column, "esperado identificador após 'definir'")
	}

	if err := p.verificarProximoToken(lexer.LPAREN); err != nil {
		return nil, err
	}

	var params []ParametroFuncao
	if p.tokenAtual().Type != lexer.RPAREN {
		for {
			idTok := p.proximoToken()
			if idTok.Type != lexer.IDENTIFIER {
				return nil, utils.NovoErro("parâmetro inválido", idTok.Position.Line, idTok.Position.Column, "esperado identificador de parâmetro")
			}

			paramNome := idTok.Value
			paramTipo := TipoInteiro // tipo padrão

			// Tipo opcional: nome: tipo (se não especificado, assume inteiro)
			if p.tokenAtual().Type == lexer.COLON {
				p.proximoToken() // consumir ':'

				tTok := p.proximoToken()
				if tTok.Type != lexer.IDENTIFIER {
					return nil, utils.NovoErro("tipo inválido", tTok.Position.Line, tTok.Position.Column, "esperado identificador de tipo")
				}
				tp, err := p.parseTipoPorNome(tTok.Value)
				if err != nil {
					return nil, utils.NovoErro("tipo inválido", tTok.Position.Line, tTok.Position.Column, err.Error())
				}
				paramTipo = tp
			}

			params = append(params, ParametroFuncao{Nome: paramNome, Tipo: paramTipo})

			if p.tokenAtual().Type == lexer.COMMA {
				p.proximoToken()
				continue
			}
			break
		}
	}

	if err := p.verificarProximoToken(lexer.RPAREN); err != nil {
		return nil, err
	}

	// Retorno opcional: ':' <tipo> (default: inteiro)
	var retorno Tipo = TipoInteiro
	if p.tokenAtual().Type == lexer.COLON {
		p.proximoToken()
		tTok := p.proximoToken()
		if tTok.Type != lexer.IDENTIFIER {
			return nil, utils.NovoErro("tipo inválido", tTok.Position.Line, tTok.Position.Column, "esperado identificador de tipo")
		}
		tp, err := p.parseTipoPorNome(tTok.Value)
		if err != nil {
			return nil, utils.NovoErro("tipo inválido", tTok.Position.Line, tTok.Position.Column, err.Error())
		}
		retorno = tp
	}

	if err := p.verificarProximoToken(lexer.LBRACE); err != nil {
		return nil, err
	}

	bloco, err := p.analisarBloco()
	if err != nil {
		return nil, err
	}

	return &FuncaoDeclaracao{Nome: nomeTok.Value, Parametros: params, Retorno: retorno, Corpo: bloco, Token: tokDef}, nil
}

// parseTipoPorNome converte o nome do tipo em Tipo
func (p *Parser) parseTipoPorNome(nome string) (Tipo, error) {
	switch nome {
	case "inteiro", "Inteiro":
		return TipoInteiro, nil
	case "decimal", "Decimal":
		return TipoDecimal, nil
	case "texto", "Texto":
		return TipoTexto, nil
	case "vazio", "Vazio":
		return TipoVazio, nil
	case "booleano", "Booleano":
		return TipoBooleano, nil
	default:
		return 0, fmt.Errorf("tipo desconhecido '%s' (suportado: inteiro, decimal, texto, vazio, booleano)", nome)
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

// esperarSemicolon verifica se o próximo token é um semicolon e o consome
func (p *Parser) esperarSemicolon() error {
	tok := p.tokenAtual()
	switch tok.Type {
	case lexer.SEMICOLON:
		p.proximoToken() // consome o semicolon
		return nil
	case lexer.RBRACE, lexer.EOF:
		// Semicolon opcional antes de '}' ou no fim do arquivo
		return nil
	default:
		return fmt.Errorf("esperado ';' em %s, encontrado '%s'", tok.Position, tok.Value)
	}
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

		// Estruturas de controle (como 'se') e declarações de função não precisam de semicolon
		if _, ehComandoSe := comando.(*ComandoSe); !ehComandoSe {
			if _, ehFuncDecl := comando.(*FuncaoDeclaracao); !ehFuncDecl {
				if err := p.esperarSemicolon(); err != nil {
					return nil, err
				}
			}
		}
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

// analisarComandoEnquanto: 'enquanto' (expr) '{' bloco '}'
func (p *Parser) analisarComandoEnquanto() (Expressao, error) {
	tok := p.proximoToken() // consumir 'enquanto'
	// condição
	cond, err := p.analisarExpressao(PRECEDENCIA_NENHUMA)
	if err != nil {
		return nil, err
	}
	if err := p.verificarProximoToken(lexer.LBRACE); err != nil {
		return nil, err
	}
	corpo, err := p.analisarBloco()
	if err != nil {
		return nil, err
	}
	return &ComandoEnquanto{Condicao: cond, Corpo: corpo, Token: tok}, nil
}

// analisarComandoPara: 'para' '(' init? ';' cond? ';' pos? ')' '{' bloco '}'
func (p *Parser) analisarComandoPara() (Expressao, error) {
	tok := p.proximoToken() // consumir 'para'
	if err := p.verificarProximoToken(lexer.LPAREN); err != nil {
		return nil, err
	}
	// init (pode ser vazio)
	var init Expressao
	if p.tokenAtual().Type != lexer.SEMICOLON {
		e, err := p.analisarAtribOuExpressao()
		if err != nil {
			return nil, err
		}
		init = e
	}
	if err := p.verificarProximoToken(lexer.SEMICOLON); err != nil {
		return nil, err
	}
	// cond (pode ser vazia)
	var cond Expressao
	if p.tokenAtual().Type != lexer.SEMICOLON {
		e, err := p.analisarExpressao(PRECEDENCIA_NENHUMA)
		if err != nil {
			return nil, err
		}
		cond = e
	}
	if err := p.verificarProximoToken(lexer.SEMICOLON); err != nil {
		return nil, err
	}
	// pos (pode ser vazia)
	var pos Expressao
	if p.tokenAtual().Type != lexer.RPAREN {
		e, err := p.analisarAtribOuExpressao()
		if err != nil {
			return nil, err
		}
		pos = e
	}
	if err := p.verificarProximoToken(lexer.RPAREN); err != nil {
		return nil, err
	}
	if err := p.verificarProximoToken(lexer.LBRACE); err != nil {
		return nil, err
	}
	corpo, err := p.analisarBloco()
	if err != nil {
		return nil, err
	}
	return &ComandoPara{Inicializacao: init, Condicao: cond, PosIteracao: pos, Corpo: corpo, Token: tok}, nil
}

// analisarAtribOuExpressao tenta analisar uma atribuição (com ou sem anotação de tipo) ou uma expressão
func (p *Parser) analisarAtribOuExpressao() (Expressao, error) {
	if p.tokenAtual().Type == lexer.IDENTIFIER {
		// snapshot da posição
		save := p.posicaoAtual
		identTok := p.proximoToken() // consome o identificador
		var tipoAnnot *Tipo
		if p.tokenAtual().Type == lexer.COLON {
			p.proximoToken()
			tTok := p.proximoToken()
			if tTok.Type != lexer.IDENTIFIER {
				return nil, utils.NovoErro("tipo inválido", tTok.Position.Line, tTok.Position.Column, "esperado identificador de tipo")
			}
			tp, err := p.parseTipoPorNome(tTok.Value)
			if err != nil {
				return nil, utils.NovoErro("tipo inválido", tTok.Position.Line, tTok.Position.Column, err.Error())
			}
			tipoAnnot = &tp
		}
		if p.tokenAtual().Type == lexer.ASSIGN {
			p.proximoToken() // consome '~>'
			valor, err := p.analisarExpressao(PRECEDENCIA_NENHUMA)
			if err != nil {
				return nil, err
			}
			return &Atribuicao{Nome: identTok.Value, Valor: valor, Token: identTok, TipoAnotado: tipoAnnot}, nil
		}
		// não era atribuição: restaura e analisa como expressão
		p.posicaoAtual = save
	}
	return p.analisarExpressao(PRECEDENCIA_NENHUMA)
}
