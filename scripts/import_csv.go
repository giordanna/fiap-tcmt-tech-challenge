package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	log.Println("=== Importador de Dados CSV para PostgreSQL ===")

	// Carrega variáveis de ambiente
	if err := godotenv.Load("../.env"); err != nil {
		log.Println("Aviso: arquivo .env não encontrado, usando variáveis de ambiente do sistema")
	}

	// Configuração do banco de dados
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "fiap")
	dbPassword := getEnv("DB_PASSWORD", "fiap123")
	dbName := getEnv("DB_NAME", "tech_challenge")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	log.Printf("Conectando ao banco de dados em %s:%s...", dbHost, dbPort)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Erro ao abrir conexão: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Erro ao conectar ao banco: %v", err)
	}
	log.Println("✓ Conectado ao banco de dados")

	// Diretório dos CSVs (projeto original)
	csvDir := "../fiap-tcmt-tech-challenge-main/scripts"

	// Importa cada arquivo CSV
	if err := importarClientes(db, filepath.Join(csvDir, "clientes.csv")); err != nil {
		log.Printf("Erro ao importar clientes: %v", err)
	}

	if err := importarProdutos(db, filepath.Join(csvDir, "produtos.csv")); err != nil {
		log.Printf("Erro ao importar produtos: %v", err)
	}

	if err := importarTransacoes(db, filepath.Join(csvDir, "transacoes.csv")); err != nil {
		log.Printf("Erro ao importar transações: %v", err)
	}

	if err := importarInteracoes(db, filepath.Join(csvDir, "interacoes.csv")); err != nil {
		log.Printf("Erro ao importar interações: %v", err)
	}

	log.Println("\n=== Importação concluída! ===")
}

func importarClientes(db *sql.DB, caminho string) error {
	log.Printf("\nImportando clientes de %s...", caminho)

	file, err := os.Open(caminho)
	if err != nil {
		return fmt.Errorf("erro ao abrir arquivo: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	header, err := reader.Read() // Pula cabeçalho
	if err != nil {
		return err
	}
	log.Printf("Cabeçalho: %v", header)

	count := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Erro ao ler linha: %v", err)
			continue
		}

		// id_cliente, nome_cliente, data_cadastro, idade, genero, renda_mensal_estimada,
		// patrimonio_total_estimado, perfil_risco, objetivo_investimento, ultima_interacao
		patrimonio, _ := strconv.ParseFloat(record[6], 64)

		query := `INSERT INTO clientes (id_cliente, nome_cliente, perfil_risco, patrimonio_total_estimado) 
		          VALUES ($1, $2, $3, $4) ON CONFLICT (id_cliente) DO NOTHING`

		_, err = db.Exec(query, record[0], record[1], record[7], patrimonio)
		if err != nil {
			log.Printf("Erro ao inserir cliente %s: %v", record[0], err)
			continue
		}
		count++
	}

	log.Printf("✓ %d clientes importados", count)
	return nil
}

func importarProdutos(db *sql.DB, caminho string) error {
	log.Printf("\nImportando produtos de %s...", caminho)

	file, err := os.Open(caminho)
	if err != nil {
		return fmt.Errorf("erro ao abrir arquivo: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	header, err := reader.Read()
	if err != nil {
		return err
	}
	log.Printf("Cabeçalho: %v", header)

	count := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Erro ao ler linha: %v", err)
			continue
		}

		// id_produto, nome_produto, tipo_produto, risco_associado, rentabilidade_historica_12m,
		// rentabilidade_historica_36m, taxa_administracao, aplicacao_minima, liquidez, status_produto
		rentabilidade12m, _ := strconv.ParseFloat(record[4], 64)
		aplicacaoMinima, _ := strconv.ParseFloat(record[7], 64)

		query := `INSERT INTO produtos (id_produto, nome_produto, risco_associado, rentabilidade_12m, aplicacao_minima, status_produto) 
		          VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (id_produto) DO NOTHING`

		_, err = db.Exec(query, record[0], record[1], record[3], rentabilidade12m, aplicacaoMinima, record[9])
		if err != nil {
			log.Printf("Erro ao inserir produto %s: %v", record[0], err)
			continue
		}
		count++
	}

	log.Printf("✓ %d produtos importados", count)
	return nil
}

func importarTransacoes(db *sql.DB, caminho string) error {
	log.Printf("\nImportando transações de %s...", caminho)

	file, err := os.Open(caminho)
	if err != nil {
		return fmt.Errorf("erro ao abrir arquivo: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	header, err := reader.Read()
	if err != nil {
		return err
	}
	log.Printf("Cabeçalho: %v", header)

	count := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Erro ao ler linha: %v", err)
			continue
		}

		// id_transacao, id_cliente, id_produto, tipo_transacao, valor_transacao,
		// data_transacao, status_transacao
		valorTransacao, _ := strconv.ParseFloat(record[4], 64)

		query := `INSERT INTO transacoes (id_transacao, id_cliente, id_produto, tipo_transacao, valor_transacao, data_transacao) 
		          VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (id_transacao) DO NOTHING`

		_, err = db.Exec(query, record[0], record[1], record[2], record[3], valorTransacao, record[5])
		if err != nil {
			log.Printf("Erro ao inserir transação %s: %v", record[0], err)
			continue
		}
		count++
	}

	log.Printf("✓ %d transações importadas", count)
	return nil
}

func importarInteracoes(db *sql.DB, caminho string) error {
	log.Printf("\nImportando interações de %s...", caminho)

	file, err := os.Open(caminho)
	if err != nil {
		return fmt.Errorf("erro ao abrir arquivo: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	header, err := reader.Read()
	if err != nil {
		return err
	}
	log.Printf("Cabeçalho: %v", header)

	count := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Erro ao ler linha: %v", err)
			continue
		}

		// id_interacao, id_cliente, id_produto, tipo_interacao, data_interacao, duracao_interacao_segundos
		query := `INSERT INTO interacoes (id_interacao, id_cliente, id_produto, tipo_interacao, data_interacao) 
		          VALUES ($1, $2, $3, $4, $5) ON CONFLICT (id_interacao) DO NOTHING`

		_, err = db.Exec(query, record[0], record[1], record[2], record[3], record[4])
		if err != nil {
			log.Printf("Erro ao inserir interação %s: %v", record[0], err)
			continue
		}
		count++
	}

	log.Printf("✓ %d interações importadas", count)
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
