package llvm

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/enum"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"

	"github.com/khevencolino/Solar/internal/debug"
	"github.com/khevencolino/Solar/internal/parser"
)

type LLVMBackend struct {
	module    *ir.Module
	block     *ir.Block
	function  *ir.Func
	variables map[string]value.Value
	tmpCount  int
}

func NewLLVMBackend() *LLVMBackend {
	return &LLVMBackend{
		variables: make(map[string]value.Value),
		tmpCount:  0,
	}
}

func (l *LLVMBackend) GetName() string      { return "LLVM IR" }
func (l *LLVMBackend) GetExtension() string { return ".ll" }

func (l *LLVMBackend) Compile(statements []parser.Expressao) error {
	debug.Printf("🔧 Compilando para LLVM IR...\n")

	// Inicializa módulo LLVM
	l.module = ir.NewModule()

	// Declara função printf para impressão
	printf := l.module.NewFunc("printf", types.I32, ir.NewParam("format", types.NewPointer(types.I8)))
	printf.Sig.Variadic = true

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
	arquivoSaida := filepath.Join("result", "programa.ll")
	if err := os.MkdirAll(filepath.Dir(arquivoSaida), 0755); err != nil {
		return fmt.Errorf("erro ao criar diretório: %v", err)
	}

	file, err := os.Create(arquivoSaida)
	if err != nil {
		return fmt.Errorf("erro ao criar arquivo: %v", err)
	}
	defer file.Close()

	if _, err := file.WriteString(l.module.String()); err != nil {
		return fmt.Errorf("erro ao escrever arquivo: %v", err)
	}

	debug.Printf("✅ Arquivo LLVM IR gerado em: %s\n", arquivoSaida)
	return nil
}

func (l *LLVMBackend) processarExpressao(expr parser.Expressao) value.Value {
	switch e := expr.(type) {
	case *parser.Constante:
		return constant.NewInt(types.I64, int64(e.Valor))

	case *parser.Variavel:
		if val, exists := l.variables[e.Nome]; exists {
			return val
		}
		fmt.Printf("⚠️  Variável '%s' não definida\n", e.Nome)
		return constant.NewInt(types.I64, 0)

	case *parser.OperacaoBinaria:
		return l.processarOperacao(e)

	case *parser.Atribuicao:
		valor := l.processarExpressao(e.Valor)
		l.variables[e.Nome] = valor
		return valor

	case *parser.ChamadaFuncao:
		return l.processarFuncao(e)

	default:
		fmt.Printf("⚠️  Tipo de expressão não suportado: %T\n", expr)
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
		// Para potenciação, usamos uma chamada para pow (que seria declarada externamente)
		pow := l.module.NewFunc("pow", types.Double,
			ir.NewParam("base", types.Double),
			ir.NewParam("exp", types.Double))

		// Converte para double
		esquerdaDouble := l.block.NewSIToFP(esquerda, types.Double)
		direitaDouble := l.block.NewSIToFP(direita, types.Double)

		powResult := l.block.NewCall(pow, esquerdaDouble, direitaDouble)
		// Converte de volta para inteiro
		return l.block.NewFPToSI(powResult, types.I64)

	default:
		fmt.Printf("⚠️  Operador não suportado: %s\n", op.Operador.String())
		return constant.NewInt(types.I64, 0)
	}
}

func (l *LLVMBackend) processarFuncao(fn *parser.ChamadaFuncao) value.Value {
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
		fmt.Printf("⚠️  Função '%s' não implementada\n", fn.Nome)
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
