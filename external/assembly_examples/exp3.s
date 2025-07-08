# comentario acima para evitar que o arquivo seja processado por go build

.text
.global _start

_start:
  # Calculando (42 - 111) * 11 + (17 * 77)

  # Primeira parte: 42 - 111
  mov $42, %rax
  mov $111, %rbx
  sub %rbx, %rax         # rax = 42 - 111 = -69

  # Multiplicação: (42 - 111) * 11
  mov $11, %rbx
  imul %rbx, %rax        # rax = -69 * 11 = -759
  mov %rax, %r11         # salva resultado em r11

  # Segunda parte: 17 * 77
  mov $17, %rax
  mov $77, %rbx
  imul %rbx, %rax        # rax = 17 * 77 = 1309

  # Soma final: (42 - 111) * 11 + (17 * 77)
  add %r11, %rax         # rax = -759 + 1309 = 550

  call imprime_num
  call sair

#
# funcoes de apoio para o codigo compilado
#
imprime_num:
  xor %r9, %r9            # rcx indice, r9 contagem
  mov $20, %rcx
  movb $10, buffer(%rcx)  # \n no final da string
  dec %rcx
  inc %r9
  mov $10, %r8
  or %rax, %rax
  jz printzero_L0
  jl mark_neg
  mov $0, %r10            # r10 flag p/ negativo
  jmp loop_L0
mark_neg:
  mov $1, %r10
  neg %rax
loop_L0:
  cqo
  idiv %r8
  addb $0x30, %dl
  movb %dl, buffer(%rcx)
  dec %rcx
  inc %r9
  or %rax, %rax
  jnz loop_L0
  test %r10, %r10
  jz print_L0
  movb $45, buffer(%rcx)
  dec %rcx
  jmp print_L0
printzero_L0:
  movb $0x30, buffer(%rcx)
  dec %rcx
  inc %r9
print_L0:
  mov $1, %rax            # sys_write
  mov $1, %rdi            # stdout
  mov $buffer, %rsi       # dados
  inc %rcx
  add %rcx, %rsi
  mov %r9, %rdx           # tamanho
  syscall
  ret
sair:
  mov $60, %rax     # sys_exit
  xor %rdi, %rdi    # codigo de saida (0)
  syscall
  .section .bss
  .lcomm buffer, 21
