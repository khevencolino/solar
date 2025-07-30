package bytecode

type OpCode byte

const (
	OP_CONST OpCode = iota // CONST valor
	OP_ADD                 // ADD
	OP_SUB                 // SUB
	OP_MUL                 // MUL
	OP_DIV                 // DIV
	OP_MOD                 // MOD
	OP_LOAD                // LOAD variavel_index
	OP_STORE               // STORE variavel_index
	OP_PRINT               // PRINT
	OP_HALT                // HALT
	OP_POW                 // POWER
)

type Instruction struct {
	OpCode  OpCode
	Operand int64
	Line    int // para debug
}

func (op OpCode) String() string {
	switch op {
	case OP_CONST:
		return "CONST"
	case OP_ADD:
		return "ADD"
	case OP_SUB:
		return "SUB"
	case OP_MUL:
		return "MUL"
	case OP_DIV:
		return "DIV"
	case OP_MOD:
		return "MOD"
	case OP_LOAD:
		return "LOAD"
	case OP_STORE:
		return "STORE"
	case OP_PRINT:
		return "PRINT"
	case OP_HALT:
		return "HALT"
	case OP_POW:
		return "POWER"
	default:
		return "UNKNOWN"
	}
}
