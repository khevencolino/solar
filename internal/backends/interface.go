package backends

import "github.com/khevencolino/Solar/internal/parser"

// Backend define a interface comum para todos os backends de compilação
type Backend interface {
	Compile(statements []parser.Expressao) error
	GetName() string
	GetExtension() string
}

// CompilationResult contém informações sobre o resultado da compilação
type CompilationResult struct {
	OutputFile string   // Arquivo de saída gerado
	ExecuteCmd []string // Comando para executar o arquivo compilado
	Success    bool     // Se a compilação foi bem-sucedida
	Message    string   // Mensagem de resultado ou erro
}

// BackendConfig centraliza configurações comuns dos backends
type BackendConfig struct {
	Debug     bool   // Habilita mensagens de debug
	OutputDir string // Diretório de saída
	Optimize  bool   // Habilita otimizações
}
