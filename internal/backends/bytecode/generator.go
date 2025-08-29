package bytecode

import (
	"fmt"

	"github.com/khevencolino/Solar/internal/debug"
	"github.com/khevencolino/Solar/internal/parser"
	"github.com/khevencolino/Solar/internal/registry"
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
	debug.Printf("🤖 Compilando para Bytecode...\n")

	// Gera bytecode
	for i, stmt := range statements {
		debug.Printf("  Processando statement %d...\n", i+1)
		b.visitarExpressao(stmt)

		// Se for a última expressão e NÃO for uma chamada de função, imprime resultado
		if i == len(statements)-1 {
			if _, ehChamadaFuncao := stmt.(*parser.ChamadaFuncao); !ehChamadaFuncao {
				b.emit(OP_PRINT, 0, 0)
			}
		}
	}

	b.emit(OP_HALT, 0, 0)

	// Executa na VM
	return b.executarVM()
}

func (b *BytecodeBackend) visitarExpressao(expr parser.Expressao) {
	// Usa o padrão visitor para gerar bytecode
	expr.Aceitar(b)
}

// Implementação da interface visitor
func (b *BytecodeBackend) Constante(constante *parser.Constante) interface{} {
	b.emit(OP_CONST, int64(constante.Valor), constante.Token.Position.Line)
	return nil
}

func (b *BytecodeBackend) Variavel(variavel *parser.Variavel) interface{} {
	varIndex := b.getVariableIndex(variavel.Nome)
	b.emit(OP_LOAD, int64(varIndex), variavel.Token.Position.Line)
	return nil
}

func (b *BytecodeBackend) Atribuicao(atribuicao *parser.Atribuicao) interface{} {
	atribuicao.Valor.Aceitar(b)
	varIndex := b.declareVariable(atribuicao.Nome)
	b.emit(OP_STORE, int64(varIndex), atribuicao.Token.Position.Line)
	return nil
}

func (b *BytecodeBackend) OperacaoBinaria(operacao *parser.OperacaoBinaria) interface{} {
	operacao.OperandoEsquerdo.Aceitar(b)
	operacao.OperandoDireito.Aceitar(b)

	switch operacao.Operador {
	case parser.ADICAO:
		b.emit(OP_ADD, 0, operacao.Token.Position.Line)
	case parser.SUBTRACAO:
		b.emit(OP_SUB, 0, operacao.Token.Position.Line)
	case parser.MULTIPLICACAO:
		b.emit(OP_MUL, 0, operacao.Token.Position.Line)
	case parser.DIVISAO:
		b.emit(OP_DIV, 0, operacao.Token.Position.Line)
	case parser.POWER:
		b.emit(OP_POW, 0, operacao.Token.Position.Line)
	}
	return nil
}

func (b *BytecodeBackend) ChamadaFuncao(chamada *parser.ChamadaFuncao) interface{} {
	// Valida a função usando o registro
	assinatura, ok := registry.RegistroGlobal.ObterAssinatura(chamada.Nome)
	if !ok {
		// Função não encontrada - gera erro de compilação
		return nil
	}

	// Para validação de argumentos em tempo de compilação, precisamos apenas verificar o número
	numArgs := len(chamada.Argumentos)
	if numArgs < assinatura.MinArgumentos {
		// Erro de validação - muito poucos argumentos
		return nil
	}
	if assinatura.MaxArgumentos != -1 && numArgs > assinatura.MaxArgumentos {
		// Erro de validação - muitos argumentos
		return nil
	}

	// Gera bytecode baseado no tipo da função
	switch assinatura.TipoFuncao {
	case registry.FUNCAO_IMPRIME:
		// Para cada argumento, gera bytecode para avaliar e imprimir
		for _, argumento := range chamada.Argumentos {
			argumento.Aceitar(b)
			b.emit(OP_PRINT, 0, chamada.Token.Position.Line)
		}
	case registry.FUNCAO_PURA:
		// Para funções puras, gera bytecode para avaliar argumentos
		for _, argumento := range chamada.Argumentos {
			argumento.Aceitar(b)
		}
		// TODO: Implementar chamada de função pura quando adicionar opcodes apropriados
		// Por enquanto, apenas deixa o resultado na pilha
	}
	return nil
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
	debug.Printf("🚀 Executando na Virtual Machine...\n")

	vm := NewVM(b.varCount)
	return vm.Execute(b.instructions)
}
