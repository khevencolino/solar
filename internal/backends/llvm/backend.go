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

type LLVMBackend struct {
	module     *ir.Module
	block      *ir.Block
	function   *ir.Func
	variables  map[string]value.Value
	varStack   []map[string]value.Value
	userFuncs  map[string]*ir.Func
	tmpCount   int
	strCount   int
	printfFn   *ir.Func              // cache para printf
	fmtGlobals map[string]*ir.Global // cache para strings de formato
}

func NewLLVMBackend() *LLVMBackend {
	return &LLVMBackend{
		variables:  make(map[string]value.Value),
		varStack:   nil,
		userFuncs:  make(map[string]*ir.Func),
		tmpCount:   0,
		strCount:   0,
		fmtGlobals: make(map[string]*ir.Global),
	}
}

func (l *LLVMBackend) GetName() string      { return "LLVM IR" }
func (l *LLVMBackend) GetExtension() string { return ".ll" }

func (l *LLVMBackend) Compile(statements []parser.Expressao) error {
	debug.Printf("Compilando para LLVM IR...\n")

	// Inicializa módulo LLVM
	l.module = ir.NewModule()

	// Declara função printf para impressão e guarda referência
	l.printfFn = l.module.NewFunc("printf", types.I32, ir.NewParam("format", types.NewPointer(types.I8)))
	l.printfFn.Sig.Variadic = true

	// Primeira passada: declarar protótipos de funções do usuário
	var funcaoPrincipal *parser.FuncaoDeclaracao
	for _, st := range statements {
		if fn, ok := st.(*parser.FuncaoDeclaracao); ok {
			l.declararFuncaoUsuario(fn)
			if fn.Nome == "principal" {
				funcaoPrincipal = fn
			}
		}
	}

	// Segunda passada: definir corpos das funções do usuário
	for _, st := range statements {
		if fn, ok := st.(*parser.FuncaoDeclaracao); ok {
			l.definirFuncaoUsuario(fn)
		}
	}

	// Declara função main
	l.function = l.module.NewFunc("main", types.I32)
	l.block = l.function.NewBlock("")

	// Se existe função principal(), chama ela. Senão, executa statements globais
	if funcaoPrincipal != nil {
		debug.Printf("  Chamando função principal()...\n")
		// Chama a função principal()
		principalFunc := l.userFuncs[funcaoPrincipal.Nome]
		l.block.NewCall(principalFunc)
	} else {
		// Processa statements globais (comportamento antigo)
		for i, stmt := range statements {
			// Pula declarações de função pois já foram processadas
			if _, ok := stmt.(*parser.FuncaoDeclaracao); !ok {
				debug.Printf("  Processando statement global %d...\n", i+1)
				l.processarExpressao(stmt)
			}
		}
	}

	// Retorna 0
	l.block.NewRet(constant.NewInt(types.I32, 0))

	// Escreve arquivo LLVM IR
	arquivoSaida := "programa.ll"
	if err := utils.EscreverArquivo(arquivoSaida, l.module.String()); err != nil {
		return err
	}

	debug.Printf("Arquivo LLVM IR gerado em: %s\n", arquivoSaida)

	// Tenta compilar para executável se disponível
	if err := l.compilarParaExecutavel(arquivoSaida); err != nil {
		debug.Printf("Aviso: Não foi possível compilar para executável: %v\n", err)
		debug.Printf("Use 'clang programa.ll -o programa' para compilar manualmente\n")
	}

	return nil
}

func (l *LLVMBackend) processarExpressao(expr parser.Expressao) value.Value {
	result := expr.Aceitar(l)
	if val, ok := result.(value.Value); ok {
		return val
	}
	return l.i64(0)
}

// Helper para processar expressões que retornam value.Value
func (l *LLVMBackend) processarExpressaoValue(expr parser.Expressao) value.Value {
	return l.processarExpressao(expr)
}

// Implementação da interface visitor
func (l *LLVMBackend) Constante(constante *parser.Constante) interface{} {
	// Suporte completo a números inteiros, incluindo negativos
	return constant.NewInt(types.I64, int64(constante.Valor))
}

