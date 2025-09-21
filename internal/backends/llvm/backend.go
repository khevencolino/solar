package llvm

import (
	"fmt"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/enum"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"

	"github.com/khevencolino/Solar/internal/debug"
	"github.com/khevencolino/Solar/internal/parser"
	"github.com/khevencolino/Solar/internal/utils"
)

type LLVMBackend struct {
	module    *ir.Module
	block     *ir.Block
	function  *ir.Func
	variables map[string]value.Value
	varStack  []map[string]value.Value
	userFuncs map[string]*ir.Func
	tmpCount  int
}

func NewLLVMBackend() *LLVMBackend {
	return &LLVMBackend{
		variables: make(map[string]value.Value),
		varStack:  nil,
		userFuncs: make(map[string]*ir.Func),
		tmpCount:  0,
	}
}

func (l *LLVMBackend) GetName() string      { return "LLVM IR" }
func (l *LLVMBackend) GetExtension() string { return ".ll" }

func (l *LLVMBackend) Compile(statements []parser.Expressao) error {
	debug.Printf("Compilando para LLVM IR...\n")

	// Inicializa módulo LLVM
	l.module = ir.NewModule()

	// Declara função printf para impressão
	printf := l.module.NewFunc("printf", types.I32, ir.NewParam("format", types.NewPointer(types.I8)))
	printf.Sig.Variadic = true

	// Primeira passada: declarar protótipos de funções do usuário
	for _, st := range statements {
		if fn, ok := st.(*parser.FuncaoDeclaracao); ok {
			l.declararFuncaoUsuario(fn)
		}
	}

	// Declara função main
	l.function = l.module.NewFunc("main", types.I32)
	l.block = l.function.NewBlock("")

	// Processa statements
	var resultado value.Value
	for i, stmt := range statements {
		debug.Printf("  Processando statement %d...\n", i+1)
		val := l.processarExpressao(stmt)
		if val != nil {
			resultado = val
		}
	}

	// Se há resultado, imprime
	if resultado != nil {
		l.imprimirValor(resultado)
	}

	// Retorna 0
	l.block.NewRet(constant.NewInt(types.I32, 0))

	// Escreve arquivo LLVM IR
	arquivoSaida := "programa.ll"
	if err := utils.EscreverArquivo(arquivoSaida, l.module.String()); err != nil {
		return err
	}

	debug.Printf("Arquivo LLVM IR gerado em: %s\n", arquivoSaida)
	return nil
}

func (l *LLVMBackend) processarExpressao(expr parser.Expressao) value.Value {
	switch e := expr.(type) {
	case *parser.Constante:
		return constant.NewInt(types.I64, int64(e.Valor))
	case *parser.Booleano:
		if e.Valor {
			return constant.NewInt(types.I64, 1)
		}
		return constant.NewInt(types.I64, 0)

	case *parser.Variavel:
		if val, ok := l.getVar(e.Nome); ok {
			return val
		}
		fmt.Printf("Variável '%s' não definida\n", e.Nome)
		return constant.NewInt(types.I64, 0)

	case *parser.OperacaoBinaria:
		return l.processarOperacao(e)

	case *parser.Atribuicao:
		valor := l.processarExpressao(e.Valor)
		l.setVar(e.Nome, valor)
		return valor

	case *parser.ChamadaFuncao:
		return l.processarFuncao(e)

	case *parser.ComandoSe:
		return l.processarComandoSe(e)
	case *parser.ComandoEnquanto:
		return l.processarEnquanto(e)
	case *parser.ComandoPara:
		return l.processarPara(e)

	case *parser.Bloco:
		return l.processarBloco(e)

	case *parser.FuncaoDeclaracao:
		l.definirFuncaoUsuario(e)
		return constant.NewInt(types.I64, 0)

	case *parser.Retorno:
		if e.Valor != nil {
			v := l.processarExpressao(e.Valor)
			if l.function != nil {
				l.block.NewRet(v)
			}
			return v
		}
		if l.function != nil {
			l.block.NewRet(constant.NewInt(types.I64, 0))
		}
		return constant.NewInt(types.I64, 0)

	default:
		fmt.Printf("Tipo de expressão não suportado: %T\n", expr)
		return constant.NewInt(types.I64, 0)
	}
}

