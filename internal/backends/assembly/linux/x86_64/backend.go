package x86_64

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
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

// Registradores para passagem de argumentos (System V ABI)
var paramRegisters = []string{"%rdi", "%rsi", "%rdx", "%rcx", "%r8", "%r9"}

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

	funcaoPrincipal := a.coletarFuncoes(statements)
	a.gerarPrologo()
	a.gerarCorpoFuncoes()
	a.gerarPontoEntrada(funcaoPrincipal, statements)
	a.gerarEpilogo()

	return a.escreverECompilar("programa.s")
}

// coletarFuncoes indexa todas as declarações de função
func (a *X86_64Backend) coletarFuncoes(statements []parser.Expressao) *parser.FuncaoDeclaracao {
	var funcaoPrincipal *parser.FuncaoDeclaracao
	for _, s := range statements {
		if fn, ok := s.(*parser.FuncaoDeclaracao); ok {
			a.functions[fn.Nome] = fn
			if fn.Nome == "principal" {
				funcaoPrincipal = fn
			}
		}
	}
	return funcaoPrincipal
}

// gerarCorpoFuncoes emite o código de todas as funções do usuário
func (a *X86_64Backend) gerarCorpoFuncoes() {
	if len(a.functions) == 0 {
		return
	}

	nomes := make([]string, 0, len(a.functions))
	for n := range a.functions {
		nomes = append(nomes, n)
	}
	sort.Strings(nomes)

	for _, nome := range nomes {
		a.gerarFuncaoUsuario(nome, a.functions[nome])
	}
}

// gerarPontoEntrada cria o _start e executa código principal
func (a *X86_64Backend) gerarPontoEntrada(funcaoPrincipal *parser.FuncaoDeclaracao, statements []parser.Expressao) {
	a.emit(".global _start")
	a.emit("_start:")

	if funcaoPrincipal != nil {
		debug.Printf("  Chamando função principal()...\n")
		a.emit("    call func_principal")
	} else {
		a.processarStatementsGlobais(statements)
	}
}

// processarStatementsGlobais executa código fora de funções
func (a *X86_64Backend) processarStatementsGlobais(statements []parser.Expressao) {
	for i, stmt := range statements {
		if _, ok := stmt.(*parser.FuncaoDeclaracao); !ok {
			debug.Printf("  Processando statement global %d...\n", i+1)
			a.checarExpressao(stmt)
		}
	}
}

// escreverECompilar salva o assembly e gera executável
func (a *X86_64Backend) escreverECompilar(nomeArquivo string) error {
	if err := utils.EscreverArquivo(nomeArquivo, a.output.String()); err != nil {
		return err
	}

	fmt.Println("Arquivo assembly criado com sucesso: ", nomeArquivo)
	return a.compilarAssembly(nomeArquivo)
}

func (a *X86_64Backend) checarExpressao(expr parser.Expressao) {
	expr.Aceitar(a)
}

// emit escreve uma linha de assembly
func (a *X86_64Backend) emit(code string) {
	a.output.WriteString(code)
	a.output.WriteString("\n")
}

// emitf escreve uma linha de assembly formatada
func (a *X86_64Backend) emitf(format string, args ...interface{}) {
	a.output.WriteString(fmt.Sprintf(format, args...))
	a.output.WriteString("\n")
}

// === Implementação da interface visitor ===

func (a *X86_64Backend) Constante(constante *parser.Constante) interface{} {
	a.emitf("    mov $%d, %%rax", constante.Valor)
	return nil
}

func (a *X86_64Backend) Booleano(b *parser.Booleano) interface{} {
	valor := 0
	if b.Valor {
		valor = 1
	}
	a.emitf("    mov $%d, %%rax", valor)
	return nil
}

func (a *X86_64Backend) LiteralTexto(literal *parser.LiteralTexto) interface{} {
	label := a.criarLabelString(literal.Valor)
	a.emitf("    lea %s(%%rip), %%rax", label)
	return nil
}

func (a *X86_64Backend) LiteralDecimal(literal *parser.LiteralDecimal) interface{} {
	label := a.criarLabelDecimal(literal.Valor)
	a.emitf("    movsd %s(%%rip), %%xmm0", label)
	a.emit("    cvttsd2si %xmm0, %rax")
	return nil
}

