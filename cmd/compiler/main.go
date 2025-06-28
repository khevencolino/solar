package main

import (
	"fmt" // Para formatação de strings e impressão no console
	"os"  // Para acessar argumentos da linha de comando e sair do programa
	"path/filepath"
	"regexp" // Para expressões regulares (usado para encontrar o número)
)

// main é o ponto de entrada principal do compilador.
// Esta função lê o nome do arquivo de entrada, extrai o primeiro número
// constante e gera um arquivo assembly AT&T que imprime essa constante.
func main() {
	// 1. Processamento de Argumentos da Linha de Comando
	args := os.Args

	// Verifica se o número correto de argumentos foi fornecido.
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "Uso: %s <arquivo_de_entrada>\n", args[0])
		fmt.Fprintf(os.Stderr, "Exemplo: %s programa.kite\n", args[0])
		os.Exit(1)
	}

	inputFilePath := args[1]
	outputFolderName := "result"
	outputFileName := filepath.Join(outputFolderName, "saida.s")

	// 2. Leitura do Arquivo de Entrada
	// Tenta ler o conteúdo completo do arquivo de entrada.
	contentBytes, err := os.ReadFile(inputFilePath)
	if err != nil {
		// Se houver um erro, imprime a mensagem de erro e sai.
		fmt.Fprintf(os.Stderr, "Erro ao ler o arquivo de entrada '%s': %v\n", inputFilePath, err)
		os.Exit(1)
	}
	content := string(contentBytes)

	// 3. Extração da Constante
	re := regexp.MustCompile(`\d+`) // Compila uma regex que busca uma ou mais ocorrências de dígitos.
	match := re.FindString(content) // Encontra a primeira correspondência

	if match == "" {
		fmt.Fprintf(os.Stderr, "Erro: Nenhum número constante encontrado no arquivo de entrada '%s'.\n", inputFilePath)
		os.Exit(1)
	}

	// O valor da constante é a string encontrada pela expressão regular.
	constantStr := match

	// 4. Geração do Código Assembly
	// O template assembly com um placeholder para a constante.
	// O placeholder "%s" será substituído pela string da constante.
	assemblyTemplate := `.section .text
.globl _start

_start:
	mov $%s, %%rax

	call imprime_num
	call sair

.include "runtime.s"
`
	// %s é o placeholder para a constante.
	assemblyCode := fmt.Sprintf(assemblyTemplate, constantStr)

	// 5. Escrita do Arquivo de Saída
	// Tenta criar (ou sobrescrever) o arquivo de saída.
	if _, err := os.Stat(outputFolderName); os.IsNotExist(err) {
		err = os.Mkdir(outputFolderName, 0755)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Erro ao criar a pasta de saída '%s': %v\n", outputFolderName, err)
			os.Exit(1)
		}
	}
	// Escreve o conteúdo no arquivo
	err = os.WriteFile(outputFileName, []byte(assemblyCode), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao escrever no arquivo de saída '%s': %v\n", outputFileName, err)
		os.Exit(1)
	}

	// Se a escrita for bem-sucedida, imprime uma mensagem de sucesso.
	fmt.Printf("Código assembly escrito com sucesso em '%s'\n", outputFileName)
}
