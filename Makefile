.PHONY: all build run assemble clean help docker-build docker-run docker-clean deps

# TODO Atualizar docker para Amd64

# Variáveis
COMPILER_NAME := kite-compiler
COMPILER_MAIN := ./cmd/compiler/main.go
OUTPUT_ASM := result/saida.s
OUTPUT_OBJ := saida.o
RUNTIME_S := external/runtime.s
EXECUTABLE_NAME := executavel

# Docker
DOCKER_IMAGE := kite-compiler
DOCKER_TAG := latest
DOCKER_CONTAINER := kite-compiler-container

# Diretórios
PROJECT_ROOT := $(shell pwd)
RESULT_DIR := result
EXTERNAL_DIR := external
EXAMPLES_DIR := examples

# Go settings
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

# --- Alvos Principais ---

# Alvo padrão: constrói o compilador
all: build

# Exibe as opções disponíveis
help:
	@echo "Makefile para o Compilador Kite"
	@echo "================================"
	@echo ""
	@echo "🏗️  Compilação Local:"
	@echo "  make build                     - Constrói o executável do compilador Go"
	@echo "  make run INPUT_FILE=<path>     - Executa o compilador Go localmente"
	@echo "  make assemble                  - Monta e linka o 'saida.s' gerado com 'runtime.s'"
	@echo ""
	@echo "🐳 Docker:"
	@echo "  make docker-build              - Constrói a imagem Docker do compilador"
	@echo "  make docker-run INPUT_FILE=<path> - Executa o compilador em container Docker"
	@echo "  make docker-clean              - Remove imagens e containers Docker"
	@echo ""
	@echo "🧪 Utilitários:"
	@echo "  make deps                      - Instala/atualiza dependências"
	@echo "  make clean                     - Remove arquivos gerados"
	@echo ""
	@echo "💡 Exemplos:"
	@echo "  make run INPUT_FILE=examples/math.kite"
	@echo "  make docker-run INPUT_FILE=examples/math.kite"

# --- Alvos Locais ---

# Verifica se Go está instalado
check-go:
	@which go > /dev/null || (echo "❌ Go não está instalado. Visite https://golang.org/doc/install" && exit 1)
	@echo "✅ Go $(shell go version | cut -d' ' -f3) detectado"

# Instala/atualiza dependências
deps: check-go
	@echo "📦 Instalando dependências..."
	go mod tidy
	go mod download
	@echo "✅ Dependências instaladas"

# Constrói o executável do compilador Go
build: check-go deps
	@echo "🏗️  Construindo o compilador Go..."
	@mkdir -p $(RESULT_DIR)
	go build -ldflags="-s -w" -o $(COMPILER_NAME) $(COMPILER_MAIN)
	@echo "✅ Compilador Go construído: $(COMPILER_NAME)"

# Executa o compilador Go localmente com um arquivo de entrada
# Uso: make run INPUT_FILE=valid_program.kite
run: build
ifndef INPUT_FILE
	@echo "❌ Erro: INPUT_FILE não está definido"
	@echo "📖 Uso: make run INPUT_FILE=<caminho/para/seu/programa.kite>"
	@echo "📖 Exemplo: make run INPUT_FILE=examples/math.kite"
	@exit 1
endif
	@echo "🚀 Executando compilador em $(INPUT_FILE)..."
	@if [ ! -f "$(INPUT_FILE)" ]; then \
		echo "❌ Erro: Arquivo $(INPUT_FILE) não encontrado"; \
		exit 1; \
	fi
	./$(COMPILER_NAME) $(INPUT_FILE)
	@echo "✅ Assembly gerado: $(OUTPUT_ASM)"

# Monta o arquivo assembly gerado (saida.s) e o linka com runtime.s
assemble: $(OUTPUT_ASM) $(RUNTIME_S)
	@echo "🔧 Montando $(OUTPUT_ASM) com GAS..."
	as --64 -o $(OUTPUT_OBJ) $(OUTPUT_ASM)
	@echo "🔗 Linkando $(OUTPUT_OBJ) com $(RUNTIME_S) usando LD..."
	ld -o $(EXECUTABLE_NAME) $(OUTPUT_OBJ)
	@echo "✅ Executável final criado: $(EXECUTABLE_NAME)"
	@echo "🏃 Para executar: ./$(EXECUTABLE_NAME)"

# Executa o programa completo (compilar + montar + executar)
run-complete: run assemble
	@echo "🏃 Executando programa..."
	./$(EXECUTABLE_NAME)

# --- Alvos Docker ---

