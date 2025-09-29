package x86_64

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"unsafe"

	"github.com/khevencolino/Solar/internal/debug"
	"github.com/khevencolino/Solar/internal/parser"
	"github.com/khevencolino/Solar/internal/registry"
	"github.com/khevencolino/Solar/internal/utils"
)

type X86_64Backend struct {
	output     strings.Builder
	variables  map[string]bool
	decimals   map[string]float64
	strings    map[string]string
	labelCount int
	functions  map[string]*parser.FuncaoDeclaracao
}

func NewX86_64Backend() *X86_64Backend {
	return &X86_64Backend{
		variables: make(map[string]bool),
		decimals:  make(map[string]float64),
		strings:   make(map[string]string),
		functions: make(map[string]*parser.FuncaoDeclaracao),
	}
}

func (a *X86_64Backend) GetName() string      { return "Assembly x86-64" }
func (a *X86_64Backend) GetExtension() string { return ".s" }

func (a *X86_64Backend) Compile(statements []parser.Expressao) error {
	debug.Printf("Compilando para Assembly x86-64...\n")

	// Primeira passada: coletar funções
	var funcaoPrincipal *parser.FuncaoDeclaracao
	for _, s := range statements {
		if fn, ok := s.(*parser.FuncaoDeclaracao); ok {
			a.functions[fn.Nome] = fn
			if fn.Nome == "principal" {
				funcaoPrincipal = fn
			}
		}
	}

	a.gerarPrologo()

	// Emite corpos de funções antes do _start
	for nome, fn := range a.functions {
		a.gerarFuncaoUsuario(nome, fn)
	}

	// Se existe função principal(), chama ela. Senão, executa statements globais
	if funcaoPrincipal != nil {
		debug.Printf("  Chamando função principal()...\n")
		// Chama a função principal()
		a.output.WriteString("    call principal\n")
	} else {
		// Processa statements globais (comportamento antigo)
		for i, stmt := range statements {
			// Pula declarações de função pois já foram processadas
			if _, ok := stmt.(*parser.FuncaoDeclaracao); !ok {
				debug.Printf("  Processando statement global %d...\n", i+1)
				a.checarExpressao(stmt)
			}
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
	// Suporte completo a números inteiros, incluindo negativos
	a.output.WriteString(fmt.Sprintf("    mov $%d, %%rax\n", constante.Valor))
	return nil
}

func (a *X86_64Backend) Booleano(b *parser.Booleano) interface{} {
	if b.Valor {
		a.output.WriteString("    mov $1, %rax\n")
	} else {
		a.output.WriteString("    mov $0, %rax\n")
	}
	return nil
}

func (a *X86_64Backend) LiteralTexto(literal *parser.LiteralTexto) interface{} {
	// Implementa suporte a strings criando um label único
	labelCount := a.labelCount
	a.labelCount++
	label := fmt.Sprintf("str_%d", labelCount)

	// Armazena a string na seção .data
	a.declararString(label, literal.Valor)

	// Carrega o endereço da string em %rax
	a.output.WriteString(fmt.Sprintf("    lea %s(%%rip), %%rax\n", label))

	return nil
}

func (a *X86_64Backend) LiteralDecimal(literal *parser.LiteralDecimal) interface{} {
	// Implementa suporte a números de ponto flutuante usando SSE
	// Cria um label único para armazenar o valor decimal na seção .data
	labelCount := a.labelCount
	a.labelCount++
	label := fmt.Sprintf("decimal_%d", labelCount)

	// Armazena o valor decimal como double (8 bytes)
	a.declararDecimal(label, literal.Valor)

	// Carrega o valor decimal no registrador XMM0 (ponto flutuante)
	a.output.WriteString(fmt.Sprintf("    movsd %s(%%rip), %%xmm0\n", label))

	// Para compatibilidade com código existente que espera inteiros em %rax,
	// converte o double para inteiro e move para %rax
	a.output.WriteString("    cvttsd2si %xmm0, %rax\n")

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
		// Gera labels únicos para evitar colisão quando POWER aparece múltiplas vezes
		id := a.labelCount
		a.labelCount++
		powLoop := fmt.Sprintf(".pow_loop_%d", id)
		powDone := fmt.Sprintf(".pow_done_%d", id)

		a.output.WriteString("    mov %rax, %rcx\n")  // copia a base de %rax para %rcx (base temporária)
		a.output.WriteString("    mov $1, %rax\n")    // inicializa o resultado em %rax com 1 (valor neutro da multiplicação)
		a.output.WriteString("    test %rbx, %rbx\n") // verifica se o expoente (%rbx) é zero
		a.output.WriteString(fmt.Sprintf("    jz %s\n", powDone))
		a.output.WriteString(fmt.Sprintf("%s:\n", powLoop))
		a.output.WriteString("    imul %rcx, %rax\n") // multiplica resultado (%rax) pela base (%rcx)
		a.output.WriteString("    dec %rbx\n")        // decrementa o expoente
		a.output.WriteString(fmt.Sprintf("    jnz %s\n", powLoop))
		a.output.WriteString(fmt.Sprintf("%s:\n", powDone)) // fim da exponenciação; %rax contém o resultado final

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
	// Função de usuário: chamada direta por label
	if _, ok := a.functions[chamada.Nome]; ok {
		// Avalia argumentos e coloca nos registradores na ordem: rdi, rsi, rdx, rcx, r8, r9
		regs := []string{"%rdi", "%rsi", "%rdx", "%rcx", "%r8", "%r9"}
		n := len(chamada.Argumentos)
		for idx := 0; idx < n && idx < len(regs); idx++ {
			chamada.Argumentos[idx].Aceitar(a)
			a.output.WriteString(fmt.Sprintf("    mov %%rax, %s\n", regs[idx]))
		}
		a.output.WriteString(fmt.Sprintf("    call func_%s\n", chamada.Nome))
		return nil
	}
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
	// Nenhuma função pura builtin implementada no momento
	// As funções puras são delegadas para o registry global
}

func (a *X86_64Backend) gerarPrologo() {
	// Código do runtime deve ficar na seção .text
	a.output.WriteString(".section .text\n")
	a.output.WriteString(`.include "external/runtime.s"`)
	// Cria um marcador da seção .data que será substituído no epílogo
	a.output.WriteString("\n.section .data\n")
	// Volta para .text para o ponto de entrada
	a.output.WriteString("\n.section .text\n")
	a.output.WriteString(".global _start\n\n")
	a.output.WriteString("_start:\n")
}

func (a *X86_64Backend) gerarEpilogo() {
	a.output.WriteString("    call sair\n\n")

	// Adiciona seção de dados para variáveis, decimais e strings
	if len(a.variables) > 0 || len(a.decimals) > 0 || len(a.strings) > 0 {
		dataSection := ".section .data\n"

		// Variáveis inteiras
		for varName := range a.variables {
			dataSection += fmt.Sprintf("%s: .quad 0\n", a.getVarName(varName))
		}

		// Decimais (double precision)
		for label, valor := range a.decimals {
			// Converte float64 para representação hexadecimal IEEE 754
			bits := fmt.Sprintf("0x%016x", *(*uint64)(unsafe.Pointer(&valor)))
			dataSection += fmt.Sprintf("%s: .quad %s\n", label, bits)
		}

		// Strings
		for label, valor := range a.strings {
			// Escapa caracteres especiais e adiciona terminador nulo
			escapedStr := strings.ReplaceAll(valor, "\\", "\\\\")
			escapedStr = strings.ReplaceAll(escapedStr, "\"", "\\\"")
			dataSection += fmt.Sprintf("%s: .ascii \"%s\\0\"\n", label, escapedStr)
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

func (a *X86_64Backend) ComandoEnquanto(cmd *parser.ComandoEnquanto) interface{} {
	// Labels únicos
	id := a.labelCount
	a.labelCount++
	lcond := fmt.Sprintf(".while_cond_%d", id)
	lbody := fmt.Sprintf(".while_body_%d", id)
	lend := fmt.Sprintf(".while_end_%d", id)

	// Jump para condição
	a.output.WriteString(fmt.Sprintf("    jmp %s\n", lcond))
	a.output.WriteString(fmt.Sprintf("%s:\n", lbody))
	// Corpo
	cmd.Corpo.Aceitar(a)
	// Volta para condição
	a.output.WriteString(fmt.Sprintf("    jmp %s\n", lcond))
	// Condição
	a.output.WriteString(fmt.Sprintf("%s:\n", lcond))
	cmd.Condicao.Aceitar(a)
	a.output.WriteString("    test %rax, %rax\n")
	a.output.WriteString(fmt.Sprintf("    jnz %s\n", lbody))
	a.output.WriteString(fmt.Sprintf("%s:\n", lend))
	return nil
}

func (a *X86_64Backend) ComandoPara(cmd *parser.ComandoPara) interface{} {
	id := a.labelCount
	a.labelCount++
	lcond := fmt.Sprintf(".for_cond_%d", id)
	lbody := fmt.Sprintf(".for_body_%d", id)
	lstep := fmt.Sprintf(".for_step_%d", id)
	lend := fmt.Sprintf(".for_end_%d", id)

	// init
	if cmd.Inicializacao != nil {
		cmd.Inicializacao.Aceitar(a)
	}

	// Condição
	a.output.WriteString(fmt.Sprintf("%s:\n", lcond))
	if cmd.Condicao != nil {
		cmd.Condicao.Aceitar(a)
		a.output.WriteString("    test %rax, %rax\n")
		a.output.WriteString(fmt.Sprintf("    jz %s\n", lend))
	}

	// Corpo
	a.output.WriteString(fmt.Sprintf("%s:\n", lbody))
	cmd.Corpo.Aceitar(a)
	a.output.WriteString(fmt.Sprintf("    jmp %s\n", lstep))

	// Passo
	a.output.WriteString(fmt.Sprintf("%s:\n", lstep))
	if cmd.PosIteracao != nil {
		cmd.PosIteracao.Aceitar(a)
	}
	a.output.WriteString(fmt.Sprintf("    jmp %s\n", lcond))

	a.output.WriteString(fmt.Sprintf("%s:\n", lend))
	return nil
}

// Declaração/definição de função do usuário (básica)
func (a *X86_64Backend) gerarFuncaoUsuario(nome string, fn *parser.FuncaoDeclaracao) {
	// Convenção System V: parâmetros em rdi, rsi, rdx, rcx, r8, r9; retorno em rax
	a.output.WriteString(fmt.Sprintf("func_%s:\n", nome))
	// Mapear parâmetros para variáveis (como globais simples com mov)
	regs := []string{"%rdi", "%rsi", "%rdx", "%rcx", "%r8", "%r9"}
	for idx, p := range fn.Parametros {
		if idx >= len(regs) {
			break // mais de 6 parâmetros não são suportados
		}
		a.declararVariavel(p.Nome)
		a.output.WriteString(fmt.Sprintf("    mov %s, %s(%%rip)\n", regs[idx], a.getVarName(p.Nome)))
	}
	// Gerar corpo (usando as variáveis globais dos params)
	fn.Corpo.Aceitar(a)
	// Resultado esperado em %rax (pela última expressão)
	a.output.WriteString("    ret\n\n")
}

func (a *X86_64Backend) FuncaoDeclaracao(fn *parser.FuncaoDeclaracao) interface{} { return nil }
func (a *X86_64Backend) Retorno(ret *parser.Retorno) interface{}                  { return nil }
func (a *X86_64Backend) Importacao(imp *parser.Importacao) interface{}            { return nil }

func (a *X86_64Backend) declararVariavel(nome string) {
	a.variables[nome] = true
}

func (a *X86_64Backend) declararDecimal(nome string, valor float64) {
	a.decimals[nome] = valor
}

func (a *X86_64Backend) declararString(nome string, valor string) {
	a.strings[nome] = valor
}

func (a *X86_64Backend) getVarName(nome string) string {
	return "var_" + nome
}

func (a *X86_64Backend) compilarAssembly(arquivoAssembly string) error {
	// Este backend gera binários ELF Linux x86_64. Em outros SOs, apenas gere o .s.
	if runtime.GOOS != "linux" {
		return fmt.Errorf("backend assembly linux só monta/linka em Linux; arquivo gerado: %s", arquivoAssembly)
	}
	debug.Printf("Criando arquivo executavel...\n")
	debug.Printf("Linkando com runtime...\n")

	objectFile := filepath.Join("result", "programa.o")
	// Garante que o diretório de saída existe
	if err := os.MkdirAll(filepath.Dir(objectFile), 0o755); err != nil {
		return fmt.Errorf("erro ao criar diretório de saída: %v", err)
	}
	cmdAs := exec.Command("as", "-I", ".", "-o", objectFile, arquivoAssembly)
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