func (l *LLVMBackend) Booleano(b *parser.Booleano) interface{} {
	if b.Valor {
		return l.i64(1)
	}
	return l.i64(0)
}

func (l *LLVMBackend) LiteralTexto(literal *parser.LiteralTexto) interface{} {
	// Implementa suporte completo a strings criando uma string global
	strValue := literal.Valor

	// Cria uma variável global para a string com terminador nulo
	globalStr := l.module.NewGlobalDef(l.getNextStringName(), constant.NewCharArrayFromString(strValue+"\x00"))
	globalStr.Immutable = true

	// Retorna um ponteiro para o primeiro caractere da string
	return l.block.NewGetElementPtr(types.NewArray(uint64(len(strValue)+1), types.I8), globalStr,
		constant.NewInt(types.I32, 0), constant.NewInt(types.I32, 0))
}

func (l *LLVMBackend) LiteralDecimal(literal *parser.LiteralDecimal) interface{} {
	// Implementa suporte completo a ponto flutuante usando double (64-bit)
	return constant.NewFloat(types.Double, literal.Valor)
}

func (l *LLVMBackend) Variavel(variavel *parser.Variavel) interface{} {
	if val, ok := l.getVar(variavel.Nome); ok {
		// Se é um ponteiro (alloca), carrega o valor
		if alloca, isAlloca := val.(*ir.InstAlloca); isAlloca {
			return l.block.NewLoad(types.I64, alloca)
		}
		return val
	}
	fmt.Printf("Variável '%s' não definida\n", variavel.Nome)
	return l.i64(0)
}

func (l *LLVMBackend) Atribuicao(atribuicao *parser.Atribuicao) interface{} {
	valor := l.processarExpressaoValue(atribuicao.Valor)

	// Verifica se a variável já existe
	if existente, ok := l.getVar(atribuicao.Nome); ok {
		// Se é um alloca existente, armazena nele
		if alloca, isAlloca := existente.(*ir.InstAlloca); isAlloca {
			l.block.NewStore(valor, alloca)
			return valor
		}
	}

	// Cria nova variável usando alloca
	alloca := l.block.NewAlloca(types.I64)
	l.block.NewStore(valor, alloca)
	l.setVar(atribuicao.Nome, alloca)
	return valor
}

func (l *LLVMBackend) OperacaoBinaria(operacao *parser.OperacaoBinaria) interface{} {
	esquerda := l.processarExpressaoValue(operacao.OperandoEsquerdo)
	direita := l.processarExpressaoValue(operacao.OperandoDireito)

	// Verifica se os operandos são válidos
	if esquerda == nil || direita == nil {
		return l.i64(0)
	}

	switch operacao.Operador {
	case parser.ADICAO:
		return l.block.NewAdd(esquerda, direita)

	case parser.SUBTRACAO:
		return l.block.NewSub(esquerda, direita)

	case parser.MULTIPLICACAO:
		return l.block.NewMul(esquerda, direita)

	case parser.DIVISAO:
		return l.divisaoSegura(esquerda, direita)

	case parser.POWER:
		// Implementação simples de potência usando loop
		return l.implementarPotencia(esquerda, direita)

	// Operações de comparação
	case parser.IGUALDADE, parser.DIFERENCA, parser.MENOR_QUE, parser.MAIOR_QUE, parser.MENOR_IGUAL, parser.MAIOR_IGUAL:
		pred := map[parser.TipoOperador]enum.IPred{
			parser.IGUALDADE:   enum.IPredEQ,
			parser.DIFERENCA:   enum.IPredNE,
			parser.MENOR_QUE:   enum.IPredSLT,
			parser.MAIOR_QUE:   enum.IPredSGT,
			parser.MENOR_IGUAL: enum.IPredSLE,
			parser.MAIOR_IGUAL: enum.IPredSGE,
		}[operacao.Operador]
		cmp := l.block.NewICmp(pred, esquerda, direita)
		return l.block.NewZExt(cmp, types.I64)

	default:
		fmt.Printf("Operador não suportado: %s\n", operacao.Operador.String())
		return l.i64(0)
	}
}

