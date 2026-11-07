package ecdsa

import (
	"errors"
	"math/big"

	"bitgoin_v1/secp256k1"
)

type Signature struct {
	r, s *big.Int
}

func NewSignature(r, s *big.Int) (*Signature, error) {
	if !inScalarRange(r, curveN) {
		return nil, errors.New("ecdsa: 서명의 r은 1 이상 n-1 이하여야 함")
	}
	if !inScalarRange(s, curveN) {
		return nil, errors.New("ecdsa: 서명의 s는 1 이상 n-1 이하여야 함")
	}
	return &Signature{r: new(big.Int).Set(r), s: new(big.Int).Set(s)}, nil
}

func inScalarRange(x, n *big.Int) bool {
	return x.Sign() > 0 && x.Cmp(n) < 0
}

func Verify(pub *secp256k1.Point, z [32]byte, sig *Signature) bool {
	zInt := new(big.Int).SetBytes(z[:])
	sInv := new(big.Int).ModInverse(sig.s, curveN)

	u1 := new(big.Int).Mul(zInt, sInv)
	u1.Mod(u1, curveN)
	u2 := new(big.Int).Mul(sig.r, sInv)
	u2.Mod(u2, curveN)

	rPoint := generator.ScalarMul(u1).Add(pub.ScalarMul(u2))
	rx, ok := rPoint.X()
	if !ok {
		return false
	}
	return new(big.Int).Mod(rx, curveN).Cmp(sig.r) == 0
}
