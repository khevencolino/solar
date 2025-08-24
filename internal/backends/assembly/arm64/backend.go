package arm64

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/khevencolino/Solar/internal/parser"
	"github.com/khevencolino/Solar/internal/registry"
	"github.com/khevencolino/Solar/internal/utils"
)

type ARM64Backend struct {
	output     strings.Builder
	variables  map[string]bool
	labelCount int
}

func NewARM64Backend() *ARM64Backend {
	return &ARM64Backend{
		variables: make(map[string]bool),
	}
}

func (a *ARM64Backend) GetName() string      { return "Assembly ARM64 (macOS)" }
func (a *ARM64Backend) GetExtension() string { return ".s" }

func (a *ARM64Backend) Compile(statements []parser.Expressao) error {
	fmt.Printf("🍎 Compilando para ARM64 (macOS)...\n")
	a.gerarPrologo()

	for i, stmt := range statements {
		fmt.Printf("  Processando statement %d...\n", i+1)
		stmt.Aceitar(a)

		// Se for a última expressão e não for uma chamada de função, imprime resultado
		if i == len(statements)-1 {
			if _, ehChamadaFuncao := stmt.(*parser.ChamadaFuncao); !ehChamadaFuncao {
				a.output.WriteString("    bl _print_num\n")
			}
		}
	}

	a.gerarEpilogo()

	arquivoSaida := filepath.Join("result", "programa.s")
	if err := utils.EscreverArquivo(arquivoSaida, a.output.String()); err != nil {
		return err
	}

	fmt.Printf("✅ Arquivo assembly criado: %s\n", arquivoSaida)
	return a.compilarAssembly(arquivoSaida)
}

// Implementação da interface visitor
func (a *ARM64Backend) Constante(constante *parser.Constante) interface{} {
	a.output.WriteString(fmt.Sprintf("    mov x0, #%d\n", constante.Valor))
	return nil
}

func (a *ARM64Backend) Variavel(variavel *parser.Variavel) interface{} {
	a.output.WriteString(fmt.Sprintf("    adrp x1, %s@PAGE\n", a.getVarName(variavel.Nome)))
	a.output.WriteString(fmt.Sprintf("    add x1, x1, %s@PAGEOFF\n", a.getVarName(variavel.Nome)))
	a.output.WriteString("    ldr x0, [x1]\n")
	return nil
}

func (a *ARM64Backend) Atribuicao(atribuicao *parser.Atribuicao) interface{} {
	a.declararVariavel(atribuicao.Nome)
	atribuicao.Valor.Aceitar(a)
	a.output.WriteString(fmt.Sprintf("    adrp x1, %s@PAGE\n", a.getVarName(atribuicao.Nome)))
	a.output.WriteString(fmt.Sprintf("    add x1, x1, %s@PAGEOFF\n", a.getVarName(atribuicao.Nome)))
	a.output.WriteString("    str x0, [x1]\n")
	return nil
}

func (a *ARM64Backend) OperacaoBinaria(operacao *parser.OperacaoBinaria) interface{} {
	// Avalia operando esquerdo
	operacao.OperandoEsquerdo.Aceitar(a)
	a.output.WriteString("    str x0, [sp, #-16]!\n") // Push x0 para stack

	// Avalia operando direito
	operacao.OperandoDireito.Aceitar(a)
	a.output.WriteString("    ldr x1, [sp], #16\n") // Pop para x1

	// Executa operação (x1 = esquerdo, x0 = direito)
	switch operacao.Operador {
	case parser.ADICAO:
		a.output.WriteString("    add x0, x1, x0\n")
	case parser.SUBTRACAO:
		a.output.WriteString("    sub x0, x1, x0\n")
	case parser.MULTIPLICACAO:
		a.output.WriteString("    mul x0, x1, x0\n")
	case parser.DIVISAO:
		a.output.WriteString("    sdiv x0, x1, x0\n")
	case parser.POWER:
		a.gerarPotencia()
	}
	return nil
}

func (a *ARM64Backend) ChamadaFuncao(chamada *parser.ChamadaFuncao) interface{} {
	// Valida a função usando o registro
	assinatura, ok := registry.RegistroGlobal.ObterAssinatura(chamada.Nome)
	if !ok {
		// Função não encontrada - erro de compilação
		return nil
	}

	// Valida número de argumentos
	numArgs := len(chamada.Argumentos)
	if numArgs < assinatura.MinArgumentos {
		return nil
	}
	if assinatura.MaxArgumentos != -1 && numArgs > assinatura.MaxArgumentos {
		return nil
	}

	// Gera assembly baseado no tipo da função
	switch assinatura.TipoFuncao {
	case registry.FUNCAO_IMPRIME:
		a.gerarAssemblyImprime(chamada.Argumentos)
	case registry.FUNCAO_PURA:
		a.gerarAssemblyFuncaoPura(chamada.Nome, chamada.Argumentos)
	}
	return nil
}

