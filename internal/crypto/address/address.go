package address

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"

	"golang.org/x/crypto/sha3"
)

type Network string

const (
	NetworkBitcoin Network = "BITCOIN"
	NetworkEthereum Network = "ETHEREUM"
	NetworkTron Network = "TRON"
	NetworkBSC Network = "BSC"
)

type Asset string

const (
	AssetUSDT  Asset = "USDT"
	AssetBTC   Asset = "BTC"
	AssetETH   Asset = "ETH"
	AssetMATIC Asset = "MATIC"
)

var (
	ErrInvalidAddress = errors.New("endereco cripto invalido")
)

type Address struct {
	Value      string
	Normalized string
}

func ResolveAddress(input string) (Network, Asset, Address, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return "", "", Address{}, ErrInvalidAddress
	}

	lower := strings.ToLower(trimmed)
	if strings.HasPrefix(lower, "bsc:") {
		raw := strings.TrimSpace(trimmed[4:])
		if isValidEVMAddress(raw) {
			return NetworkBSC, AssetUSDT, Address{Value: trimmed, Normalized: normalizeEVM(raw)}, nil
		}
		return "", "", Address{}, ErrInvalidAddress
	}

	if strings.HasPrefix(lower, "eth:") {
		raw := strings.TrimSpace(trimmed[4:])
		if isValidEVMAddress(raw) {
			return NetworkEthereum, AssetETH, Address{Value: trimmed, Normalized: normalizeEVM(raw)}, nil
		}
		return "", "", Address{}, ErrInvalidAddress
	}

	if isValidBTCAddress(trimmed) {
		return NetworkBitcoin, AssetBTC, Address{Value: trimmed, Normalized: trimmed}, nil
	}
	if isValidEVMAddress(trimmed) {
		return NetworkEthereum, AssetETH, Address{Value: trimmed, Normalized: normalizeEVM(trimmed)}, nil
	}
	if isValidTronAddress(trimmed) {
		return NetworkTron, AssetUSDT, Address{Value: trimmed, Normalized: trimmed}, nil
	}
	return "", "", Address{}, ErrInvalidAddress
}

func normalizeEVM(value string) string {
	trimmed := strings.TrimSpace(value)
	if strings.HasPrefix(trimmed, "0x") || strings.HasPrefix(trimmed, "0X") {
		return "0x" + strings.ToLower(trimmed[2:])
	}
	return "0x" + strings.ToLower(trimmed)
}

func isValidBTCAddress(value string) bool {
	if strings.HasPrefix(strings.ToLower(value), "bc1") {
		return isValidBech32(value, "bc")
	}
	return isValidBase58Check(value, []byte{0x00, 0x05})
}

func isValidTronAddress(value string) bool {
	if !strings.HasPrefix(value, "T") {
		return false
	}
	payload, ok := decodeBase58Check(value)
	if !ok || len(payload) != 21 {
		return false
	}
	return payload[0] == 0x41
}

func isValidEVMAddress(value string) bool {
	normalized := normalizeEVM(value)
	hexBody := normalized[2:]
	if len(hexBody) != 40 {
		return false
	}
	if _, err := hex.DecodeString(hexBody); err != nil {
		return false
	}
	return validateEIP55(value)
}

func validateEIP55(value string) bool {
	normalized := normalizeEVM(value)
	raw := normalized[2:]
	if raw == strings.ToLower(raw) || raw == strings.ToUpper(raw) {
		return true
	}
	hash := keccak256(strings.ToLower(raw))
	for i := 0; i < len(raw); i++ {
		r := raw[i]
		hashNibble := hash[i]
		if r >= 'a' && r <= 'f' {
			if hashNibble >= '8' {
				return false
			}
		}
		if r >= 'A' && r <= 'F' {
			if hashNibble < '8' {
				return false
			}
		}
	}
	return true
}

func keccak256(input string) string {
	hasher := sha3.NewLegacyKeccak256()
	_, _ = hasher.Write([]byte(input))
	return hex.EncodeToString(hasher.Sum(nil))
}

const base58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

func isValidBase58Check(value string, allowedVersions []byte) bool {
	payload, ok := decodeBase58Check(value)
	if !ok || len(payload) < 1 {
		return false
	}
	version := payload[0]
	for _, v := range allowedVersions {
		if v == version {
			return true
		}
	}
	return false
}

func decodeBase58Check(value string) ([]byte, bool) {
	decoded := decodeBase58(value)
	if len(decoded) < 4 {
		return nil, false
	}
	payload := decoded[:len(decoded)-4]
	checksum := decoded[len(decoded)-4:]
	hash := sha256.Sum256(payload)
	hash = sha256.Sum256(hash[:])
	if !equalBytes(checksum, hash[:4]) {
		return nil, false
	}
	return payload, true
}

func decodeBase58(value string) []byte {
	var zeros int
	for zeros < len(value) && value[zeros] == '1' {
		zeros++
	}

	result := make([]byte, 0, len(value))
	for i := 0; i < len(value); i++ {
		char := value[i]
		index := strings.IndexByte(base58Alphabet, char)
		if index < 0 {
			return nil
		}
		carry := index
		for j := len(result) - 1; j >= 0; j-- {
			carry += int(result[j]) * 58
			result[j] = byte(carry & 0xff)
			carry >>= 8
		}
		for carry > 0 {
			result = append([]byte{byte(carry & 0xff)}, result...)
			carry >>= 8
		}
	}

	for i := 0; i < zeros; i++ {
		result = append([]byte{0x00}, result...)
	}
	return result
}

func equalBytes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func isValidBech32(value, hrp string) bool {
	lower := strings.ToLower(value)
	if !strings.HasPrefix(lower, hrp+"1") {
		return false
	}
	pos := strings.LastIndexByte(lower, '1')
	if pos < 1 || pos+7 > len(lower) {
		return false
	}
	data := lower[pos+1:]
	if !bech32VerifyChecksum(hrp, data) {
		return false
	}
	return true
}

func bech32VerifyChecksum(hrp, data string) bool {
	expanded := bech32HrpExpand(hrp)
	var values []int
	values = append(values, expanded...)
	for i := 0; i < len(data); i++ {
		idx := strings.IndexByte(bech32Charset, data[i])
		if idx == -1 {
			return false
		}
		values = append(values, idx)
	}
	polymod := bech32Polymod(values)
	return polymod == 1 || polymod == 0x2bc830a3
}

const bech32Charset = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"

func bech32HrpExpand(hrp string) []int {
	result := make([]int, 0, len(hrp)*2+1)
	for _, r := range hrp {
		result = append(result, int(r>>5))
	}
	result = append(result, 0)
	for _, r := range hrp {
		result = append(result, int(r&31))
	}
	return result
}

func bech32Polymod(values []int) int {
	generator := []int{0x3b6a57b2, 0x26508e6d, 0x1ea119fa, 0x3d4233dd, 0x2a1462b3}
	chk := 1
	for _, v := range values {
		top := chk >> 25
		chk = (chk&0x1ffffff)<<5 ^ v
		for i := 0; i < 5; i++ {
			if ((top >> i) & 1) == 1 {
				chk ^= generator[i]
			}
		}
	}
	return chk
}
