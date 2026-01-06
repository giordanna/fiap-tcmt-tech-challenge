package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	casodeuso "github.com/giordanna/fiap-tcmt-tech-challenge/backend/interno/casos_de_uso"
	"github.com/giordanna/fiap-tcmt-tech-challenge/backend/interno/handlers"
	"github.com/giordanna/fiap-tcmt-tech-challenge/backend/interno/repositorio"
	"github.com/giordanna/fiap-tcmt-tech-challenge/backend/pkg/logger"
)

func main() {
	// Inicializa logger estruturado
	logger.InitLogger()
	slog.Info("Iniciando API de Recomendações...")

	// Carrega variáveis de ambiente (ignora erro se .env não existir)
	_ = godotenv.Load()

	// Configuração do banco de dados
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "fiap")
	dbPassword := getEnv("DB_PASSWORD", "fiap123")
	dbName := getEnv("DB_NAME", "tech_challenge")
	apiPort := getEnv("API_PORT", "8080")

	// String de conexão PostgreSQL
	// Suporta tanto Unix socket (Cloud SQL) quanto TCP (desenvolvimento local)
	var dsn string
	if strings.HasPrefix(dbHost, "/cloudsql/") {
		// Cloud Run com Cloud SQL (Unix socket)
		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
			dbHost, dbUser, dbPassword, dbName)
		slog.Info("Usando conexão Unix socket para Cloud SQL", "socket", dbHost)
	} else {
		// Desenvolvimento local (TCP)
		dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			dbHost, dbPort, dbUser, dbPassword, dbName)
		slog.Info("Usando conexão TCP", "host", dbHost, "port", dbPort)
	}

	// Conecta ao banco de dados
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		slog.Error("Erro ao abrir conexão com banco de dados", "erro", err)
		os.Exit(1)
	}
	defer db.Close()

	// Verifica se a conexão está funcionando
	if err := db.Ping(); err != nil {
		slog.Error("Erro ao conectar ao banco de dados", "erro", err)
		os.Exit(1)
	}

	slog.Info("Conexão com banco de dados estabelecida com sucesso")

	// Configuração do pool de conexões
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Injeção de dependências (DI)
	repo := repositorio.NovoRepositorioPostgres(db)
	servico := casodeuso.NovoServicoRecomendacao(repo)
	handler := handlers.NovoHandlerRecomendacoes(servico)

	// Configuração do Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Middlewares
	router.Use(gin.Recovery())
	router.Use(loggerMiddleware())

	// Rotas
	router.GET("/healthcheck", handler.HealthCheck)
	router.GET("/recomendacoes/:clienteId", handler.BuscarRecomendacoes)
	router.POST("/recomendacoes/:clienteId", handler.GerarRecomendacoes)

	// Configuração do servidor HTTP
	srv := &http.Server{
		Addr:         ":" + apiPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Inicia servidor em goroutine separada
	go func() {
		slog.Info("Servidor HTTP iniciado", "porta", apiPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Erro ao iniciar servidor", "erro", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Desligando servidor...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Erro ao desligar servidor", "erro", err)
	}

	slog.Info("Servidor desligado com sucesso")
}

// getEnv retorna o valor da variável de ambiente ou um valor padrão
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// loggerMiddleware adiciona logging para cada requisição HTTP
func loggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		duration := time.Since(start)
		statusCode := c.Writer.Status()

		slog.Info("HTTP Request",
			"method", method,
			"path", path,
			"status", statusCode,
			"duration_ms", duration.Milliseconds(),
			"client_ip", c.ClientIP(),
		)
	}
}
