package usecase

const maxInt64 = int64(^uint64(0) >> 1)

func convertFiatToCrypto(fiatCents, priceCents int64, scale int32) (int64, bool) {
	if priceCents <= 0 || scale < 0 {
		return 0, false
	}
	multiplier, ok := pow10Int(scale)
	if !ok {
		return 0, false
	}
	if fiatCents > (maxInt64 / multiplier) {
		return 0, false
	}
	numerator := fiatCents * multiplier
	return numerator / priceCents, true
}

func convertCryptoToFiat(amountCrypto, priceCents int64, scale int32) (int64, bool) {
	if priceCents <= 0 || scale < 0 {
		return 0, false
	}
	multiplier, ok := pow10Int(scale)
	if !ok {
		return 0, false
	}
	if amountCrypto > (maxInt64 / priceCents) {
		return 0, false
	}
	numerator := amountCrypto * priceCents
	return numerator / multiplier, true
}

func pow10Int(exp int32) (int64, bool) {
	if exp < 0 || exp > 18 {
		return 0, false
	}
	var result int64 = 1
	for i := int32(0); i < exp; i++ {
		if result > maxInt64/10 {
			return 0, false
		}
		result *= 10
	}
	return result, true
}
