package main

import (
	"fmt"
	"os"

	"github.com/khevencolino/Kite/internal/compiler"
)

func main() {
	arquivoEntrada, backend, showHelp, err := processarArgumentos()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro: %v\n", err)
		os.Exit(1)
	}

	if showHelp {
		mostrarAjuda()
		return
	}

	compilador := compiler.NovoCompilador()

	if err := compilador.CompilarArquivo(arquivoEntrada, backend); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Erro de compilação: %v\n", err)
		os.Exit(1)
	}
}

func processarArgumentos() (string, string, bool, error) {
	args := os.Args

	if len(args) < 2 {
		return "", "", false, fmt.Errorf("argumentos insuficientes")
	}

	// Verifica help
	if args[1] == "--help" || args[1] == "-h" {
		return "", "", true, nil
	}

	arquivo := args[1]
	backend := "interpreter"

	if len(args) >= 3 {
		backend = args[2]
	}

	return arquivo, backend, false, nil
}

func mostrarAjuda() {
	fmt.Printf(`Compilador Kite - Sistema de Backends Múltiplos

USO:
    kite-compiler <arquivo> [backend]

BACKENDS DISPONÍVEIS:

🔍 interpreter, interp, ast (PADRÃO)
   - Interpretação direta da AST
   - Mais rápido para desenvolvimento e debug
   - Mostra árvore sintática

🤖 bytecode, vm, bc
   - Compilação para bytecode + Virtual Machine
   - Mostra instruções geradas
   - Boa performance, fácil debug

🔧 assembly, asm, native
   - Compilação para Assembly x86-64 nativo
   - Gera executável standalone
   - Máxima performance

EXEMPLOS:
    kite-compiler programa.kite                    # Usa interpretador (padrão)
    kite-compiler programa.kite interpreter        # Interpretação direta
    kite-compiler programa.kite bytecode           # Bytecode + VM
    kite-compiler programa.kite assembly           # Assembly nativo

ARQUIVOS DE TESTE:
    exemplos/constante/valido.kite                 # Número simples
    exemplos/operadores/valido.kite                # Expressões
    exemplos/variaveis/valido.kite                 # Variáveis
    exemplos/aninhados/valido.kite                 # Expressões complexas
`)
}
