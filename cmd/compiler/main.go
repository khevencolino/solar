package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/khevencolino/Solar/internal/compiler"
)

func main() {
	arquivoEntrada, backend, arch, showHelp, err := processarArgumentos()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro: %v\n", err)
		os.Exit(1)
	}

	if showHelp {
		mostrarAjuda()
		return
	}

	compilador := compiler.NovoCompilador()

	if err := compilador.CompilarArquivo(arquivoEntrada, backend, arch); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Erro de compilação: %v\n", err)
		os.Exit(1)
	}
}

func processarArgumentos() (string, string, string, bool, error) {
	// Define flags
	backend := flag.String("backend", "interpreter", "Backend a ser usado (interpreter, bytecode, assembly)")
	arch := flag.String("arch", "x86_64", "Arquitetura para assembly (x86_64, arm64)")
	help := flag.Bool("help", false, "Mostra ajuda")

	// Parse flags
	flag.Parse()

	// Verifica se help foi solicitado
	if *help {
		return "", "", "", true, nil
	}

	// Verifica se arquivo foi fornecido
	args := flag.Args()
	if len(args) < 1 {
		return "", "", "", false, fmt.Errorf("arquivo de entrada requerido")
	}

	arquivo := args[0]

	return arquivo, *backend, *arch, false, nil
}

func mostrarAjuda() {
	fmt.Printf(`Compilador Solar - Sistema de Backends Múltiplos

USO:
    solar-compiler [flags] <arquivo>

FLAGS:
    -backend=<tipo>     Backend a ser usado (padrão: interpreter)
    -arch=<arquitetura> Arquitetura para assembly (padrão: x86_64)
    -help               Mostra esta ajuda

BACKENDS DISPONÍVEIS:

🔍 interpreter, interp, ast
    - Interpretação direta da AST
    - Mostra árvore sintática

🤖 bytecode, vm, bc
    - Compilação para bytecode + Virtual Machine
    - Mostra instruções geradas
    - Boa performance, fácil debug

🔧 assembly, asm, native
    - Compilação para Assembly nativo
    - Gera executável standalone*
    - Máxima performance

ARQUITETURAS SUPORTADAS PARA ASSEMBLY:
    - x86_64 (Linux - padrão)
    - arm64 (macOS)

EXEMPLOS:
    solar-compiler programa.solar                            # Usa interpretador (padrão)
    solar-compiler -backend=interpreter programa.solar       # Interpretação direta
    solar-compiler -backend=bytecode programa.solar          # Bytecode + VM
    solar-compiler -backend=assembly programa.solar          # Assembly x86_64 (padrão)
    solar-compiler -backend=assembly -arch=arm64 programa.solar # Assembly ARM64
`)
}
