# Etapa de construção
FROM golang:1.19 as builder

# Diretório de trabalho dentro do container
WORKDIR /app

# Copiando o código do Go para o Container
COPY . .

# Configurando credenciais privadas para acesso ao github
ARG GO_GIT_CRED__HTTPS__GITHUB__COM
ARG GO_PRIVATE

RUN echo "machine github.com login $GO_GIT_CRED__HTTPS__GITHUB__COM password x-oauth-basic" > ~/.netrc

ENV GOPRIVATE=$GO_PRIVATE

# Baixando dependências
RUN go mod download

# Compilando o projeto
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Etapa de execução
FROM alpine:latest

WORKDIR /root/

# Copiando o binário compilado do builder para o container principal
COPY --from=builder /app/main .

# Executando o binário
CMD ["./main"]
