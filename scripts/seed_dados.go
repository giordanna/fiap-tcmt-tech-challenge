package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL nao definida")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	fmt.Println("Iniciando carga de dados...")

	// Carga de Clientes
	file, err := os.Open("scripts/datalate_poc/clientes.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, _ := reader.ReadAll()

	for _, record := range records[1:] { // Pula header
		_, err := db.Exec(`
			INSERT INTO clientes (id_cliente, nome_cliente, perfil_risco, patrimonio_total_estimado)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (id_cliente) DO NOTHING`,
			record[0], record[1], record[2], record[3],
		)
		if err != nil {
			log.Printf("Erro ao inserir cliente %s: %v", record[0], err)
		}
	}

	// Carga de Produtos
	file2, err := os.Open("scripts/datalate_poc/produtos.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer file2.Close()

	reader2 := csv.NewReader(file2)
	records2, _ := reader2.ReadAll()

	for _, record := range records2[1:] {
		_, err := db.Exec(`
			INSERT INTO produtos (id_produto, nome_produto, risco_associado, rentabilidade_12m, aplicacao_minima)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (id_produto) DO NOTHING`,
			record[0], record[1], record[2], record[3], record[4],
		)
		if err != nil {
			log.Printf("Erro ao inserir produto %s: %v", record[0], err)
		}
	}

	fmt.Println("Carga de dados finalizada.")
}
