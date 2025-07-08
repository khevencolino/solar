# comentario acima para evitar que o arquivo seja processado por go build

.text
.global _start

_start:
  # Calculando (8 * 7 * 6) / (5 * 4 * 3 * 2)

  # Numerador: 8 * 7 * 6
  mov $8, %rax
  mov $7, %rbx
  imul %rbx, %rax        # rax = 8 * 7 = 56
  mov $6, %rbx
  imul %rbx, %rax        # rax = 56 * 6 = 336
  mov %rax, %r11         # salva numerador em r11

  # Denominador: 5 * 4 * 3 * 2
  mov $5, %rax
  mov $4, %rbx
  imul %rbx, %rax        # rax = 5 * 4 = 20
  mov $3, %rbx
  imul %rbx, %rax        # rax = 20 * 3 = 60
  mov $2, %rbx
  imul %rbx, %rax        # rax = 60 * 2 = 120

  # Divisão: numerador / denominador
  mov %r11, %rax         # rax = 336 (numerador)
  mov %rax, %rdx         # salva numerador em rdx
  mov $120, %rbx         # rbx = 120 (denominador)

  # Preparar para divisão
  mov %rdx, %rax         # rax = numerador
  cqo                    # estende sinal para rdx:rax
  idiv %rbx              # rax = 336 / 120 = 2

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
