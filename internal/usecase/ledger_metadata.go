package usecase

import (
	"strings"

	"z-finance-api/internal/entity"
)

func normalizeExternalRef(value string, txType entity.TransactionType, fallback string) string {
	if strings.TrimSpace(value) != "" {
		return value
	}
	if strings.TrimSpace(fallback) != "" {
		return fallback
	}
	return string(txType)
}
