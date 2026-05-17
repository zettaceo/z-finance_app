package address

import (
	"crypto/sha256"
	"math/big"
	"testing"
)

func TestResolveAddress(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		network Network
		asset   Asset
		ok      bool
	}{
		{
			name:    "btc_legacy",
			input:   "1BoatSLRHtKNngkdXEeobR76b53LETtpyT",
			network: NetworkBitcoin,
			asset:   AssetBTC,
			ok:      true,
		},
		{
			name:    "btc_bech32",
			input:   makeBech32Address(),
			network: NetworkBitcoin,
			asset:   AssetBTC,
			ok:      true,
		},
		{
			name:    "eth_eip55",
			input:   "0x52908400098527886E0F7030069857D2E4169EE7",
			network: NetworkEthereum,
			asset:   AssetETH,
			ok:      true,
		},
		{
			name:    "tron_usdt",
			input:   makeTronAddress(),
			network: NetworkTron,
			asset:   AssetUSDT,
			ok:      true,
		},
		{
			name:    "bsc_prefix",
			input:   "bsc:0x52908400098527886E0F7030069857D2E4169EE7",
			network: NetworkBSC,
			asset:   AssetUSDT,
			ok:      true,
		},
		{
			name:  "invalid",
			input: "not-an-address",
			ok:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			network, asset, _, err := ResolveAddress(tt.input)
			if tt.ok && err != nil {
				t.Fatalf("expected ok, got err: %v", err)
			}
			if !tt.ok && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if tt.ok {
				if network != tt.network {
					t.Fatalf("network mismatch: expected %s, got %s", tt.network, network)
				}
				if asset != tt.asset {
					t.Fatalf("asset mismatch: expected %s, got %s", tt.asset, asset)
				}
			}
		})
	}
}

func makeTronAddress() string {
	payload := make([]byte, 21)
	payload[0] = 0x41
	for i := 1; i < len(payload); i++ {
		payload[i] = byte(i + 1)
	}
	checksum := doubleSha256(payload)[:4]
	full := append(payload, checksum...)
	return base58Encode(full)
}

func makeBech32Address() string {
	hrp := "bc"
	data := make([]byte, 20)
	for i := range data {
		data[i] = byte(i + 1)
	}
	fiveBit := convertBits(data, 8, 5, true)
	checksum := bech32CreateChecksum(hrp, fiveBit)
	values := append(fiveBit, checksum...)
	return hrp + "1" + bech32Encode(values)
}

func convertBits(data []byte, from, to uint, pad bool) []int {
	var acc uint
	var bits uint
	var result []int
	maxv := uint((1 << to) - 1)
	for _, value := range data {
		acc = (acc << from) | uint(value)
		bits += from
		for bits >= to {
			bits -= to
			result = append(result, int((acc>>bits)&maxv))
		}
	}
	if pad && bits > 0 {
		result = append(result, int((acc<<(to-bits))&maxv))
	}
	return result
}

func bech32CreateChecksum(hrp string, data []int) []int {
	values := append(bech32HrpExpandTest(hrp), data...)
	values = append(values, []int{0, 0, 0, 0, 0, 0}...)
	polymod := bech32PolymodTest(values) ^ 1
	var checksum []int
	for i := 0; i < 6; i++ {
		checksum = append(checksum, (polymod>>uint(5*(5-i)))&31)
	}
	return checksum
}

func bech32Encode(values []int) string {
	out := make([]byte, len(values))
	for i, v := range values {
		out[i] = testBech32Charset[v]
	}
	return string(out)
}

const testBech32Charset = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"

func bech32HrpExpandTest(hrp string) []int {
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

func bech32PolymodTest(values []int) int {
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

func doubleSha256(data []byte) []byte {
	sum := sha256.Sum256(data)
	sum = sha256.Sum256(sum[:])
	return sum[:]
}

const testBase58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

func base58Encode(input []byte) string {
	var zeros int
	for zeros < len(input) && input[zeros] == 0 {
		zeros++
	}
	x := new(big.Int).SetBytes(input)
	base := big.NewInt(58)
	mod := new(big.Int)
	var result []byte
	for x.Sign() > 0 {
		x.DivMod(x, base, mod)
		result = append(result, testBase58Alphabet[mod.Int64()])
	}
	for i := 0; i < zeros; i++ {
		result = append(result, testBase58Alphabet[0])
	}
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}
	return string(result)
}
