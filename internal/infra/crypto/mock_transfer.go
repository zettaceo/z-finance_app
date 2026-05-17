package crypto

import (
	"context"
	"fmt"
)

type MockTransferGateway struct{}

func NewMockTransferGateway() *MockTransferGateway {
	return &MockTransferGateway{}
}

func (m *MockTransferGateway) Send(ctx context.Context, network, asset, address string, amount int64) (string, error) {
	if address == "" || amount <= 0 {
		return "", fmt.Errorf("transferencia invalida")
	}
	return "mock-transfer-hash", nil
}
