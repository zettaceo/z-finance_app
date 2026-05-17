package ports

import "context"

type TransferGateway interface {
	Send(ctx context.Context, network, asset, address string, amount int64) (string, error)
}
