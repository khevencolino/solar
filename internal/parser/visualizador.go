package parser

import (
	"fmt"
	"strconv"

	"github.com/m1gwings/treedrawer/tree"
)

// VisualizadorArvore cria representações visuais da AST
type VisualizadorArvore struct{}

// NovoVisualizador cria um novo visualizador
func NovoVisualizador() *VisualizadorArvore {
	return &VisualizadorArvore{}
}

// CriarArvore converte a AST para o formato do treedrawer
func (v *VisualizadorArvore) CriarArvore(expressao Expressao) *tree.Tree {
	switch expr := expressao.(type) {
	case *Constante:
		// Cria árvore com apenas um nó (constante)
		return tree.NewTree(tree.NodeString(strconv.Itoa(expr.Valor)))

	case *OperacaoBinaria:
		// Cria árvore com operador como raiz
		arvore := tree.NewTree(tree.NodeString(expr.Operador.String()))

		// Adiciona filhos usando AddChild
		arvoreEsquerda := v.CriarArvore(expr.OperandoEsquerdo)
		arvoreDireita := v.CriarArvore(expr.OperandoDireito)

		// AddChild retorna ponteiro para o filho adicionado
		arvore.AddChild(arvoreEsquerda.Val()) // Adiciona o valor da árvore esquerda
		arvore.AddChild(arvoreDireita.Val())  // Adiciona o valor da árvore direita

		return arvore

	default:
		return tree.NewTree(tree.NodeString("?"))
	}
}

// ImprimirArvore imprime a árvore no console
func (v *VisualizadorArvore) ImprimirArvore(expressao Expressao) {
	fmt.Println("=== Árvore Sintática ===")
	arvore := v.criarArvoreRecursiva(expressao)
	fmt.Println(arvore)
	fmt.Println()
}

// criarArvoreRecursiva cria a árvore de forma recursiva
func (v *VisualizadorArvore) criarArvoreRecursiva(expressao Expressao) *tree.Tree {
	switch expr := expressao.(type) {
	case *Constante:
		// Folha da árvore: apenas o número
		return tree.NewTree(tree.NodeString(strconv.Itoa(expr.Valor)))

	case *OperacaoBinaria:
		// Nó interno: operador com dois filhos
		arvore := tree.NewTree(tree.NodeString(expr.Operador.String()))

		// Cria subárvores para os operandos
		subarvoreEsquerda := v.criarArvoreRecursiva(expr.OperandoEsquerdo)
		subarvoreDireita := v.criarArvoreRecursiva(expr.OperandoDireito)

		// Adiciona os valores das subárvores como filhos
		v.adicionarSubarvore(arvore, subarvoreEsquerda)
		v.adicionarSubarvore(arvore, subarvoreDireita)

		return arvore

	default:
		return tree.NewTree(tree.NodeString("ERRO"))
	}
}

// adicionarSubarvore adiciona uma subárvore como filho
func (v *VisualizadorArvore) adicionarSubarvore(pai *tree.Tree, filho *tree.Tree) {
	// Adiciona o valor do nó raiz do filho
	novoFilho := pai.AddChild(filho.Val())

	// Se o filho tem seus próprios filhos, adiciona recursivamente
	v.copiarFilhos(filho, novoFilho)
}

// copiarFilhos copia todos os filhos de uma árvore para outra
func (v *VisualizadorArvore) copiarFilhos(origem *tree.Tree, destino *tree.Tree) {
	// Percorre todos os filhos da origem
	for i := 0; ; i++ {
		filho, err := origem.Child(i)
		if err != nil {
			break // Não há mais filhos
		}

		// Adiciona o filho ao destino
		novoFilho := destino.AddChild(filho.Val())

		// Recursivamente copia os filhos do filho
		v.copiarFilhos(filho, novoFilho)
	}
}
