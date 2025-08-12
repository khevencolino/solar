// +build ignore
// comentario acima para evitar que o arquivo seja processado por go build

// funcoes de apoio para o codigo compilado

imprime_num:
    mov x11, #0            // x11: character count
    mov x10, #20           // x10: buffer index
    adrp x9, buffer         // x9: buffer address

    mov w12, #10
    strb w12, [x9, x10]
    sub x10, x10, #1
    add x11, x11, #1

    mov x12, #10           // x12: divisor
    cmp x0, #0
    b.eq printzero_arm64

    mov x13, #0            // x13: negative flag
    cmp x0, #0
    b.ge loop_arm64

    mov x13, #1
    neg x0, x0

loop_arm64:
    udiv x2, x0, x12
    msub x1, x0, x2, x12
    add x1, x1, #'0'
    strb w1, [x9, x10]
    sub x10, x10, #1
    add x11, x11, #1
    mov x0, x2
    cmp x0, #0
    b.ne loop_arm64

    cmp x13, #1
    b.eq mark_neg_arm64
    b print_arm64

mark_neg_arm64:
    mov w1, #'-'
    strb w1, [x9, x10]
    sub x10, x10, #1
    add x11, x11, #1

printzero_arm64:
    mov w1, #'0'
    strb w1, [x9, x10]
    sub x10, x10, #1
    add x11, x11, #1

print_arm64:
    add x10, x10, #1

    mov x8, #64            // sys_write
    mov x0, #1             // stdout
    add x1, x9, x10        // buffer address
    mov x2, x11            // length
    svc #0
    ret

sair:
    mov x8, #93            // sys_exit
    mov x0, #0             // exit code
    svc #0


.comm buffer, 21
