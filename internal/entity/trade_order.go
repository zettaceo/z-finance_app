package entity

import "time"

type TradeSide string

const (
	TradeSideBuy  TradeSide = "BUY"
	TradeSideSell TradeSide = "SELL"
)

type TradeStatus string

const (
	TradeStatusCreated  TradeStatus = "CREATED"
	TradeStatusExecuted TradeStatus = "EXECUTED"
	TradeStatusFailed   TradeStatus = "FAILED"
)

type TradeOrder struct {
	ID                string
	UserID            string
	IdempotencyKey    string
	Status            TradeStatus
	Side              TradeSide
	BaseCurrency      string
	QuoteCurrency     string
	Price             int64
	Quantity          int64
	Fee               int64
	ExternalRef       string
	DebitAccountID    string
	CreditAccountID   string
	DebitTransaction  string
	CreditTransaction string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
