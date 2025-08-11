package compiler

import (
	"fmt"

	"github.com/khevencolino/Solar/internal/backends"
	"github.com/khevencolino/Solar/internal/backends/assembly"
	"github.com/khevencolino/Solar/internal/backends/bytecode"
	"github.com/khevencolino/Solar/internal/backends/interpreter"
	"github.com/khevencolino/Solar/internal/lexer"
	"github.com/khevencolino/Solar/internal/parser"
	"github.com/khevencolino/Solar/internal/utils"
)

type Compiler struct {
	lexer  *lexer.Lexer
	parser *parser.Parser
}

func NovoCompilador() *Compiler {
	return &Compiler{}
}

func (c *Compiler) CompilarArquivo(arquivoEntrada string, backendType string) error {
	// Lê o arquivo
	conteudo, err := utils.LerArquivo(arquivoEntrada)
	if err != nil {
		return err
	}

	// Análise léxica
	tokens, err := c.tokenizar(conteudo)
	if err != nil {
		return err
	}

	// Imprime tokens
	fmt.Printf("Tokens encontrados:\n")
	lexer.ImprimirTokens(tokens)
	fmt.Println()

	// Análise sintática
	statements, err := c.analisarSintaxe(tokens)
	if err != nil {
		return err
	}

	// Seleciona e executa backend
	return c.executarBackend(statements, backendType)
}

func (c *Compiler) executarBackend(statements []parser.Expressao, backendType string) error {
	var backend backends.Backend

	switch backendType {
	case "interpreter", "interp", "ast":
		backend = interpreter.NewInterpreterBackend()

	case "bytecode", "vm", "bc":
		backend = bytecode.NewBytecodeBackend()

	case "assembly", "asm", "native":
		backend = assembly.NewAssemblyBackend()

	default:
		return fmt.Errorf(`backend desconhecido: %s

Backends disponíveis:
  interpreter, interp, ast  - Interpretação direta da AST (padrão)
  bytecode, vm, bc         - Compilação para Bytecode + VM
  assembly, asm, native    - Compilação para Assembly x86-64

Exemplo: ./solar-compiler programa.solar interpreter`, backendType)
	}

	fmt.Printf("🎯 Backend selecionado: %s\n\n", backend.GetName())

	return backend.Compile(statements)
}

func (c *Compiler) tokenizar(conteudo string) ([]lexer.Token, error) {
	c.lexer = lexer.NovoLexer(conteudo)
	tokens, err := c.lexer.Tokenizar()
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

func (c *Compiler) analisarSintaxe(tokens []lexer.Token) ([]parser.Expressao, error) {
	c.parser = parser.NovoParser(tokens)
	return c.parser.AnalisarPrograma()
}
