package registry

import (
	"fmt"
)

// TipoArgumento define os tipos de argumentos aceitos
type TipoArgumento int

const (
	TIPO_INTEIRO TipoArgumento = iota
	TIPO_QUALQUER
)

// TipoFuncao define como a função se comporta
type TipoFuncao int

const (
	FUNCAO_IMPRIME TipoFuncao = iota // Função que imprime (tem efeito colateral)
	FUNCAO_PURA                      // Função que só retorna valor
)

// AssinaturaFuncao define a assinatura de uma função builtin
type AssinaturaFuncao struct {
	Nome           string
	MinArgumentos  int
	MaxArgumentos  int // -1 para ilimitado
	TiposArgumento []TipoArgumento
	TipoFuncao     TipoFuncao
	Descricao      string
}

// FuncaoBuiltin representa uma função builtin
type FuncaoBuiltin struct {
	Assinatura AssinaturaFuncao
	Executar   func(argumentos []interface{}) (interface{}, error)
}

// RegistroBuiltin mantém todas as funções builtin registradas
type RegistroBuiltin struct {
	funcoes map[string]*FuncaoBuiltin
}

// NovoRegistroBuiltin cria um novo registro de funções builtin
func NovoRegistroBuiltin() *RegistroBuiltin {
	registro := &RegistroBuiltin{
		funcoes: make(map[string]*FuncaoBuiltin),
	}

	// Registra funções builtin padrão
	registro.registrarFuncoesBasicas()

	return registro
}

// Definições de funções builtin (otimizadas - definidas como constantes)
var funcoesBuiltinPadroes = map[string]FuncaoBuiltin{
	"imprime": {
		Assinatura: AssinaturaFuncao{
			Nome:           "imprime",
			MinArgumentos:  1,
			MaxArgumentos:  -1, // ilimitado
			TiposArgumento: []TipoArgumento{TIPO_QUALQUER},
			TipoFuncao:     FUNCAO_IMPRIME,
			Descricao:      "Imprime valores na saída padrão",
		},
		Executar: func(argumentos []interface{}) (interface{}, error) {
			return nil, nil
		},
	},
}

// registrarFuncoesBasicas registra as funções builtin básicas de forma otimizada
func (r *RegistroBuiltin) registrarFuncoesBasicas() {
	// Copia as funções pré-definidas para o registro
	for nome, funcao := range funcoesBuiltinPadroes {
		r.funcoes[nome] = &FuncaoBuiltin{
			Assinatura: funcao.Assinatura,
			Executar:   funcao.Executar,
		}
	}
}

// RegistrarFuncao adiciona uma nova função builtin ao registro
func (r *RegistroBuiltin) RegistrarFuncao(nome string, assinatura AssinaturaFuncao, executar func([]interface{}) (interface{}, error)) {
	r.funcoes[nome] = &FuncaoBuiltin{
		Assinatura: assinatura,
		Executar:   executar,
	}
}

// ObterAssinatura retorna a assinatura de uma função
func (r *RegistroBuiltin) ObterAssinatura(nome string) (AssinaturaFuncao, bool) {
	funcao, ok := r.funcoes[nome]
	if !ok {
		return AssinaturaFuncao{}, false
	}
	return funcao.Assinatura, true
}

// obterFuncaoValidada retorna uma função validada ou erro
func (r *RegistroBuiltin) obterFuncaoValidada(nome string) (*FuncaoBuiltin, error) {
	funcao, ok := r.funcoes[nome]
	if !ok {
		return nil, fmt.Errorf("função '%s' não encontrada", nome)
	}
	return funcao, nil
}

// ValidarChamada valida uma chamada de função
func (r *RegistroBuiltin) ValidarChamada(nome string, argumentos []interface{}) error {
	funcao, err := r.obterFuncaoValidada(nome)
	if err != nil {
		return err
	}

	assinatura := funcao.Assinatura
	numArgs := len(argumentos)

	// Verifica número mínimo de argumentos
	if numArgs < assinatura.MinArgumentos {
		return fmt.Errorf("função '%s' requer pelo menos %d argumentos, mas recebeu %d",
			nome, assinatura.MinArgumentos, numArgs)
	}

	// Verifica número máximo de argumentos (se limitado)
	if assinatura.MaxArgumentos != -1 && numArgs > assinatura.MaxArgumentos {
		return fmt.Errorf("função '%s' aceita no máximo %d argumentos, mas recebeu %d",
			nome, assinatura.MaxArgumentos, numArgs)
	}

	return nil
}

// EhFuncaoBuiltin verifica se um nome é uma função builtin
func (r *RegistroBuiltin) EhFuncaoBuiltin(nome string) bool {
	_, existe := r.funcoes[nome]
	return existe
}

// ObterFuncao retorna uma função builtin pelo nome
func (r *RegistroBuiltin) ObterFuncao(nome string) (*FuncaoBuiltin, bool) {
	funcao, existe := r.funcoes[nome]
	return funcao, existe
}

// ListarFuncoes retorna uma lista de todas as funções builtin
func (r *RegistroBuiltin) ListarFuncoes() []string {
	nomes := make([]string, 0, len(r.funcoes))
	for nome := range r.funcoes {
		nomes = append(nomes, nome)
	}
	return nomes
}

// ExecutarFuncao executa uma função builtin
func (r *RegistroBuiltin) ExecutarFuncao(nome string, argumentos []interface{}) (interface{}, error) {
	funcao, err := r.obterFuncaoValidada(nome)
	if err != nil {
		return nil, err
	}

	return funcao.Executar(argumentos)
}

// Instância global do registro
var RegistroGlobal = NovoRegistroBuiltin()
