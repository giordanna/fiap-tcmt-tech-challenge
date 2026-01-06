# estágio de build
FROM golang:1.21 AS builder

WORKDIR /app

# cache de modulos
COPY go.mod go.sum ./
RUN go mod download

# compilação estática
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o api-gateway ./main.go

# estágio final (imagem leve)
FROM alpine:latest

WORKDIR /root/
COPY --from=builder /app/api-gateway .

# certificados para chamadas https externas
RUN apk --no-cache add ca-certificates

EXPOSE 8080
CMD ["./api-gateway"]