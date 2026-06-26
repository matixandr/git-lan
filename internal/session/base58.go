package session

import (
	"math/big"
)

// base58 uses the Bitcoin alphabet: no 0, O, I, or l, so tokens are safe to
// read aloud and type. Implemented here to avoid an external dependency.
const base58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

var base58Index = func() map[byte]int {
	m := make(map[byte]int, len(base58Alphabet))
	for i := 0; i < len(base58Alphabet); i++ {
		m[base58Alphabet[i]] = i
	}
	return m
}()

// base58Encode encodes bytes, preserving leading-zero bytes as leading '1's.
func base58Encode(input []byte) string {
	var zeros int
	for zeros < len(input) && input[zeros] == 0 {
		zeros++
	}

	num := new(big.Int).SetBytes(input)
	radix := big.NewInt(58)
	mod := new(big.Int)

	var out []byte
	for num.Sign() > 0 {
		num.DivMod(num, radix, mod)
		out = append(out, base58Alphabet[mod.Int64()])
	}
	for i := 0; i < zeros; i++ {
		out = append(out, base58Alphabet[0])
	}
	// Reverse.
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return string(out)
}

// base58Decode reverses base58Encode. It returns false on any invalid char.
func base58Decode(s string) ([]byte, bool) {
	num := big.NewInt(0)
	radix := big.NewInt(58)
	for i := 0; i < len(s); i++ {
		idx, ok := base58Index[s[i]]
		if !ok {
			return nil, false
		}
		num.Mul(num, radix)
		num.Add(num, big.NewInt(int64(idx)))
	}
	decoded := num.Bytes()

	var zeros int
	for zeros < len(s) && s[zeros] == base58Alphabet[0] {
		zeros++
	}
	out := make([]byte, zeros+len(decoded))
	copy(out[zeros:], decoded)
	return out, true
}
