package interpreter

import (
	"fmt"
	"math"
	"strconv"

	"github.com/khevencolino/Solar/internal/debug"
	"github.com/khevencolino/Solar/internal/lexer"
	"github.com/khevencolino/Solar/internal/parser"
	"github.com/khevencolino/Solar/internal/prelude"
	"github.com/khevencolino/Solar/internal/registry"
	"github.com/khevencolino/Solar/internal/utils"
)

// Valor representa um valor tipado no interpretador
type Valor struct {
	Tipo  parser.Tipo
	Dados interface{}
}

type InterpreterBackend struct {
	variaveis map[string]Valor
	funcoes   map[string]*parser.FuncaoDeclaracao
	prelude   *prelude.Prelude
}

func NewInterpreterBackend() *InterpreterBackend {
	return &InterpreterBackend{
		variaveis: make(map[string]Valor),
		funcoes:   make(map[string]*parser.FuncaoDeclaracao),
		prelude:   prelude.NewPrelude(),
	}
}

func (i *InterpreterBackend) GetName() string      { return "Interpretador AST" }
func (i *InterpreterBackend) GetExtension() string { return "" }

func (i *InterpreterBackend) Compile(statements []parser.Expressao) error {
	debug.Printf("Interpretando diretamente da AST...\n")

	var ultimoResultado interface{}

	// Primeira passada: registrar declarações de funções para permitir chamadas antes da definição
	var funcaoPrincipal *parser.FuncaoDeclaracao
	for _, stmt := range statements {
		if fn, ok := stmt.(*parser.FuncaoDeclaracao); ok {
			i.funcoes[fn.Nome] = fn
			if fn.Nome == "principal" {
				funcaoPrincipal = fn
			}
		}
	}

	// Se existe função principal(), chama ela. Senão, executa statements globais
	if funcaoPrincipal != nil {
		debug.Printf("--- Executando função principal() ---\n")

		// Imprime a árvore da função principal (opcional)
		if debug.Enabled {
			fmt.Printf("\nÁrvore da função principal:\n")
			visualizador := parser.NovoVisualizador()
			visualizador.ImprimirArvore(funcaoPrincipal)
		}

		// Executa a função principal
		resultado, err := i.interpretar(funcaoPrincipal.Corpo)
		if err != nil {
			return err
		}
		ultimoResultado = resultado
	} else {
		// Executa statements globais (comportamento antigo, vou manter por compatibilidade, ou remover dps)
		for idx, stmt := range statements {
			// Pula declarações de função pois já foram processadas
			if _, ok := stmt.(*parser.FuncaoDeclaracao); ok {
				continue
			}

			debug.Printf("--- Statement global %d ---\n", idx+1)

			// Imprime a árvore (opcional)
			if debug.Enabled {
				fmt.Printf("\nÁrvore da expressão:\n")
				visualizador := parser.NovoVisualizador()
				visualizador.ImprimirArvore(stmt)
			}

			// Interpreta
			resultado, err := i.interpretar(stmt)
			if err != nil {
				return err
			}

			ultimoResultado = resultado
		}
	}

	debug.Printf("\n Interpretação concluída! Resultado final: %v\n", ultimoResultado)
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
	return b.Valor
}

func (i *InterpreterBackend) LiteralTexto(literal *parser.LiteralTexto) interface{} {
	return literal.Valor
}

func (i *InterpreterBackend) LiteralDecimal(literal *parser.LiteralDecimal) interface{} {
	return literal.Valor
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
	return valor.Dados
}

