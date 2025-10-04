package llvm

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/enum"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"

	"github.com/khevencolino/Solar/internal/debug"
	"github.com/khevencolino/Solar/internal/parser"
	"github.com/khevencolino/Solar/internal/registry"
	"github.com/khevencolino/Solar/internal/utils"
)

// LLVMBackend gera código LLVM IR a partir da AST
type LLVMBackend struct {
	module     *ir.Module
	block      *ir.Block
	function   *ir.Func
	variables  map[string]value.Value
	varStack   []map[string]value.Value
	userFuncs  map[string]*ir.Func
	printfFn   *ir.Func
	fmtGlobals map[string]*ir.Global
	tmpCount   int
	strCount   int
}

func NewLLVMBackend() *LLVMBackend {
	return &LLVMBackend{
		variables:  make(map[string]value.Value),
		userFuncs:  make(map[string]*ir.Func),
		fmtGlobals: make(map[string]*ir.Global),
	}
}

func (l *LLVMBackend) GetName() string      { return "LLVM IR" }
func (l *LLVMBackend) GetExtension() string { return ".ll" }

// Compile gera código LLVM IR
func (l *LLVMBackend) Compile(statements []parser.Expressao) error {
	debug.Printf("Compilando para LLVM IR...\n")

	l.initializeModule()

	mainFunc := l.findMainFunction(statements)
	l.declareUserFunctions(statements)
	l.defineUserFunctions(statements)
	l.generateMainFunction(mainFunc, statements)

	return l.writeAndCompile()
}

// initializeModule configura o módulo LLVM e funções internas
func (l *LLVMBackend) initializeModule() {
	l.module = ir.NewModule()

	// Declara printf para I/O
	l.printfFn = l.module.NewFunc("printf", types.I32,
		ir.NewParam("format", types.NewPointer(types.I8)))
	l.printfFn.Sig.Variadic = true
}

// findMainFunction localiza a função principal se existir
func (l *LLVMBackend) findMainFunction(statements []parser.Expressao) *parser.FuncaoDeclaracao {
	for _, st := range statements {
		if fn, ok := st.(*parser.FuncaoDeclaracao); ok && fn.Nome == "principal" {
			return fn
		}
	}
	return nil
}

// declareUserFunctions cria protótipos de todas as funções do usuário
func (l *LLVMBackend) declareUserFunctions(statements []parser.Expressao) {
	for _, st := range statements {
		if fn, ok := st.(*parser.FuncaoDeclaracao); ok {
			l.declareFunction(fn)
		}
	}
}

// defineUserFunctions gera o corpo de todas as funções do usuário
func (l *LLVMBackend) defineUserFunctions(statements []parser.Expressao) {
	for _, st := range statements {
		if fn, ok := st.(*parser.FuncaoDeclaracao); ok {
			l.defineFunction(fn)
		}
	}
}

// generateMainFunction cria a função main que é o ponto de entrada
func (l *LLVMBackend) generateMainFunction(mainFunc *parser.FuncaoDeclaracao, statements []parser.Expressao) {
	l.function = l.module.NewFunc("main", types.I32)
	l.block = l.function.NewBlock("")

	if mainFunc != nil {
		debug.Printf("  Chamando função principal()...\n")
		principalFunc := l.userFuncs[mainFunc.Nome]
		l.block.NewCall(principalFunc)
	} else {
		l.executeGlobalStatements(statements)
	}

	l.block.NewRet(l.constInt(0))
}

// executeGlobalStatements processa statements no escopo global
func (l *LLVMBackend) executeGlobalStatements(statements []parser.Expressao) {
	for i, stmt := range statements {
		if _, ok := stmt.(*parser.FuncaoDeclaracao); !ok {
			debug.Printf("  Processando statement global %d...\n", i+1)
			l.processExpression(stmt)
		}
	}
}

