package compiler

import (
	"fmt"

	"github.com/khevencolino/Solar/internal/parser"
)

// TypeChecker realiza inferência e checagem real de tipos
type TypeChecker struct {
	scopes       []map[string]parser.Tipo
	funcs        map[string]*funcSig
	funcRetStack []parser.Tipo
	builtins     map[string]builtinSig
}

type funcSig struct {
	name   string
	params []parser.Tipo
	ret    parser.Tipo
}

type builtinSig struct {
	// Se varargs for true, usa params[0] como tipo do argumento repetido
	params  []parser.Tipo
	varargs bool
	minArgs int
	ret     parser.Tipo
	accept  func(parser.Tipo) bool
}

func NovoTypeChecker() *TypeChecker {
	tc := &TypeChecker{
		scopes:       []map[string]parser.Tipo{make(map[string]parser.Tipo)},
		funcs:        make(map[string]*funcSig),
		funcRetStack: []parser.Tipo{},
		builtins: map[string]builtinSig{
			"imprime": {params: []parser.Tipo{parser.TipoInteiro}, varargs: true, minArgs: 1, ret: parser.TipoInteiro, accept: func(tp parser.Tipo) bool {
				return tp == parser.TipoInteiro || tp == parser.TipoBooleano
			}},
			"soma": {params: []parser.Tipo{parser.TipoInteiro}, varargs: true, minArgs: 2, ret: parser.TipoInteiro},
			"abs":  {params: []parser.Tipo{parser.TipoInteiro}, varargs: false, minArgs: 1, ret: parser.TipoInteiro},
		},
	}
	return tc
}

func (t *TypeChecker) Check(stmts []parser.Expressao) error {
	// Primeira passada: coletar assinaturas de funções de nível superior
	for _, s := range stmts {
		if fn, ok := s.(*parser.FuncaoDeclaracao); ok {
			sig := &funcSig{name: fn.Nome, ret: fn.Retorno}
			sig.params = append(sig.params, fn.ParamTipos...)
			t.funcs[fn.Nome] = sig
		}
	}

	// Checar statements top-level
	for _, s := range stmts {
		if _, err := t.inferirExpr(s); err != nil {
			return err
		}
	}
	return nil
}

func (t *TypeChecker) pushScope() { t.scopes = append(t.scopes, make(map[string]parser.Tipo)) }
func (t *TypeChecker) popScope() {
	if len(t.scopes) > 1 {
		t.scopes = t.scopes[:len(t.scopes)-1]
	}
}

func (t *TypeChecker) getVar(nome string) (parser.Tipo, bool) {
	for i := len(t.scopes) - 1; i >= 0; i-- {
		if tp, ok := t.scopes[i][nome]; ok {
			return tp, true
		}
	}
	return 0, false
}

func (t *TypeChecker) setVar(nome string, tp parser.Tipo) {
	for i := len(t.scopes) - 1; i >= 0; i-- {
		if _, ok := t.scopes[i][nome]; ok {
			t.scopes[i][nome] = tp
			return
		}
	}
	t.scopes[len(t.scopes)-1][nome] = tp
}