func (a *X86_64Backend) Variavel(variavel *parser.Variavel) interface{} {
	a.emitf("    mov %s(%%rip), %%rax", a.getVarName(variavel.Nome))
	return nil
}

func (a *X86_64Backend) Atribuicao(atribuicao *parser.Atribuicao) interface{} {
	a.declararVariavel(atribuicao.Nome)
	atribuicao.Valor.Aceitar(a)
	a.emitf("    mov %%rax, %s(%%rip)", a.getVarName(atribuicao.Nome))
	return nil
}

func (a *X86_64Backend) OperacaoBinaria(operacao *parser.OperacaoBinaria) interface{} {
	operacao.OperandoEsquerdo.Aceitar(a)
	a.emit("    push %rax")

	operacao.OperandoDireito.Aceitar(a)
	a.emit("    mov %rax, %rbx")
	a.emit("    pop %rax")

	switch operacao.Operador {
	case parser.ADICAO:
		a.emit("    add %rbx, %rax")
	case parser.SUBTRACAO:
		a.emit("    sub %rbx, %rax")
	case parser.MULTIPLICACAO:
		a.emit("    imul %rbx, %rax")
	case parser.DIVISAO:
		a.emit("    cqo")
		a.emit("    idiv %rbx")
	case parser.POWER:
		a.gerarPotencia()
	case parser.IGUALDADE, parser.DIFERENCA, parser.MENOR_QUE, parser.MAIOR_QUE, parser.MENOR_IGUAL, parser.MAIOR_IGUAL:
		a.gerarComparacao(operacao.Operador)
	}

	return nil
}

// gerarPotencia implementa exponenciação via loop
func (a *X86_64Backend) gerarPotencia() {
	id := a.reserveID()
	loopLabel := fmt.Sprintf(".pow_loop_%d", id)
	doneLabel := fmt.Sprintf(".pow_done_%d", id)

	a.emit("    mov %rax, %rcx")
	a.emit("    mov $1, %rax")
	a.emit("    test %rbx, %rbx")
	a.emitf("    jz %s", doneLabel)
	a.emitf("%s:", loopLabel)
	a.emit("    imul %rcx, %rax")
	a.emit("    dec %rbx")
	a.emitf("    jnz %s", loopLabel)
	a.emitf("%s:", doneLabel)
}

// gerarComparacao gera código para operadores relacionais
func (a *X86_64Backend) gerarComparacao(op parser.TipoOperador) {
	instrucoes := map[parser.TipoOperador]string{
		parser.IGUALDADE:   "sete",
		parser.DIFERENCA:   "setne",
		parser.MENOR_QUE:   "setl",
		parser.MAIOR_QUE:   "setg",
		parser.MENOR_IGUAL: "setle",
		parser.MAIOR_IGUAL: "setge",
	}

	a.emit("    cmp %rbx, %rax")
	a.emitf("    %s %%al", instrucoes[op])
	a.emit("    movzx %al, %rax")
}

func (a *X86_64Backend) ChamadaFuncao(chamada *parser.ChamadaFuncao) interface{} {
	if _, ok := a.functions[chamada.Nome]; ok {
		a.gerarChamadaFuncaoUsuario(chamada)
		return nil
	}

	assinatura, ok := registry.RegistroGlobal.ObterAssinatura(chamada.Nome)
	if !ok {
		return nil
	}

	numArgs := len(chamada.Argumentos)
	if numArgs < assinatura.MinArgumentos {
		return nil
	}
	if assinatura.MaxArgumentos != -1 && numArgs > assinatura.MaxArgumentos {
		return nil
	}

	switch assinatura.TipoFuncao {
	case registry.FUNCAO_IMPRIME:
		a.gerarAssemblyImprime(chamada.Argumentos)
	case registry.FUNCAO_PURA:
		a.gerarAssemblyFuncaoPura(chamada.Nome, chamada.Argumentos)
	}
	return nil
}

// gerarChamadaFuncaoUsuario gera código para chamar funções definidas pelo usuário
func (a *X86_64Backend) gerarChamadaFuncaoUsuario(chamada *parser.ChamadaFuncao) {
	n := len(chamada.Argumentos)
	maxRegs := len(paramRegisters)

	for idx := 0; idx < n && idx < maxRegs; idx++ {
		chamada.Argumentos[idx].Aceitar(a)
		a.emitf("    mov %%rax, %s", paramRegisters[idx])
	}
	a.emitf("    call func_%s", chamada.Nome)
}

