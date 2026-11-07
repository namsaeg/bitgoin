package base58

import (
	"math/big"
	"slices"

	"bitgoin_v1/hash"
)

const base58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

func Encode(b []byte) string {
	zeros := 0
	for zeros < len(b) && b[zeros] == 0 {
		zeros++
	}
	n := new(big.Int).SetBytes(b)
	base, mod := big.NewInt(58), new(big.Int)
	var out []byte
	for n.Sign() > 0 {
		n.DivMod(n, base, mod)
		out = append(out, base58Alphabet[mod.Int64()])
	}
	for i := 0; i < zeros; i++ {
		out = append(out, '1')
	}
	slices.Reverse(out)
	return string(out)
}

func CheckEncode(b []byte) string {
	checksum := hash.Hash256(b)
	buf := make([]byte, 0, len(b)+4)
	buf = append(buf, b...)
	buf = append(buf, checksum[:4]...)
	return Encode(buf)
}