func (l *LLVMBackend) processarFuncao(fn *parser.ChamadaFuncao) value.Value {
	// Chamada de função de usuário
	if uf, ok := l.userFuncs[fn.Nome]; ok {
		// Avalia argumentos
		var args []value.Value
		for _, a := range fn.Argumentos {
			args = append(args, l.processarExpressao(a))
		}
		call := l.block.NewCall(uf, args...)
		return call
	}
	// Verifica se é função builtin no registry
	if assinatura, ok := registry.RegistroGlobal.ObterAssinatura(fn.Nome); ok {
		switch assinatura.TipoFuncao {
		case registry.FUNCAO_IMPRIME:
			// Implementa função imprime diretamente
			for _, arg := range fn.Argumentos {
				valor := l.processarExpressao(arg)
				l.imprimirValor(valor)
			}
			return l.i64(0)

		case registry.FUNCAO_PURA:
			// Para funções puras, tenta extrair valores constantes
			var args []interface{}
			for _, arg := range fn.Argumentos {
				valor := l.processarExpressao(arg)
				// Converte value.Value para interface{} extraindo o valor
				if constVal, ok := valor.(*constant.Int); ok {
					args = append(args, int(constVal.X.Int64()))
				} else {
					// Para valores não constantes, não podemos processar em tempo de compilação
					fmt.Printf("Aviso: função builtin '%s' com argumentos não constantes\n", fn.Nome)
					return l.i64(0)
				}
			}

			// Executa função pura e retorna resultado como constante
			resultado, err := registry.RegistroGlobal.ExecutarFuncao(fn.Nome, args)
			if err != nil {
				fmt.Printf("Erro ao executar função '%s': %v\n", fn.Nome, err)
				return constant.NewInt(types.I64, 0)
			}
			return l.i64(int64(resultado.(int)))
		}
	}

	fmt.Printf("Função '%s' não implementada\n", fn.Nome)
	return l.i64(0)
}

func (l *LLVMBackend) imprimirValor(valor value.Value) {
	var formatStr string
	var printValue value.Value

	// Determina o formato baseado no tipo do valor
	valorType := valor.Type()
	switch {
	case valorType == types.Double:
		// Números decimais (double)
		formatStr = "%g\n"
		printValue = valor
	case valorType.Equal(types.NewPointer(types.I8)):
		// Strings (ponteiro para char)
		formatStr = "%s\n"
		printValue = valor
	case valorType == types.I64:
		// Inteiros (incluindo booleanos convertidos)
		formatStr = "%ld\n"
		printValue = valor
	default:
		// Conversão padrão para inteiro
		formatStr = "%ld\n"
		printValue = l.block.NewSExt(valor, types.I64)
	}

	// Reuso de globals de formato para evitar duplicações
	formatGlobal, ok := l.fmtGlobals[formatStr]
	if !ok {
		l.tmpCount++
		formatGlobal = l.module.NewGlobalDef(fmt.Sprintf("fmt%d", l.tmpCount), constant.NewCharArrayFromString(formatStr))
		formatGlobal.Immutable = true
		l.fmtGlobals[formatStr] = formatGlobal
	}
	formatPtr := l.block.NewGetElementPtr(types.NewArray(uint64(len(formatStr)), types.I8), formatGlobal, l.i64(0), l.i64(0))
	l.block.NewCall(l.printfFn, formatPtr, printValue)
}