func (a *X86_64Backend) gerarAssemblyImprime(argumentos []parser.Expressao) {
	for _, argumento := range argumentos {
		argumento.Aceitar(a)
		a.emit("    call imprime_num")
	}
}

func (a *X86_64Backend) gerarAssemblyFuncaoPura(nome string, argumentos []parser.Expressao) {
	// Funções puras builtin delegadas ao registry
}

func (a *X86_64Backend) gerarPrologo() {
	a.emit(".section .text")
	a.emit(`.include "external/runtime.s"`)
	a.emit("")
	a.emit(".section .data")
	a.emit("")
	a.emit(".section .text")
}

func (a *X86_64Backend) gerarEpilogo() {
	a.emit("    call sair")
	a.emit("")

	if len(a.variables) == 0 && len(a.decimals) == 0 && len(a.strings) == 0 {
		return
	}

	dataSection := a.construirSecaoDados()
	fullCode := strings.Replace(a.output.String(), ".section .data\n", dataSection, 1)
	a.output.Reset()
	a.output.WriteString(fullCode)
}

// construirSecaoDados gera a seção .data completa
func (a *X86_64Backend) construirSecaoDados() string {
	var sb strings.Builder
	sb.WriteString(".section .data\n")

	for varName := range a.variables {
		sb.WriteString(fmt.Sprintf("%s: .quad 0\n", a.getVarName(varName)))
	}

	for label, valor := range a.decimals {
		bits := fmt.Sprintf("0x%016x", *(*uint64)(unsafe.Pointer(&valor)))
		sb.WriteString(fmt.Sprintf("%s: .quad %s\n", label, bits))
	}

	for label, valor := range a.strings {
		escapedStr := strings.ReplaceAll(valor, "\\", "\\\\")
		escapedStr = strings.ReplaceAll(escapedStr, "\"", "\\\"")
		sb.WriteString(fmt.Sprintf("%s: .ascii \"%s\\0\"\n", label, escapedStr))
	}

	return sb.String()
}

func (a *X86_64Backend) ComandoSe(comando *parser.ComandoSe) interface{} {
	id := a.reserveID()
	labelFim := fmt.Sprintf(".if_fim_%d", id)
	labelSenao := fmt.Sprintf(".if_senao_%d", id)

	comando.Condicao.Aceitar(a)
	a.emit("    test %rax, %rax")

	if comando.BlocoSenao != nil {
		a.emitf("    jz %s", labelSenao)
	} else {
		a.emitf("    jz %s", labelFim)
	}

	posBefore := a.output.Len()
	comando.BlocoSe.Aceitar(a)

	blockCode := a.output.String()[posBefore:]
	hasReturn := strings.HasSuffix(strings.TrimSpace(blockCode), "ret")

	if comando.BlocoSenao != nil {
		if !hasReturn {
			a.emitf("    jmp %s", labelFim)
		}
		a.emitf("%s:", labelSenao)
		comando.BlocoSenao.Aceitar(a)
	}

	a.emitf("%s:", labelFim)
	return nil
}

func (a *X86_64Backend) Bloco(bloco *parser.Bloco) interface{} {
	for _, comando := range bloco.Comandos {
		comando.Aceitar(a)
		if _, isReturn := comando.(*parser.Retorno); isReturn {
			break
		}
	}
	return nil
}

func (a *X86_64Backend) ComandoEnquanto(cmd *parser.ComandoEnquanto) interface{} {
	id := a.reserveID()
	lcond := fmt.Sprintf(".while_cond_%d", id)
	lbody := fmt.Sprintf(".while_body_%d", id)
	lend := fmt.Sprintf(".while_end_%d", id)

	a.emitf("    jmp %s", lcond)
	a.emitf("%s:", lbody)
	cmd.Corpo.Aceitar(a)
	a.emitf("    jmp %s", lcond)
	a.emitf("%s:", lcond)
	cmd.Condicao.Aceitar(a)
	a.emit("    test %rax, %rax")
	a.emitf("    jnz %s", lbody)
	a.emitf("%s:", lend)
	return nil
}