// writeAndCompile escreve o IR e tenta compilar para executável
func (l *LLVMBackend) writeAndCompile() error {
	arquivoSaida := "programa.ll"
	if err := utils.EscreverArquivo(arquivoSaida, l.module.String()); err != nil {
		return err
	}

	debug.Printf("Arquivo LLVM IR gerado em: %s\n", arquivoSaida)

	if err := l.compileToExecutable(arquivoSaida); err != nil {
		debug.Printf("Aviso: Não foi possível compilar para executável: %v\n", err)
		debug.Printf("Use 'clang programa.ll -o programa' para compilar manualmente\n")
	}

	return nil
}

// processExpression processa uma expressão e retorna seu valor
func (l *LLVMBackend) processExpression(expr parser.Expressao) value.Value {
	result := expr.Aceitar(l)
	if val, ok := result.(value.Value); ok {
		return val
	}
	return l.constInt(0)
}

// === Implementação da interface Visitor ===

func (l *LLVMBackend) Constante(c *parser.Constante) interface{} {
	return l.constInt(int64(c.Valor))
}

func (l *LLVMBackend) Booleano(b *parser.Booleano) interface{} {
	if b.Valor {
		return l.constInt(1)
	}
	return l.constInt(0)
}

func (l *LLVMBackend) LiteralTexto(literal *parser.LiteralTexto) interface{} {
	return l.createStringConstant(literal.Valor)
}

func (l *LLVMBackend) LiteralDecimal(literal *parser.LiteralDecimal) interface{} {
	return constant.NewFloat(types.Double, literal.Valor)
}

func (l *LLVMBackend) Variavel(v *parser.Variavel) interface{} {
	val, ok := l.getVariable(v.Nome)
	if !ok {
		fmt.Printf("Variável '%s' não definida\n", v.Nome)
		return l.constInt(0)
	}

	// Carrega valor se for alloca
	if alloca, isAlloca := val.(*ir.InstAlloca); isAlloca {
		return l.block.NewLoad(types.I64, alloca)
	}
	return val
}

func (l *LLVMBackend) Atribuicao(a *parser.Atribuicao) interface{} {
	valor := l.processExpression(a.Valor)

	// Atualiza variável existente ou cria nova
	if existente, ok := l.getVariable(a.Nome); ok {
		if alloca, isAlloca := existente.(*ir.InstAlloca); isAlloca {
			l.block.NewStore(valor, alloca)
			return valor
		}
	}

	// Cria nova variável
	alloca := l.block.NewAlloca(types.I64)
	l.block.NewStore(valor, alloca)
	l.setVariable(a.Nome, alloca)
	return valor
}

func (l *LLVMBackend) OperacaoBinaria(op *parser.OperacaoBinaria) interface{} {
	left := l.processExpression(op.OperandoEsquerdo)
	right := l.processExpression(op.OperandoDireito)

	if left == nil || right == nil {
		return l.constInt(0)
	}

	switch op.Operador {
	case parser.ADICAO:
		return l.block.NewAdd(left, right)
	case parser.SUBTRACAO:
		return l.block.NewSub(left, right)
	case parser.MULTIPLICACAO:
		return l.block.NewMul(left, right)
	case parser.DIVISAO:
		return l.safeDivision(left, right)
	case parser.POWER:
		return l.generatePowerOperation(left, right)
	default:
		return l.generateComparison(op.Operador, left, right)
	}
}

func (l *LLVMBackend) ChamadaFuncao(fn *parser.ChamadaFuncao) interface{} {
	// Função do usuário
	if userFunc, ok := l.userFuncs[fn.Nome]; ok {
		return l.callUserFunction(userFunc, fn.Argumentos)
	}

	// Função builtin
	if sig, ok := registry.RegistroGlobal.ObterAssinatura(fn.Nome); ok {
		return l.handleBuiltinFunction(fn, sig)
	}

	fmt.Printf("Função '%s' não implementada\n", fn.Nome)
	return l.constInt(0)
}

func (l *LLVMBackend) ComandoSe(cmd *parser.ComandoSe) interface{} {
	return l.generateIfStatement(cmd)
}

func (l *LLVMBackend) ComandoEnquanto(cmd *parser.ComandoEnquanto) interface{} {
	return l.generateWhileLoop(cmd)
}

func (l *LLVMBackend) ComandoPara(cmd *parser.ComandoPara) interface{} {
	return l.generateForLoop(cmd)
}

