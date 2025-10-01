package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/khevencolino/Solar/internal/compiler"
)

// Config centraliza as configurações do compilador
type Config struct {
	ArquivoEntrada string
	Backend        string
	Arch           string
	Debug          bool
	ShowHelp       bool
}

func main() {
	config, err := processarArgumentos()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro: %v\n", err)
		os.Exit(1)
	}

	if config.ShowHelp {
		mostrarAjuda()
		return
	}

	compilador := compiler.NovoCompilador()

	// Converte para a estrutura de configuração do compiler
	compileConfig := &compiler.CompileConfig{
		ArquivoEntrada: config.ArquivoEntrada,
		Backend:        config.Backend,
		Arch:           config.Arch,
		Debug:          config.Debug,
	}

	if err := compilador.CompilarArquivo(compileConfig); err != nil {
		fmt.Fprintf(os.Stderr, "Erro de compilação: %v\n", err)
		os.Exit(1)
	}
}

func processarArgumentos() (*Config, error) {
	// Define flags
	backend := flag.String("backend", "interpreter", "Backend a ser usado (interpreter, assembly, llvm)")
	arch := flag.String("arch", "x86_64", "Arquitetura para assembly (x86_64)")
	debug := flag.Bool("debug", false, "Ativar mensagens de debug")
	help := flag.Bool("help", false, "Mostra ajuda")

	// Parse flags
	flag.Parse()

	config := &Config{
		Backend:  *backend,
		Arch:     *arch,
		Debug:    *debug,
		ShowHelp: *help,
	}

	// Verifica se help foi solicitado
	if *help {
		return config, nil
	}

	// Verifica se arquivo foi fornecido
	args := flag.Args()
	if len(args) < 1 {
		return nil, fmt.Errorf("arquivo de entrada requerido")
	}

	config.ArquivoEntrada = args[0]
	return config, nil
}

func mostrarAjuda() {
	fmt.Printf(`Compilador Solar - Sistema de Backends Múltiplos

USO:
    solar-compiler [flags] <arquivo>

FLAGS:
    -backend=<tipo>     Backend a ser usado (padrão: interpreter)
    -arch=<arquitetura> Arquitetura para assembly (padrão: x86_64)
    -debug              Ativar mensagens de debug
    -help               Mostra esta ajuda

BACKENDS DISPONÍVEIS:

interpreter, interp, ast
    - Interpretação direta da AST
    - Mostra árvore sintática

assembly, asm, native
    - Compilação para Assembly nativo
    - Gera executável standalone*
    - Máxima performance

llvm, llvmir, ir
    - Compilação para LLVM IR
    - Pode ser compilado para executável com clang/llc
    - Otimizações LLVM disponíveis

ARQUITETURAS SUPORTADAS PARA ASSEMBLY:
    - x86_64 (padrão)

EXEMPLOS:
    solar-compiler programa.solar                            # Usa interpretador (padrão)
    solar-compiler -backend=interpreter programa.solar       # Interpretação direta
    solar-compiler -backend=assembly programa.solar          # Assembly x86_64
    solar-compiler -backend=llvm programa.solar              # LLVM IR
    solar-compiler -debug programa.solar                     # Com mensagens de debug
`)
}
