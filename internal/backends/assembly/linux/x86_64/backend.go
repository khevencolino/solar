package x86_64

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/khevencolino/Solar/internal/debug"
	"github.com/khevencolino/Solar/internal/parser"
	"github.com/khevencolino/Solar/internal/registry"
	"github.com/khevencolino/Solar/internal/utils"
)

type X86_64Backend struct {
	output     strings.Builder
	variables  map[string]bool
	labelCount int
}

func NewX86_64Backend() *X86_64Backend {
	return &X86_64Backend{
		variables: make(map[string]bool),
	}
}

func (a *X86_64Backend) GetName() string      { return "Assembly x86-64" }
func (a *X86_64Backend) GetExtension() string { return ".s" }

func (a *X86_64Backend) Compile(statements []parser.Expressao) error {
	debug.Printf("🔧 Compilando para Assembly x86-64...\n")

	a.gerarPrologo()

	// Processa statements
	for i, stmt := range statements {
		debug.Printf("  Processando statement %d...\n", i+1)
		a.checarExpressao(stmt)

		// Se for a última expressão, imprime resultado
		if i == len(statements)-1 {
			a.output.WriteString("    call imprime_num\n")
		}
	}

	a.gerarEpilogo()

	// Escreve arquivo assembly
	arquivoSaida := "programa.s"
	if err := utils.EscreverArquivo(arquivoSaida, a.output.String()); err != nil {
		return err
	}

	fmt.Println("Arquivo assembly criado com sucesso: ", arquivoSaida)

	// Compila assembly para executável
	return a.compilarAssembly(arquivoSaida)
}

func (a *X86_64Backend) checarExpressao(expr parser.Expressao) {
	// Usa o padrão visitor para gerar código assembly
	expr.Aceitar(a)
}

// Implementação da interface visitor
func (a *X86_64Backend) Constante(constante *parser.Constante) interface{} {
	a.output.WriteString(fmt.Sprintf("    mov $%d, %%rax\n", constante.Valor))
	return nil
}

func (a *X86_64Backend) Variavel(variavel *parser.Variavel) interface{} {
	a.output.WriteString(fmt.Sprintf("    mov %s(%%rip), %%rax\n", a.getVarName(variavel.Nome)))
	return nil
}

func (a *X86_64Backend) Atribuicao(atribuicao *parser.Atribuicao) interface{} {
	a.declararVariavel(atribuicao.Nome)
	atribuicao.Valor.Aceitar(a)
	a.output.WriteString(fmt.Sprintf("    mov %%rax, %s(%%rip)\n", a.getVarName(atribuicao.Nome)))
	return nil
}

func (a *X86_64Backend) OperacaoBinaria(operacao *parser.OperacaoBinaria) interface{} {
	// Operando esquerdo
	operacao.OperandoEsquerdo.Aceitar(a)
	a.output.WriteString("    push %rax\n")

	// Operando direito
	operacao.OperandoDireito.Aceitar(a)
	a.output.WriteString("    mov %rax, %rbx\n")
	a.output.WriteString("    pop %rax\n")

	// Operação
	switch operacao.Operador {
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

	// Operações de comparação
	case parser.IGUALDADE:
		a.output.WriteString("    cmp %rbx, %rax\n")
		a.output.WriteString("    sete %al\n")
		a.output.WriteString("    movzx %al, %rax\n")
	case parser.DIFERENCA:
		a.output.WriteString("    cmp %rbx, %rax\n")
		a.output.WriteString("    setne %al\n")
		a.output.WriteString("    movzx %al, %rax\n")
	case parser.MENOR_QUE:
		a.output.WriteString("    cmp %rbx, %rax\n")
		a.output.WriteString("    setl %al\n")
		a.output.WriteString("    movzx %al, %rax\n")
	case parser.MAIOR_QUE:
		a.output.WriteString("    cmp %rbx, %rax\n")
		a.output.WriteString("    setg %al\n")
		a.output.WriteString("    movzx %al, %rax\n")
	case parser.MENOR_IGUAL:
		a.output.WriteString("    cmp %rbx, %rax\n")
		a.output.WriteString("    setle %al\n")
		a.output.WriteString("    movzx %al, %rax\n")
	case parser.MAIOR_IGUAL:
		a.output.WriteString("    cmp %rbx, %rax\n")
		a.output.WriteString("    setge %al\n")
		a.output.WriteString("    movzx %al, %rax\n")
	}

	return nil
}

