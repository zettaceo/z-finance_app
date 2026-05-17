package ports

import "context"

// ExecutionIntent representa uma intencao de execucao aprovada externamente.
// Invariantes:
// - Nunca altera saldo diretamente; apenas descreve a acao.
// - Deve ser imutavel e idempotente (ID unico por execucao).
// - Requer confirmacao explicita do usuario/operacao antes de executar.
// - O core deve tratar intents de forma deterministica e auditavel.
type ExecutionIntent interface {
	ID() string
	UserID() string
	Kind() string
	Payload() map[string]any
}

// AIProvider apenas analisa e propoe intents; nao executa saldo.
// Qualquer execucao real deve passar pelos use cases do core.
type AIProvider interface {
	Analyze(ctx context.Context, userID string, input map[string]any) (map[string]any, error)
	ProposeIntent(ctx context.Context, userID string, input map[string]any) (ExecutionIntent, error)
}

// WalletProvider define operacoes de carteira/custodia.
// Implementacoes devem expor health-checks e nunca alterar ledger direto.
type WalletProvider interface {
	Health(ctx context.Context) error
}

// DexProvider define operacoes em DEXes.
// Operacoes de troca devem retornar apenas dados de mercado; execucao real via core.
type DexProvider interface {
	Health(ctx context.Context) error
}

// LiquidityProvider define fontes externas de liquidez.
// Deve suportar monitoramento e fallback sem bloquear o core.
type LiquidityProvider interface {
	Health(ctx context.Context) error
}
