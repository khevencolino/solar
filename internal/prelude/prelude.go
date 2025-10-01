package prelude

import "fmt"

// Prelude contém símbolos sempre disponíveis (sem import)
type Prelude struct {
	funcoes map[string]*FuncaoPrelude
}

type FuncaoPrelude struct {
	Nome        string
	Implementar func(args []interface{}) interface{}
	Varargs     bool
	MinArgs     int
	MaxArgs     int // -1 para ilimitado
}

// NewPrelude cria o prelude padrão com funções básicas
func NewPrelude() *Prelude {
	p := &Prelude{
		funcoes: make(map[string]*FuncaoPrelude),
	}

	// Adiciona funções básicas sempre disponíveis
	p.funcoes["imprime"] = &FuncaoPrelude{
		Nome:    "imprime",
		Varargs: true,
		MinArgs: 1,
		MaxArgs: -1,
		Implementar: func(args []interface{}) interface{} {
			for i, arg := range args {
				if i > 0 {
					fmt.Print(" ")
				}
				// Formata diferentes tipos apropriadamente
				switch v := arg.(type) {
				case int:
					fmt.Print(v)
				case float64:
					fmt.Printf("%g", v)
				case string:
					fmt.Print(v)
				case bool:
					if v {
						fmt.Print("verdadeiro")
					} else {
						fmt.Print("falso")
					}
				default:
					fmt.Print(v)
				}
			}
			fmt.Println()
			return 0
		},
	}

	return p
}

// EhFuncaoPrelude verifica se uma função está no prelude
func (p *Prelude) EhFuncaoPrelude(nome string) bool {
	_, existe := p.funcoes[nome]
	return existe
}

// ExecutarFuncaoPrelude executa uma função do prelude
func (p *Prelude) ExecutarFuncaoPrelude(nome string, args []interface{}) interface{} {
	if fn, ok := p.funcoes[nome]; ok {
		return fn.Implementar(args)
	}
	return fmt.Errorf("função '%s' não encontrada no prelude", nome)
}

// ObterFuncaoPrelude retorna os metadados de uma função do prelude
func (p *Prelude) ObterFuncaoPrelude(nome string) (*FuncaoPrelude, bool) {
	fn, ok := p.funcoes[nome]
	return fn, ok
}
