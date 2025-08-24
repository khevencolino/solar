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

// inicializarPadroes inicializa os padrões regex para cada tipo de token
func (l *Lexer) inicializarPadroes() {
	l.padroes = map[TokenType]*regexp.Regexp{
		NUMBER:     regexp.MustCompile(`^\d+`),                  // Números: 123, 456
		PLUS:       regexp.MustCompile(`^\+`),                   // Adição: +
		MINUS:      regexp.MustCompile(`^-`),                    // Subtraço: -
		MULTIPLY:   regexp.MustCompile(`^\*`),                   // Multiplicação: *
		POWER:      regexp.MustCompile(`^\*\*`),                 // Potência: **
		DIVIDE:     regexp.MustCompile(`^/`),                    // Divisão
		LPAREN:     regexp.MustCompile(`^\(`),                   // Parêntese esquerdo: (
		RPAREN:     regexp.MustCompile(`^\)`),                   // Parêntese direito: )
		ASSIGN:     regexp.MustCompile(`^~>`),                   // Simbolo para alocar variavel ~>
		IDENTIFIER: regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9]*`), // Palavras permitidas para nomear variaveis
		COMMA:      regexp.MustCompile(`^,`),                    // Vírgula: ,
		WHITESPACE: regexp.MustCompile(`^\s+`),                  // Espaços em branco
		COMMENT:    regexp.MustCompile(`^//.*`),                 // Comentarios //
	}
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

	// Tenta fazer match com cada padrão (ordem importa para ** vs *)
	tiposToken := []TokenType{COMMENT, ASSIGN, IDENTIFIER, POWER, NUMBER, PLUS, MINUS, DIVIDE, MULTIPLY, LPAREN, RPAREN, COMMA, WHITESPACE}

	for _, tipoToken := range tiposToken {
		if match := l.padroes[tipoToken].FindString(restante); match != "" {
			token := NovoToken(tipoToken, match, posicaoAtual)

			// Se é um identificador, verifica se é uma função builtin
			if tipoToken == IDENTIFIER {
				if l.ehFuncaoBuiltin(match) {
					token.Type = FUNCTION
				}
			}

			l.avancar(len(match))
			return token, nil
		}
	}

	// Caractere inválido
	caractereInvalido := string(l.espiar())
	l.avancar(1)
	return NovoToken(INVALID, caractereInvalido, posicaoAtual),
		fmt.Errorf("caractere inválido '%s' em %s", caractereInvalido, posicaoAtual)
}

// ehFuncaoBuiltin verifica se um identificador é uma função builtin
func (l *Lexer) ehFuncaoBuiltin(nome string) bool {
	return registry.RegistroGlobal.EhFuncaoBuiltin(nome)
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