// gerarAssemblyImprime gera código ARM64 para a função imprime
func (a *ARM64Backend) gerarAssemblyImprime(argumentos []parser.Expressao) {
	for i, argumento := range argumentos {
		argumento.Aceitar(a)
		a.output.WriteString("    bl _print_num\n")

		// Se não for o último argumento, imprime espaço
		if i < len(argumentos)-1 {
			a.output.WriteString("    mov x0, #32\n") // ASCII espaço
			a.output.WriteString("    bl _print_char\n")
		}
	}
	// Imprime nova linha no final
	a.output.WriteString("    mov x0, #10\n") // ASCII newline
	a.output.WriteString("    bl _print_char\n")
}

// gerarAssemblyFuncaoPura gera código ARM64 para funções puras
func (a *ARM64Backend) gerarAssemblyFuncaoPura(nome string, argumentos []parser.Expressao) {
	// Para funções puras, avalia argumentos e chama função específica
	switch nome {
	case "soma":
		a.gerarAssemblySoma(argumentos)
	case "abs":
		a.gerarAssemblyAbs(argumentos)
	}
}

// gerarAssemblySoma gera assembly ARM64 para soma de múltiplos valores
func (a *ARM64Backend) gerarAssemblySoma(argumentos []parser.Expressao) {
	if len(argumentos) == 0 {
		a.output.WriteString("    mov x0, #0\n")
		return
	}

	// Avalia primeiro argumento
	argumentos[0].Aceitar(a)

	// Soma os demais argumentos
	for i := 1; i < len(argumentos); i++ {
		a.output.WriteString("    str x0, [sp, #-16]!\n") // Salva resultado atual
		argumentos[i].Aceitar(a)
		a.output.WriteString("    ldr x1, [sp], #16\n") // Recupera resultado anterior
		a.output.WriteString("    add x0, x1, x0\n")    // Soma
	}
}

// gerarAssemblyAbs gera assembly ARM64 para valor absoluto
func (a *ARM64Backend) gerarAssemblyAbs(argumentos []parser.Expressao) {
	if len(argumentos) != 1 {
		return // Erro: abs requer exatamente 1 argumento
	}

	argumentos[0].Aceitar(a)
	a.output.WriteString("    cmp x0, #0\n")
	a.output.WriteString("    cneg x0, x0, mi\n") // Nega se for negativo
}

// Gera código para potenciação
func (a *ARM64Backend) gerarPotencia() {
	labelLoop := fmt.Sprintf(".pow_loop_%d", a.labelCount)
	labelDone := fmt.Sprintf(".pow_done_%d", a.labelCount)
	labelZero := fmt.Sprintf(".pow_zero_%d", a.labelCount)
	a.labelCount++

	a.output.WriteString("    mov x2, #1\n")                      // resultado = 1
	a.output.WriteString("    cmp x0, #0\n")                      // compara expoente com 0
	a.output.WriteString(fmt.Sprintf("    b.eq %s\n", labelZero)) // se expoente = 0, resultado = 1
	a.output.WriteString(fmt.Sprintf("%s:\n", labelLoop))
	a.output.WriteString("    mul x2, x2, x1\n") // resultado *= base
	a.output.WriteString("    sub x0, x0, #1\n") // expoente--
	a.output.WriteString("    cmp x0, #0\n")
	a.output.WriteString(fmt.Sprintf("    b.ne %s\n", labelLoop)) // se expoente != 0, continua loop
	a.output.WriteString(fmt.Sprintf("%s:\n", labelZero))
	a.output.WriteString("    mov x0, x2\n") // move resultado para x0
	a.output.WriteString(fmt.Sprintf("%s:\n", labelDone))
}

func (a *ARM64Backend) gerarPrologo() {
	a.output.WriteString("// ARM64 Assembly for macOS\n")
	a.output.WriteString(".section __DATA,__data\n")
	a.output.WriteString(".align 3\n")
	// Variáveis serão adicionadas dinamicamente aqui

	a.output.WriteString("\n.section __TEXT,__text\n")
	a.output.WriteString(".globl _main\n")
	a.output.WriteString(".align 2\n\n")

	a.output.WriteString("_main:\n")
	a.output.WriteString("    // Salva frame pointer\n")
	a.output.WriteString("    stp x29, x30, [sp, #-16]!\n")
	a.output.WriteString("    mov x29, sp\n\n")
}