func (i *InterpreterBackend) Atribuicao(atribuicao *parser.Atribuicao) interface{} {
	// Avalia o valor da expressão (pode ser int, bool, float64, string)
	valorInterface := atribuicao.Valor.Aceitar(i)
	if erro, ok := valorInterface.(error); ok {
		return erro
	}

	// Detecta tipo dinamicamente (até ter tipagem estática mais forte aqui)
	switch v := valorInterface.(type) {
	case int:
		i.variaveis[atribuicao.Nome] = Valor{Tipo: parser.TipoInteiro, Dados: v}
	case bool:
		i.variaveis[atribuicao.Nome] = Valor{Tipo: parser.TipoBooleano, Dados: v}
	case float64:
		i.variaveis[atribuicao.Nome] = Valor{Tipo: parser.TipoDecimal, Dados: v}
	case string:
		i.variaveis[atribuicao.Nome] = Valor{Tipo: parser.TipoTexto, Dados: v}
	default:
		return utils.NovoErro(
			"tipo de valor não suportado na atribuição",
			atribuicao.Token.Position.Line,
			atribuicao.Token.Position.Column,
			fmt.Sprintf("Tipo: %T", valorInterface),
		)
	}
	return valorInterface
}

func (i *InterpreterBackend) OperacaoBinaria(operacao *parser.OperacaoBinaria) interface{} {
	// Avalia operandos com helper
	esqVal, err := i.evaluateOperand(operacao.OperandoEsquerdo)
	if err != nil {
		return err
	}
	dirVal, err2 := i.evaluateOperand(operacao.OperandoDireito)
	if err2 != nil {
		return err2
	}

	switch operacao.Operador {
	case parser.ADICAO:
		return esqVal + dirVal
	case parser.SUBTRACAO:
		return esqVal - dirVal
	case parser.MULTIPLICACAO:
		return esqVal * dirVal
	case parser.DIVISAO:
		if dirVal == 0 {
			return utils.NovoErro("divisão por zero", operacao.Token.Position.Line, operacao.Token.Position.Column, "")
		}
		return esqVal / dirVal
	case parser.POWER:
		return int(math.Pow(float64(esqVal), float64(dirVal)))
	case parser.IGUALDADE:
		return i.compareInts(esqVal == dirVal)
	case parser.DIFERENCA:
		return i.compareInts(esqVal != dirVal)
	case parser.MENOR_QUE:
		return i.compareInts(esqVal < dirVal)
	case parser.MAIOR_QUE:
		return i.compareInts(esqVal > dirVal)
	case parser.MENOR_IGUAL:
		return i.compareInts(esqVal <= dirVal)
	case parser.MAIOR_IGUAL:
		return i.compareInts(esqVal >= dirVal)
	default:
		return utils.NovoErro("operador desconhecido", operacao.Token.Position.Line, operacao.Token.Position.Column, "")
	}
}

