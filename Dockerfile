# Dockerfile para Kite Compiler
FROM golang:1.21-alpine AS builder

# Instala dependências necessárias
RUN apk add --no-cache gcc musl-dev binutils

# Define diretório de trabalho
WORKDIR /app

# Copia arquivos do projeto
COPY go.mod go.sum ./
RUN go mod download

# Copia código fonte
COPY . .

# Compila o compilador
RUN go build -ldflags="-s -w" -o kite-compiler ./main.go

# Imagem final
FROM alpine:latest

# Instala ferramentas necessárias para assembly
RUN apk add --no-cache gcc musl-dev binutils

# Define diretório de trabalho
WORKDIR /workspace

# Copia o compilador da imagem builder
COPY --from=builder /app/kite-compiler .

# Define usuário não-root
RUN addgroup -g 1000 kite && adduser -D -u 1000 -G kite kite
USER kite
