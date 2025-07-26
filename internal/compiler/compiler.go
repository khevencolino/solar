package compiler

import (
	"fmt"

	"github.com/khevencolino/Kite/internal/lexer"
	"github.com/khevencolino/Kite/internal/parser"
	"github.com/khevencolino/Kite/internal/utils"
)

// Compiler representa o compilador principal
type Compiler struct {
	lexer         *lexer.Lexer          // Analisador léxico
	gerador       *Generator            // Gerador de código
	parser        *parser.Parser        // Parser
	interpretador *parser.Interpretador // Interpretador
}

// NovoCompilador cria um novo compilador
func NovoCompilador() *Compiler {
	return &Compiler{
		gerador:       NovoGerador(),
		interpretador: parser.NovoInterpretador(),
	}
}

// CompilarArquivo compila um arquivo fonte
func (c *Compiler) CompilarArquivo(arquivoEntrada string) error {
	// Lê o arquivo de entrada
	conteudo, err := utils.LerArquivo(arquivoEntrada)
	if err != nil {
		return err
	}

	// Realiza análise léxica
	tokens, err := c.tokenizar(conteudo)
	if err != nil {
		return err
	}

	// Imprime tokens para depuração
	fmt.Printf("Tokens encontrados:\n")
	lexer.ImprimirTokens(tokens)

	ast, err := c.analisarSintaxe(tokens)
	if err != nil {
		return err
	}

	// Imprime a arvore
	visualizador := parser.NovoVisualizador()
	visualizador.ImprimirArvore(ast)

	// Roda o interpretador sobre a arvore
	resultado, err := c.interpretador.Interpretar(ast)
	if err != nil {
		return err
	}

	fmt.Println("Resultado da expressão: ", resultado)

	// // Extrai o primeiro número (lógica temporária)
	// primeiroNumero, err := c.extrairPrimeiroNumero(tokens)
	// if err != nil {
	// 	return err
	// }

	// // Gera código assembly, defasado desde atv02
	// assembly := c.gerador.GerarAssembly(primeiroNumero)

	// // Escreve arquivo de saída
	// arquivoSaida := filepath.Join("result", "saida.s")
	// if err := utils.EscreverArquivo(arquivoSaida, assembly); err != nil {
	// 	return err
	// }

	// fmt.Printf("Código assembly escrito em '%s'\n", arquivoSaida)
	return nil
}

// tokenizar realiza análise léxica
func (c *Compiler) tokenizar(conteudo string) ([]lexer.Token, error) {
	c.lexer = lexer.NovoLexer(conteudo)
	tokens, err := c.lexer.Tokenizar()
	if err != nil {
		return nil, err
	}

	// Valida a expressão
	if err := c.lexer.ValidarExpressao(tokens); err != nil {
		return nil, err
	}

	return tokens, nil
}

// extrairPrimeiroNumero extrai o primeiro número dos tokens (temporário)
func (c *Compiler) extrairPrimeiroNumero(tokens []lexer.Token) (string, error) {
	for _, token := range tokens {
		if token.ENumero() {
			return token.Value, nil
		}
	}
	return "", utils.NovoErro("nenhum número encontrado", 0, 0, "")
}

func (c *Compiler) analisarSintaxe(tokens []lexer.Token) (parser.Expressao, error) {
	c.parser = parser.NovoParser(tokens)
	return c.parser.AnalisarPrograma()
}
