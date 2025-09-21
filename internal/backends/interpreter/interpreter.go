package interpreter

import (
	"fmt"
	"math"

	"github.com/khevencolino/Solar/internal/debug"
	"github.com/khevencolino/Solar/internal/lexer"
	"github.com/khevencolino/Solar/internal/parser"
	"github.com/khevencolino/Solar/internal/registry"
	"github.com/khevencolino/Solar/internal/utils"
)

type InterpreterBackend struct {
	variaveis map[string]int
	funcoes   map[string]*parser.FuncaoDeclaracao
}

func NewInterpreterBackend() *InterpreterBackend {
	return &InterpreterBackend{
		variaveis: make(map[string]int),
		funcoes:   make(map[string]*parser.FuncaoDeclaracao),
	}
}

func (i *InterpreterBackend) GetName() string      { return "Interpretador AST" }
func (i *InterpreterBackend) GetExtension() string { return "" }

func (i *InterpreterBackend) Compile(statements []parser.Expressao) error {
	debug.Printf("Interpretando diretamente da AST...\n")

	var ultimoResultado interface{}

	// Primeira passada: registrar declarações de funções para permitir chamadas antes da definição
	for _, stmt := range statements {
		if fn, ok := stmt.(*parser.FuncaoDeclaracao); ok {
			i.funcoes[fn.Nome] = fn
		}
	}

	for idx, stmt := range statements {
		debug.Printf("--- Statement %d ---\n", idx+1)

		// Imprime a árvore (opcional - pode ser configurável)
		visualizador := parser.NovoVisualizador()
		visualizador.ImprimirArvore(stmt)

		// Interpreta
		resultado, err := i.interpretar(stmt)
		if err != nil {
			return err
		}

		fmt.Printf("Resultado: %d\n", resultado)
		ultimoResultado = resultado
	}

	debug.Printf("\n Interpretação concluída! Resultado final: %d\n", ultimoResultado)
	return nil
}

// Interpretar executa uma expressão e retorna o resultado
func (i *InterpreterBackend) interpretar(expressao parser.Expressao) (interface{}, error) {
	resultado := expressao.Aceitar(i)
	if erro, ok := resultado.(error); ok {
		return 0, erro
	}
	return resultado, nil
}

// Implementa interface Node (visitor pattern)
func (i *InterpreterBackend) Constante(constante *parser.Constante) interface{} {
	return constante.Valor
}

func (i *InterpreterBackend) Booleano(b *parser.Booleano) interface{} {
	if b.Valor {
		return 1
	}
	return 0
}

func (i *InterpreterBackend) Variavel(variavel *parser.Variavel) interface{} {
	valor, existe := i.variaveis[variavel.Nome]
	if !existe {
		return utils.NovoErro(
			fmt.Sprintf("variável '%s' não definida", variavel.Nome),
			variavel.Token.Position.Line,
			variavel.Token.Position.Column,
			"",
		)
	}
	return valor
}

func (i *InterpreterBackend) Atribuicao(atribuicao *parser.Atribuicao) interface{} {
	// Avalia o valor da expressão
	valorInterface := atribuicao.Valor.Aceitar(i)
	if erro, ok := valorInterface.(error); ok {
		return erro
	}
	valor := valorInterface.(int)

	// Armazena na tabela de símbolos
	i.variaveis[atribuicao.Nome] = valor
	return valor
}

func (i *InterpreterBackend) OperacaoBinaria(operacao *parser.OperacaoBinaria) interface{} {
	// Interpreta operando esquerdo
	esquerdoInterface := operacao.OperandoEsquerdo.Aceitar(i)
	if erro, ok := esquerdoInterface.(error); ok {
		return erro
	}
	esquerdo := esquerdoInterface.(int)

	// Interpreta operando direito
	direitoInterface := operacao.OperandoDireito.Aceitar(i)
	if erro, ok := direitoInterface.(error); ok {
		return erro
	}
	direito := direitoInterface.(int)

	// Executa operação
	switch operacao.Operador {
	case parser.ADICAO:
		return esquerdo + direito
	case parser.SUBTRACAO:
		return esquerdo - direito
	case parser.MULTIPLICACAO:
		return esquerdo * direito
	case parser.DIVISAO:
		if direito == 0 {
			return utils.NovoErro(
				"divisão por zero",
				operacao.Token.Position.Line,
				operacao.Token.Position.Column,
				"",
			)
		}
		return esquerdo / direito
	case parser.POWER:
		return int(math.Pow(float64(esquerdo), float64(direito)))

	// Operações de comparação
	case parser.IGUALDADE:
		if esquerdo == direito {
			return 1
		}
		return 0
	case parser.DIFERENCA:
		if esquerdo != direito {
			return 1
		}
		return 0
	case parser.MENOR_QUE:
		if esquerdo < direito {
			return 1
		}
		return 0
	case parser.MAIOR_QUE:
		if esquerdo > direito {
			return 1
		}
		return 0
	case parser.MENOR_IGUAL:
		if esquerdo <= direito {
			return 1
		}
		return 0
	case parser.MAIOR_IGUAL:
		if esquerdo >= direito {
			return 1
		}
		return 0
	default:
		return utils.NovoErro(
			"operador desconhecido",
			operacao.Token.Position.Line,
			operacao.Token.Position.Column,
			"",
		)
	}
}