func (a *X86_64Backend) ComandoPara(cmd *parser.ComandoPara) interface{} {
	id := a.reserveID()
	lcond := fmt.Sprintf(".for_cond_%d", id)
	lbody := fmt.Sprintf(".for_body_%d", id)
	lstep := fmt.Sprintf(".for_step_%d", id)
	lend := fmt.Sprintf(".for_end_%d", id)

	if cmd.Inicializacao != nil {
		cmd.Inicializacao.Aceitar(a)
	}

	a.emitf("%s:", lcond)
	if cmd.Condicao != nil {
		cmd.Condicao.Aceitar(a)
		a.emit("    test %rax, %rax")
		a.emitf("    jz %s", lend)
	}

	a.emitf("%s:", lbody)
	cmd.Corpo.Aceitar(a)
	a.emitf("    jmp %s", lstep)

	a.emitf("%s:", lstep)
	if cmd.PosIteracao != nil {
		cmd.PosIteracao.Aceitar(a)
	}
	a.emitf("    jmp %s", lcond)

	a.emitf("%s:", lend)
	return nil
}

// gerarFuncaoUsuario emite código para uma função definida pelo usuário
func (a *X86_64Backend) gerarFuncaoUsuario(nome string, fn *parser.FuncaoDeclaracao) {
	a.emitf("func_%s:", nome)

	maxRegs := len(paramRegisters)
	for idx, p := range fn.Parametros {
		if idx >= maxRegs {
			break
		}
		a.declararVariavel(p.Nome)
		a.emitf("    mov %s, %s(%%rip)", paramRegisters[idx], a.getVarName(p.Nome))
	}

	fn.Corpo.Aceitar(a)
	a.emit("    ret")
	a.emit("")
}

func (a *X86_64Backend) FuncaoDeclaracao(fn *parser.FuncaoDeclaracao) interface{} {
	return nil
}

func (a *X86_64Backend) Retorno(ret *parser.Retorno) interface{} {
	if ret.Valor != nil {
		ret.Valor.Aceitar(a)
	}
	a.emit("    ret")
	return nil
}

func (a *X86_64Backend) Importacao(imp *parser.Importacao) interface{} {
	return nil
}

// === Helpers para gerenciamento de símbolos ===

func (a *X86_64Backend) declararVariavel(nome string) {
	a.variables[nome] = true
}

func (a *X86_64Backend) getVarName(nome string) string {
	return "var_" + nome
}

func (a *X86_64Backend) criarLabelString(valor string) string {
	id := a.reserveID()
	label := fmt.Sprintf("str_%d", id)
	a.strings[label] = valor
	return label
}

func (a *X86_64Backend) criarLabelDecimal(valor float64) string {
	id := a.reserveID()
	label := fmt.Sprintf("decimal_%d", id)
	a.decimals[label] = valor
	return label
}

func (a *X86_64Backend) reserveID() int {
	id := a.labelCount
	a.labelCount++
	return id
}

// === Compilação e linkagem ===

func (a *X86_64Backend) compilarAssembly(arquivoAssembly string) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("backend assembly linux só monta/linka em Linux; arquivo gerado: %s", arquivoAssembly)
	}

	debug.Printf("Criando arquivo executavel...\n")
	debug.Printf("Linkando com runtime...\n")

	objectFile := filepath.Join("result", "programa.o")
	if err := os.MkdirAll(filepath.Dir(objectFile), 0o755); err != nil {
		return fmt.Errorf("erro ao criar diretório de saída: %v", err)
	}

	if err := a.montar(arquivoAssembly, objectFile); err != nil {
		return err
	}

	executavel := filepath.Join("result", "programa")
	if err := a.linkar(objectFile, executavel); err != nil {
		return err
	}

	debug.Printf("Executável gerado: %s\n", executavel)
	debug.Printf("Para executar: ./%s\n", executavel)

	return nil
}

func (a *X86_64Backend) montar(arquivoAssembly, objectFile string) error {
	cmdAs := exec.Command("as", "-I", ".", "-o", objectFile, arquivoAssembly)
	if err := cmdAs.Run(); err != nil {
		return fmt.Errorf("erro ao montar (as): %v", err)
	}
	return nil
}

func (a *X86_64Backend) linkar(objectFile, executavel string) error {
	cmdLd := exec.Command("ld", "-o", executavel, objectFile)
	if err := cmdLd.Run(); err != nil {
		return fmt.Errorf("erro ao ligar (ld): %v", err)
	}
	return nil
}
