package lexer

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/khevencolino/Solar/internal/registry"
)

// Lexer representa o analisador léxico
type Lexer struct {
	entrada string                       // Código fonte de entrada
	posicao int                          // Posição atual no código
	linha   int                          // Linha atual
	coluna  int                          // Coluna atual
	padroes map[TokenType]*regexp.Regexp // Padrões regex para cada tipo de token
}

// NovoLexer cria um novo analisador léxico
func NovoLexer(entrada string) *Lexer {
	lexer := &Lexer{
		entrada: entrada,
		linha:   1,
		coluna:  1,
	}
	lexer.inicializarPadroes()
	return lexer
}

// Padrões regex pré-compilados (otimização - compilados apenas uma vez)
var padroesCompiledos = map[TokenType]*regexp.Regexp{
	NUMBER:        regexp.MustCompile(`^\d+`),                    // Números inteiros: 123, 456
	FLOAT:         regexp.MustCompile(`^\d+\.\d+`),               // Números decimais: 123.45, 0.5
	STRING:        regexp.MustCompile(`^"[^"]*"`),                // Strings: "texto", "olá mundo"
	PLUS:          regexp.MustCompile(`^\+`),                     // Adição: +
	MINUS:         regexp.MustCompile(`^-`),                      // Subtração: -
	MULTIPLY:      regexp.MustCompile(`^\*`),                     // Multiplicação: *
	POWER:         regexp.MustCompile(`^\*\*`),                   // Potência: **
	DIVIDE:        regexp.MustCompile(`^/`),                      // Divisão: /
	LPAREN:        regexp.MustCompile(`^\(`),                     // Parêntese esquerdo: (
	RPAREN:        regexp.MustCompile(`^\)`),                     // Parêntese direito: )
	ASSIGN:        regexp.MustCompile(`^~>`),                     // Símbolo para alocar variável: ~>
	IDENTIFIER:    regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*`), // Identificadores válidos (com underscore)
	COMMA:         regexp.MustCompile(`^,`),                      // Vírgula: ,
	SEMICOLON:     regexp.MustCompile(`^;`),                      // Ponto e vírgula: ;
	COLON:         regexp.MustCompile(`^:`),                      // Dois pontos: :
	WHITESPACE:    regexp.MustCompile(`^\s+`),                    // Espaços em branco
	COMMENT:       regexp.MustCompile(`^//.*`),                   // Comentários: //
	LBRACE:        regexp.MustCompile(`^\{`),                     // Chave esquerda: {
	RBRACE:        regexp.MustCompile(`^\}`),                     // Chave direita: }
	EQUAL:         regexp.MustCompile(`^==`),                     // Operador de igualdade: ==
	NOT_EQUAL:     regexp.MustCompile(`^!=`),                     // Operador de diferença: !=
	LESS_EQUAL:    regexp.MustCompile(`^<=`),                     // Operador menor ou igual: <=
	GREATER_EQUAL: regexp.MustCompile(`^>=`),                     // Operador maior ou igual: >=
	LESS:          regexp.MustCompile(`^<`),                      // Operador menor que: <
	GREATER:       regexp.MustCompile(`^>`),                      // Operador maior que: >
}

// ordemTiposToken define a ordem de tentativa de matching dos tokens.
// A ordem é importante para evitar conflitos (ex: FLOAT antes de NUMBER, POWER antes de MULTIPLY, >= antes de >, etc.)
var ordemTiposToken = []TokenType{
	COMMENT,
	ASSIGN,
	IDENTIFIER,
	POWER,
	GREATER_EQUAL,
	LESS_EQUAL,
	NOT_EQUAL,
	EQUAL,
	STRING,
	FLOAT,
	NUMBER,
	PLUS,
	MINUS,
	DIVIDE,
	MULTIPLY,
	LPAREN,
	RPAREN,
	LBRACE,
	RBRACE,
	LESS,
	GREATER,
	COMMA,
	SEMICOLON,
	COLON,
	WHITESPACE,
}

// inicializarPadroes atribui os padrões pré-compilados
func (l *Lexer) inicializarPadroes() {
	l.padroes = padroesCompiledos
}

