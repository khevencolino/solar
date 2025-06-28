.PHONY: all build run assemble clean test help

# Variáveis
COMPILER_NAME := kite-compiler
COMPILER_MAIN := ./cmd/compiler/main.go
OUTPUT_ASM := result/saida.s
OUTPUT_OBJ := saida.o
RUNTIME_S := external/runtime.s
EXECUTABLE_NAME := executavel

# --- Alvos Principais ---

# Alvo padrão: constrói o compilador.
all: build

# Exibe as opções disponíveis.
help:
	@echo "Makefile para o Compilador Kite"
	@echo "-----------------------------------"
	@echo "make build             - Constrói o executável do compilador Go."
	@echo "make run INPUT_FILE=<path> - Executa o compilador Go localmente."
	@echo "make assemble          - Monta e linka o 'saida.s' gerado com 'runtime.s'."
	@echo "make clean             - Remove arquivos gerados (executável, .s, .o)."

# --- Alvos Locais ---

# Constrói o executável do compilador Go.
build:
	@echo "Construindo o compilador Go..."
	go build -o $(COMPILER_NAME) $(COMPILER_MAIN)
	@echo "Compilador Go construído: $(COMPILER_NAME)"

# Executa o compilador Go localmente com um arquivo de entrada.
# Uso: make run INPUT_FILE=valid_program.kite
run: build
ifndef INPUT_FILE
	@echo "Erro: INPUT_FILE não está definido. Uso: make run INPUT_FILE=<caminho/para/seu/programa.kite>"
	@exit 1
endif
	@echo "Executando compilador em $(INPUT_FILE)..."
	./$(COMPILER_NAME) $(INPUT_FILE)
	@echo "Assembly gerado: $(OUTPUT_ASM)"

# Monta o arquivo assembly gerado (saida.s) e o linka com runtime.s.
# Requer que 'saida.s' e 'runtime.s' estejam presentes no diretório de execução.
assemble: $(OUTPUT_ASM) $(RUNTIME_S)
	@echo "Montando $(OUTPUT_ASM) com GAS..."
	as --64 -o $(OUTPUT_OBJ) $(OUTPUT_ASM)
	@echo "Linkando $(OUTPUT_OBJ) com $(RUNTIME_S) usando LD..."
	ld -o $(EXECUTABLE_NAME) $(OUTPUT_OBJ)
	@echo "Executável final criado: $(EXECUTABLE_NAME)"

# Garante que 'saida.s' e 'runtime.s' existam para o alvo 'assemble'.
$(OUTPUT_ASM):
	@echo "Erro: Arquivo $(OUTPUT_ASM) não encontrado. Por favor, execute 'make run INPUT_FILE=<seu_arquivo.kite>' primeiro para gerá-lo."
	@exit 1

$(RUNTIME_S):
	@echo "Erro: Arquivo $(RUNTIME_S) não encontrado. Por favor, certifique-se de que ele esteja na raiz do projeto."
	@exit 1

# Limpa os arquivos gerados (executável do compilador, .s).
clean:
	@echo "Limpando arquivos gerados..."
	rm -f $(COMPILER_NAME) $(OUTPUT_ASM)
	@echo "Limpeza concluída."