func (l *LLVMBackend) Bloco(b *parser.Bloco) interface{} {
	return l.processBlock(b)
}

func (l *LLVMBackend) FuncaoDeclaracao(fn *parser.FuncaoDeclaracao) interface{} {
	l.defineFunction(fn)
	return l.constInt(0)
}

func (l *LLVMBackend) Retorno(ret *parser.Retorno) interface{} {
	if ret.Valor != nil {
		val := l.processExpression(ret.Valor)
		if l.function != nil {
			l.block.NewRet(val)
		}
		return val
	}

	if l.function != nil {
		l.block.NewRet(l.constInt(0))
	}
	return l.constInt(0)
}

func (l *LLVMBackend) Importacao(imp *parser.Importacao) interface{} {
	return nil
}

// === Geração de Comparações ===

func (l *LLVMBackend) generateComparison(op parser.TipoOperador, left, right value.Value) value.Value {
	predicates := map[parser.TipoOperador]enum.IPred{
		parser.IGUALDADE:   enum.IPredEQ,
		parser.DIFERENCA:   enum.IPredNE,
		parser.MENOR_QUE:   enum.IPredSLT,
		parser.MAIOR_QUE:   enum.IPredSGT,
		parser.MENOR_IGUAL: enum.IPredSLE,
		parser.MAIOR_IGUAL: enum.IPredSGE,
	}

	if pred, ok := predicates[op]; ok {
		cmp := l.block.NewICmp(pred, left, right)
		return l.block.NewZExt(cmp, types.I64)
	}

	fmt.Printf("Operador não suportado: %s\n", op.String())
	return l.constInt(0)
}

// === Geração de Estruturas de Controle ===

func (l *LLVMBackend) generateIfStatement(cmd *parser.ComandoSe) value.Value {
	condition := l.processExpression(cmd.Condicao)
	condBool := l.block.NewICmp(enum.IPredNE, condition, l.constInt(0))

	thenBlock := l.function.NewBlock("")
	mergeBlock := l.function.NewBlock("")

	var elseBlock *ir.Block
	if cmd.BlocoSenao != nil {
		elseBlock = l.function.NewBlock("")
		l.block.NewCondBr(condBool, thenBlock, elseBlock)
	} else {
		l.block.NewCondBr(condBool, thenBlock, mergeBlock)
	}

	// Bloco then
	l.block = thenBlock
	thenValue := l.processBlock(cmd.BlocoSe)
	thenTerminated := l.block.Term != nil
	if !thenTerminated {
		l.block.NewBr(mergeBlock)
	}

	// Bloco else (se existir)
	var elseValue value.Value
	var elseTerminated bool
	if cmd.BlocoSenao != nil {
		l.block = elseBlock
		elseValue = l.processBlock(cmd.BlocoSenao)
		elseTerminated = l.block.Term != nil
		if !elseTerminated {
			l.block.NewBr(mergeBlock)
		}
	}

	l.block = mergeBlock

	// Ambos terminados = merge inalcançável
	if thenTerminated && (cmd.BlocoSenao == nil || elseTerminated) {
		return thenValue
	}

	// PHI node se ambos não terminaram
	if cmd.BlocoSenao != nil && !thenTerminated && !elseTerminated {
		return mergeBlock.NewPhi(
			ir.NewIncoming(thenValue, thenBlock),
			ir.NewIncoming(elseValue, elseBlock))
	}

	return thenValue
}

func (l *LLVMBackend) generateWhileLoop(cmd *parser.ComandoEnquanto) value.Value {
	condBlock := l.function.NewBlock("while.cond")
	bodyBlock := l.function.NewBlock("while.body")
	endBlock := l.function.NewBlock("while.end")

	l.block.NewBr(condBlock)

	// Condição
	l.block = condBlock
	condVal := l.processExpression(cmd.Condicao)
	condBool := l.block.NewICmp(enum.IPredNE, condVal, l.constInt(0))
	l.block.NewCondBr(condBool, bodyBlock, endBlock)

	// Corpo
	l.block = bodyBlock
	lastVal := l.processBlock(cmd.Corpo)
	if l.block.Term == nil {
		l.block.NewBr(condBlock)
	}

	l.block = endBlock
	return lastVal
}