// Tokenizar converte a entrada em uma lista de tokens
func (l *Lexer) Tokenizar() ([]Token, error) {
	var tokens []Token

	for {
		token, err := l.proximoToken()
		if err != nil {
			return nil, err
		}

		// Pula espaços em branco mas adiciona outros tokens
		if token.Type != WHITESPACE && token.Type != COMMENT {
			tokens = append(tokens, token)
		}

		if token.Type == EOF {
			break
		}
	}

	return tokens, nil
}

// proximoToken encontra o próximo token
func (l *Lexer) proximoToken() (Token, error) {
	if !l.temMais() {
		return NovoToken(EOF, "", l.obterPosicaoAtual()), nil
	}

	posicaoAtual := l.obterPosicaoAtual()
	restante := l.entrada[l.posicao:]

	// Tenta fazer match com cada padrão respeitando a ordem definida globalmente
	for _, tipoToken := range ordemTiposToken {
		if match := l.padroes[tipoToken].FindString(restante); match != "" {
			token := NovoToken(tipoToken, match, posicaoAtual)

			// Se é um identificador, verifica se é uma função builtin ou palavra-chave
			if tipoToken == IDENTIFIER {
				if l.ehPalavraChave(match) {
					token.Type = l.obterTipoPalavraChave(match)
				} else if l.ehFuncaoBuiltin(match) {
					token.Type = FUNCTION
				}
			}

			l.avancar(len(match))
			return token, nil
		}
	}

	// Caractere inválido: consome um byte e retorna erro
	caractereInvalido := string(l.espiar())
	l.avancar(1)
	return NovoToken(INVALID, caractereInvalido, posicaoAtual), fmt.Errorf("caractere inválido '%s' em %s", caractereInvalido, posicaoAtual)
}

// ehFuncaoBuiltin verifica se um identificador é uma função builtin
func (l *Lexer) ehFuncaoBuiltin(nome string) bool {
	return registry.RegistroGlobal.EhFuncaoBuiltin(nome)
}

// palavrasChave é um mapa pré-definido das palavras-chave
var palavrasChave = map[string]TokenType{
	"se":         SE,
	"senao":      SENAO,
	"definir":    DEFINIR,
	"retornar":   RETORNAR,
	"verdadeiro": VERDADEIRO,
	"falso":      FALSO,
	"para":       PARA,
	"enquanto":   ENQUANTO,
	"importar":   IMPORTAR,
	"de":         DE,
}

// ehPalavraChave verifica se um identificador é uma palavra-chave
func (l *Lexer) ehPalavraChave(nome string) bool {
	_, existe := palavrasChave[nome]
	return existe
}

// obterTipoPalavraChave retorna o tipo de token para uma palavra-chave
func (l *Lexer) obterTipoPalavraChave(nome string) TokenType {
	if tipoToken, existe := palavrasChave[nome]; existe {
		return tipoToken
	}
	return IDENTIFIER
}

// obterPosicaoAtual retorna a posição atual no código fonte
func (l *Lexer) obterPosicaoAtual() Position {
	return NovaPosicao(l.linha, l.coluna, l.posicao)
}

// avancar move a posição do lexer para frente
func (l *Lexer) avancar(comprimento int) {
	for i := 0; i < comprimento; i++ {
		if l.posicao < len(l.entrada) {
			if l.entrada[l.posicao] == '\n' {
				l.linha++
				l.coluna = 1
			} else {
				l.coluna++
			}
			l.posicao++
		}
	}
}

// espiar retorna o caractere atual sem avançar
func (l *Lexer) espiar() byte {
	if l.posicao >= len(l.entrada) {
		return 0
	}
	return l.entrada[l.posicao]
}

// temMais verifica se há mais caracteres para processar
func (l *Lexer) temMais() bool {
	return l.posicao < len(l.entrada)
}

// ImprimirTokens imprime todos os tokens de forma formatada
func ImprimirTokens(tokens []Token) {
	fmt.Printf("%-10s %-15s %-20s\n", "TIPO", "VALOR", "POSIÇÃO")
	fmt.Println(strings.Repeat("-", 50))

	for _, token := range tokens {
		if token.Type != EOF {
			fmt.Printf("%-10s %-15s %-20s\n", token.Type, token.Value, token.Position)
		}
	}
}
