# Compilador Kite

Um compilador experimental para a linguagem Kite, escrito em Go, que gera código assembly x86-64.

## 📋 Índice

- [Visão Geral](#-visão-geral)
- [Recursos](#-recursos)
- [Pré-requisitos](#-pré-requisitos)
- [Instalação](#-instalação)
- [Como Usar](#-como-usar)
- [Exemplos](#-exemplos)
- [Estrutura do Projeto](#-estrutura-do-projeto)
- [Desenvolvimento](#-desenvolvimento)
- [Docker](#-docker)
- [Arquitetura](#-arquitetura)

## 🎯 Visão Geral

O Kite é um compilador que converte código fonte da linguagem Kite em assembly x86-64. Atualmente suporta:

- **Análise Léxica**: Tokenização de números e operadores matemáticos
- **Validação Sintática**: Verificação de parênteses balanceados
- **Geração de Assembly**: Produção de código assembly x86-64
- **Runtime**: Funções de suporte para impressão e saída

## ✨ Recursos

- 🔢 **Números inteiros**: Suporte a constantes numéricas
- 🧮 **Operadores**: `+`, `-`, `*`, `**` (potência)
- 📐 **Parênteses**: Agrupamento de expressões
- 🔧 **Assembly x86-64**: Geração de código nativo
- 🐳 **Docker**: Ambiente containerizado
- 📊 **Debugging**: Visualização de tokens

## 📦 Pré-requisitos

### Desenvolvimento Local
- **Go 1.21+**: [Instalar Go](https://golang.org/doc/install)
- **GAS Assembler**: Parte do GNU Binutils
- **GNU Linker (ld)**: Para linking do executável
- **Make**: Para automação de build

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

2. **Compilar**:
```bash
# Local
make run INPUT_FILE=meu_programa.kite

# Docker
make docker-run INPUT_FILE=meu_programa.kite
```

3. **Montar e executar**:
```bash
# Gera o executável final
make assemble

# Executar
./executavel
```

### Fluxo Completo
```bash
# Compilar + Montar + Executar em um comando
make run-complete INPUT_FILE=meu_programa.kite

# Com Docker
make docker-run-complete INPUT_FILE=meu_programa.kite
```

### Linha de Comando Direta
```bash
# Depois do build
./kite-compiler meu_programa.kite

# Assembly gerado em: result/saida.s
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
make assemble
./executavel
```

**Saída esperada**: `123`

### Exemplo 2: Expressão com Parênteses
**Arquivo**: `exemplos/stage02/valido.kite`
```
(11 + 2)
```

**Compilação**:
```bash
make run INPUT_FILE=exemplos/stage02/valido.kite
```

**Nota**: Atualmente o compilador extrai apenas o primeiro número (11), mas tokeniza toda a expressão.

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

## 📁 Estrutura do Projeto

```
Kite/
├── cmd/compiler/main.go          # Ponto de entrada alternativo
├── exemplos/                     # Exemplos de código Kite
│   ├── stage01/                  # Números simples
│   └── stage02/                  # Expressões com parênteses
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
│   └── utils/                   # Utilitários
│       ├── error.go            # Sistema de erros
│       └── file.go             # Manipulação de arquivos
├── result/                      # Arquivos gerados
│   └── saida.s                 # Assembly gerado
├── Dockerfile                   # Container Docker
├── Makefile                     # Automação de build
├── go.mod                      # Dependências Go
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
3. **Geração de Código**: Edite `internal/compiler/generator.go`
4. **Testes**: Crie arquivos em `exemplos/`

### Debug e Análise

O compilador mostra informações detalhadas durante a execução:

```bash
make run INPUT_FILE=exemplos/stage02/valido.kite
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
4. **generator.go** → Gera código assembly x86-64
5. **runtime.s** → Fornece funções de runtime (impressão, saída)

### Componentes Principais

- **Lexer**: Análise léxica com regex patterns
- **Compiler**: Coordenação entre lexer e generator
- **Generator**: Template-based assembly generation
- **Runtime**: Assembly functions for I/O operations
- **Utils**: File I/O and error handling

### Estado Atual vs. Planejado

**✅ Implementado:**
- Tokenização completa de expressões matemáticas
- Validação de parênteses balanceados
- Geração básica de assembly
- Sistema de runtime funcional
- Suporte a Docker e Make

**🚧 Em Desenvolvimento:**
- Parser para análise sintática completa
- Avaliação de expressões matemáticas
- Suporte a variáveis e funções
- Otimizações de código

## 🤝 Contribuição

1. Fork o projeto
2. Crie uma branch para sua feature (`git checkout -b feature/nova-feature`)
3. Commit suas mudanças (`git commit -am 'Adiciona nova feature'`)
4. Push para a branch (`git push origin feature/nova-feature`)
5. Abra um Pull Request

## 📝 Licença

Este projeto está sob licença MIT. Veja o arquivo `LICENSE` para mais detalhes.

## 🐛 Problemas Conhecidos

- O compilador atualmente extrai apenas o primeiro número das expressões
- Operadores são tokenizados mas não processados
- Análise sintática está em desenvolvimento

## 📞 Suporte

Para dúvidas e problemas:
- Abra uma [Issue](https://github.com/khevencolino/Kite/issues)
- Consulte a documentação dos comandos: `make help`

---

**Desenvolvido com ❤️ em Go**
