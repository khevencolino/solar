package utils

import (
	"os"
)

// LerArquivo lê um arquivo e retorna seu conteúdo
func LerArquivo(nomeArquivo string) (string, error) {
	bytesConteudo, err := os.ReadFile(nomeArquivo)
	if err != nil {
		return "", NovoErro("erro ao ler arquivo", 0, 0, err.Error())
	}
	return string(bytesConteudo), nil
}

// EscreverArquivo escreve conteúdo em um arquivo
func EscreverArquivo(nomeArquivo string, conteudo string) error {
	// Escreve o arquivo
	if err := os.WriteFile(nomeArquivo, []byte(conteudo), 0644); err != nil {
		return NovoErro("erro ao escrever arquivo", 0, 0, err.Error())
	}

	return nil
}
