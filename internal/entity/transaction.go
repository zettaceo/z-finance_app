package entity

import "time"

type TransactionType string

const (
	TransactionTypeDeposit    TransactionType = "DEPOSIT"
	TransactionTypeWithdrawal TransactionType = "WITHDRAWAL"
	TransactionTypePayment    TransactionType = "PAYMENT"
	TransactionTypeTradeBuy   TransactionType = "TRADE_BUY"
	TransactionTypeTradeSell  TransactionType = "TRADE_SELL"
	TransactionTypeCardAuth   TransactionType = "CARD_AUTH"
	TransactionTypeReversal   TransactionType = "REVERSAL"
)

type TransactionStatus string

const (
	TransactionStatusCreated        TransactionStatus = "CREATED"
	TransactionStatusPendingPartner TransactionStatus = "PENDING_PARTNER"
	TransactionStatusHold           TransactionStatus = "HOLD"
	TransactionStatusConfirmed      TransactionStatus = "CONFIRMED"
	TransactionStatusRejected       TransactionStatus = "REJECTED"
	TransactionStatusReversed       TransactionStatus = "REVERSED"
)

type Transaction struct {
	ID             string
	AccountID      string
	UserID         string
	Type           TransactionType
	Status         TransactionStatus
	Amount         int64
	Fee            int64
	NetAmount      int64
	IdempotencyKey string
	ExternalRef    string
	ReversalOf     string
	OccurredAt     time.Time
	CreatedAt      time.Time
}
