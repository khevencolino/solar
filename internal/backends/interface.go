package backends

import "github.com/khevencolino/Kite/internal/parser"

type Backend interface {
	Compile(statements []parser.Expressao) error
	GetName() string
	GetExtension() string
}

type CompilationResult struct {
	OutputFile string
	ExecuteCmd []string
	Success    bool
	Message    string
}
