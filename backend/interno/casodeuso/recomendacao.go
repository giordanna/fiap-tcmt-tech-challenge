package casodeuso

import (
	"log/slog"
	"sort"
	"sync"

	"backend/interno/dominio"
)

type ServicoRecomendacao struct {
	repo       dominio.RepositorioDados
	publicador dominio.Publicador
}

func NovoServicoRecomendacao(r dominio.RepositorioDados, p dominio.Publicador) *ServicoRecomendacao {
	return &ServicoRecomendacao{repo: r, publicador: p}
}

// Executar roda a lógica de scoring definida no projeto
func (s *ServicoRecomendacao) Executar(clienteID string) (*dominio.ResultadoRecomendacao, error) {
	slog.Info("Iniciando cálculo de recomendação", "cliente_id", clienteID)

	cliente, err := s.repo.ObterCliente(clienteID)
	if err != nil {
		slog.Error("Falha ao obter dados do cliente", "erro", err, "cliente_id", clienteID)
		return nil, err
	}

	produtos, err := s.repo.ListarProdutosAtivos()
	if err != nil {
		slog.Error("Falha ao listar produtos ativos", "erro", err)
		return nil, err
	}

	// estrutura para coletar resultados das goroutines
	type resultadoScore struct {
		item dominio.RecomendacaoItem
		ok   bool
	}

	// canal bufferizado para evitar bloqueio
	canal := make(chan resultadoScore, len(produtos))
	var wg sync.WaitGroup

	// processa cada produto em paralelo (goroutines)
	for _, p := range produtos {
		wg.Add(1)
		go func(prod dominio.Produto) {
			defer wg.Done()
			score := 0.0
			motivo := ""

			// 1. regra de compatibilidade de perfil
			matchRisco := false
			if cliente.PerfilRisco == "Conservador" && prod.RiscoAssociado == "Baixo" {
				score += 0.3
				matchRisco = true
			} else if cliente.PerfilRisco == "Moderado" && prod.RiscoAssociado == "Médio" {
				score += 0.25
				matchRisco = true
			} else if cliente.PerfilRisco == "Arrojado" && prod.RiscoAssociado == "Alto" {
				score += 0.2
				matchRisco = true
			}
			if matchRisco {
				motivo += "[perfil compativel] "
			}

			// 2. regra de rentabilidade (>10%)
			if prod.Rentabilidade12m > 10.0 {
				score += 0.1
				motivo += "[boa rentabilidade] "
			}

			// 3. regra de acessibilidade (< 5% patrimonio)
			if cliente.Patrimonio > 0 && prod.AplicacaoMinima < (cliente.Patrimonio*0.05) {
				score += 0.1
				motivo += "[acessivel] "
			}

			// 4. regra de diversificacao (penaliza se ja tem)
			jaTem, _ := s.repo.VerificarPosseProduto(cliente.ID, prod.ID)
			if jaTem {
				score -= 0.2
			}

			// 5. regra de interesse (se interagiu recentemente)
			interagiu, _ := s.repo.VerificarInteracaoRecente(cliente.ID, prod.ID)
			if interagiu {
				score += 0.15
				motivo += "[interesse recente] "
			}

			// envia para o canal se tiver pontuação positiva
			if score > 0 {
				canal <- resultadoScore{
					item: dominio.RecomendacaoItem{Produto: prod, Pontuacao: score, Motivo: motivo},
					ok:   true,
				}
			} else {
				canal <- resultadoScore{ok: false}
			}
		}(p)
	}

	// fecha o canal quando todas as goroutines terminarem
	go func() {
		wg.Wait()
		close(canal)
	}()

	var recomendacoes []dominio.RecomendacaoItem
	for res := range canal {
		if res.ok {
			recomendacoes = append(recomendacoes, res.item)
		}
	}

	// ordena por pontuação decrescente
	sort.Slice(recomendacoes, func(i, j int) bool {
		return recomendacoes[i].Pontuacao > recomendacoes[j].Pontuacao
	})

	slog.Info("Cálculo finalizado",
		"cliente_id", clienteID,
		"total_recomendacoes", len(recomendacoes),
		"produtos_analisados", len(produtos),
	)

	// persiste no banco (auditoria) e recupera uuid
	uuid, err := s.repo.SalvarRecomendacao(clienteID, recomendacoes)
	if err != nil {
		slog.Error("Erro ao salvar recomendação no banco (auditoria)", "erro", err, "cliente_id", clienteID)
		return nil, err
	}

	return &dominio.ResultadoRecomendacao{
		ID:            uuid,
		ClienteID:     cliente.ID,
		Recomendacoes: recomendacoes,
	}, nil
}

// BuscarUltima recupera a última recomendação gerada para o cliente
func (s *ServicoRecomendacao) BuscarUltima(clienteID string) (*dominio.ResultadoRecomendacao, error) {
	return s.repo.BuscarUltimaRecomendacao(clienteID)
}

// GerarEmMassa dispara o processo de recomendação para todos os clientes de forma assíncrona
func (s *ServicoRecomendacao) GerarEmMassa() error {
	clientes, err := s.repo.ListarTodosClientes()
	if err != nil {
		slog.Error("Erro ao listar clientes para geração em massa", "erro", err)
		return err
	}

	slog.Info("Iniciando geração em massa de recomendações", "total_clientes", len(clientes))

	for _, cliente := range clientes {
		// Publica evento para cada cliente
		// O tópico deve coincidir com o que o worker escuta: "gerar-recomendacao"
		s.publicador.Publicar("gerar-recomendacao", cliente.ID)
	}

	return nil
}
