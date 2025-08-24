.PHONY: help build run clean fmt lint test

# Variáveis
COMPILER_NAME := solar-compiler
COMPILER_MAIN := ./cmd/compiler/main.go

# Alvo padrão
all: build

# Ajuda
help:
	@echo "Compilador Solar ☀️"
	@echo "==================="
	@echo ""
	@echo "🏗️  make build                         - Construir compilador"
	@echo "🚀 make run FILE=<arquivo>             - Executar (interpretador)"
	@echo "🤖 make run FILE=<arquivo> BACKEND=bytecode"
	@echo "🔧 make run FILE=<arquivo> BACKEND=assembly"
	@echo "🎨 make fmt                            - Formatar código"
	@echo "🔍 make lint                           - Verificar código"
	@echo "🧪 make test                           - Executar testes"
	@echo "🧹 make clean                          - Limpar arquivos"
	@echo ""
	@echo "Exemplos:"
	@echo "  make run FILE=exemplos/operacao/valido.solar"
	@echo "  make run FILE=exemplos/funcoes_builtin/teste_simples.solar BACKEND=bytecode"

# Construir
build:
	@echo "🏗️ Construindo..."
	go build -o $(COMPILER_NAME) $(COMPILER_MAIN)
	@echo "✅ Pronto: $(COMPILER_NAME)"

# Executar
run: build
ifndef FILE
	@echo "❌ Erro: FILE não definido"
	@echo "📖 Uso: make run FILE=<arquivo> [BACKEND=<tipo>]"
	@echo "📖 Exemplo: make run FILE=exemplos/operacao/valido.solar"
	@exit 1
endif
	@echo "🚀 Executando $(FILE)..."
ifdef BACKEND
	./$(COMPILER_NAME) -backend=$(BACKEND) $(FILE)
else
	./$(COMPILER_NAME) $(FILE)
endif

# Desenvolvimento
fmt:
	@echo "🎨 Formatando código..."
	go fmt ./...
	@echo "✅ Código formatado"

lint:
	@echo "🔍 Verificando código..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "⚠️  golangci-lint não instalado"; \
		echo "📦 Instale com: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

test:
	@echo "🧪 Executando testes..."
	go test ./...
	@echo "✅ Testes concluídos"

# Limpeza
clean:
	@echo "🧹 Limpando arquivos..."
	rm -f $(COMPILER_NAME)
	@echo "✅ Limpo"
