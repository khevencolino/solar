package lexer

import "fmt"

// TokenType representa o tipo de token
type TokenType int

const (
	// Tipos de tokens
	NUMBER     TokenType = iota // Números
	PLUS                        // Operador de adição (+)
	MINUS                       // Operador de subtração (-)
	MULTIPLY                    // Operador de multiplicação (*)
	POWER                       // Operador de potência (**)
	DIVIDE                      // Operador de divisão
	LPAREN                      // Parêntese esquerdo (()
	RPAREN                      // Parêntese direito ())
	ASSIGN                      // Assign para variavel ~>
	COMMENT                     // Comentarios
	IDENTIFIER                  // Identificador da variavel
	FUNCTION                    // Função builtin
	COMMA                       // Vírgula (,)
	SEMICOLON                   // Ponto e vírgula (;)
	COLON                       // Dois pontos (:)
	WHITESPACE                  // Espaços em branco
	EOF                         // Fim do arquivo
	INVALID                     // Token inválido

	// Tokens para estruturas de controle
	SE            // Comando "se" (if)
	SENAO         // Comando "senao" (else)
	DEFINIR       // Declaração de função "definir"
	RETORNAR      // Comando de retorno "retornar"
	LBRACE        // Chave esquerda {
	RBRACE        // Chave direita }
	EQUAL         // Operador de igualdade ==
	NOT_EQUAL     // Operador de diferença !=
	LESS          // Operador menor que <
	GREATER       // Operador maior que >
	LESS_EQUAL    // Operador menor ou igual <=
	GREATER_EQUAL // Operador maior ou igual >=
	// Boolean literals
	VERDADEIRO // verdadeiro
	FALSO      // falso
	// Loops
	PARA     // for
	ENQUANTO // while
)

// String retorna uma representação em string do tipo de token
func (t TokenType) String() string {
	switch t {
	case NUMBER:
		return "NUMBER"
	case PLUS:
		return "PLUS"
	case MINUS:
		return "MINUS"
	case MULTIPLY:
		return "MULTIPLY"
	case POWER:
		return "POWER"
	case DIVIDE:
		return "DIVIDE"
	case LPAREN:
		return "LPAREN"
	case RPAREN:
		return "RPAREN"
	case WHITESPACE:
		return "WHITESPACE"
	case COMMENT:
		return "COMMENT"
	case EOF:
		return "EOF"
	case ASSIGN:
		return "ASSIGN"
	case IDENTIFIER:
		return "IDENTIFIER"
	case FUNCTION:
		return "FUNCTION"
	case COMMA:
		return "COMMA"
	case SEMICOLON:
		return "SEMICOLON"
	case COLON:
		return "COLON"
	case INVALID:
		return "INVALID"
	case SE:
		return "SE"
	case SENAO:
		return "SENAO"
	case LBRACE:
		return "LBRACE"
	case RBRACE:
		return "RBRACE"
	case DEFINIR:
		return "DEFINIR"
	case RETORNAR:
		return "RETORNAR"
	case EQUAL:
		return "EQUAL"
	case NOT_EQUAL:
		return "NOT_EQUAL"
	case LESS:
		return "LESS"
	case GREATER:
		return "GREATER"
	case LESS_EQUAL:
		return "LESS_EQUAL"
	case GREATER_EQUAL:
		return "GREATER_EQUAL"
	case VERDADEIRO:
		return "VERDADEIRO"
	case FALSO:
		return "FALSO"
	case PARA:
		return "PARA"
	case ENQUANTO:
		return "ENQUANTO"
	default:
		return "UNKNOWN"
	}
}

// Token representa um token encontrado no código fonte
type Token struct {
	Type     TokenType // Tipo do token
	Value    string    // Valor do token
	Position Position  // Posição no código fonte
}

// String retorna uma representação em string do token
func (t Token) String() string {
	return fmt.Sprintf("%s('%s') em %s", t.Type, t.Value, t.Position)
}

// NovoToken cria um novo token
func NovoToken(tipoToken TokenType, valor string, posicao Position) Token {
	return Token{
		Type:     tipoToken,
		Value:    valor,
		Position: posicao,
	}
}
