package arm64

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/khevencolino/Solar/internal/parser"
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

func (a *ARM64Backend) GetName() string      { return "Assembly ARM64" }
func (a *ARM64Backend) GetExtension() string { return ".s" }

func (a *ARM64Backend) Compile(statements []parser.Expressao) error {
	fmt.Printf("🔧 Compilando para Assembly ARM64...\n")
	a.gerarPrologo()

	for i, stmt := range statements {
		fmt.Printf(" Processando statement %d...\n", i+1)
		a.checarExpressao(stmt)
		if i == len(statements)-1 {
			// Placeholder: Call print function
			a.output.WriteString("    bl imprime_num\n")
		}
	}

	a.gerarEpilogo()

	arquivoSaida := filepath.Join("result", "programa.s")
	if err := utils.EscreverArquivo(arquivoSaida, a.output.String()); err != nil {
		return err
	}
	fmt.Println("Arquivo assembly criado com sucesso: ", arquivoSaida)
	return a.compilarAssembly(arquivoSaida)
}

func (a *ARM64Backend) checarExpressao(expr parser.Expressao) {
	switch e := expr.(type) {
	case *parser.Constante:
		a.output.WriteString(fmt.Sprintf("    mov x0, #%d\n", e.Valor))
	case *parser.Variavel:
		a.output.WriteString(fmt.Sprintf("    ldr x0, =%s\n", a.getVarName(e.Nome)))
		a.output.WriteString("    ldr x0, [x0]\n")
	case *parser.Atribuicao:
		a.declararVariavel(e.Nome)
		a.checarExpressao(e.Valor)
		a.output.WriteString(fmt.Sprintf("    ldr x1, =%s\n", a.getVarName(e.Nome)))
		a.output.WriteString("    str x0, [x1]\n")
	case *parser.OperacaoBinaria:
		a.checarExpressao(e.OperandoEsquerdo)
		a.output.WriteString("    str x0, [sp, #-16]!\n") // Push x0 onto the stack
		a.checarExpressao(e.OperandoDireito)
		a.output.WriteString("    ldr x1, [sp], #16\n") // Pop x0 into x1
		switch e.Operador {
		case parser.ADICAO:
			a.output.WriteString("    add x0, x1, x0\n")
		case parser.SUBTRACAO:
			a.output.WriteString("    sub x0, x1, x0\n")
		case parser.MULTIPLICACAO:
			a.output.WriteString("    mul x0, x1, x0\n")
		case parser.DIVISAO:
			a.output.WriteString("    sdiv x0, x1, x0\n")
		case parser.POWER:
			// This is a simplified power implementation for illustration
			// A more robust implementation would be needed
			a.output.WriteString("    mov x2, #1\n")
			a.output.WriteString(fmt.Sprintf("    cmp x0, #0\n    b.eq .pow_done_%d\n", a.labelCount))
			a.output.WriteString(fmt.Sprintf(".pow_loop_%d:\n", a.labelCount))
			a.output.WriteString("    mul x2, x2, x1\n")
			a.output.WriteString("    sub x0, x0, #1\n")
			a.output.WriteString(fmt.Sprintf("    cmp x0, #0\n    b.ne .pow_loop_%d\n", a.labelCount))
			a.output.WriteString(fmt.Sprintf(".pow_done_%d:\n", a.labelCount))
			a.output.WriteString("    mov x0, x2\n")
			a.labelCount++
		}
	}
}

func (a *ARM64Backend) gerarPrologo() {
	a.output.WriteString(".data\n")
	a.output.WriteString("\n.text\n")
	a.output.WriteString(".global _start\n\n")
	a.output.WriteString("_start:\n")
}

func (a *ARM64Backend) gerarEpilogo() {
	a.output.WriteString("    bl sair\n\n")

	if len(a.variables) > 0 {
		dataSection := ".data\n"
		for varName := range a.variables {
			dataSection += fmt.Sprintf("%s: .quad 0\n", a.getVarName(varName))
		}
		fullCode := strings.Replace(a.output.String(), ".data\n", dataSection, 1)
		a.output.Reset()
		a.output.WriteString(fullCode)
	}

	a.output.WriteString(".include \"external/runtime_arm.s\"\n")
}

func (a *ARM64Backend) declararVariavel(nome string) {
	a.variables[nome] = true
}

func (a *ARM64Backend) getVarName(nome string) string {
	return "var_" + nome
}

func (a *ARM64Backend) compilarAssembly(arquivoAssembly string) error {
	fmt.Printf("🧑‍💻 Criando arquivo executavel...\n")
	fmt.Printf("🔗 Linkando com runtime...\n")

	objectFile := filepath.Join("result", "programa.o")
	cmdAs := exec.Command("as", "-o", objectFile, arquivoAssembly)
	if err := cmdAs.Run(); err != nil {
		return fmt.Errorf("erro ao montar (as): %v", err)
	}

	executavel := filepath.Join("result", "programa")
	cmdLd := exec.Command("ld", "-o", executavel, objectFile)
	if err := cmdLd.Run(); err != nil {
		return fmt.Errorf("erro ao ligar (ld): %v", err)
	}

	fmt.Printf("✅ Executável gerado: %s\n", executavel)
	fmt.Printf("🏃 Para executar: ./%s\n", executavel)
	return nil
}
