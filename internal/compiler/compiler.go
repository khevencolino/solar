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
	"github.com/khevencolino/Solar/internal/prelude"
	"github.com/khevencolino/Solar/internal/utils"
)

type Compiler struct {
	lexer          *lexer.Lexer
	parser         *parser.Parser
	moduleResolver *ModuleResolver
	prelude        *prelude.Prelude
	debug          bool
}

func NovoCompilador() *Compiler {
	return &Compiler{
		moduleResolver: NewModuleResolver(),
		prelude:        prelude.NewPrelude(),
	}
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

	// Processamento de imports
	statements, err = c.processarImports(statements)
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

// processarImports resolve e incorpora módulos importados
func (c *Compiler) processarImports(statements []parser.Expressao) ([]parser.Expressao, error) {
	var novosStatements []parser.Expressao
	var importsEncontrados []*parser.Importacao

	// Separa imports dos outros statements
	for _, stmt := range statements {
		if imp, ehImport := stmt.(*parser.Importacao); ehImport {
			importsEncontrados = append(importsEncontrados, imp)
		} else {
			novosStatements = append(novosStatements, stmt)
		}
	}

	// Processa cada import
	for _, imp := range importsEncontrados {
		if c.debug {
			fmt.Printf("Processando import: %s de %s\n", imp.Simbolos, imp.Modulo)
		}

		// Resolve o módulo
		_, err := c.moduleResolver.ResolverModulo(imp.Modulo)
		if err != nil {
			return nil, fmt.Errorf("erro ao resolver módulo '%s': %v", imp.Modulo, err)
		}

		// Valida e incorpora os símbolos solicitados
		for _, simbolo := range imp.Simbolos {
			sim, err := c.moduleResolver.ResolverSimbolo(imp.Modulo, simbolo)
			if err != nil {
				return nil, fmt.Errorf("erro ao resolver símbolo '%s' do módulo '%s': %v", simbolo, imp.Modulo, err)
			}

			// Adiciona o nó AST do símbolo importado (apenas se não for built-in)
			if sim.Node != nil {
				novosStatements = append(novosStatements, sim.Node)
			}

			if c.debug {
				tipoStr := "AST"
				if sim.Node == nil {
					tipoStr = "built-in"
				}
				fmt.Printf("  Símbolo '%s' importado com sucesso (%s)\n", simbolo, tipoStr)
			}
		}
	}

	return novosStatements, nil
}

// checagemTipos executa a validação de tipos sobre a AST
func (c *Compiler) checagemTipos(statements []parser.Expressao) error {
	tc := NovoTypeChecker()
	return tc.Check(statements)
}
