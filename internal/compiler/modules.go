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
	// Caminho do arquivo fonte atual (para resolver imports relativos)
	arquivoFonteAtual string
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

// ResolverModulo encontra e carrega um módulo pelo nome ou caminho
func (mr *ModuleResolver) ResolverModulo(nomeModulo string) (*ModuloCarregado, error) {
	// Normaliza o caminho para usar como chave no cache
	caminhoNormalizado, _ := filepath.Abs(nomeModulo)
	cacheKey := caminhoNormalizado

	// Verifica se já foi carregado
	if modulo, existe := mr.modulosCarregados[cacheKey]; existe {
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

	// Cache do módulo usando caminho normalizado
	mr.modulosCarregados[cacheKey] = modulo
	return modulo, nil
}

// encontrarArquivoModulo procura o arquivo correspondente ao módulo
func (mr *ModuleResolver) encontrarArquivoModulo(nomeModulo string) (string, error) {
	// Possíveis extensões de arquivo
	extensoes := []string{".solar", ".sl"}

	// Verifica se é um caminho explícito (relativo ou absoluto)
	if mr.ehCaminhoExplicito(nomeModulo) {
		return mr.resolverCaminhoExplicito(nomeModulo, extensoes)
	}

	// Caso contrário, procura nos caminhos de busca padrão
	return mr.procurarNoCaminhosBusca(nomeModulo, extensoes)
}

// ehCaminhoExplicito verifica se o nome do módulo é um caminho explícito
func (mr *ModuleResolver) ehCaminhoExplicito(nomeModulo string) bool {
	// Caminho absoluto (começa com / no Unix ou C:\ no Windows)
	if filepath.IsAbs(nomeModulo) {
		return true
	}

	// Caminho relativo (começa com ./ ou ../)
	if len(nomeModulo) >= 2 {
		if nomeModulo[0:2] == "./" || nomeModulo[0:2] == ".." {
			return true
		}
	}

	return false
}

// resolverCaminhoExplicito resolve um caminho explícito (relativo ou absoluto)
func (mr *ModuleResolver) resolverCaminhoExplicito(caminho string, extensoes []string) (string, error) {
	// Se é um caminho relativo, resolve em relação ao arquivo fonte atual
	if !filepath.IsAbs(caminho) && mr.arquivoFonteAtual != "" {
		dirFonte := filepath.Dir(mr.arquivoFonteAtual)
		caminho = filepath.Join(dirFonte, caminho)
	}

	// Tenta o caminho exato primeiro
	if _, err := os.Stat(caminho); err == nil {
		return filepath.Abs(caminho)
	}

	// Tenta adicionar extensões
	for _, extensao := range extensoes {
		caminhoComExt := caminho
		if filepath.Ext(caminho) == "" {
			caminhoComExt = caminho + extensao
		}
		if _, err := os.Stat(caminhoComExt); err == nil {
			return filepath.Abs(caminhoComExt)
		}
	}

	// Tenta nome/nome.extensao (para pacotes)
	base := filepath.Base(caminho)
	for _, extensao := range extensoes {
		caminhoCompleto := filepath.Join(caminho, base+extensao)
		if _, err := os.Stat(caminhoCompleto); err == nil {
			return filepath.Abs(caminhoCompleto)
		}
	}

	// Tenta nome/index.extensao
	for _, extensao := range extensoes {
		caminhoCompleto := filepath.Join(caminho, "index"+extensao)
		if _, err := os.Stat(caminhoCompleto); err == nil {
			return filepath.Abs(caminhoCompleto)
		}
	}

	return "", fmt.Errorf("arquivo não encontrado: %s (tentativas com extensões %v)", caminho, extensoes)
}

// procurarNoCaminhosBusca procura o módulo nos caminhos de busca configurados
func (mr *ModuleResolver) procurarNoCaminhosBusca(nomeModulo string, extensoes []string) (string, error) {
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

// SetArquivoFonteAtual define o arquivo fonte atual para resolução de imports relativos
func (mr *ModuleResolver) SetArquivoFonteAtual(caminho string) {
	mr.arquivoFonteAtual = caminho
}
