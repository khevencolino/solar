package compiler

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/khevencolino/Solar/internal/lexer"
	"github.com/khevencolino/Solar/internal/parser"
)

// ModuleResolver gerencia resolução e carregamento de módulos
type ModuleResolver struct {
	// Cache de módulos já carregados para evitar recarregamento
	modulosCarregados map[string]*ModuloCarregado
	// Caminhos onde buscar módulos
	caminhosBusca []string
}

// ModuloCarregado representa um módulo já processado
type ModuloCarregado struct {
	Nome     string
	Caminho  string
	AST      []parser.Expressao
	Simbolos map[string]*SimboloExportado // funções/variáveis exportadas
}

// SimboloExportado representa uma função ou variável exportada de um módulo
type SimboloExportado struct {
	Nome string
	Tipo TipoSimbolo
	Node parser.Expressao // O nó AST correspondente
}

type TipoSimbolo int

const (
	SIMBOLO_FUNCAO TipoSimbolo = iota
	SIMBOLO_VARIAVEL
	SIMBOLO_BUILTIN // símbolos especiais implementados pelo compilador
)

// NewModuleResolver cria um novo resolvedor de módulos
func NewModuleResolver() *ModuleResolver {
	return &ModuleResolver{
		modulosCarregados: make(map[string]*ModuloCarregado),
		caminhosBusca: []string{
			".",         // diretório atual
			"./stdlib",  // biblioteca padrão local
			"./pacotes", // pacotes do usuário
		},
	}
}

// ResolverModulo encontra e carrega um módulo pelo nome
func (mr *ModuleResolver) ResolverModulo(nomeModulo string) (*ModuloCarregado, error) {
	// Verifica se já foi carregado
	if modulo, existe := mr.modulosCarregados[nomeModulo]; existe {
		return modulo, nil
	}

	// Procura o arquivo do módulo
	caminhoArquivo, err := mr.encontrarArquivoModulo(nomeModulo)
	if err != nil {
		return nil, err
	}

	// Carrega e analisa o arquivo
	modulo, err := mr.carregarModulo(nomeModulo, caminhoArquivo)
	if err != nil {
		return nil, err
	}

	// Cache do módulo
	mr.modulosCarregados[nomeModulo] = modulo
	return modulo, nil
}

// encontrarArquivoModulo procura o arquivo correspondente ao módulo
func (mr *ModuleResolver) encontrarArquivoModulo(nomeModulo string) (string, error) {
	// Possíveis extensões de arquivo
	extensoes := []string{".solar", ".sl"}

	for _, caminhoBusca := range mr.caminhosBusca {
		for _, extensao := range extensoes {
			// Tenta nome.extensao
			caminhoCompleto := filepath.Join(caminhoBusca, nomeModulo+extensao)
			if _, err := os.Stat(caminhoCompleto); err == nil {
				return caminhoCompleto, nil
			}

			// Tenta nome/nome.extensao (para pacotes)
			caminhoCompleto = filepath.Join(caminhoBusca, nomeModulo, nomeModulo+extensao)
			if _, err := os.Stat(caminhoCompleto); err == nil {
				return caminhoCompleto, nil
			}

			// Tenta nome/index.extensao
			caminhoCompleto = filepath.Join(caminhoBusca, nomeModulo, "index"+extensao)
			if _, err := os.Stat(caminhoCompleto); err == nil {
				return caminhoCompleto, nil
			}
		}
	}

	return "", fmt.Errorf("módulo '%s' não encontrado nos caminhos de busca: %v", nomeModulo, mr.caminhosBusca)
}

// carregarModulo lê e analisa um arquivo de módulo
func (mr *ModuleResolver) carregarModulo(nomeModulo, caminhoArquivo string) (*ModuloCarregado, error) {
	// Lê o conteúdo do arquivo
	conteudo, err := os.ReadFile(caminhoArquivo)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler módulo '%s': %v", caminhoArquivo, err)
	}

	// Tokeniza o código
	l := lexer.NovoLexer(string(conteudo))
	tokens, err := l.Tokenizar()
	if err != nil {
		return nil, fmt.Errorf("erro ao tokenizar módulo '%s': %v", nomeModulo, err)
	}

	// Analisa sintáticamente
	p := parser.NovoParser(tokens)
	ast, err := p.AnalisarPrograma()
	if err != nil {
		return nil, fmt.Errorf("erro ao analisar módulo '%s': %v", nomeModulo, err)
	}

	// Extrai símbolos exportados
	simbolos := mr.extrairSimbolosExportados(ast)

	// Adiciona símbolos built-in especiais para módulos da biblioteca padrão
	mr.adicionarSimbolosBuiltin(nomeModulo, simbolos)

	return &ModuloCarregado{
		Nome:     nomeModulo,
		Caminho:  caminhoArquivo,
		AST:      ast,
		Simbolos: simbolos,
	}, nil
}

// extrairSimbolosExportados identifica funções e variáveis que podem ser exportadas
func (mr *ModuleResolver) extrairSimbolosExportados(ast []parser.Expressao) map[string]*SimboloExportado {
	simbolos := make(map[string]*SimboloExportado)

	for _, expr := range ast {
		switch node := expr.(type) {
		case *parser.FuncaoDeclaracao:
			simbolos[node.Nome] = &SimboloExportado{
				Nome: node.Nome,
				Tipo: SIMBOLO_FUNCAO,
				Node: expr,
			}
		case *parser.Atribuicao:
			// Variáveis globais podem ser exportadas
			simbolos[node.Nome] = &SimboloExportado{
				Nome: node.Nome,
				Tipo: SIMBOLO_VARIAVEL,
				Node: expr,
			}
		}
	}

	return simbolos
}

// adicionarSimbolosBuiltin adiciona símbolos built-in especiais para módulos da stdlib
func (mr *ModuleResolver) adicionarSimbolosBuiltin(nomeModulo string, simbolos map[string]*SimboloExportado) {
	switch nomeModulo {
	case "io":
		// Função imprime é implementada como built-in pelo compilador
		simbolos["imprime"] = &SimboloExportado{
			Nome: "imprime",
			Tipo: SIMBOLO_BUILTIN,
			Node: nil, // Não há nó AST para funções built-in
		}
	}
}

// ResolverSimbolo encontra um símbolo específico em um módulo
func (mr *ModuleResolver) ResolverSimbolo(nomeModulo, nomeSimbolo string) (*SimboloExportado, error) {
	modulo, err := mr.ResolverModulo(nomeModulo)
	if err != nil {
		return nil, err
	}

	simbolo, existe := modulo.Simbolos[nomeSimbolo]
	if !existe {
		return nil, fmt.Errorf("símbolo '%s' não encontrado no módulo '%s'", nomeSimbolo, nomeModulo)
	}

	return simbolo, nil
}

// ListarSimbolosModulo retorna todos os símbolos disponíveis em um módulo
func (mr *ModuleResolver) ListarSimbolosModulo(nomeModulo string) ([]string, error) {
	modulo, err := mr.ResolverModulo(nomeModulo)
	if err != nil {
		return nil, err
	}

	var nomes []string
	for nome := range modulo.Simbolos {
		nomes = append(nomes, nome)
	}

	return nomes, nil
}

// AdicionarCaminhoBusca adiciona um novo caminho para busca de módulos
func (mr *ModuleResolver) AdicionarCaminhoBusca(caminho string) {
	for _, existente := range mr.caminhosBusca {
		if existente == caminho {
			return // já existe
		}
	}
	mr.caminhosBusca = append(mr.caminhosBusca, caminho)
}
