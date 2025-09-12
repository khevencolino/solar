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

	// Operações de comparação
	OP_EQ // EQUAL (==)
	OP_NE // NOT_EQUAL (!=)
	OP_LT // LESS_THAN (<)
	OP_GT // GREATER_THAN (>)
	OP_LE // LESS_EQUAL (<=)
	OP_GE // GREATER_EQUAL (>=)

	// Estruturas de controle
	OP_JMP // JMP endereço (pulo incondicional)
	OP_JF  // JF endereço (pulo se falso)
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
	case OP_EQ:
		return "EQ"
	case OP_NE:
		return "NE"
	case OP_LT:
		return "LT"
	case OP_GT:
		return "GT"
	case OP_LE:
		return "LE"
	case OP_GE:
		return "GE"
	case OP_JMP:
		return "JMP"
	case OP_JF:
		return "JF"
	default:
		return "UNKNOWN"
	}
}