func (a *ARM64Backend) gerarEpilogo() {
	a.output.WriteString("\n    // Limpa e retorna\n")
	a.output.WriteString("    mov x0, #0\n") // exit code 0
	a.output.WriteString("    ldp x29, x30, [sp], #16\n")
	a.output.WriteString("    ret\n\n")

	// Adiciona funções auxiliares para impressão
	a.gerarFuncoesAuxiliares()

	// Adiciona seção de dados das variáveis
	a.gerarSecaoVariaveis()
}

func (a *ARM64Backend) gerarFuncoesAuxiliares() {
	a.output.WriteString("// Função para imprimir número\n")
	a.output.WriteString("_print_num:\n")
	a.output.WriteString("    stp x29, x30, [sp, #-16]!\n")
	a.output.WriteString("    mov x29, sp\n")
	a.output.WriteString("    \n")
	a.output.WriteString("    // Converter número para string e imprimir\n")
	a.output.WriteString("    // Implementação simplificada - usa printf do sistema\n")
	a.output.WriteString("    adrp x1, format_str@PAGE\n")
	a.output.WriteString("    add x1, x1, format_str@PAGEOFF\n")
	a.output.WriteString("    mov x1, x0\n")
	a.output.WriteString("    adrp x0, format_str@PAGE\n")
	a.output.WriteString("    add x0, x0, format_str@PAGEOFF\n")
	a.output.WriteString("    bl _printf\n")
	a.output.WriteString("    \n")
	a.output.WriteString("    ldp x29, x30, [sp], #16\n")
	a.output.WriteString("    ret\n\n")

	a.output.WriteString("// Função para imprimir caractere\n")
	a.output.WriteString("_print_char:\n")
	a.output.WriteString("    stp x29, x30, [sp, #-16]!\n")
	a.output.WriteString("    mov x29, sp\n")
	a.output.WriteString("    \n")
	a.output.WriteString("    adrp x1, char_format@PAGE\n")
	a.output.WriteString("    add x1, x1, char_format@PAGEOFF\n")
	a.output.WriteString("    mov x1, x0\n")
	a.output.WriteString("    adrp x0, char_format@PAGE\n")
	a.output.WriteString("    add x0, x0, char_format@PAGEOFF\n")
	a.output.WriteString("    bl _printf\n")
	a.output.WriteString("    \n")
	a.output.WriteString("    ldp x29, x30, [sp], #16\n")
	a.output.WriteString("    ret\n\n")

	// Strings de formato
	a.output.WriteString(".section __DATA,__cstring\n")
	a.output.WriteString("format_str: .asciz \"%ld\"\n")
	a.output.WriteString("char_format: .asciz \"%c\"\n")
}

func (a *ARM64Backend) gerarSecaoVariaveis() {
	if len(a.variables) > 0 {
		a.output.WriteString("\n.section __DATA,__data\n")
		a.output.WriteString(".align 3\n")
		for varName := range a.variables {
			a.output.WriteString(fmt.Sprintf("%s: .quad 0\n", a.getVarName(varName)))
		}
	}
}

func (a *ARM64Backend) declararVariavel(nome string) {
	a.variables[nome] = true
}

func (a *ARM64Backend) getVarName(nome string) string {
	return "_var_" + nome
}

func (a *ARM64Backend) compilarAssembly(arquivoAssembly string) error {
	fmt.Printf("🔨 Compilando assembly para executável...\n")

	arquivoObj := strings.TrimSuffix(arquivoAssembly, ".s") + ".o"
	arquivoExe := strings.TrimSuffix(arquivoAssembly, ".s")

	// Compilar assembly para objeto
	cmdAssemble := exec.Command("as", "-64", "-o", arquivoObj, arquivoAssembly)
	if err := cmdAssemble.Run(); err != nil {
		return fmt.Errorf("erro ao compilar assembly: %v", err)
	}

	// Linkar com bibliotecas do sistema (inclui printf)
	cmdLink := exec.Command("ld", "-o", arquivoExe, arquivoObj, "-lSystem", "-syslibroot", "/Library/Developer/CommandLineTools/SDKs/MacOSX.sdk", "-e", "_main", "-arch", "arm64")
	if err := cmdLink.Run(); err != nil {
		return fmt.Errorf("erro ao linkar: %v", err)
	}

	fmt.Printf("✅ Executável criado: %s\n", arquivoExe)
	fmt.Printf("💡 Execute com: %s\n", arquivoExe)

	return nil
}
