package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"z-finance-api/internal/conversion/engine"
	"z-finance-api/internal/infra/observability"
	"z-finance-api/internal/core/ports"
	"z-finance-api/internal/entity"
	"z-finance-api/internal/repository"
)

var ErrCardJitDeclined = errors.New("cartao recusado")

type AuthorizeCardInput struct {
	UserID       string
	AccountID    string
	Amount       int64
	Fee          int64
	NetAmount    int64
	MerchantName string
	MerchantMCC  string
	AuthCode     string
	ExternalRef  string
	IdempotencyKey string
}

type CardAuthorizationResult struct {
	Approved        bool
	AuthorizationID string
	Reason          string
}

type AuthorizeCardJITUseCase struct {
	uow      ports.UnitOfWork
	exchange ports.ExchangeGateway
	settings repository.UserSettingsRepository
	engine   *engine.Engine
	pricing  *ResolvePricingUseCase
	clock    Clock
	timeout  time.Duration
}

func NewAuthorizeCardJITUseCase(uow ports.UnitOfWork, exchange ports.ExchangeGateway, settings repository.UserSettingsRepository, engine *engine.Engine, pricing *ResolvePricingUseCase) *AuthorizeCardJITUseCase {
	return &AuthorizeCardJITUseCase{
		uow:      uow,
		exchange: exchange,
		settings: settings,
		engine:   engine,
		pricing:  pricing,
		clock:    NewRealClock(),
		timeout:  180 * time.Millisecond,
	}
}

func NewAuthorizeCardJITUseCaseWithTimeout(uow ports.UnitOfWork, exchange ports.ExchangeGateway, settings repository.UserSettingsRepository, engine *engine.Engine, pricing *ResolvePricingUseCase, timeout time.Duration) *AuthorizeCardJITUseCase {
	return &AuthorizeCardJITUseCase{
		uow:      uow,
		exchange: exchange,
		settings: settings,
		engine:   engine,
		pricing:  pricing,
		clock:    NewRealClock(),
		timeout:  timeout,
	}
}

func (uc *AuthorizeCardJITUseCase) Execute(ctx context.Context, input AuthorizeCardInput) (CardAuthorizationResult, error) {
	ctx, end := observability.StartUseCaseSpan(ctx, "AuthorizeCardJIT")
	defer end()

	ctx, cancel := context.WithTimeout(ctx, uc.timeout)
	defer cancel()

	resultCh := make(chan CardAuthorizationResult, 1)
	errCh := make(chan error, 1)
	go func() {
		result, err := uc.execute(ctx, input)
		if err != nil {
			errCh <- err
			return
		}
		resultCh <- result
	}()

	select {
	case result := <-resultCh:
		return result, nil
	case err := <-errCh:
		return CardAuthorizationResult{}, err
	case <-ctx.Done():
		return CardAuthorizationResult{Approved: false, Reason: "timeout"}, nil
	}
}

