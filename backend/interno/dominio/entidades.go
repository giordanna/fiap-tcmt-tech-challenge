package dominio

// entidades de dominio que espelham o banco
type Cliente struct {
	ID          string  `json:"id_cliente"`
	PerfilRisco string  `json:"perfil_risco"`
	Patrimonio  float64 `json:"patrimonio_total_estimado"`
}

type Produto struct {
	ID               string  `json:"id_produto"`
	Nome             string  `json:"nome_produto"`
	RiscoAssociado   string  `json:"risco_associado"`
	Rentabilidade12m float64 `json:"rentabilidade_12m"`
	AplicacaoMinima  float64 `json:"aplicacao_minima"`
}

// estruturas de retorno da api
type RecomendacaoItem struct {
	Produto   Produto `json:"produto"`
	Pontuacao float64 `json:"pontuacao"`
	Motivo    string  `json:"motivo"`
}

type ResultadoRecomendacao struct {
	ID            string             `json:"id_recomendacao"` // uuid gerado
	ClienteID     string             `json:"id_cliente"`
	Recomendacoes []RecomendacaoItem `json:"recomendacoes"`
}

// interface do repositorio (inversão de dependência)
type RepositorioDados interface {
	ObterCliente(id string) (*Cliente, error)
	ListarProdutosAtivos() ([]Produto, error)
	VerificarPosseProduto(clienteID, produtoID string) (bool, error)
	VerificarInteracaoRecente(clienteID, produtoID string) (bool, error)
	SalvarRecomendacao(clienteID string, itens []RecomendacaoItem) (string, error)
	BuscarUltimaRecomendacao(clienteID string) (*ResultadoRecomendacao, error)
	ListarTodosClientes() ([]Cliente, error)
}

type Publicador interface {
	Publicar(topico string, payload interface{})
}
