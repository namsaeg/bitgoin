package ecdsa

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"math/big"
)

func nonceRFC6979(d *big.Int, z [32]byte) *big.Int {
	var x, h1 [32]byte
	d.FillBytes(x[:])
	zInt := new(big.Int).SetBytes(z[:])
	zInt.Mod(zInt, curveN).FillBytes(h1[:])

	mac := func(key []byte, parts ...[]byte) []byte {
		m := hmac.New(sha256.New, key)
		for _, p := range parts {
			m.Write(p)
		}
		return m.Sum(nil)
	}

	v := bytes.Repeat([]byte{0x01}, 32)
	k := make([]byte, 32)

	k = mac(k, v, []byte{0x00}, x[:], h1[:])
	v = mac(k, v)
	k = mac(k, v, []byte{0x01}, x[:], h1[:])
	v = mac(k, v)

	for {
		v = mac(k, v)
		cand := new(big.Int).SetBytes(v)
		if cand.Sign() > 0 && cand.Cmp(curveN) < 0 {
			return cand
		}
		k = mac(k, v, []byte{0x00})
		v = mac(k, v)
	}
}