func (l *LLVMBackend) processarOperacao(op *parser.OperacaoBinaria) value.Value {
	esquerda := l.processarExpressao(op.OperandoEsquerdo)
	direita := l.processarExpressao(op.OperandoDireito)

	// Verifica se os operandos são válidos
	if esquerda == nil || direita == nil {
		return constant.NewInt(types.I64, 0)
	}

	l.tmpCount++

	switch op.Operador {
	case parser.ADICAO:
		return l.block.NewAdd(esquerda, direita)

	case parser.SUBTRACAO:
		return l.block.NewSub(esquerda, direita)

	case parser.MULTIPLICACAO:
		return l.block.NewMul(esquerda, direita)

	case parser.DIVISAO:
		return l.block.NewSDiv(esquerda, direita)

	case parser.POWER:
		pow := l.module.NewFunc("pow", types.Double,
			ir.NewParam("base", types.Double),
			ir.NewParam("exp", types.Double))

		esquerdaDouble := l.block.NewSIToFP(esquerda, types.Double)
		direitaDouble := l.block.NewSIToFP(direita, types.Double)

		powResult := l.block.NewCall(pow, esquerdaDouble, direitaDouble)

		return l.block.NewFPToSI(powResult, types.I64)

	// Operações de comparação
	case parser.IGUALDADE:
		cmp := l.block.NewICmp(enum.IPredEQ, esquerda, direita)
		return l.block.NewZExt(cmp, types.I64)

	case parser.DIFERENCA:
		cmp := l.block.NewICmp(enum.IPredNE, esquerda, direita)
		return l.block.NewZExt(cmp, types.I64)

	case parser.MENOR_QUE:
		cmp := l.block.NewICmp(enum.IPredSLT, esquerda, direita)
		return l.block.NewZExt(cmp, types.I64)

	case parser.MAIOR_QUE:
		cmp := l.block.NewICmp(enum.IPredSGT, esquerda, direita)
		return l.block.NewZExt(cmp, types.I64)

	case parser.MENOR_IGUAL:
		cmp := l.block.NewICmp(enum.IPredSLE, esquerda, direita)
		return l.block.NewZExt(cmp, types.I64)

	case parser.MAIOR_IGUAL:
		cmp := l.block.NewICmp(enum.IPredSGE, esquerda, direita)
		return l.block.NewZExt(cmp, types.I64)

	default:
		fmt.Printf("Operador não suportado: %s\n", op.Operador.String())
		return constant.NewInt(types.I64, 0)
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
	switch fn.Nome {
	case "imprime":
		if len(fn.Argumentos) > 0 {
			for _, arg := range fn.Argumentos {
				valor := l.processarExpressao(arg)
				l.imprimirValor(valor)
			}
		}
		return constant.NewInt(types.I64, 0)

	case "soma":
		if len(fn.Argumentos) >= 2 {
			resultado := l.processarExpressao(fn.Argumentos[0])
			for i := 1; i < len(fn.Argumentos); i++ {
				arg := l.processarExpressao(fn.Argumentos[i])
				resultado = l.block.NewAdd(resultado, arg)
			}
			return resultado
		}
		return constant.NewInt(types.I64, 0)

	case "abs":
		if len(fn.Argumentos) == 1 {
			valor := l.processarExpressao(fn.Argumentos[0])
			zero := constant.NewInt(types.I64, 0)

			// Verifica se valor < 0
			cond := l.block.NewICmp(enum.IPredSLT, valor, zero)

			// Cria blocos para if/else
			thenBlock := l.function.NewBlock("")
			elseBlock := l.function.NewBlock("")
			mergeBlock := l.function.NewBlock("")

			l.block.NewCondBr(cond, thenBlock, elseBlock)

			// Bloco then: retorna -valor
			thenBlock.NewBr(mergeBlock)
			negValue := thenBlock.NewSub(zero, valor)

			// Bloco else: retorna valor
			elseBlock.NewBr(mergeBlock)

			// Bloco merge: phi node para resultado
			l.block = mergeBlock
			phi := mergeBlock.NewPhi(ir.NewIncoming(negValue, thenBlock), ir.NewIncoming(valor, elseBlock))

			return phi
		}
		return constant.NewInt(types.I64, 0)

	default:
		fmt.Printf("Função '%s' não implementada\n", fn.Nome)
		return constant.NewInt(types.I64, 0)
	}
}

func (l *LLVMBackend) imprimirValor(valor value.Value) {
	// String format para printf
	formatStr := "%ld\n"

	// Converte para i64 se necessário
	printValue := valor
	if valor.Type() != types.I64 {
		printValue = l.block.NewSExt(valor, types.I64)
	}

	// Cria string global para formato
	l.tmpCount++
	formatGlobal := l.module.NewGlobalDef(fmt.Sprintf("fmt%d", l.tmpCount),
		constant.NewCharArrayFromString(formatStr))

	// Obtém ponteiro para string
	formatPtr := l.block.NewGetElementPtr(types.NewArray(uint64(len(formatStr)), types.I8),
		formatGlobal, constant.NewInt(types.I64, 0), constant.NewInt(types.I64, 0))

	// Chama printf
	printf := l.module.Funcs[0] // printf é a primeira função declarada
	l.block.NewCall(printf, formatPtr, printValue)
}

// processarComandoSe processa comandos if/else
func (l *LLVMBackend) processarComandoSe(comando *parser.ComandoSe) value.Value {
	// Avalia a condição
	condicao := l.processarExpressao(comando.Condicao)

	// Converte para i1 (boolean)
	zero := constant.NewInt(types.I64, 0)
	cond := l.block.NewICmp(enum.IPredNE, condicao, zero)

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
		elseValue = constant.NewInt(types.I64, 0)
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
	var ultimoValor value.Value = constant.NewInt(types.I64, 0)

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
	zero := constant.NewInt(types.I64, 0)
	condI1 := l.block.NewICmp(enum.IPredNE, condVal, zero)
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
		zero := constant.NewInt(types.I64, 0)
		condI1 = l.block.NewICmp(enum.IPredNE, condVal, zero)
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
		params[i] = ir.NewParam(p, types.I64)
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