func (l *LLVMBackend) generateForLoop(cmd *parser.ComandoPara) value.Value {
	// Inicialização
	if cmd.Inicializacao != nil {
		l.processExpression(cmd.Inicializacao)
	}

	condBlock := l.function.NewBlock("for.cond")
	bodyBlock := l.function.NewBlock("for.body")
	stepBlock := l.function.NewBlock("for.step")
	endBlock := l.function.NewBlock("for.end")

	l.block.NewBr(condBlock)

	// Condição
	l.block = condBlock
	var condBool value.Value
	if cmd.Condicao != nil {
		condVal := l.processExpression(cmd.Condicao)
		condBool = l.block.NewICmp(enum.IPredNE, condVal, l.constInt(0))
	} else {
		condBool = constant.NewInt(types.I1, 1) // Loop infinito se não há condição
	}
	l.block.NewCondBr(condBool, bodyBlock, endBlock)

	// Corpo
	l.block = bodyBlock
	lastVal := l.processBlock(cmd.Corpo)
	if l.block.Term == nil {
		l.block.NewBr(stepBlock)
	}

	// Incremento
	l.block = stepBlock
	if cmd.PosIteracao != nil {
		l.processExpression(cmd.PosIteracao)
	}
	if l.block.Term == nil {
		l.block.NewBr(condBlock)
	}

	l.block = endBlock
	return lastVal
}

// === Funções do Usuário ===

func (l *LLVMBackend) declareFunction(fn *parser.FuncaoDeclaracao) {
	params := make([]*ir.Param, len(fn.Parametros))
	for i, p := range fn.Parametros {
		params[i] = ir.NewParam(p.Nome, types.I64)
	}

	funcDef := l.module.NewFunc(fn.Nome, types.I64, params...)
	l.userFuncs[fn.Nome] = funcDef
}

func (l *LLVMBackend) defineFunction(fn *parser.FuncaoDeclaracao) {
	funcDef, ok := l.userFuncs[fn.Nome]
	if !ok {
		l.declareFunction(fn)
		funcDef = l.userFuncs[fn.Nome]
	}

	// Salva contexto atual
	prevFunc := l.function
	prevBlock := l.block

	// Define corpo da função
	l.function = funcDef
	entry := funcDef.NewBlock("entry")
	l.block = entry

	// Novo escopo com parâmetros
	l.pushScope()
	for _, p := range funcDef.Params {
		l.variables[p.Name()] = p
	}

	// Processa corpo
	result := l.processBlock(fn.Corpo)
	if l.block.Term == nil {
		l.block.NewRet(result)
	}

	l.popScope()

	// Restaura contexto
	l.function = prevFunc
	l.block = prevBlock
}

func (l *LLVMBackend) callUserFunction(fn *ir.Func, args []parser.Expressao) value.Value {
	evaluatedArgs := make([]value.Value, len(args))
	for i, arg := range args {
		evaluatedArgs[i] = l.processExpression(arg)
	}
	return l.block.NewCall(fn, evaluatedArgs...)
}

// === Funções Builtin ===

func (l *LLVMBackend) handleBuiltinFunction(fn *parser.ChamadaFuncao, sig registry.AssinaturaFuncao) value.Value {
	switch sig.TipoFuncao {
	case registry.FUNCAO_IMPRIME:
		return l.generatePrintCall(fn.Argumentos)
	case registry.FUNCAO_PURA:
		return l.evaluatePureFunction(fn)
	}
	return l.constInt(0)
}

func (l *LLVMBackend) generatePrintCall(args []parser.Expressao) value.Value {
	for _, arg := range args {
		valor := l.processExpression(arg)
		l.printValue(valor)
	}
	return l.constInt(0)
}

