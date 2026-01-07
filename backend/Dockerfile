# estágio de build
FROM golang:1.25-alpine AS builder

WORKDIR /app

# cache de modulos
COPY go.mod go.sum ./
RUN go mod download

# compilação estática com otimizações
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o api-gateway ./main.go

# estágio final (imagem leve)
FROM alpine:3.23

# Instala certificados e timezone
RUN apk --no-cache add ca-certificates tzdata

# Configura timezone padrão
ENV TZ=America/Sao_Paulo

# Cria usuário não-root para segurança
RUN adduser -D -g '' appuser
USER appuser

WORKDIR /app
COPY --from=builder /app/api-gateway .

EXPOSE 8080
CMD ["./api-gateway"]