package assembly

import (
	"fmt"

	"github.com/khevencolino/Solar/internal/backends"
	"github.com/khevencolino/Solar/internal/backends/assembly/arm64"
	"github.com/khevencolino/Solar/internal/backends/assembly/x86_64"
	"github.com/khevencolino/Solar/internal/parser"
)

type AssemblyBackend interface {
	backends.Backend
	Compile(statements []parser.Expressao) error
}

func NewAssemblyBackend(arch string) (backends.Backend, error) {
	switch arch {
	case "x86_64", "amd64":
		return x86_64.NewX86_64Backend(), nil
	case "arm64", "aarch64":
		return arm64.NewARM64Backend(), nil
	default:
		return nil, fmt.Errorf("unsupported assembly architecture: %s", arch)
	}
}
