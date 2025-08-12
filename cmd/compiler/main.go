package main

import (
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
	args := os.Args

	if len(args) < 2 {
		return "", "", "", false, fmt.Errorf("argumentos insuficientes")
	}

	// Verifica help
	if args[1] == "--help" || args[1] == "-h" {
		return "", "", "", true, nil
	}

	arquivo := args[1]
	backend := "interpreter"
	arch := "x86_64"

	if len(args) >= 3 {
		backend = args[2]
	}

	if len(args) >= 4 {
		arch = args[3]
	}

	return arquivo, backend, arch, false, nil
}

func mostrarAjuda() {
	fmt.Printf(`Compilador Solar - Sistema de Backends Múltiplos

USO:
    solar-compiler <arquivo> [backend] [arquitetura]

BACKENDS DISPONÍVEIS:

🔍 interpreter, interp, ast (PADRÃO)
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
    - x86_64 (padrão)
    - arm64

EXEMPLOS:
    solar-compiler programa.solar                            # Usa interpretador (padrão)
    solar-compiler programa.solar interpreter                # Interpretação direta
    solar-compiler programa.solar bytecode                   # Bytecode + VM
    solar-compiler programa.solar assembly                   # Assembly x86_64 (padrão)
    solar-compiler programa.solar assembly arm64             # Assembly ARM64
`)
}
