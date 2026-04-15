# Etapa de construção
FROM golang:1.22.0-alpine AS builder

# Instalando o git
RUN apk add --no-cache git

# Diretório de trabalho dentro do container
WORKDIR /app

# Configurando credenciais privadas para acesso ao github
ARG GO_GIT_CRED__HTTPS__GITHUB__COM
ARG GO_PRIVATE

RUN git config --global url."https://$GO_GIT_CRED__HTTPS__GITHUB__COM@github.com/".insteadOf "https://github.com/"
RUN go env -w GOPRIVATE=$GO_PRIVATE
RUN go env -w GOPROXY=https://proxy.golang.org,direct

# Copiando apenas os arquivos de dependências primeiro para aproveitar o cache
COPY go.mod go.sum ./

# Baixando dependências
RUN go mod download

# Copiando o resto do código
COPY . .

# Compilando o projeto
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Etapa de execução
FROM alpine:latest
WORKDIR /root/

# Copiando o binário compilado do builder para o container principal
COPY --from=builder /app/main .

# Executando o binário
CMD ["./main"]
