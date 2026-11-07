package secp256k1

import (
	"errors"
	"math/big"
)

var (
	curveN     = mustHexToBig("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141")
	curveB     = newFieldElem(big.NewInt(7))
	fieldTwo   = newFieldElem(big.NewInt(2))
	fieldThree = newFieldElem(big.NewInt(3))
	g          = mustNewPoint("79BE667EF9DCBBAC55A06295CE870B07029BFCDB2DCE28D959F2815B16F81798", "483ADA7726A3C4655DA4FBFC0E1108A8FD17B448A68554199C47D08FFB10D4B8")
)

type Point struct {
	x, y *fieldElem
	inf  bool
}

func NewPoint(x, y *big.Int) (*Point, error) {
	if x.Sign() < 0 || x.Cmp(fieldP) >= 0 || y.Sign() < 0 || y.Cmp(fieldP) >= 0 {
		return nil, errors.New("좌표가 0 이상 p 미만이 아님")
	}
	fx, fy := newFieldElem(x), newFieldElem(y)
	if !fy.Mul(fy).Equal(fx.Mul(fx.Mul(fx)).Add(curveB)) {
		return nil, errors.New("점이 곡선 위에 없음")
	}
	return &Point{x: fx, y: fy}, nil
}

func (p *Point) X() (*big.Int, bool) {
	if p.inf {
		return nil, false
	}
	return new(big.Int).Set(p.x.num), true
}

func (p *Point) Y() (*big.Int, bool) {
	if p.inf {
		return nil, false
	}
	return new(big.Int).Set(p.y.num), true
}

func (p *Point) Clone() *Point {
	if p.inf {
		return Infinity()
	}
	return &Point{x: p.x.Clone(), y: p.y.Clone()}
}

func (p *Point) Equal(q *Point) bool {
	if p.inf || q.inf {
		return p.inf == q.inf
	}
	return p.x.Equal(q.x) && p.y.Equal(q.y)
}

func (p *Point) Add(q *Point) *Point {
	switch {
	case p.inf:
		return q.Clone()
	case q.inf:
		return p.Clone()
	case !p.x.Equal(q.x):
		m, err := q.y.Sub(p.y).Div(q.x.Sub(p.x))
		if err != nil {
			panic("secp256k1: 버그 - 일반 덧셈에서 x1 == x2")
		}
		x3 := m.Mul(m).Sub(p.x).Sub(q.x)
		y3 := m.Mul(p.x.Sub(x3)).Sub(p.y)
		return &Point{x: x3, y: y3}
	case !p.y.Equal(q.y):
		return Infinity()
	default:
		m, err := p.x.Mul(p.x).Mul(fieldThree).Div(p.y.Mul(fieldTwo))
		if err != nil {
			panic("secp256k1: 버그 - 같은 점 덧셈에서 y1 * 2 == 0")
		}
		x3 := m.Mul(m).Sub(p.x.Mul(fieldTwo))
		y3 := m.Mul(p.x.Sub(x3)).Sub(p.y)
		return &Point{x: x3, y: y3}
	}
}

func (p *Point) ScalarMul(k *big.Int) *Point {
	kModN := new(big.Int).Mod(k, curveN)
	r := Infinity()
	for i := kModN.BitLen() - 1; i >= 0; i-- {
		r = r.Add(r)
		if kModN.Bit(i) == 1 {
			r = r.Add(p)
		}
	}
	return r
}

func (p *Point) IsInfinity() bool { return p.inf }

func Infinity() *Point { return &Point{inf: true} }

func mustNewPoint(xHex, yHex string) *Point {
	p, err := NewPoint(mustHexToBig(xHex), mustHexToBig(yHex))
	if err != nil {
		panic("secp256k1: 잘못된 상수 점: " + err.Error())
	}
	return p
}

func G() *Point { return g.Clone() }

func N() *big.Int { return new(big.Int).Set(curveN) }