// processarComandoSe processa comandos if/else
func (l *LLVMBackend) processarComandoSe(comando *parser.ComandoSe) value.Value {
	// Avalia a condição
	condicao := l.processarExpressao(comando.Condicao)

	// Converte para i1 (boolean)
	cond := l.block.NewICmp(enum.IPredNE, condicao, l.i64(0))

	// Cria blocos
	thenBlock := l.function.NewBlock("")
	var elseBlock *ir.Block
	mergeBlock := l.function.NewBlock("")

	if comando.BlocoSenao != nil {
		elseBlock = l.function.NewBlock("")
		l.block.NewCondBr(cond, thenBlock, elseBlock)
	} else {
		l.block.NewCondBr(cond, thenBlock, mergeBlock)
	}

	// Processa bloco "se"
	l.block = thenBlock
	thenValue := l.processarBloco(comando.BlocoSe)
	l.block.NewBr(mergeBlock)

	var elseValue value.Value
	if comando.BlocoSenao != nil {
		// Processa bloco "senao"
		l.block = elseBlock
		elseValue = l.processarBloco(comando.BlocoSenao)
		l.block.NewBr(mergeBlock)
	} else {
		elseValue = l.i64(0)
	}

	// Merge block
	l.block = mergeBlock
	if comando.BlocoSenao != nil {
		phi := mergeBlock.NewPhi(ir.NewIncoming(thenValue, thenBlock), ir.NewIncoming(elseValue, elseBlock))
		return phi
	}

	return thenValue
}

// processarBloco processa um bloco de comandos
func (l *LLVMBackend) processarBloco(bloco *parser.Bloco) value.Value {
	// Novo escopo de variáveis
	l.pushScope()
	var ultimoValor value.Value = l.i64(0)

	for _, comando := range bloco.Comandos {
		val := l.processarExpressao(comando)
		// Se encontrou retorno, encerra bloco cedo
		if term := l.block.Term; term != nil {
			l.popScope()
			return val
		}
		if val != nil {
			ultimoValor = val
		}
	}

	l.popScope()
	return ultimoValor
}

func (l *LLVMBackend) processarEnquanto(cmd *parser.ComandoEnquanto) value.Value {
	funcBlock := l.function
	// Cria blocos
	condBlock := funcBlock.NewBlock("while.cond")
	bodyBlock := funcBlock.NewBlock("while.body")
	endBlock := funcBlock.NewBlock("while.end")

	// Branch para condição
	l.block.NewBr(condBlock)
	l.block = condBlock
	condVal := l.processarExpressao(cmd.Condicao)
	condI1 := l.block.NewICmp(enum.IPredNE, condVal, l.i64(0))
	l.block.NewCondBr(condI1, bodyBlock, endBlock)

	// Corpo
	l.block = bodyBlock
	last := l.processarBloco(cmd.Corpo)
	// Se corpo não retornou, volta para cond
	if l.block.Term == nil {
		l.block.NewBr(condBlock)
	}

	l.block = endBlock
	return last
}

func (l *LLVMBackend) processarPara(cmd *parser.ComandoPara) value.Value {
	funcBlock := l.function
	// init
	if cmd.Inicializacao != nil {
		l.processarExpressao(cmd.Inicializacao)
	}
	// Blocos
	condBlock := funcBlock.NewBlock("for.cond")
	bodyBlock := funcBlock.NewBlock("for.body")
	stepBlock := funcBlock.NewBlock("for.step")
	endBlock := funcBlock.NewBlock("for.end")

	l.block.NewBr(condBlock)
	l.block = condBlock
	// condição (vazia => true)
	var condI1 value.Value
	if cmd.Condicao != nil {
		condVal := l.processarExpressao(cmd.Condicao)
		condI1 = l.block.NewICmp(enum.IPredNE, condVal, l.i64(0))
	} else {
		condI1 = constant.NewInt(types.I1, 1)
	}
	l.block.NewCondBr(condI1, bodyBlock, endBlock)

	// body
	l.block = bodyBlock
	last := l.processarBloco(cmd.Corpo)
	if l.block.Term == nil {
		l.block.NewBr(stepBlock)
	}

	// step
	l.block = stepBlock
	if cmd.PosIteracao != nil {
		l.processarExpressao(cmd.PosIteracao)
	}
	if l.block.Term == nil {
		l.block.NewBr(condBlock)
	}

	l.block = endBlock
	return last
}

// Suporte a funções do usuário
func (l *LLVMBackend) declararFuncaoUsuario(fn *parser.FuncaoDeclaracao) {
	// Assinaturas somente de inteiros i64
	// TODO: tipos variados
	params := make([]*ir.Param, len(fn.Parametros))
	for i, p := range fn.Parametros {
		params[i] = ir.NewParam(p.Nome, types.I64)
	}
	f := l.module.NewFunc(fn.Nome, types.I64, params...)
	l.userFuncs[fn.Nome] = f
}

