package assembly

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/khevencolino/Solar/internal/parser"
	"github.com/khevencolino/Solar/internal/utils"
)

type AssemblyBackend struct {
	output     strings.Builder
	variables  map[string]bool
	labelCount int
}

func NewAssemblyBackend() *AssemblyBackend {
	return &AssemblyBackend{
		variables: make(map[string]bool),
	}
}

func (a *AssemblyBackend) GetName() string      { return "Assembly x86-64" }
func (a *AssemblyBackend) GetExtension() string { return ".s" }

func (a *AssemblyBackend) Compile(statements []parser.Expressao) error {
	fmt.Printf("🔧 Compilando para Assembly x86-64...\n")

	a.gerarPrologo()

	// Processa statements
	for i, stmt := range statements {
		fmt.Printf("  Processando statement %d...\n", i+1)
		a.checarExpressao(stmt)

		// Se for a última expressão, imprime resultado
		if i == len(statements)-1 {
			a.output.WriteString("    call imprime_num\n")
		}
	}

	a.gerarEpilogo()

	// Escreve arquivo assembly
	arquivoSaida := filepath.Join("result", "programa.s")
	if err := utils.EscreverArquivo(arquivoSaida, a.output.String()); err != nil {
		return err
	}

	fmt.Println("✅ Arquivo assembly criado com sucesso: ", arquivoSaida)

	// Compila assembly para executável
	return a.compilarAssembly(arquivoSaida)
}

func (a *AssemblyBackend) checarExpressao(expr parser.Expressao) {
	switch e := expr.(type) {
	case *parser.Constante:
		a.output.WriteString(fmt.Sprintf("    mov $%d, %%rax\n", e.Valor))

	case *parser.Variavel:
		a.output.WriteString(fmt.Sprintf("    mov %s(%%rip), %%rax\n", a.getVarName(e.Nome)))

	case *parser.Atribuicao:
		a.declararVariavel(e.Nome)
		a.checarExpressao(e.Valor)
		a.output.WriteString(fmt.Sprintf("    mov %%rax, %s(%%rip)\n", a.getVarName(e.Nome)))

	case *parser.OperacaoBinaria:
		// Operando esquerdo
		a.checarExpressao(e.OperandoEsquerdo)
		a.output.WriteString("    push %rax\n")

		// Operando direito
		a.checarExpressao(e.OperandoDireito)
		a.output.WriteString("    mov %rax, %rbx\n")
		a.output.WriteString("    pop %rax\n")

		// Operação
		switch e.Operador {
		case parser.ADICAO:
			a.output.WriteString("    add %rbx, %rax\n")
		case parser.SUBTRACAO:
			a.output.WriteString("    sub %rbx, %rax\n")
		case parser.MULTIPLICACAO:
			a.output.WriteString("    imul %rbx, %rax\n")
		case parser.DIVISAO:
			a.output.WriteString("    cqo\n")
			a.output.WriteString("    idiv %rbx\n")
		case parser.POWER:
			a.output.WriteString("    mov %rax, %rcx\n")  // copia a base de %rax para %rcx (base temporária)
			a.output.WriteString("    mov $1, %rax\n")    // inicializa o resultado em %rax com 1 (valor neutro da multiplicação)
			a.output.WriteString("    test %rbx, %rbx\n") // verifica se o expoente (%rbx) é zero
			a.output.WriteString("    jz .pow_done\n")    // se for zero, pula para o final (qualquer número^0 = 1)
			a.output.WriteString(".pow_loop:\n")
			a.output.WriteString("    imul %rax, %rcx\n") // multiplica resultado (%rax) pela base (%rcx)
			a.output.WriteString("    dec %rbx\n")        // decrementa o expoente
			a.output.WriteString("    jnz .pow_loop\n")   // se o expoente ainda não for zero, repete o loop
			a.output.WriteString(".pow_done:\n")          // fim da exponenciação; %rax contém o resultado final

		}

	}
}

func (a *AssemblyBackend) gerarPrologo() {
	a.output.WriteString(".section .data\n")
	// Variáveis serão adicionadas dinamicamente
	a.output.WriteString("\n.section .text\n")
	a.output.WriteString(".global _start\n\n")
	a.output.WriteString("_start:\n")
}

func (a *AssemblyBackend) gerarEpilogo() {
	a.output.WriteString("    call sair\n\n")

	// Adiciona seção de dados para variáveis
	if len(a.variables) > 0 {
		dataSection := ".section .data\n"
		for varName := range a.variables {
			dataSection += fmt.Sprintf("%s: .quad 0\n", a.getVarName(varName))
		}
		// Substitui seção de dados no início
		fullCode := strings.Replace(a.output.String(), ".section .data\n", dataSection, 1)
		a.output.Reset()
		a.output.WriteString(fullCode)
	}

	// Inclui runtime
	a.output.WriteString(".include \"external/runtime.s\"\n")
}

func (a *AssemblyBackend) declararVariavel(nome string) {
	a.variables[nome] = true
}

func (a *AssemblyBackend) getVarName(nome string) string {
	return "var_" + nome
}

func (a *AssemblyBackend) compilarAssembly(arquivoAssembly string) error {
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
