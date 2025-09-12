# solar

Compilador para a linguagem Solar com múltiplos backends de execução.

## Índice

- [Instalação](#-instalação)
- [Como Usar](#-como-usar)
- [Exemplos](#-exemplos)
- [Funções Builtin](#-funções-builtin)
- [Backends](#-backends)
- [Desenvolvimento](#-desenvolvimento)

## Instalação

```bash
# Clonar repositório
git clone https://github.com/khevencolino/solar.git
cd solar

# Construir compilador
make build
```

## Como Usar

```bash
# Ajuda
make help

# Executar com interpretador (padrão)
make run FILE=exemplos/operacao/valido.solar

# Executar com bytecode + VM
make run FILE=exemplos/operacao/valido.solar BACKEND=bytecode

# Executar com assembly
make run FILE=exemplos/operacao/valido.solar BACKEND=assembly

# Executar com LLVM IR
make run FILE=exemplos/operacao/valido.solar BACKEND=llvm
```

### CLI Direta

```bash
# Interpretador
go run cmd/compiler/main.go arquivo.solar

# Bytecode
go run cmd/compiler/main.go -backend=bytecode arquivo.solar

# LLVM IR
go run cmd/compiler/main.go -backend=llvm arquivo.solar

# Assembly x86-64
go run cmd/compiler/main.go -backend=assembly arquivo.solar

# Com debug habilitado
go run cmd/compiler/main.go -debug arquivo.solar
```

## Exemplos

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

## Backends

### Interpretador

Execução direta da AST.

### Bytecode + VM

Compilação para bytecode próprio com máquina virtual.

### Assembly (x86-64)

Geração de código nativo.

### LLVM IR

Compilação para LLVM Intermediate Representation.

```bash
make run FILE=arquivo.solar BACKEND=bytecode
make run FILE=arquivo.solar BACKEND=assembly
make run FILE=arquivo.solar BACKEND=llvm
```

## Desenvolvimento

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

## Licença

MIT License