func (l *LLVMBackend) definirFuncaoUsuario(fn *parser.FuncaoDeclaracao) {
	f, ok := l.userFuncs[fn.Nome]
	if !ok {
		l.declararFuncaoUsuario(fn)
		f = l.userFuncs[fn.Nome]
	}
	// Cria bloco de entrada
	prevFunc := l.function
	prevBlock := l.block
	l.function = f
	entry := f.NewBlock("entry")
	l.block = entry

	// Novo escopo e bind de parâmetros
	l.pushScope()
	for _, p := range f.Params {
		l.variables[p.Name()] = p
	}

	// Processa corpo: retorno implícito = última expressão
	result := l.processarBloco(fn.Corpo)
	if l.block.Term == nil {
		l.block.NewRet(result)
	}
	l.popScope()

	// Restaura função/bloco anterior
	l.function = prevFunc
	l.block = prevBlock
}

// Escopos de variáveis
func (l *LLVMBackend) pushScope() {
	// Copia raso para permitir shadowing isolado
	// TODO: otimizar com linked list?
	novo := make(map[string]value.Value)
	l.varStack = append(l.varStack, l.variables)
	l.variables = novo
}

func (l *LLVMBackend) popScope() {
	if len(l.varStack) == 0 {
		return
	}
	topo := l.varStack[len(l.varStack)-1]
	l.varStack = l.varStack[:len(l.varStack)-1]
	l.variables = topo
}

// Variável: set no escopo atual
func (l *LLVMBackend) setVar(name string, val value.Value) {
	l.variables[name] = val
}

// Variável: busca do escopo atual para os anteriores
func (l *LLVMBackend) getVar(name string) (value.Value, bool) {
	if v, ok := l.variables[name]; ok {
		return v, true
	}
	for i := len(l.varStack) - 1; i >= 0; i-- {
		if v, ok := l.varStack[i][name]; ok {
			return v, true
		}
	}
	return nil, false
}

func (l *LLVMBackend) ChamadaFuncao(chamada *parser.ChamadaFuncao) interface{} {
	return l.processarFuncao(chamada)
}

func (l *LLVMBackend) ComandoSe(comando *parser.ComandoSe) interface{} {
	return l.processarComandoSe(comando)
}

func (l *LLVMBackend) ComandoEnquanto(cmd *parser.ComandoEnquanto) interface{} {
	return l.processarEnquanto(cmd)
}

func (l *LLVMBackend) ComandoPara(cmd *parser.ComandoPara) interface{} {
	return l.processarPara(cmd)
}

func (l *LLVMBackend) Bloco(bloco *parser.Bloco) interface{} {
	return l.processarBloco(bloco)
}

func (l *LLVMBackend) FuncaoDeclaracao(fn *parser.FuncaoDeclaracao) interface{} {
	l.definirFuncaoUsuario(fn)
	return l.i64(0)
}

func (l *LLVMBackend) Retorno(ret *parser.Retorno) interface{} {
	if ret.Valor != nil {
		v := l.processarExpressaoValue(ret.Valor)
		if l.function != nil {
			l.block.NewRet(v)
		}
		return v
	}
	if l.function != nil {
		l.block.NewRet(l.i64(0))
	}
	return l.i64(0)
}

// Suporte a importação
func (l *LLVMBackend) Importacao(imp *parser.Importacao) interface{} {
	// Imports são processados antes da geração de código
	return nil
}

