package ports

import (
	"context"

	"z-finance-api/internal/entity"
)

type CustodyGateway interface {
	CreateDepositAddress(ctx context.Context, userID, asset, network string) (entity.CustodyAddress, error)
	SendTransfer(ctx context.Context, transfer entity.CustodyTransfer) (entity.CustodyTransfer, error)
	GetTransferStatus(ctx context.Context, providerID string) (entity.CustodyTransferStatus, error)
}
