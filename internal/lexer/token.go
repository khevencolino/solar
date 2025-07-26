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
	COMMENT                     // Comentarios
	WHITESPACE                  // Espaços em branco
	EOF                         // Fim do arquivo
	INVALID                     // Token inválido
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
	case INVALID:
		return "INVALID"
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

// EOperador verifica se o token é um operador
func (t Token) EOperador() bool {
	return t.Type == PLUS || t.Type == MINUS || t.Type == MULTIPLY || t.Type == POWER || t.Type == DIVIDE
}

// ENumero verifica se o token é um número
func (t Token) ENumero() bool {
	return t.Type == NUMBER
}

// EParenteses verifica se o token é um parêntese
func (t Token) EParenteses() bool {
	return t.Type == LPAREN || t.Type == RPAREN
}