func (l *LLVMBackend) evaluatePureFunction(fn *parser.ChamadaFuncao) value.Value {
	// Extrai valores constantes
	args := make([]interface{}, len(fn.Argumentos))
	for i, arg := range fn.Argumentos {
		valor := l.processExpression(arg)
		if constVal, ok := valor.(*constant.Int); ok {
			args[i] = int(constVal.X.Int64())
		} else {
			fmt.Printf("Aviso: função builtin '%s' com argumentos não constantes\n", fn.Nome)
			return l.constInt(0)
		}
	}

	// Executa função
	result, err := registry.RegistroGlobal.ExecutarFuncao(fn.Nome, args)
	if err != nil {
		fmt.Printf("Erro ao executar função '%s': %v\n", fn.Nome, err)
		return l.constInt(0)
	}

	return l.constInt(int64(result.(int)))
}

func (l *LLVMBackend) printValue(valor value.Value) {
	formatStr, printVal := l.getPrintFormat(valor)

	// Reusa ou cria global de formato
	formatGlobal, ok := l.fmtGlobals[formatStr]
	if !ok {
		l.tmpCount++
		formatGlobal = l.module.NewGlobalDef(
			fmt.Sprintf("fmt%d", l.tmpCount),
			constant.NewCharArrayFromString(formatStr))
		formatGlobal.Immutable = true
		l.fmtGlobals[formatStr] = formatGlobal
	}

	formatPtr := l.block.NewGetElementPtr(
		types.NewArray(uint64(len(formatStr)), types.I8),
		formatGlobal, l.constInt(0), l.constInt(0))

	l.block.NewCall(l.printfFn, formatPtr, printVal)
}

func (l *LLVMBackend) getPrintFormat(valor value.Value) (string, value.Value) {
	switch valorType := valor.Type(); {
	case valorType == types.Double:
		return "%g\n", valor
	case valorType.Equal(types.NewPointer(types.I8)):
		return "%s\n", valor
	case valorType == types.I64:
		return "%ld\n", valor
	default:
		return "%ld\n", l.block.NewSExt(valor, types.I64)
	}
}

// === Processamento de Blocos ===

func (l *LLVMBackend) processBlock(bloco *parser.Bloco) value.Value {
	l.pushScope()
	defer l.popScope()

	var lastValue value.Value = l.constInt(0)

	for _, cmd := range bloco.Comandos {
		val := l.processExpression(cmd)

		// Retorno antecipado
		if l.block.Term != nil {
			return val
		}

		if val != nil {
			lastValue = val
		}
	}

	return lastValue
}

// === Operações Matemáticas ===

func (l *LLVMBackend) generatePowerOperation(base, exp value.Value) value.Value {
	// Implementação iterativa de potência
	chkBlock := l.function.NewBlock("pow_chk")
	loopBlock := l.function.NewBlock("pow_loop")
	endBlock := l.function.NewBlock("pow_end")

	// Allocas para estado
	resAlloca := l.block.NewAlloca(types.I64)
	expAlloca := l.block.NewAlloca(types.I64)
	baseAlloca := l.block.NewAlloca(types.I64)

	l.block.NewStore(l.constInt(1), resAlloca)
	l.block.NewStore(exp, expAlloca)
	l.block.NewStore(base, baseAlloca)
	l.block.NewBr(chkBlock)

	// Verificação
	l.block = chkBlock
	curExp := l.block.NewLoad(types.I64, expAlloca)
	cond := l.block.NewICmp(enum.IPredSGT, curExp, l.constInt(0))
	l.block.NewCondBr(cond, loopBlock, endBlock)

	// Loop principal
	l.block = loopBlock
	curExp2 := l.block.NewLoad(types.I64, expAlloca)
	isOdd := l.block.NewICmp(enum.IPredNE,
		l.block.NewAnd(curExp2, l.constInt(1)), l.constInt(0))

	mulBlock := l.function.NewBlock("pow_mul")
	contBlock := l.function.NewBlock("pow_cont")
	l.block.NewCondBr(isOdd, mulBlock, contBlock)

	// Multiplicação se expoente ímpar
	l.block = mulBlock
	curRes := l.block.NewLoad(types.I64, resAlloca)
	curBase := l.block.NewLoad(types.I64, baseAlloca)
	l.block.NewStore(l.block.NewMul(curRes, curBase), resAlloca)
	l.block.NewBr(contBlock)

	// Continuação: eleva base ao quadrado e divide expoente por 2
	l.block = contBlock
	baseVal := l.block.NewLoad(types.I64, baseAlloca)
	l.block.NewStore(l.block.NewMul(baseVal, baseVal), baseAlloca)
	curExp3 := l.block.NewLoad(types.I64, expAlloca)
	l.block.NewStore(l.block.NewAShr(curExp3, l.constInt(1)), expAlloca)
	l.block.NewBr(chkBlock)

	// Fim
	l.block = endBlock
	return l.block.NewLoad(types.I64, resAlloca)
}

