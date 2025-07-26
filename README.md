# Compilador Kite

Um compilador experimental para a linguagem Kite, escrito em Go, que gera código assembly x86-64 com análise sintática completa e interpretação de expressões.

## 📋 Índice

- [Visão Geral](#-visão-geral)
- [Recursos](#-recursos)
- [Pré-requisitos](#-pré-requisitos)
- [Instalação](#-instalação)
- [Como Usar](#-como-usar)
- [Exemplos](#-exemplos)
- [Análise Sintática e Interpretação](#-análise-sintática-e-interpretação)
- [Estrutura do Projeto](#-estrutura-do-projeto)
- [Desenvolvimento](#-desenvolvimento)
- [Docker](#-docker)
- [Arquitetura](#-arquitetura)

## 🎯 Visão Geral

O Kite é um compilador que converte código fonte da linguagem Kite em assembly x86-64. Atualmente suporta:

- **Análise Léxica**: Tokenização de números e operadores matemáticos
- **Parser**: Análise sintática com construção de AST
- **Interpretador**: Execução e avaliação de expressões matemáticas
- **Visualização de AST**: Representação gráfica da árvore sintática
- **Validação Sintática**: Verificação de parênteses balanceados
- **Geração de Assembly**: Produção de código assembly x86-64
- **Runtime**: Funções de suporte para impressão e saída

## ✨ Recursos

- 🔢 **Números inteiros**: Suporte a constantes numéricas
- 🧮 **Operadores**: `+`, `-`, `*`, `**` (potência)
- 📐 **Parênteses**: Agrupamento de expressões
- 🌳 **AST**: Construção de árvore sintática abstrata
- 🔍 **Interpretador**: Avaliação completa de expressões
- 📈 **Visualização**: Representação gráfica da AST
- 🔧 **Assembly x86-64**: Geração de código nativo
- 🐳 **Docker**: Ambiente containerizado
- 📊 **Debugging**: Visualização de tokens

## 📦 Pré-requisitos

### Desenvolvimento Local
- **Go 1.21+**: [Instalar Go](https://golang.org/doc/install)
- **GAS Assembler**: Parte do GNU Binutils
- **GNU Linker (ld)**: Para linking do executável
- **Make**: Para automação de build

### Dependências Go
O projeto utiliza a biblioteca `treedrawer` para visualização da AST:
```bash
go mod download
```

### Ubuntu/Debian
```bash
sudo apt update
sudo apt install golang-go build-essential binutils make
```

### Arch Linux
```bash
sudo pacman -S go base-devel binutils make
```

### Docker (Alternativa)
- **Docker**: [Instalar Docker](https://docs.docker.com/get-docker/)

## 🚀 Instalação

### Clonagem do Repositório
```bash
git clone https://github.com/khevencolino/Kite.git
cd Kite
```

### Build Local
```bash
# Build do compilador
make build

# Ou manualmente
go build -o kite-compiler ./main.go
```

### Build com Docker
```bash
make docker-build
```

## 📖 Como Usar

### Uso Básico

1. **Criar um arquivo `.kite`**:
```bash
echo "123" > meu_programa.kite
```

2. **Compilar e Interpretar**:
```bash
# Local
make run INPUT_FILE=meu_programa.kite

# Docker
make docker-run INPUT_FILE=meu_programa.kite
```

3. **Montar e executar** (assembly desabilitado temporariamente):
```bash
# Gera o executável final
make assemble

# Executar
./executavel
```

### Fluxo Completo
```bash
# Compilar + Interpretar em um comando
make run INPUT_FILE=meu_programa.kite

# Com Docker
make docker-run INPUT_FILE=meu_programa.kite
```

### Linha de Comando Direta
```bash
# Depois do build
./kite-compiler meu_programa.kite

# Mostra tokens, AST e resultado da interpretação
```

## 🧪 Exemplos

### Exemplo 1: Número Simples
**Arquivo**: `exemplos/stage01/valido.kite`
```
123
```

**Compilação**:
```bash
make run INPUT_FILE=exemplos/stage01/valido.kite
```

**Saída esperada**:
```
Tokens encontrados:
TIPO       VALOR           POSIÇÃO
--------------------------------------------------
NUMBER     123             linha 1, coluna 1

=== Árvore Sintática ===
123

Resultado da expressão: 123
✅ Compilação concluída com sucesso!
```

### Exemplo 2: Expressão com Parênteses
**Arquivo**: `exemplos/stage02/valido.kite`
```
(11 + 2)
```

**Compilação**:
```bash
make run INPUT_FILE=exemplos/stage02/valido.kite
```

**Saída esperada**:
```
Tokens encontrados:
TIPO       VALOR           POSIÇÃO
--------------------------------------------------
LPAREN     (               linha 1, coluna 1
NUMBER     11              linha 1, coluna 2
PLUS       +               linha 1, coluna 5
NUMBER     2               linha 1, coluna 7
RPAREN     )               linha 1, coluna 8

=== Árvore Sintática ===
+
├── 11
└── 2

Resultado da expressão: 13
✅ Compilação concluída com sucesso!
```

### Exemplo 3: Expressão Inválida
**Arquivo**: `exemplos/stage02/invalido.kite`
```
(11 + A + 23 + B)
```

**Resultado**: Erro de tokenização devido aos caracteres inválidos `A` e `B`.

### Testando Exemplos
```bash
# Testar arquivo válido do stage 1
make run INPUT_FILE=exemplos/stage01/valido.kite

# Testar arquivo inválido do stage 1
make run INPUT_FILE=exemplos/stage01/invalido.kite

# Testar arquivo válido do stage 2
make run INPUT_FILE=exemplos/stage02/valido.kite

# Testar arquivo inválido do stage 2
make run INPUT_FILE=exemplos/stage02/invalido.kite
```

### Exemplos Avançados (Stage 3)
```bash
# Testar expressão complexa válida
make run INPUT_FILE=exemplos/stage03/valido.kite

# Testar expressão com erro de sintaxe
make run INPUT_FILE=exemplos/stage03/invalido.kite
```

## 🌳 Análise Sintática e Interpretação

### Exemplo Completo com AST
**Arquivo**: `exemplos/stage03/valido.kite`
```
((11 + 2) + (8 * 9))
```

**Compilação**:
```bash
make run INPUT_FILE=exemplos/stage03/valido.kite
```

**Saída esperada**:
```
Tokens encontrados:
TIPO       VALOR           POSIÇÃO
--------------------------------------------------
LPAREN     (               linha 1, coluna 1
LPAREN     (               linha 1, coluna 2
NUMBER     11              linha 1, coluna 3
PLUS       +               linha 1, coluna 6
NUMBER     2               linha 1, coluna 8
RPAREN     )               linha 1, coluna 9
PLUS       +               linha 1, coluna 11
LPAREN     (               linha 1, coluna 13
NUMBER     8               linha 1, coluna 14
MULTIPLY   *               linha 1, coluna 16
NUMBER     9               linha 1, coluna 18
RPAREN     )               linha 1, coluna 19
RPAREN     )               linha 1, coluna 20

=== Árvore Sintática ===
+
├── +
│   ├── 11
│   └── 2
└── *
    ├── 8
    └── 9

Resultado da expressão: 85
✅ Compilação concluída com sucesso!
```

### Exemplo com Erro de Sintaxe
**Arquivo**: `exemplos/stage03/invalido.kite`
```
(11 + 2))
```

**Resultado**: Erro de parênteses não balanceados detectado durante a análise léxica.

## 📁 Estrutura do Projeto

```
Kite/
├── cmd/compiler/main.go          # Ponto de entrada alternativo
├── exemplos/                     # Exemplos de código Kite
│   ├── stage01/                  # Números simples
│   ├── stage02/                  # Expressões com parênteses
│   └── stage03/                  # Expressões complexas aninhadas
├── external/                     # Arquivos de suporte
│   ├── assembly_examples/        # Exemplos de assembly
│   └── runtime.s                 # Runtime do assembly
├── internal/                     # Código interno do compilador
│   ├── compiler/                 # Lógica principal do compilador
│   │   ├── compiler.go          # Coordenador principal
│   │   └── generator.go         # Gerador de assembly
│   ├── lexer/                   # Analisador léxico
│   │   ├── lexer.go            # Tokenização
│   │   ├── position.go         # Posicionamento no código
│   │   └── token.go            # Definições de tokens
│   ├── parser/                  # Analisador sintático
│   │   ├── ast.go              # Definições da AST
│   │   ├── parser.go           # Parser descendente recursivo
│   │   ├── interpretador.go    # Interpretador de expressões
│   │   └── visualizador.go     # Visualização da AST
│   └── utils/                   # Utilitários
│       ├── error.go            # Sistema de erros
│       └── file.go             # Manipulação de arquivos
├── result/                      # Arquivos gerados
│   └── saida.s                 # Assembly gerado
├── Dockerfile                   # Container Docker
├── Makefile                     # Automação de build
├── go.mod                      # Dependências Go
├── go.sum                      # Checksums das dependências
└── main.go                     # Ponto de entrada principal
```

## 🛠️ Desenvolvimento

### Comandos Make Disponíveis

```bash
# Mostrar ajuda
make help

# Desenvolvimento
make build                        # Build do compilador
make run INPUT_FILE=<arquivo>     # Executar compilador
make assemble                     # Montar assembly
make run-complete INPUT_FILE=<arquivo> # Fluxo completo

# Docker
make docker-build                 # Build da imagem
make docker-run INPUT_FILE=<arquivo> # Executar no Docker
make docker-clean                 # Limpar recursos Docker

# Utilitários
make clean                        # Limpar arquivos gerados
make deps                         # Instalar dependências
make fmt                          # Formatar código
make lint                         # Executar linter
make info                         # Informações do projeto
```

### Adicionando Novos Recursos

1. **Novos Tokens**: Adicione em `internal/lexer/token.go`
2. **Análise Léxica**: Modifique `internal/lexer/lexer.go`
3. **AST**: Edite `internal/parser/ast.go` para novos tipos de nós
4. **Parser**: Modifique `internal/parser/parser.go` para nova sintaxe
5. **Interpretador**: Atualize `internal/parser/interpretador.go` para nova semântica
6. **Geração de Código**: Edite `internal/compiler/generator.go`
7. **Testes**: Crie arquivos em `exemplos/`

### Debug e Análise

O compilador mostra informações detalhadas durante a execução:

```bash
make run INPUT_FILE=exemplos/stage03/valido.kite
```

**Saída de exemplo**:
```
Tokens encontrados:
TIPO       VALOR           POSIÇÃO
--------------------------------------------------
LPAREN     (               linha 1, coluna 1
NUMBER     11              linha 1, coluna 2
PLUS       +               linha 1, coluna 5
NUMBER     2               linha 1, coluna 7
RPAREN     )               linha 1, coluna 8

=== Árvore Sintática ===
+
├── 11
└── 2

Resultado da expressão: 13
✅ Compilação concluída com sucesso!
```

## 🐳 Docker

### Build e Execução
```bash
# Build da imagem
make docker-build

# Executar compilador
make docker-run INPUT_FILE=exemplos/stage01/valido.kite

# Execução completa
make docker-run-complete INPUT_FILE=exemplos/stage01/valido.kite

# Limpeza
make docker-clean
```

### Uso Manual do Docker
```bash
# Build
docker build -t kite-compiler .

# Executar
docker run --rm -v $(pwd):/workspace -w /workspace \
  kite-compiler ./kite-compiler exemplos/stage01/valido.kite
```

## 🏗️ Arquitetura

### Fluxo de Compilação

1. **main.go** → Ponto de entrada, processa argumentos
2. **compiler.go** → Coordena o processo de compilação
3. **lexer.go** → Tokeniza o código fonte
4. **parser.go** → Constrói a AST (Abstract Syntax Tree)
5. **interpretador.go** → Avalia a AST e calcula resultado
6. **visualizador.go** → Gera representação gráfica da AST
7. **generator.go** → Gera código assembly x86-64 (desabilitado)
8. **runtime.s** → Fornece funções de runtime (impressão, saída)

### Componentes Principais

- **Lexer**: Análise léxica com regex patterns
- **Parser**: Análise sintática descendente recursiva
- **AST**: Abstract Syntax Tree para representação estrutural
- **Interpretador**: Padrão Visitor para avaliação de expressões
- **Visualizador**: Representação gráfica da árvore sintática
- **Compiler**: Coordenação entre lexer, parser e generator
- **Generator**: Template-based assembly generation
- **Runtime**: Assembly functions for I/O operations
- **Utils**: File I/O and error handling

### Estado Atual vs. Planejado

**✅ Implementado:**
- Tokenização completa de expressões matemáticas
- Validação de parênteses balanceados
- **Parser completo com análise sintática descendente recursiva**
- **Construção de AST (Abstract Syntax Tree)**
- **Interpretador funcional com avaliação de expressões**
- **Visualização gráfica da árvore sintática**
- Geração básica de assembly (desabilitada temporariamente)
- Sistema de runtime funcional
- Suporte a Docker e Make

**🚧 Em Desenvolvimento:**
- Reativação da geração de assembly baseada na AST
- Análise de precedência de operadores
- Suporte a variáveis e funções
- Otimizações de código
- Operador de divisão no lexer
- Mapeamento correto do operador de potência

## 🤝 Contribuição

1. Fork o projeto
2. Crie uma branch para sua feature (`git checkout -b feature/nova-feature`)
3. Commit suas mudanças (`git commit -am 'Adiciona nova feature'`)
4. Push para a branch (`git push origin feature/nova-feature`)
5. Abra um Pull Request

## 📝 Licença

Este projeto está sob licença MIT. Veja o arquivo `LICENSE` para mais detalhes.

## 🐛 Problemas Conhecidos

- Geração de assembly foi desabilitada temporariamente (comentada no código)
- Operador de potência (**) é tokenizado mas mapeado como multiplicação no parser
- Falta suporte a divisão no lexer (implementado apenas no parser)
- Análise de precedência de operadores ainda não implementada

## 📞 Suporte

Para dúvidas e problemas:
- Abra uma [Issue](https://github.com/khevencolino/Kite/issues)
- Consulte a documentação dos comandos: `make help`

---

**Desenvolvido com ❤️ em Go**
