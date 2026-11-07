package secp256k1

import (
	"errors"
	"math/big"
)

var (
	fieldP       = mustHexToBig("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEFFFFFC2F")
	fieldPMinus1 = new(big.Int).Sub(fieldP, big.NewInt(1))
)

type fieldElem struct {
	num *big.Int
}

func newFieldElem(n *big.Int) *fieldElem {
	return &fieldElem{num: new(big.Int).Mod(n, fieldP)}
}

func (p *fieldElem) Clone() *fieldElem {
	return &fieldElem{num: new(big.Int).Set(p.num)}
}

func (p *fieldElem) Equal(q *fieldElem) bool {
	return p.num.Cmp(q.num) == 0
}

func (p *fieldElem) Add(q *fieldElem) *fieldElem {
	r := new(big.Int).Add(p.num, q.num)
	r.Mod(r, fieldP)
	return &fieldElem{num: r}
}

func (p *fieldElem) Sub(q *fieldElem) *fieldElem {
	r := new(big.Int).Sub(p.num, q.num)
	r.Mod(r, fieldP)
	return &fieldElem{num: r}
}

func (p *fieldElem) Mul(q *fieldElem) *fieldElem {
	r := new(big.Int).Mul(p.num, q.num)
	r.Mod(r, fieldP)
	return &fieldElem{num: r}
}

func (p *fieldElem) Pow(e *big.Int) *fieldElem {
	if p.num.Sign() == 0 {
		return &fieldElem{num: new(big.Int)}
	}
	reducedE := new(big.Int).Mod(e, fieldPMinus1)
	r := new(big.Int).Exp(p.num, reducedE, fieldP)
	return &fieldElem{num: r}
}

func (p *fieldElem) Div(q *fieldElem) (*fieldElem, error) {
	qInv := new(big.Int).ModInverse(q.num, fieldP)
	if qInv == nil {
		return nil, errors.New("0으로 나눌 수 없음")
	}
	r := new(big.Int).Mul(p.num, qInv)
	r.Mod(r, fieldP)
	return &fieldElem{num: r}, nil
}

func (p *fieldElem) Sqrt() (*fieldElem, bool) {
	r := new(big.Int).ModSqrt(p.num, fieldP)
	if r == nil {
		return nil, false
	}
	return &fieldElem{num: r}, true
}

func mustHexToBig(s string) *big.Int {
	n, ok := new(big.Int).SetString(s, 16)
	if !ok {
		panic("잘못된 16진수 문자열: " + s)
	}
	return n
}