func (l *LLVMBackend) safeDivision(a, b value.Value) value.Value {
	// Proteção contra divisão por zero
	zero := l.constInt(0)
	cond := l.block.NewICmp(enum.IPredEQ, b, zero)

	divZeroBlock := l.function.NewBlock("div_zero")
	divOkBlock := l.function.NewBlock("div_ok")
	mergeBlock := l.function.NewBlock("div_merge")

	l.block.NewCondBr(cond, divZeroBlock, divOkBlock)

	// Divisão por zero retorna 0
	divZeroBlock.NewBr(mergeBlock)

	// Divisão normal
	l.block = divOkBlock
	divResult := l.block.NewSDiv(a, b)
	l.block.NewBr(mergeBlock)

	// Merge com PHI
	l.block = mergeBlock
	return mergeBlock.NewPhi(
		ir.NewIncoming(zero, divZeroBlock),
		ir.NewIncoming(divResult, divOkBlock))
}

// === Gerenciamento de Escopos ===

func (l *LLVMBackend) pushScope() {
	l.varStack = append(l.varStack, l.variables)
	l.variables = make(map[string]value.Value)
}

func (l *LLVMBackend) popScope() {
	if len(l.varStack) == 0 {
		return
	}
	l.variables = l.varStack[len(l.varStack)-1]
	l.varStack = l.varStack[:len(l.varStack)-1]
}

func (l *LLVMBackend) setVariable(name string, val value.Value) {
	l.variables[name] = val
}

func (l *LLVMBackend) getVariable(name string) (value.Value, bool) {
	// Busca no escopo atual
	if v, ok := l.variables[name]; ok {
		return v, true
	}

	// Busca em escopos anteriores
	for i := len(l.varStack) - 1; i >= 0; i-- {
		if v, ok := l.varStack[i][name]; ok {
			return v, true
		}
	}

	return nil, false
}

// === Helpers ===

func (l *LLVMBackend) constInt(v int64) *constant.Int {
	return constant.NewInt(types.I64, v)
}

func (l *LLVMBackend) createStringConstant(str string) value.Value {
	globalStr := l.module.NewGlobalDef(
		l.nextStringName(),
		constant.NewCharArrayFromString(str+"\x00"))
	globalStr.Immutable = true

	return l.block.NewGetElementPtr(
		types.NewArray(uint64(len(str)+1), types.I8),
		globalStr, l.constInt(0), l.constInt(0))
}

func (l *LLVMBackend) nextStringName() string {
	name := fmt.Sprintf("str_%d", l.strCount)
	l.strCount++
	return name
}

// === Compilação para Executável ===

func (l *LLVMBackend) compileToExecutable(llvmFile string) error {
	fmt.Printf("Tentando compilar LLVM IR para executável...\n")

	if err := os.MkdirAll("resultado", 0755); err != nil {
		return fmt.Errorf("erro ao criar diretório resultado: %v", err)
	}

	executable := "resultado/programa"
	cmd := exec.Command("clang", "-O2", "-o", executable, llvmFile)

	if err := cmd.Run(); err != nil {
		return l.tryLLI(llvmFile)
	}

	fmt.Printf("Executável gerado em: %s\n", executable)
	fmt.Printf("Para executar: ./%s\n", executable)
	return nil
}

func (l *LLVMBackend) tryLLI(llvmFile string) error {
	fmt.Printf("clang não disponível, tentando lli...\n")

	cmd := exec.Command("lli", llvmFile)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("nem clang nem lli estão disponíveis")
	}

	fmt.Printf("Executado via lli (interpretador LLVM)\n")
	return nil
}
