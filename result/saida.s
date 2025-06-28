.section .text
.globl _start

_start:
	mov $123, %rax

	call imprime_num
	call sair

.include "runtime.s"