func (uc *AuthorizeCardJITUseCase) execute(ctx context.Context, input AuthorizeCardInput) (CardAuthorizationResult, error) {
	if input.UserID == "" || input.AccountID == "" || input.Amount <= 0 {
		return CardAuthorizationResult{Approved: false, Reason: "dados_invalidos"}, nil
	}
	feeResult, rule, err := uc.resolveFee(ctx, input)
	if err != nil {
		return CardAuthorizationResult{}, err
	}
	netAmount := feeResult.NetAmount
	if netAmount <= 0 {
		return CardAuthorizationResult{Approved: false, Reason: "valor_invalido"}, nil
	}

	uowTx, err := uc.uow.Begin(ctx)
	if err != nil {
		return CardAuthorizationResult{}, err
	}
	defer func() { _ = uowTx.Rollback(ctx) }()

	account, err := uowTx.AccountRepository().GetByIDForUpdate(ctx, input.AccountID)
	if err != nil {
		return CardAuthorizationResult{}, err
	}
	if account.UserID != input.UserID || account.Status != entity.AccountStatusActive {
		return CardAuthorizationResult{Approved: false, Reason: "conta_inativa"}, nil
	}

	ledgerBalance, err := uowTx.TransactionRepository().GetLedgerBalance(ctx, input.AccountID)
	if err != nil {
		return CardAuthorizationResult{}, err
	}
	holdBalance, err := uowTx.TransactionRepository().GetHoldBalance(ctx, input.AccountID)
	if err != nil {
		return CardAuthorizationResult{}, err
	}
	available := ledgerBalance - holdBalance

	if available < netAmount {
		deficit := netAmount - available
		assets := uc.resolveFallbackAssets(ctx, input.UserID)
		remaining, err := liquidateCryptoForDeficit(ctx, uowTx, uc.exchange, uc.clock, input.UserID, input.AccountID, fmt.Sprintf("card-jit-sell-%s", input.IdempotencyKey), "CARD_JIT_SELL", deficit, assets, uc.pricing)
		if err != nil {
			return CardAuthorizationResult{}, err
		}

		if remaining > 0 {
			return CardAuthorizationResult{Approved: false, Reason: "saldo_insuficiente"}, nil
		}
	}

	now := uc.clock.Now().UTC()
	externalRef := normalizeExternalRef(input.ExternalRef, entity.TransactionTypeCardAuth, "CARD_AUTH")
	authTx := &entity.Transaction{
		ID:             uuid.NewString(),
		AccountID:      input.AccountID,
		UserID:         input.UserID,
		Type:           entity.TransactionTypeCardAuth,
		Status:         entity.TransactionStatusHold,
		Amount:         input.Amount,
		Fee:            feeResult.Fee,
		NetAmount:      netAmount,
		IdempotencyKey: input.IdempotencyKey,
		ExternalRef:    externalRef,
		OccurredAt:     now,
		CreatedAt:      now,
	}
	if err := uowTx.TransactionRepository().Create(ctx, authTx); err != nil {
		return CardAuthorizationResult{}, err
	}

	auth := &entity.CardAuthorization{
		ID:            uuid.NewString(),
		UserID:        input.UserID,
		AccountID:     input.AccountID,
		Status:        entity.CardAuthStatusHold,
		Amount:        input.Amount,
		Fee:           feeResult.Fee,
		NetAmount:     netAmount,
		MerchantName:  input.MerchantName,
		MerchantMCC:   input.MerchantMCC,
		AuthCode:      input.AuthCode,
		ExternalRef:   externalRef,
		TransactionID: authTx.ID,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := uowTx.CardAuthorizationRepository().Create(ctx, auth); err != nil {
		return CardAuthorizationResult{}, err
	}

	_ = appendAudit(ctx, uowTx, uc.clock, input.UserID, "CARD_AUTH_HOLD", "card_authorization", auth.ID, map[string]any{
		"status": "HOLD",
	})
	if feeResult.Fee > 0 {
		_ = appendAudit(ctx, uowTx, uc.clock, input.UserID, "PRICING_APPLIED", "transaction", authTx.ID, map[string]any{
			"operation":  string(entity.PricingOperationCardCrypto),
			"amount":     feeResult.Amount,
			"fee":        feeResult.Fee,
			"net_amount": feeResult.NetAmount,
			"rule_id":    ruleID(rule),
		})
	}

	if err := uowTx.Commit(ctx); err != nil {
		return CardAuthorizationResult{}, err
	}

	return CardAuthorizationResult{Approved: true, AuthorizationID: auth.ID}, nil
}

func (uc *AuthorizeCardJITUseCase) resolveFee(ctx context.Context, input AuthorizeCardInput) (entity.FeeResult, *entity.PricingRule, error) {
	if uc.pricing == nil {
		netAmount := input.NetAmount
		if netAmount == 0 && input.Amount > 0 {
			netAmount = input.Amount - input.Fee
		}
		return entity.FeeResult{Amount: input.Amount, Fee: input.Fee, NetAmount: netAmount}, nil, nil
	}
	return uc.pricing.Execute(ctx, PricingInput{
		UserID:        input.UserID,
		OperationType: entity.PricingOperationCardCrypto,
		Asset:         "BRL",
		GrossAmount:   input.Amount,
		FeatureCode:   string(entity.PricingOperationCardCrypto),
	})
}

func (uc *AuthorizeCardJITUseCase) resolveFallbackAssets(ctx context.Context, userID string) []string {
	assets := []string{"USDT", "ETH", "BTC", "MATIC"}
	if uc.engine != nil {
		resolved, err := uc.engine.ResolveAssets(ctx, userID, entity.ConversionTriggerCardJIT)
		if err == nil && len(resolved) > 0 {
			return resolved
		}
	}
	if uc.settings == nil {
		return assets
	}
	settings, err := uc.settings.GetByUserID(ctx, userID)
	if err != nil {
		return assets
	}
	if len(settings.ConversionPriority) > 0 {
		assets = make([]string, 0, len(settings.ConversionPriority))
		for _, asset := range settings.ConversionPriority {
			asset = strings.ToUpper(strings.TrimSpace(asset))
			if asset != "" {
				assets = append(assets, asset)
			}
		}
	}
	if settings.FallbackAsset != "" {
		fallback := strings.ToUpper(strings.TrimSpace(settings.FallbackAsset))
		if fallback != "" {
			assets = prependIfMissing(assets, fallback)
		}
	}
	return assets
}

func prependIfMissing(values []string, value string) []string {
	for _, item := range values {
		if item == value {
			return values
		}
	}
	return append([]string{value}, values...)
}
