package ports

import (
	"context"

	"z-finance-api/internal/entity"
)

type PixPartnerClient interface {
	Send(ctx context.Context, transfer *entity.PixTransfer) error
}
