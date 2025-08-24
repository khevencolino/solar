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

// registrarFuncoesBasicas registra as funções builtin básicas
func (r *RegistroBuiltin) registrarFuncoesBasicas() {
	// Função imprime
	r.RegistrarFuncao("imprime", AssinaturaFuncao{
		Nome:           "imprime",
		MinArgumentos:  1,
		MaxArgumentos:  -1, // ilimitado
		TiposArgumento: []TipoArgumento{TIPO_INTEIRO},
		TipoFuncao:     FUNCAO_IMPRIME,
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

	// Função soma (exemplo de como adicionar novas funções)
	r.RegistrarFuncao("soma", AssinaturaFuncao{
		Nome:           "soma",
		MinArgumentos:  2,
		MaxArgumentos:  -1, // ilimitado
		TiposArgumento: []TipoArgumento{TIPO_INTEIRO},
		TipoFuncao:     FUNCAO_PURA,
		Descricao:      "Soma dois ou mais números",
	}, func(argumentos []interface{}) (interface{}, error) {
		resultado := 0
		for _, arg := range argumentos {
			if num, ok := arg.(int); ok {
				resultado += num
			}
		}
		return resultado, nil
	})

	// Função abs (valor absoluto)
	r.RegistrarFuncao("abs", AssinaturaFuncao{
		Nome:           "abs",
		MinArgumentos:  1,
		MaxArgumentos:  1,
		TiposArgumento: []TipoArgumento{TIPO_INTEIRO},
		TipoFuncao:     FUNCAO_PURA,
		Descricao:      "Retorna o valor absoluto de um número",
	}, func(argumentos []interface{}) (interface{}, error) {
		if num, ok := argumentos[0].(int); ok {
			if num < 0 {
				return -num, nil
			}
			return num, nil
		}
		return 0, fmt.Errorf("argumento deve ser um número")
	})
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

// ValidarChamada valida uma chamada de função
func (r *RegistroBuiltin) ValidarChamada(nome string, argumentos []interface{}) error {
	funcao, ok := r.funcoes[nome]
	if !ok {
		return fmt.Errorf("função '%s' não encontrada", nome)
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
	funcao, existe := r.funcoes[nome]
	if !existe {
		return nil, fmt.Errorf("função '%s' não está definida", nome)
	}

	return funcao.Executar(argumentos)
}

// Instância global do registro
var RegistroGlobal = NovoRegistroBuiltin()