// ChamadaFuncao implementa chamadas de função builtin
func (i *InterpreterBackend) ChamadaFuncao(chamada *parser.ChamadaFuncao) interface{} {
	// Primeiro verifica se é uma função definida pelo usuário
	if fn, ok := i.funcoes[chamada.Nome]; ok {
		return i.executarFuncaoUsuario(fn, chamada)
	}

	// Caso contrário, tenta como builtin
	assinatura, ok := registry.RegistroGlobal.ObterAssinatura(chamada.Nome)
	if !ok {
		return utils.NovoErro(
			"função desconhecida",
			chamada.Token.Position.Line,
			chamada.Token.Position.Column,
			fmt.Sprintf("Função '%s' não encontrada", chamada.Nome),
		)
	}

	// Avalia todos os argumentos primeiro
	argumentos := make([]interface{}, len(chamada.Argumentos))
	for idx, argumento := range chamada.Argumentos {
		valorInterface := argumento.Aceitar(i)
		if erro, ok := valorInterface.(error); ok {
			return erro
		}
		argumentos[idx] = valorInterface
	}

	// Valida argumentos (apenas verificação de quantidade em tempo de compilação)
	numArgs := len(chamada.Argumentos)
	if numArgs < assinatura.MinArgumentos {
		return utils.NovoErro(
			"erro na função",
			chamada.Token.Position.Line,
			chamada.Token.Position.Column,
			fmt.Sprintf("Função '%s' requer pelo menos %d argumentos, mas recebeu %d",
				chamada.Nome, assinatura.MinArgumentos, numArgs),
		)
	}
	if assinatura.MaxArgumentos != -1 && numArgs > assinatura.MaxArgumentos {
		return utils.NovoErro(
			"erro na função",
			chamada.Token.Position.Line,
			chamada.Token.Position.Column,
			fmt.Sprintf("Função '%s' aceita no máximo %d argumentos, mas recebeu %d",
				chamada.Nome, assinatura.MaxArgumentos, numArgs),
		)
	}

	// Executa baseado no tipo da função
	return i.executarFuncaoBuiltin(chamada.Nome, assinatura.TipoFuncao, argumentos, chamada.Token.Position)
}

// executarFuncaoBuiltin executa uma função builtin baseada no seu tipo
func (i *InterpreterBackend) executarFuncaoBuiltin(nome string, tipo registry.TipoFuncao, args []interface{}, pos lexer.Position) interface{} {
	switch tipo {
	case registry.FUNCAO_IMPRIME:
		return i.executarImprime(args)
	case registry.FUNCAO_PURA:
		// Para funções puras, use a implementação do registro
		resultado, err := registry.RegistroGlobal.ExecutarFuncao(nome, args)
		if err != nil {
			return utils.NovoErro(
				"erro na função",
				pos.Line,
				pos.Column,
				err.Error(),
			)
		}
		return resultado
	default:
		return utils.NovoErro(
			"erro na função",
			pos.Line,
			pos.Column,
			fmt.Sprintf("Tipo de função não suportado: %v", tipo),
		)
	}
}

// executarImprime implementa a função imprime específica do interpretador
func (i *InterpreterBackend) executarImprime(args []interface{}) interface{} {
	for idx, arg := range args {
		if idx > 0 {
			fmt.Print(" ")
		}
		fmt.Print(arg)
	}
	fmt.Println()
	return 0
}

// ComandoSe implementa o comando if/else
func (i *InterpreterBackend) ComandoSe(comando *parser.ComandoSe) interface{} {
	// Avalia a condição
	condicaoInterface := comando.Condicao.Aceitar(i)
	if erro, ok := condicaoInterface.(error); ok {
		return erro
	}
	condicao := condicaoInterface.(int)

	// Se a condição for verdadeira (não zero), executa o bloco "se"
	if condicao != 0 {
		return comando.BlocoSe.Aceitar(i)
	} else if comando.BlocoSenao != nil {
		// Se há bloco "senao" e a condição é falsa, executa o bloco "senao"
		return comando.BlocoSenao.Aceitar(i)
	}

	// Se não há bloco "senao" e a condição é falsa, retorna 0
	return 0
}

