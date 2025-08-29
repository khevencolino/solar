package llvm

import (
	"github.com/khevencolino/Solar/internal/backends"
	"github.com/khevencolino/Solar/internal/parser"
)

type LLVMGenerator interface {
	backends.Backend
	Compile(statements []parser.Expressao) error
}

// GetGeneratorInfo retorna informações sobre o gerador LLVM
func GetGeneratorInfo() map[string]interface{} {
	return map[string]interface{}{
		"name":        "LLVM IR",
		"description": "Gerador de código LLVM IR",
		"extension":   ".ll",
		"executable":  true,
		"requires": []string{
			"llvm-tools (opcional, para compilação)",
			"clang (para linking)",
		},
	}
}