// Inferência e checagem
func (t *TypeChecker) inferirExpr(e parser.Expressao) (parser.Tipo, error) {
	switch n := e.(type) {
	case *parser.Constante:
		return parser.TipoInteiro, nil
	case *parser.Booleano:
		return parser.TipoBooleano, nil

	case *parser.Variavel:
		if tp, ok := t.getVar(n.Nome); ok {
			return tp, nil
		}
		return 0, fmt.Errorf("variável '%s' não declarada", n.Nome)

	case *parser.Atribuicao:
		// Inferir tipo do valor
		vtp, err := t.inferirExpr(n.Valor)
		if err != nil {
			return 0, err
		}
		if n.TipoAnotado != nil {
			if !t.mesmoTipo(*n.TipoAnotado, vtp) {
				return 0, fmt.Errorf("atribuição incompatível: variável '%s' anotada como %s, valor é %s", n.Nome, n.TipoAnotado.String(), vtp.String())
			}
			t.setVar(n.Nome, *n.TipoAnotado)
			return *n.TipoAnotado, nil
		}
		// sem anotação: variável assume o tipo do valor
		t.setVar(n.Nome, vtp)
		return vtp, nil

	case *parser.OperacaoBinaria:
		lt, err := t.inferirExpr(n.OperandoEsquerdo)
		if err != nil {
			return 0, err
		}
		rt, err := t.inferirExpr(n.OperandoDireito)
		if err != nil {
			return 0, err
		}
		switch n.Operador {
		case parser.ADICAO, parser.SUBTRACAO, parser.MULTIPLICACAO, parser.DIVISAO, parser.POWER:
			if !(t.ehNumerico(lt) && t.ehNumerico(rt)) {
				return 0, fmt.Errorf("operador aritmético requer operandos numéricos, recebeu %s e %s", lt.String(), rt.String())
			}
			if !t.mesmoTipo(lt, rt) {
				return 0, fmt.Errorf("tipos incompatíveis: %s %s %s (sem coerção)", lt.String(), n.Operador.String(), rt.String())
			}
			return lt, nil
		case parser.IGUALDADE, parser.DIFERENCA, parser.MENOR_QUE, parser.MAIOR_QUE, parser.MENOR_IGUAL, parser.MAIOR_IGUAL:
			if n.Operador == parser.IGUALDADE || n.Operador == parser.DIFERENCA {
				if !t.mesmoTipo(lt, rt) {
					return 0, fmt.Errorf("comparação entre tipos incompatíveis: %s e %s", lt.String(), rt.String())
				}
				return parser.TipoBooleano, nil
			}
			if !t.mesmoTipo(lt, rt) || !t.ehNumerico(lt) {
				return 0, fmt.Errorf("comparação relacional requer tipos numéricos iguais, recebeu %s e %s", lt.String(), rt.String())
			}
			return parser.TipoBooleano, nil
		default:
			return 0, fmt.Errorf("operador desconhecido")
		}

	case *parser.ChamadaFuncao:
		// Função do usuário?
		if sig, ok := t.funcs[n.Nome]; ok {
			if len(n.Argumentos) != len(sig.params) {
				return 0, fmt.Errorf("função '%s' espera %d argumentos, recebeu %d", sig.name, len(sig.params), len(n.Argumentos))
			}
			for i, arg := range n.Argumentos {
				at, err := t.inferirExpr(arg)
				if err != nil {
					return 0, err
				}
				if !t.mesmoTipo(at, sig.params[i]) {
					return 0, fmt.Errorf("argumento %d de '%s' incompatível: esperado %s, recebeu %s", i+1, sig.name, sig.params[i].String(), at.String())
				}
			}
			return sig.ret, nil
		}
		// Builtin conhecido?
		if b, ok := t.builtins[n.Nome]; ok {
			if len(n.Argumentos) < b.minArgs {
				return 0, fmt.Errorf("função '%s' requer pelo menos %d argumento(s)", n.Nome, b.minArgs)
			}
			if b.varargs {
				for i, arg := range n.Argumentos {
					at, err := t.inferirExpr(arg)
					if err != nil {
						return 0, err
					}
					if b.accept != nil {
						if !b.accept(at) {
							return 0, fmt.Errorf("argumento %d de '%s' incompatível: recebido %s", i+1, n.Nome, at.String())
						}
					} else if !t.mesmoTipo(at, b.params[0]) {
						return 0, fmt.Errorf("argumento %d de '%s' incompatível: esperado %s, recebeu %s", i+1, n.Nome, b.params[0].String(), at.String())
					}
				}
			} else {
				if len(n.Argumentos) != len(b.params) {
					return 0, fmt.Errorf("função '%s' espera %d argumentos, recebeu %d", n.Nome, len(b.params), len(n.Argumentos))
				}
				for i, arg := range n.Argumentos {
					at, err := t.inferirExpr(arg)
					if err != nil {
						return 0, err
					}
					if b.accept != nil {
						if !b.accept(at) {
							return 0, fmt.Errorf("argumento %d de '%s' incompatível: recebido %s", i+1, n.Nome, at.String())
						}
					} else if !t.mesmoTipo(at, b.params[i]) {
						return 0, fmt.Errorf("argumento %d de '%s' incompatível: esperado %s, recebeu %s", i+1, n.Nome, b.params[i].String(), at.String())
					}
				}
			}
			return b.ret, nil
		}
		return 0, fmt.Errorf("função '%s' não encontrada", n.Nome)

	case *parser.ComandoSe:
		ct, err := t.inferirExpr(n.Condicao)
		if err != nil {
			return 0, err
		}
		if !(t.mesmoTipo(ct, parser.TipoBooleano) || t.mesmoTipo(ct, parser.TipoInteiro)) {
			return 0, fmt.Errorf("condição do 'se' deve ser booleano, recebeu %s", ct.String())
		}
		if _, err := t.inferirBloco(n.BlocoSe); err != nil {
			return 0, err
		}
		if n.BlocoSenao != nil {
			if _, err := t.inferirBloco(n.BlocoSenao); err != nil {
				return 0, err
			}
		}
		return parser.TipoVazio, nil

	case *parser.ComandoEnquanto:
		ct, err := t.inferirExpr(n.Condicao)
		if err != nil {
			return 0, err
		}
		if !(t.mesmoTipo(ct, parser.TipoBooleano) || t.mesmoTipo(ct, parser.TipoInteiro)) {
			return 0, fmt.Errorf("condição do 'enquanto' deve ser booleano, recebeu %s", ct.String())
		}
		if _, err := t.inferirBloco(n.Corpo); err != nil {
			return 0, err
		}
		return parser.TipoVazio, nil

	case *parser.ComandoPara:
		t.pushScope()
		defer t.popScope()
		if n.Inicializacao != nil {
			if _, err := t.inferirExpr(n.Inicializacao); err != nil {
				return 0, err
			}
		}
		if n.Condicao != nil {
			ct, err := t.inferirExpr(n.Condicao)
			if err != nil {
				return 0, err
			}
			if !(t.mesmoTipo(ct, parser.TipoBooleano) || t.mesmoTipo(ct, parser.TipoInteiro)) {
				return 0, fmt.Errorf("condição do 'para' deve ser booleano, recebeu %s", ct.String())
			}
		}
		if _, err := t.inferirBloco(n.Corpo); err != nil {
			return 0, err
		}
		if n.PosIteracao != nil {
			if _, err := t.inferirExpr(n.PosIteracao); err != nil {
				return 0, err
			}
		}
		return parser.TipoVazio, nil

	case *parser.Bloco:
		return t.inferirBloco(n)

	case *parser.FuncaoDeclaracao:
		return t.checkFuncDecl(n)

	case *parser.Retorno:
		if len(t.funcRetStack) == 0 {
			return 0, fmt.Errorf("'retornar' só é permitido dentro de funções")
		}
		declRet := t.funcRetStack[len(t.funcRetStack)-1]
		if n.Valor == nil {
			if !t.mesmoTipo(declRet, parser.TipoVazio) {
				return 0, fmt.Errorf("retorno vazio incompatível: função declara retorno %s", declRet.String())
			}
			return parser.TipoVazio, nil
		}
		vt, err := t.inferirExpr(n.Valor)
		if err != nil {
			return 0, err
		}
		if !t.mesmoTipo(vt, declRet) {
			return 0, fmt.Errorf("tipo de retorno incompatível: esperado %s, recebeu %s", declRet.String(), vt.String())
		}
		return parser.TipoVazio, nil

	default:
		return 0, fmt.Errorf("nó não suportado na checagem de tipos")
	}
}

