package entity

// LedgerEffect retorna o sinal do efeito no saldo para cada tipo de transacao.
// +1 soma, -1 subtrai, 0 neutro/indefinido.
func LedgerEffect(txType TransactionType) int64 {
	switch txType {
	case TransactionTypeDeposit,
		TransactionTypeTradeSell,
		TransactionTypeReversal:
		return 1
	case TransactionTypeWithdrawal,
		TransactionTypePayment,
		TransactionTypeTradeBuy,
		TransactionTypeCardAuth:
		return -1
	default:
		return 0
	}
}

func IsDebitType(txType TransactionType) bool {
	return LedgerEffect(txType) < 0
}