// Enquanto (while)
func (i *InterpreterBackend) ComandoEnquanto(cmd *parser.ComandoEnquanto) interface{} {
	var ultimo interface{} = 0
	for {
		c := cmd.Condicao.Aceitar(i)
		if erro, ok := c.(error); ok {
			return erro
		}
		if c.(int) == 0 {
			break
		}
		r := cmd.Corpo.Aceitar(i)
		if erro, ok := r.(error); ok {
			return erro
		}
		if rv, ok := r.(retornoValor); ok {
			return rv
		}
		ultimo = r
	}
	return ultimo
}

// Para (for)
func (i *InterpreterBackend) ComandoPara(cmd *parser.ComandoPara) interface{} {
	if cmd.Inicializacao != nil {
		v := cmd.Inicializacao.Aceitar(i)
		if erro, ok := v.(error); ok {
			return erro
		}
	}
	var ultimo interface{} = 0
	for {
		// condicao vazia => verdadeiro
		cond := 1
		if cmd.Condicao != nil {
			c := cmd.Condicao.Aceitar(i)
			if erro, ok := c.(error); ok {
				return erro
			}
			cond = c.(int)
		}
		if cond == 0 {
			break
		}
		r := cmd.Corpo.Aceitar(i)
		if erro, ok := r.(error); ok {
			return erro
		}
		if rv, ok := r.(retornoValor); ok {
			return rv
		}
		ultimo = r
		if cmd.PosIteracao != nil {
			p := cmd.PosIteracao.Aceitar(i)
			if erro, ok := p.(error); ok {
				return erro
			}
		}
	}
	return ultimo
}

// Bloco implementa um bloco de comandos
func (i *InterpreterBackend) Bloco(bloco *parser.Bloco) interface{} {
	var ultimoResultado interface{} = 0

	// Executa todos os comandos do bloco
	for _, comando := range bloco.Comandos {
		// Se for um retorno, propaga uma interrupção especial
		if ret, ok := comando.(*parser.Retorno); ok {
			if ret.Valor == nil {
				return retornoValor{valor: 0}
			}
			val := ret.Valor.Aceitar(i)
			if erro, ok := val.(error); ok {
				return erro
			}
			return retornoValor{valor: val.(int)}
		}

		resultado := comando.Aceitar(i)
		if erro, ok := resultado.(error); ok {
			return erro
		}
		// Se um bloco interno retornou, propaga
		if rv, ok := resultado.(retornoValor); ok {
			return rv
		}
		ultimoResultado = resultado
	}

	return ultimoResultado
}

// Suporte a declaração de função do usuário
func (i *InterpreterBackend) FuncaoDeclaracao(fn *parser.FuncaoDeclaracao) interface{} {
	i.funcoes[fn.Nome] = fn
	return 0
}

// Suporte a retorno (em nível de execução de bloco)
func (i *InterpreterBackend) Retorno(ret *parser.Retorno) interface{} {
	if ret.Valor == nil {
		return retornoValor{valor: 0}
	}
	val := ret.Valor.Aceitar(i)
	if erro, ok := val.(error); ok {
		return erro
	}
	return retornoValor{valor: val.(int)}
}

// Estrutura para propagar retorno através do visitor
type retornoValor struct{ valor int }

// Executa função definida pelo usuário com escopo local
func (i *InterpreterBackend) executarFuncaoUsuario(fn *parser.FuncaoDeclaracao, chamada *parser.ChamadaFuncao) interface{} {
	// Verifica argumentos
	if len(chamada.Argumentos) != len(fn.Parametros) {
		return utils.NovoErro(
			"erro na função",
			chamada.Token.Position.Line,
			chamada.Token.Position.Column,
			fmt.Sprintf("Função '%s' espera %d argumento(s), recebeu %d", fn.Nome, len(fn.Parametros), len(chamada.Argumentos)),
		)
	}

	// Salva contexto de variáveis e cria escopo local
	antigo := i.variaveis
	local := make(map[string]int)
	i.variaveis = local

	// Avalia e vincula parâmetros
	for idx, param := range fn.Parametros {
		v := chamada.Argumentos[idx].Aceitar(i)
		if erro, ok := v.(error); ok {
			i.variaveis = antigo
			return erro
		}
		local[param] = v.(int)
	}

	// Executa corpo
	resultado := fn.Corpo.Aceitar(i)

	// Restaura escopo
	i.variaveis = antigo

	// Trata retorno explícito
	if rv, ok := resultado.(retornoValor); ok {
		return rv.valor
	}

	// Retorno implícito: valor da última expressão do bloco (0 se vazio)
	switch r := resultado.(type) {
	case int:
		return r
	default:
		return 0
	}
}
