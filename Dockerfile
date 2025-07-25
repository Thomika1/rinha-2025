# Estágio 1: Build
FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Entra na pasta que contém o main.go
WORKDIR /app/cmd

# Compila o projeto. O binário será criado em /app/cmd/api
RUN go build -o /app/cmd/api .

# Estágio 2: Final
FROM scratch

# Copia o binário /app/cmd/api do estágio anterior para a raiz da nova imagem com o nome "api"
COPY --from=builder /app/cmd/api /api

# Expõe a porta que a aplicação realmente escuta (ajuste se necessário)
EXPOSE 8080

# Define o comando para executar o binário
CMD ["/api"]