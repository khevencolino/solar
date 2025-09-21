package compiler

import (
	"fmt"

	"github.com/khevencolino/Solar/internal/backends"
	"github.com/khevencolino/Solar/internal/backends/assembly"
	"github.com/khevencolino/Solar/internal/backends/interpreter"
	"github.com/khevencolino/Solar/internal/backends/llvm"
	"github.com/khevencolino/Solar/internal/debug"
	"github.com/khevencolino/Solar/internal/lexer"
	"github.com/khevencolino/Solar/internal/parser"
	"github.com/khevencolino/Solar/internal/utils"
)

type Compiler struct {
	lexer  *lexer.Lexer
	parser *parser.Parser
	debug  bool
}

func NovoCompilador() *Compiler {
	return &Compiler{}
}

func (c *Compiler) CompilarArquivo(arquivoEntrada string, backendType string, arch string, debugEnabled bool) error {
	c.debug = debugEnabled
	debug.Enabled = debugEnabled

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

	// Imprime tokens apenas se debug estiver ativo
	if c.debug {
		fmt.Printf("Tokens encontrados:\n")
		lexer.ImprimirTokens(tokens)
		fmt.Println()
	}

	// Análise sintática
	statements, err := c.analisarSintaxe(tokens)
	if err != nil {
		return err
	}

	// Checagem de tipos (semântica)
	if err := c.checagemTipos(statements); err != nil {
		if c.debug {
			fmt.Printf("Erro na checagem de tipos: %v\n", err)
		}
		return err
	}

	// Seleciona e executa backend
	return c.executarBackend(statements, backendType, arch)
}

func (c *Compiler) executarBackend(statements []parser.Expressao, backendType string, arch string) error {
	var backend backends.Backend

	switch backendType {
	case "interpreter", "interp", "ast":
		backend = interpreter.NewInterpreterBackend()

	case "assembly", "asm", "native":
		var err error
		backend, err = assembly.NewAssemblyBackend(arch)
		if err != nil {
			return err
		}

	case "llvm", "llvmir", "ir":
		backend = llvm.NewLLVMBackend()

	default:
		return fmt.Errorf(`backend desconhecido: %s

Backends disponíveis:
  interpreter, interp, ast  - Interpretação direta da AST (padrão)
  assembly, asm, native    - Compilação para Assembly x86-64
  llvm, llvmir, ir         - Compilação para LLVM IR
  `, backendType)
	}

	if c.debug {
		fmt.Printf("Backend selecionado: %s\n\n", backend.GetName())
	}

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
	statements, err := c.parser.AnalisarPrograma()
	if err != nil {
		if c.debug {
			fmt.Printf("Erro no parser: %v\n", err)
		}
		return nil, err
	}
	return statements, nil
}

// checagemTipos executa a validação de tipos sobre a AST
func (c *Compiler) checagemTipos(statements []parser.Expressao) error {
	tc := NovoTypeChecker()
	return tc.Check(statements)
}
