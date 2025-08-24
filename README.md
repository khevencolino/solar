# Compilador Solar ☀️

Um compilador moderno para a linguagem Solar com múltiplos backends de execução.

## 🌟 Destaques

- **🎯 Múltiplos Backends**: Interpretador, Bytecode + VM, Assembly nativo
- **🧮 Precedência de Operadores**: Parsing inteligente sem parênteses obrigatórios
- **🔧 Funções Builtin**: `imprime()`, `soma()`, `abs()` extensíveis
- **🏗️ Arquiteturas**: x86-64 (Linux) e ARM64 (macOS)
- **🌳 Visualização AST**: Árvore sintática gráfica

## 📋 Índice

- [Instalação](#-instalação)
- [Como Usar](#-como-usar)
- [Exemplos](#-exemplos)
- [Funções Builtin](#-funções-builtin)
- [Backends](#-backends)
- [Desenvolvimento](#-desenvolvimento)

## 🚀 Instalação

```bash
# Clonar repositório
git clone https://github.com/khevencolino/Solar.git
cd Solar

# Construir compilador
make build
```

## 📖 Como Usar

```bash
# Ajuda
make help

# Executar com interpretador (padrão)
make run FILE=exemplos/operacao/valido.solar

# Executar com bytecode + VM
make run FILE=exemplos/operacao/valido.solar BACKEND=bytecode

# Executar com assembly
make run FILE=exemplos/operacao/valido.solar BACKEND=assembly
```

### CLI Direta

```bash
# Interpretador
go run cmd/compiler/main.go arquivo.solar

# Bytecode
go run cmd/compiler/main.go -backend=bytecode arquivo.solar

# Assembly ARM64
go run cmd/compiler/main.go -backend=assembly -arch=arm64 arquivo.solar
```

## 🧪 Exemplos

### Precedência de Operadores

```solar
2 + 3 * 4
```

**Resultado**: `14` (não `20` - precedência correta!)

```
=== Árvore Sintática ===
    +
   ╱ ╲
  2   *
     ╱ ╲
    3   4
```

### Funções Builtin

```solar
imprime(42)
imprime(10, 20, 30)
soma(5, 10)
abs(15)
```

**Saída**:

```
42
10 20 30
Resultado: 15
Resultado: 15
```

### Expressões Complexas

```solar
imprime(soma(2, 3), abs(-7))
```

### Testando

```bash
# Interpretador
make run FILE=exemplos/operacao/valido.solar

# Bytecode
make run FILE=exemplos/funcoes_builtin/teste_simples.solar BACKEND=bytecode

# Assembly
make run FILE=exemplos/power/valido.solar BACKEND=assembly
```

## 🔧 Backends

### Interpretador

Execução direta da AST.

### Bytecode + VM

Compilação para bytecode próprio com máquina virtual.

### Assembly (x86-64/ARM64)

Geração de código nativo.

```bash
# Escolher backend
make run FILE=arquivo.solar BACKEND=bytecode
make run FILE=arquivo.solar BACKEND=assembly
```

## 🛠️ Desenvolvimento

### Build

```bash
make build       # Compilar
make run FILE=exemplo.solar    # Executar
make clean       # Limpar
```

### Estrutura

- `cmd/` - CLI principal
- `internal/` - Core do compilador
- `exemplos/` - Casos de teste
- `external/` - Runtime assembly

### Contribuir

1. Fork do projeto
2. Branch para feature
3. Commit com testes
4. Pull request

---

## � Licença

MIT License
