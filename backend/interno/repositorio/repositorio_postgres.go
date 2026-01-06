package repositorio

import (
	"database/sql"
	"encoding/json"

	"github.com/giordanna/fiap-tcmt-tech-challenge/backend/interno/dominio"
)

type RepositorioPostgres struct {
	db *sql.DB
}

func NovoRepositorioPostgres(db *sql.DB) *RepositorioPostgres {
	return &RepositorioPostgres{db: db}
}

func (r *RepositorioPostgres) ObterCliente(id string) (*dominio.Cliente, error) {
	query := `SELECT id_cliente, nome_cliente, perfil_risco, patrimonio_total_estimado FROM clientes WHERE id_cliente = $1`
	var c dominio.Cliente
	err := r.db.QueryRow(query, id).Scan(&c.ID, &c.Nome, &c.PerfilRisco, &c.Patrimonio)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *RepositorioPostgres) ListarProdutosAtivos() ([]dominio.Produto, error) {
	query := `SELECT id_produto, nome_produto, risco_associado, rentabilidade_12m, aplicacao_minima FROM produtos WHERE status_produto = 'Ativo'`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var produtos []dominio.Produto
	for rows.Next() {
		var p dominio.Produto
		if err := rows.Scan(&p.ID, &p.Nome, &p.RiscoAssociado, &p.Rentabilidade12m, &p.AplicacaoMinima); err != nil {
			return nil, err
		}
		produtos = append(produtos, p)
	}
	return produtos, nil
}

func (r *RepositorioPostgres) VerificarPosseProduto(clienteID, produtoID string) (bool, error) {
	var existe bool
	// query otimizada usando EXISTS
	query := `SELECT EXISTS(SELECT 1 FROM transacoes WHERE id_cliente=$1 AND id_produto=$2 AND tipo_transacao='Aplicacao')`
	err := r.db.QueryRow(query, clienteID, produtoID).Scan(&existe)
	return existe, err
}

func (r *RepositorioPostgres) VerificarInteracaoRecente(clienteID, produtoID string) (bool, error) {
	var existe bool
	query := `SELECT EXISTS(SELECT 1 FROM interacoes WHERE id_cliente=$1 AND id_produto=$2)`
	err := r.db.QueryRow(query, clienteID, produtoID).Scan(&existe)
	return existe, err
}

func (r *RepositorioPostgres) SalvarRecomendacao(clienteID string, itens []dominio.RecomendacaoItem) (string, error) {
	// converte o slice de structs para jsonb
	jsonBytes, err := json.Marshal(itens)
	if err != nil {
		return "", err
	}

	query := `INSERT INTO recomendacoes (id_cliente, produtos_json) VALUES ($1, $2) RETURNING id`

	var uuidGerado string
	// executa insert e já retorna o uuid gerado pelo banco
	err = r.db.QueryRow(query, clienteID, jsonBytes).Scan(&uuidGerado)
	if err != nil {
		return "", err
	}
	return uuidGerado, nil
}