// evaluateOperand avalia uma expressão e garante retorno int (simplifiquei, talvez altero na prox)
func (i *InterpreterBackend) evaluateOperand(expr parser.Expressao) (int, error) {
	v := expr.Aceitar(i)
	if erro, ok := v.(error); ok {
		return 0, erro
	}
	switch val := v.(type) {
	case int:
		return val, nil
	case bool:
		// Permite usar booleano em contexto numérico (true=1,false=0)
		if val {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, utils.NovoErro(
			"operando não numérico",
			0, 0,
			fmt.Sprintf("Tipo: %T", v),
		)
	}
}

// compareInts converte bool para 1/0
func (i *InterpreterBackend) compareInts(cond bool) int {
	if cond {
		return 1
	}
	return 0
}

// ChamadaFuncao implementa chamadas de função builtin
func (i *InterpreterBackend) ChamadaFuncao(chamada *parser.ChamadaFuncao) interface{} {
	// 1. Verifica prelude primeiro (funções sempre disponíveis)
	if i.prelude.EhFuncaoPrelude(chamada.Nome) {
		// Avalia argumentos
		argumentos := make([]interface{}, len(chamada.Argumentos))
		for idx, argumento := range chamada.Argumentos {
			valorInterface := argumento.Aceitar(i)
			if erro, ok := valorInterface.(error); ok {
				return erro
			}
			argumentos[idx] = valorInterface
		}
		return i.prelude.ExecutarFuncaoPrelude(chamada.Nome, argumentos)
	}

	// 2. Verifica se é uma função definida pelo usuário
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
	// Suporte a múltiplos tipos básicos: inteiro, booleano, decimal, texto
	for idx, arg := range args {
		if idx > 0 {
			fmt.Print(" ")
		}
		fmt.Print(i.formatarValor(arg))
	}
	fmt.Println()
	return 0 // retorno neutro para chamadas encadeadas
}

// formatarValor converte um valor para sua representação textual idiomática em Solar
func (i *InterpreterBackend) formatarValor(v interface{}) string {
	switch val := v.(type) {
	case int:
		return fmt.Sprintf("%d", val)
	case bool:
		if val {
			return "verdadeiro"
		}
		return "falso"
	case float64:
		return strconv.FormatFloat(val, 'g', -1, 64)
	case string:
		return val
	default:
		return fmt.Sprintf("%v", val)
	}
}

// ComandoSe implementa o comando if/else
func (i *InterpreterBackend) ComandoSe(comando *parser.ComandoSe) interface{} {
	condRaw := comando.Condicao.Aceitar(i)
	if erro, ok := condRaw.(error); ok {
		return erro
	}
	verdade := i.isTruthy(condRaw)
	if verdade {
		return comando.BlocoSe.Aceitar(i)
	}
	if comando.BlocoSenao != nil {
		return comando.BlocoSenao.Aceitar(i)
	}
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
		if !i.isTruthy(c) {
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
		condOk := true
		if cmd.Condicao != nil {
			c := cmd.Condicao.Aceitar(i)
			if erro, ok := c.(error); ok {
				return erro
			}
			condOk = i.isTruthy(c)
		}
		if !condOk {
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
			// Converte retorno não-int para int por enquanto (bool -> 0/1) até ter suporte tipado
			switch x := val.(type) {
			case int:
				return retornoValor{valor: x}
			case bool:
				if x {
					return retornoValor{valor: 1}
				}
				return retornoValor{valor: 0}
			default:
				return retornoValor{valor: 0}
			}
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
	switch x := val.(type) {
	case int:
		return retornoValor{valor: x}
	case bool:
		if x {
			return retornoValor{valor: 1}
		}
		return retornoValor{valor: 0}
	default:
		return retornoValor{valor: 0}
	}
}

// Suporte a importação (processadas antes da interpretação)
func (i *InterpreterBackend) Importacao(imp *parser.Importacao) interface{} {
	// Imports já foram processados pelo compilador antes de chegar aqui
	// Não precisamos fazer nada específico no interpreter
	return 0
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
	local := make(map[string]Valor)
	// Copia variáveis do escopo anterior para permitir acesso a funções globais
	for k, v := range i.variaveis {
		local[k] = v
	}
	i.variaveis = local

	// Avalia e vincula parâmetros (no contexto do chamador, não do chamado)
	for idx, param := range fn.Parametros {
		// Temporariamente restaura o contexto anterior para avaliar argumentos
		i.variaveis = antigo
		v := chamada.Argumentos[idx].Aceitar(i)
		i.variaveis = local

		if erro, ok := v.(error); ok {
			i.variaveis = antigo
			return erro
		}
		// Armazena dinamicamente conforme tipo recebido
		switch x := v.(type) {
		case int:
			local[param.Nome] = Valor{Tipo: parser.TipoInteiro, Dados: x}
		case bool:
			local[param.Nome] = Valor{Tipo: parser.TipoBooleano, Dados: x}
		case float64:
			local[param.Nome] = Valor{Tipo: parser.TipoDecimal, Dados: x}
		case string:
			local[param.Nome] = Valor{Tipo: parser.TipoTexto, Dados: x}
		default:
			local[param.Nome] = Valor{Tipo: parser.TipoVazio, Dados: 0}
		}
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

// isTruthy aplica a regra de verdade: int != 0, bool == valor, outros => falso
func (i *InterpreterBackend) isTruthy(v interface{}) bool {
	switch x := v.(type) {
	case int:
		return x != 0
	case bool:
		return x
	default:
		return false
	}
}
