package bytecode

import (
	"fmt"
	"math"
)

type VM struct {
	stack     []int64
	stackTop  int
	variables []int64
	pc        int // program counter
}

func NewVM(varCount int) *VM {
	return &VM{
		stack:     make([]int64, 256), // stack fixo
		stackTop:  0,
		variables: make([]int64, varCount),
		pc:        0,
	}
}

func (vm *VM) Execute(instructions []Instruction) error {
	fmt.Printf("📊 Bytecode gerado (%d instruções):\n", len(instructions))
	for i, instr := range instructions {
		fmt.Printf("  %03d: %s %d\n", i, instr.OpCode, instr.Operand)
	}
	fmt.Println()

	fmt.Printf("🏃 Executando...\n")

	for vm.pc < len(instructions) {
		instr := instructions[vm.pc]

		switch instr.OpCode {
		case OP_CONST:
			vm.push(instr.Operand)

		case OP_ADD:
			b := vm.pop()
			a := vm.pop()
			vm.push(a + b)

		case OP_SUB:
			b := vm.pop()
			a := vm.pop()
			vm.push(a - b)

		case OP_MUL:
			b := vm.pop()
			a := vm.pop()
			vm.push(a * b)

		case OP_POW:
			b := vm.pop()
			a := vm.pop()
			vm.push(int64(math.Pow(float64(b), float64(a))))

		case OP_DIV:
			b := vm.pop()
			a := vm.pop()
			if b == 0 {
				return fmt.Errorf("divisão por zero na linha %d", instr.Line)
			}
			vm.push(a / b)

		case OP_LOAD:
			if int(instr.Operand) >= len(vm.variables) {
				return fmt.Errorf("índice de variável inválido: %d", instr.Operand)
			}
			vm.push(vm.variables[instr.Operand])

		case OP_STORE:
			value := vm.pop()
			if int(instr.Operand) >= len(vm.variables) {
				return fmt.Errorf("índice de variável inválido: %d", instr.Operand)
			}
			vm.variables[instr.Operand] = value
			vm.push(value) // atribuição retorna o valor

		case OP_PRINT:
			fmt.Printf("Resultado: %d\n", vm.peek())

		case OP_HALT:
			fmt.Printf("✅ Execução concluída!\n")
			return nil

		default:
			return fmt.Errorf("opcode desconhecido: %d", instr.OpCode)
		}

		vm.pc++
	}

	return nil
}

func (vm *VM) push(value int64) {
	if vm.stackTop >= len(vm.stack) {
		panic("stack overflow")
	}
	vm.stack[vm.stackTop] = value
	vm.stackTop++
}

func (vm *VM) pop() int64 {
	if vm.stackTop <= 0 {
		panic("stack underflow")
	}
	vm.stackTop--
	return vm.stack[vm.stackTop]
}

func (vm *VM) peek() int64 {
	if vm.stackTop <= 0 {
		panic("stack empty")
	}
	return vm.stack[vm.stackTop-1]
}
