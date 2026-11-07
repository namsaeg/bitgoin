package ecdsa

import (
	"errors"
	"math/big"

	"bitgoin_v1/secp256k1"
)

var (
	curveN    = secp256k1.N()
	generator = secp256k1.G()
)

type PrivateKey struct {
	d   *big.Int
	pub *secp256k1.Point
}

func NewPrivateKey(d *big.Int) (*PrivateKey, error) {
	if d.Sign() <= 0 || d.Cmp(curveN) >= 0 {
		return nil, errors.New("ecdsa: 비밀키는 1 이상 n-1 이하여야 함")
	}
	return &PrivateKey{
		d:   new(big.Int).Set(d),
		pub: generator.ScalarMul(d),
	}, nil
}

func (key *PrivateKey) Bytes() [32]byte {
	var b [32]byte
	key.d.FillBytes(b[:])
	return b
}

func (p *PrivateKey) PublicKey() *secp256k1.Point {
	return p.pub.Clone()
}

func (p *PrivateKey) Sign(z [32]byte) *Signature {
	k := nonceRFC6979(p.d, z)

	rx, _ := generator.ScalarMul(k).X()
	r := new(big.Int).Mod(rx, curveN)

	kInv := new(big.Int).ModInverse(k, curveN)
	s := new(big.Int).Mul(r, p.d)
	s.Add(s, new(big.Int).SetBytes(z[:]))
	s.Mul(s, kInv)
	s.Mod(s, curveN)

	if s.Cmp(new(big.Int).Rsh(curveN, 1)) > 0 {
		s.Sub(curveN, s)
	}
	return &Signature{r: r, s: s}
}