// Implementa potência usando loop iterativo
func (l *LLVMBackend) implementarPotencia(base, exp value.Value) value.Value {
	chk := l.function.NewBlock("pow_chk")
	loop := l.function.NewBlock("pow_loop")
	end := l.function.NewBlock("pow_end")
	resAlloca := l.block.NewAlloca(types.I64)
	expAlloca := l.block.NewAlloca(types.I64)
	baseAlloca := l.block.NewAlloca(types.I64)
	l.block.NewStore(l.i64(1), resAlloca)
	l.block.NewStore(exp, expAlloca)
	l.block.NewStore(base, baseAlloca)
	l.block.NewBr(chk)
	// check
	l.block = chk
	curExp := l.block.NewLoad(types.I64, expAlloca)
	cond := l.block.NewICmp(enum.IPredSGT, curExp, l.i64(0))
	l.block.NewCondBr(cond, loop, end)
	// loop
	l.block = loop
	curExp2 := l.block.NewLoad(types.I64, expAlloca)
	one := l.i64(1)
	isOdd := l.block.NewICmp(enum.IPredNE, l.block.NewAnd(curExp2, one), l.i64(0))
	mulBlock := l.function.NewBlock("pow_mul")
	cont := l.function.NewBlock("pow_cont")
	l.block.NewCondBr(isOdd, mulBlock, cont)
	// mul path
	l.block = mulBlock
	curRes := l.block.NewLoad(types.I64, resAlloca)
	curBase := l.block.NewLoad(types.I64, baseAlloca)
	l.block.NewStore(l.block.NewMul(curRes, curBase), resAlloca)
	l.block.NewBr(cont)
	// cont
	l.block = cont
	baseVal := l.block.NewLoad(types.I64, baseAlloca)
	l.block.NewStore(l.block.NewMul(baseVal, baseVal), baseAlloca)
	curExp3 := l.block.NewLoad(types.I64, expAlloca)
	l.block.NewStore(l.block.NewAShr(curExp3, one), expAlloca) // shift right aritmético
	l.block.NewBr(chk)
	// end
	l.block = end
	return l.block.NewLoad(types.I64, resAlloca)
}

// Tenta compilar o LLVM IR para um executável usando clang
func (l *LLVMBackend) compilarParaExecutavel(arquivoLLVM string) error {
	debug.Printf("Tentando compilar LLVM IR para executável...\n")

	// Cria diretório result se não existir
	if err := os.MkdirAll("result", 0755); err != nil {
		return fmt.Errorf("erro ao criar diretório result: %v", err)
	}

	// Tenta usar clang para compilar
	executavel := "result/programa"
	cmd := exec.Command("clang", "-O2", "-o", executavel, arquivoLLVM)

	if err := cmd.Run(); err != nil {
		// Se clang não estiver disponível, tenta lli para interpretação
		debug.Printf("clang não disponível, tentando lli...\n")
		cmdLli := exec.Command("lli", arquivoLLVM)
		if err := cmdLli.Run(); err != nil {
			return fmt.Errorf("nem clang nem lli estão disponíveis")
		}
		debug.Printf("Executado via lli (interpretador LLVM)\n")
		return nil
	}

	debug.Printf("Executável gerado em: %s\n", executavel)
	debug.Printf("Para executar: ./%s\n", executavel)
	return nil
}

// Gera um nome único para strings globais
func (l *LLVMBackend) getNextStringName() string {
	name := fmt.Sprintf("str_%d", l.strCount)
	l.strCount++
	return name
}

// i64 cria constante inteira de 64 bits.
func (l *LLVMBackend) i64(v int64) *constant.Int { return constant.NewInt(types.I64, v) }

// divisaoSegura gera código de divisão com proteção contra divisor zero (retorna 0 se divisor==0).
func (l *LLVMBackend) divisaoSegura(a, b value.Value) value.Value {
	zero := l.i64(0)
	cond := l.block.NewICmp(enum.IPredEQ, b, zero)
	divZero := l.function.NewBlock("div_zero")
	divOk := l.function.NewBlock("div_ok")
	merge := l.function.NewBlock("div_merge")
	l.block.NewCondBr(cond, divZero, divOk)
	// zero path
	divZero.NewBr(merge)
	// ok path
	l.block = divOk
	divRes := l.block.NewSDiv(a, b)
	l.block.NewBr(merge)
	// merge
	l.block = merge
	return merge.NewPhi(ir.NewIncoming(zero, divZero), ir.NewIncoming(divRes, divOk))
}
