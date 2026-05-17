package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"z-finance-api/internal/core/ports"
	"z-finance-api/internal/entity"
)

func liquidateCryptoForDeficit(
	ctx context.Context,
	uowTx ports.UnitOfWorkTx,
	exchange ports.ExchangeGateway,
	clock Clock,
	userID string,
	accountID string,
	idempotencyPrefix string,
	externalRef string,
	deficit int64,
	assets []string,
	pricing *ResolvePricingUseCase,
) (int64, error) {
	remaining := deficit
	if remaining <= 0 {
		return 0, nil
	}
	for _, asset := range assets {
		if remaining <= 0 {
			break
		}
		normalized := strings.ToUpper(strings.TrimSpace(asset))
		if normalized == "" {
			continue
		}
		idempotencyKey := ""
		if idempotencyPrefix != "" {
			idempotencyKey = fmt.Sprintf("%s-%s", idempotencyPrefix, strings.ToLower(normalized))
		}
		if idempotencyKey != "" {
			existingTx, err := uowTx.TransactionRepository().GetByIdempotencyKey(ctx, accountID, idempotencyKey)
			if err != nil {
				return remaining, err
			}
			if existingTx != nil {
				remaining -= existingTx.NetAmount
				if remaining < 0 {
					remaining = 0
				}
				continue
			}
		}

		confirmed, err := uowTx.CryptoTransferRepository().SumConfirmedByUserAsset(ctx, userID, normalized)
		if err != nil {
			return remaining, err
		}
		if confirmed <= 0 {
			continue
		}
		quote, err := exchange.Quote(ctx, normalized)
		if err != nil {
			return remaining, NewExternalDependencyError("exchange", "quote", err)
		}
		brlCap, ok := convertCryptoToFiat(confirmed, quote.PriceInBRLCents, 2)
		if !ok || brlCap <= 0 {
			continue
		}

		creditBRL := remaining
		sellAmount := confirmed
		if brlCap > remaining {
			sellAmount, ok = convertFiatToCrypto(remaining, quote.PriceInBRLCents, 2)
			if !ok || sellAmount <= 0 {
				continue
			}
		} else {
			creditBRL = brlCap
		}

		now := clock.Now().UTC()
		cryptoSell := &entity.CryptoTransfer{
			ID:            uuid.NewString(),
			UserID:        userID,
			AccountID:     accountID,
			TransactionID: uuid.NewString(),
			Asset:         normalized,
			Network:       "INTERNAL",
			Address:       externalRef,
			Amount:        sellAmount,
			Fee:           0,
			Status:        entity.CryptoTransferPendingExchange,
			Direction:     "SELL",
			CreatedAt:     now,
		}
		if err := uowTx.CryptoTransferRepository().Create(ctx, cryptoSell); err != nil {
			return remaining, err
		}
		txHash, err := exchange.Execute(ctx, normalized, sellAmount, "SELL")
		if err != nil {
			return remaining, NewExternalDependencyError("exchange", "execute", err)
		}
		if err := uowTx.CryptoTransferRepository().UpdateStatus(ctx, cryptoSell.ID, entity.CryptoTransferConfirmed, txHash); err != nil {
			return remaining, err
		}

		feeResult := entity.FeeResult{Amount: creditBRL, Fee: 0, NetAmount: creditBRL}
		var rule *entity.PricingRule
		if pricing != nil {
			resolved, matchedRule, err := pricing.Execute(ctx, PricingInput{
				UserID:        userID,
				OperationType: entity.PricingOperationSwap,
				Asset:         "BRL",
				GrossAmount:   creditBRL,
				FeatureCode:   string(entity.PricingOperationSwap),
			})
			if err != nil {
				return remaining, err
			}
			feeResult = resolved
			rule = matchedRule
		}

		sellTx := &entity.Transaction{
			ID:             cryptoSell.TransactionID,
			AccountID:      accountID,
			UserID:         userID,
			Type:           entity.TransactionTypeTradeSell,
			Status:         entity.TransactionStatusConfirmed,
			Amount:         creditBRL,
			Fee:            feeResult.Fee,
			NetAmount:      feeResult.NetAmount,
			IdempotencyKey: idempotencyKey,
			ExternalRef:    normalizeExternalRef(externalRef, entity.TransactionTypeTradeSell, "LIQUIDATE_CRYPTO"),
			OccurredAt:     now,
			CreatedAt:      now,
		}
		if err := uowTx.TransactionRepository().Create(ctx, sellTx); err != nil {
			return remaining, err
		}
		if err := appendConversionAudit(ctx, uowTx, clock, conversionAuditInput{
			UserID:          userID,
			OperationType:   entity.PricingOperationSwap,
			Asset:           normalized,
			GrossAmount:     feeResult.Amount,
			Fee:             feeResult.Fee,
			NetAmount:       feeResult.NetAmount,
			QuotePrice:      quote.PriceInBRLCents,
			SpreadBps:       0,
			LiquiditySource: "EXCHANGE",
			RelatedType:     "transaction",
			RelatedID:       sellTx.ID,
			QuotedAt:        &now,
		}); err != nil {
			return remaining, err
		}
		if feeResult.Fee > 0 {
			_ = appendAudit(ctx, uowTx, clock, userID, "PRICING_APPLIED", "transaction", sellTx.ID, map[string]any{
				"operation":  string(entity.PricingOperationSwap),
				"amount":     feeResult.Amount,
				"fee":        feeResult.Fee,
				"net_amount": feeResult.NetAmount,
				"rule_id":    ruleID(rule),
			})
		}

		remaining -= feeResult.NetAmount
	}

	return remaining, nil
}
