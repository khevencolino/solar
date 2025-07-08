package compiler

import "fmt"

// Generator manipula a geração de código assembly
type Generator struct {
	template string // Template do código assembly
}

// NovoGerador cria um novo gerador de código
func NovoGerador() *Generator {
	return &Generator{
		template: `.section .text
.globl _start
_start:
	mov $%s, %%rax
	call imprime_num
	call sair
.include "runtime.s"
`,
	}
}

// GerarAssembly gera código assembly para uma constante
func (g *Generator) GerarAssembly(constante string) string {
	return fmt.Sprintf(g.template, constante)
}
