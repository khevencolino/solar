package main

import (
	"fmt"
	"os"

	"github.com/khevencolino/Kite/internal/compiler"
	"github.com/khevencolino/Kite/internal/utils"
)

func main() {
	// Processa argumentos da linha de comando
	arquivoEntrada, err := processarArgumentos()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro: %v\n", err)
		os.Exit(1)
	}

	// Cria uma instância do compilador
	compilador := compiler.NovoCompilador()

	// Compila o arquivo de entrada
	if err := compilador.CompilarArquivo(arquivoEntrada); err != nil {
		fmt.Fprintf(os.Stderr, "Erro de compilação: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Compilação concluída com sucesso!\n")
}

// processarArgumentos processa os argumentos da linha de comando
func processarArgumentos() (string, error) {
	args := os.Args
	if len(args) < 2 {
		return "", utils.NovoErro("argumentos insuficientes", 0, 0,
			fmt.Sprintf("Uso: %s <arquivo_de_entrada>\nExemplo: %s programa.kite", args[0], args[0]))
	}
	return args[1], nil
}
