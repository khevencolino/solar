package builtin

import (
	"fmt"

	"github.com/khevencolino/Solar/internal/parser"
)

// TipoArgumento define os tipos de argumentos aceitos
type TipoArgumento int

const (
	TIPO_INTEIRO TipoArgumento = iota
	TIPO_QUALQUER
)

// AssinaturaFuncao define a assinatura de uma função builtin
type AssinaturaFuncao struct {
	Nome           string
	MinArgumentos  int
	MaxArgumentos  int // -1 para ilimitado
	TiposArgumento []TipoArgumento
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

// registrarFuncoesBasicas registra as funções builtin básicas
func (r *RegistroBuiltin) registrarFuncoesBasicas() {
	// Função imprime
	r.RegistrarFuncao("imprime", AssinaturaFuncao{
		Nome:           "imprime",
		MinArgumentos:  1,
		MaxArgumentos:  -1, // ilimitado
		TiposArgumento: []TipoArgumento{TIPO_INTEIRO},
		Descricao:      "Imprime um ou mais valores separados por espaço",
	}, func(argumentos []interface{}) (interface{}, error) {
		for i, arg := range argumentos {
			if i > 0 {
				fmt.Print(" ")
			}
			fmt.Print(arg)
		}
		fmt.Println()
		return 0, nil
	})
}

// RegistrarFuncao adiciona uma nova função builtin ao registro
func (r *RegistroBuiltin) RegistrarFuncao(nome string, assinatura AssinaturaFuncao, executar func([]interface{}) (interface{}, error)) {
	r.funcoes[nome] = &FuncaoBuiltin{
		Assinatura: assinatura,
		Executar:   executar,
	}
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

// ValidarChamada valida uma chamada de função
func (r *RegistroBuiltin) ValidarChamada(chamada *parser.ChamadaFuncao) error {
	funcao, existe := r.funcoes[chamada.Nome]
	if !existe {
		return fmt.Errorf("função '%s' não está definida", chamada.Nome)
	}

	numArgs := len(chamada.Argumentos)

	// Verifica número mínimo de argumentos
	if numArgs < funcao.Assinatura.MinArgumentos {
		return fmt.Errorf("função '%s' espera pelo menos %d argumentos, recebeu %d",
			chamada.Nome, funcao.Assinatura.MinArgumentos, numArgs)
	}

	// Verifica número máximo de argumentos (se não for ilimitado)
	if funcao.Assinatura.MaxArgumentos != -1 && numArgs > funcao.Assinatura.MaxArgumentos {
		return fmt.Errorf("função '%s' espera no máximo %d argumentos, recebeu %d",
			chamada.Nome, funcao.Assinatura.MaxArgumentos, numArgs)
	}

	return nil
}

// ExecutarFuncao executa uma função builtin
func (r *RegistroBuiltin) ExecutarFuncao(nome string, argumentos []interface{}) (interface{}, error) {
	funcao, existe := r.funcoes[nome]
	if !existe {
		return nil, fmt.Errorf("função '%s' não está definida", nome)
	}

	return funcao.Executar(argumentos)
}

// Instância global do registro
var RegistroGlobal = NovoRegistroBuiltin()