func (t *TypeChecker) inferirBloco(b *parser.Bloco) (parser.Tipo, error) {
	t.pushScope()
	defer t.popScope()
	lastType := parser.TipoVazio
	for _, cmd := range b.Comandos {
		tp, err := t.inferirExpr(cmd)
		if err != nil {
			return 0, err
		}
		lastType = tp
	}
	return lastType, nil
}

func (t *TypeChecker) checkFuncDecl(fn *parser.FuncaoDeclaracao) (parser.Tipo, error) {
	t.funcRetStack = append(t.funcRetStack, fn.Retorno)
	defer func() { t.funcRetStack = t.funcRetStack[:len(t.funcRetStack)-1] }()

	t.pushScope()
	for i, nome := range fn.Parametros {
		t.setVar(nome, fn.ParamTipos[i])
	}
	lastType, err := t.inferirBloco(fn.Corpo)
	if err != nil {
		t.popScope()
		return 0, err
	}
	t.popScope()

	// Se a função declara retorno não-vazio, deve haver um retorno compatível no final
	if fn.Retorno != parser.TipoVazio {
		if !t.hasReturnInBlock(fn.Corpo) {
			if !t.mesmoTipo(lastType, fn.Retorno) {
				return 0, fmt.Errorf("retorno implícito incompatível na função '%s': esperado %s, obteve %s", fn.Nome, fn.Retorno.String(), lastType.String())
			}
		}
	}
	return parser.TipoVazio, nil
}

func (t *TypeChecker) mesmoTipo(a, b parser.Tipo) bool { return a == b }
func (t *TypeChecker) ehNumerico(tp parser.Tipo) bool {
	return tp == parser.TipoInteiro || tp == parser.TipoDecimal
}

func (t *TypeChecker) hasReturnInBlock(b *parser.Bloco) bool {
	for _, cmd := range b.Comandos {
		switch n := cmd.(type) {
		case *parser.Retorno:
			return true
		case *parser.Bloco:
			if t.hasReturnInBlock(n) {
				return true
			}
		case *parser.ComandoSe:
			if t.hasReturnInBlock(n.BlocoSe) {
				return true
			}
			if n.BlocoSenao != nil && t.hasReturnInBlock(n.BlocoSenao) {
				return true
			}
		case *parser.ComandoEnquanto:
			if t.hasReturnInBlock(n.Corpo) {
				return true
			}
		case *parser.ComandoPara:
			if t.hasReturnInBlock(n.Corpo) {
				return true
			}
		}
	}
	return false
}
