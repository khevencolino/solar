package bytecode

import (
	"fmt"

	"github.com/khevencolino/Kite/internal/parser"
)

type BytecodeBackend struct {
	instructions []Instruction
	variables    map[string]int // nome -> índice
	varCount     int
}

func NewBytecodeBackend() *BytecodeBackend {
	return &BytecodeBackend{
		instructions: make([]Instruction, 0),
		variables:    make(map[string]int),
	}
}

func (b *BytecodeBackend) GetName() string      { return "Bytecode + VM" }
func (b *BytecodeBackend) GetExtension() string { return ".bc" }

func (b *BytecodeBackend) Compile(statements []parser.Expressao) error {
	fmt.Printf("🤖 Compilando para Bytecode...\n")

	// Gera bytecode
	for i, stmt := range statements {
		fmt.Printf("  Processando statement %d...\n", i+1)
		b.visitarExpressao(stmt)

		// Se for a última expressão, imprime resultado
		if i == len(statements)-1 {
			b.emit(OP_PRINT, 0, 0)
		}
	}

	b.emit(OP_HALT, 0, 0)

	// Executa na VM
	return b.executarVM()
}

func (b *BytecodeBackend) visitarExpressao(expr parser.Expressao) {
	switch e := expr.(type) {
	case *parser.Constante:
		b.emit(OP_CONST, int64(e.Valor), e.Token.Position.Line)

	case *parser.Variavel:
		varIndex := b.getVariableIndex(e.Nome)
		b.emit(OP_LOAD, int64(varIndex), e.Token.Position.Line)

	case *parser.Atribuicao:
		b.visitarExpressao(e.Valor)
		varIndex := b.declareVariable(e.Nome)
		b.emit(OP_STORE, int64(varIndex), e.Token.Position.Line)

	case *parser.OperacaoBinaria:
		b.visitarExpressao(e.OperandoEsquerdo)
		b.visitarExpressao(e.OperandoDireito)

		switch e.Operador {
		case parser.ADICAO:
			b.emit(OP_ADD, 0, e.Token.Position.Line)
		case parser.SUBTRACAO:
			b.emit(OP_SUB, 0, e.Token.Position.Line)
		case parser.MULTIPLICACAO:
			b.emit(OP_MUL, 0, e.Token.Position.Line)
		case parser.DIVISAO:
			b.emit(OP_DIV, 0, e.Token.Position.Line)
		case parser.POWER:
			b.emit(OP_POW, 0, e.Token.Position.Line)
		}
	}
}

func (b *BytecodeBackend) emit(op OpCode, operand int64, line int) {
	b.instructions = append(b.instructions, Instruction{
		OpCode:  op,
		Operand: operand,
		Line:    line,
	})
}

func (b *BytecodeBackend) declareVariable(nome string) int {
	if index, exists := b.variables[nome]; exists {
		return index
	}

	index := b.varCount
	b.variables[nome] = index
	b.varCount++
	return index
}

func (b *BytecodeBackend) getVariableIndex(nome string) int {
	if index, exists := b.variables[nome]; exists {
		return index
	}
	panic(fmt.Sprintf("Variável '%s' não definida", nome))
}

func (b *BytecodeBackend) executarVM() error {
	fmt.Printf("🚀 Executando na Virtual Machine...\n")

	vm := NewVM(b.varCount)
	return vm.Execute(b.instructions)
}