# Constrói a imagem Docker
docker-build:
	@echo "🐳 Construindo imagem Docker..."
	@if [ ! -f "Dockerfile" ]; then \
		echo "📝 Criando Dockerfile..."; \
		$(MAKE) create-dockerfile; \
	fi
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "✅ Imagem Docker construída: $(DOCKER_IMAGE):$(DOCKER_TAG)"

# Executa o compilador em container Docker
# Uso: make docker-run INPUT_FILE=examples/math.kite
docker-run: docker-build
ifndef INPUT_FILE
	@echo "❌ Erro: INPUT_FILE não está definido"
	@echo "📖 Uso: make docker-run INPUT_FILE=<caminho/para/seu/programa.kite>"
	@echo "📖 Exemplo: make docker-run INPUT_FILE=examples/math.kite"
	@exit 1
endif
	@echo "🐳 Executando compilador em Docker com $(INPUT_FILE)..."
	@if [ ! -f "$(INPUT_FILE)" ]; then \
		echo "❌ Erro: Arquivo $(INPUT_FILE) não encontrado"; \
		exit 1; \
	fi
	@# Remove container se existir
	-docker rm -f $(DOCKER_CONTAINER) 2>/dev/null || true
	@# Executa o container
	docker run --name $(DOCKER_CONTAINER) \
		-v $(PROJECT_ROOT):/workspace \
		-w /workspace \
		$(DOCKER_IMAGE):$(DOCKER_TAG) \
		./$(COMPILER_NAME) $(INPUT_FILE)
	@echo "✅ Compilação Docker concluída"
	@echo "📁 Resultados disponíveis em: $(RESULT_DIR)/"

# Executa programa completo no Docker (compilar + montar + executar)
docker-run-complete: docker-run
	@echo "🐳 Executando programa completo no Docker..."
	@if [ ! -f "$(RUNTIME_S)" ]; then \
		echo "❌ Erro: Arquivo $(RUNTIME_S) não encontrado"; \
		exit 1; \
	fi
	docker run --name $(DOCKER_CONTAINER)-exec \
		-v $(PROJECT_ROOT):/workspace \
		-w /workspace \
		--rm \
		$(DOCKER_IMAGE):$(DOCKER_TAG) \
		sh -c "as --64 -o $(OUTPUT_OBJ) $(OUTPUT_ASM) && ld -o $(EXECUTABLE_NAME) $(OUTPUT_OBJ) && ./$(EXECUTABLE_NAME)"
	@echo "✅ Execução completa no Docker concluída"

# Remove imagens e containers Docker
docker-clean:
	@echo "🧹 Limpando recursos Docker..."
	-docker rm -f $(DOCKER_CONTAINER) 2>/dev/null || true
	-docker rm -f $(DOCKER_CONTAINER)-exec 2>/dev/null || true
	-docker rmi $(DOCKER_IMAGE):$(DOCKER_TAG) 2>/dev/null || true
	@echo "✅ Limpeza Docker concluída"

# --- Alvos de Limpeza ---

# Limpa arquivos gerados localmente
clean:
	@echo "🧹 Limpando arquivos gerados..."
	rm -f $(COMPILER_NAME)
	rm -f $(OUTPUT_OBJ)
	rm -f $(EXECUTABLE_NAME)
	rm -rf $(RESULT_DIR)
	@echo "✅ Limpeza local concluída"

# Limpeza completa (local + Docker)
clean-all: clean docker-clean
	@echo "🧹 Limpeza completa concluída"

# --- Alvos de Desenvolvimento ---

# Formata o código
fmt: check-go
	@echo "🎨 Formatando código..."
	go fmt ./...
	@echo "✅ Código formatado"

# Executa linter
lint: check-go
	@echo "🔍 Executando linter..."
	@if ! command -v golangci-lint &> /dev/null; then \
		echo "📦 Instalando golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	golangci-lint run
	@echo "✅ Linter concluído"


# Mostra informações do projeto
info:
	@echo "📊 Informações do Projeto Kite Compiler"
	@echo "========================================"
	@echo "🏗️  Compilador: $(COMPILER_NAME)"
	@echo "📁 Diretório: $(PROJECT_ROOT)"
	@echo "🐳 Imagem Docker: $(DOCKER_IMAGE):$(DOCKER_TAG)"
	@echo "🖥️  SO/Arch: $(GOOS)/$(GOARCH)"
	@echo "🔧 Go Version: $(shell go version 2>/dev/null || echo 'não instalado')"
	@echo "📦 Docker: $(shell docker --version 2>/dev/null || echo 'não instalado')"