func (a *X86_64Backend) ChamadaFuncao(chamada *parser.ChamadaFuncao) interface{} {
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

// gerarAssemblyImprime gera código assembly para a função imprime
func (a *X86_64Backend) gerarAssemblyImprime(argumentos []parser.Expressao) {
	for _, argumento := range argumentos {
		argumento.Aceitar(a)
		a.output.WriteString("    call imprime_num\n")
	}
}

// gerarAssemblyFuncaoPura gera código assembly para funções puras
func (a *X86_64Backend) gerarAssemblyFuncaoPura(nome string, argumentos []parser.Expressao) {
	// Avalia argumentos
	for _, argumento := range argumentos {
		argumento.Aceitar(a)
		a.output.WriteString("    push %rax\n") // Salva argumento na pilha
	}

	// Chama função específica
	switch nome {
	case "soma":
		a.gerarAssemblySoma(len(argumentos))
	case "abs":
		a.gerarAssemblyAbs()
	}
}

// gerarAssemblySoma gera assembly para soma de múltiplos valores
func (a *X86_64Backend) gerarAssemblySoma(numArgs int) {
	if numArgs == 0 {
		a.output.WriteString("    mov $0, %rax\n")
		return
	}

	// Pop primeiro argumento
	a.output.WriteString("    pop %rax\n")

	// Soma os demais argumentos
	for i := 1; i < numArgs; i++ {
		a.output.WriteString("    pop %rbx\n")
		a.output.WriteString("    add %rbx, %rax\n")
	}
}

// gerarAssemblyAbs gera assembly para valor absoluto
func (a *X86_64Backend) gerarAssemblyAbs() {
	a.output.WriteString("    pop %rax\n")
	a.output.WriteString("    test %rax, %rax\n")
	a.output.WriteString("    jns abs_positive\n")
	a.output.WriteString("    neg %rax\n")
	a.output.WriteString("abs_positive:\n")
}

func (a *X86_64Backend) gerarPrologo() {
	a.output.WriteString(".section .data\n")
	// Variáveis serão adicionadas dinamicamente
	a.output.WriteString("\n.section .text\n")
	a.output.WriteString(".global _start\n\n")
	a.output.WriteString("_start:\n")
}

func (a *X86_64Backend) gerarEpilogo() {
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
}

func (a *X86_64Backend) ComandoSe(comando *parser.ComandoSe) interface{} {
	// Gera um label único para este comando if
	labelFim := fmt.Sprintf(".if_fim_%d", a.labelCount)
	labelSenao := fmt.Sprintf(".if_senao_%d", a.labelCount)
	a.labelCount++

	// Avalia a condição
	comando.Condicao.Aceitar(a)

	// Testa se o resultado da condição é 0 (falso)
	a.output.WriteString("    test %rax, %rax\n")

	if comando.BlocoSenao != nil {
		// Se há bloco senao, pula para o label senao se for falso
		a.output.WriteString(fmt.Sprintf("    jz %s\n", labelSenao))
	} else {
		// Se não há bloco senao, pula para o fim se for falso
		a.output.WriteString(fmt.Sprintf("    jz %s\n", labelFim))
	}

	// Executa o bloco do "se"
	comando.BlocoSe.Aceitar(a)

	if comando.BlocoSenao != nil {
		// Pula para o fim após executar o bloco "se"
		a.output.WriteString(fmt.Sprintf("    jmp %s\n", labelFim))

		// Label para o bloco "senao"
		a.output.WriteString(fmt.Sprintf("%s:\n", labelSenao))
		comando.BlocoSenao.Aceitar(a)
	}

	// Label para o fim do comando if
	a.output.WriteString(fmt.Sprintf("%s:\n", labelFim))

	return nil
}

func (a *X86_64Backend) Bloco(bloco *parser.Bloco) interface{} {
	// Executa todos os comandos do bloco
	for _, comando := range bloco.Comandos {
		comando.Aceitar(a)
	}
	return nil
}

func (a *X86_64Backend) declararVariavel(nome string) {
	a.variables[nome] = true
}

func (a *X86_64Backend) getVarName(nome string) string {
	return "var_" + nome
}

func (a *X86_64Backend) compilarAssembly(arquivoAssembly string) error {
	debug.Printf("Criando arquivo executavel...\n")
	debug.Printf("Linkando com runtime...\n")

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

	debug.Printf("Executável gerado: %s\n", executavel)
	debug.Printf("Para executar: ./%s\n", executavel)

	return nil
}
