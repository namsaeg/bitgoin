package secp256k1

import (
	"errors"
	"math/big"
)

func (p *Point) SEC(compressed bool) []byte {
	if p.inf {
		panic("secp256k1: 무한원점을 SEC로 직렬화할 수 없음")
	}
	if compressed {
		b := make([]byte, 33)
		b[0] = 0x02 | byte(p.y.num.Bit(0))
		p.x.num.FillBytes(b[1:])
		return b
	}
	b := make([]byte, 65)
	b[0] = 0x04
	p.x.num.FillBytes(b[1:33])
	p.y.num.FillBytes(b[33:])
	return b
}

func ParseSEC(b []byte) (*Point, error) {
	switch {
	case len(b) == 65 && b[0] == 0x04:
		x := new(big.Int).SetBytes(b[1:33])
		y := new(big.Int).SetBytes(b[33:])
		return NewPoint(x, y)

	case len(b) == 33 && (b[0] == 0x02 || b[0] == 0x03):
		x := new(big.Int).SetBytes(b[1:])
		if x.Cmp(fieldP) >= 0 {
			return nil, errors.New("secp256k1: x 좌표가 p 이상")
		}
		fx := newFieldElem(x)
		fy, ok := fx.Mul(fx).Mul(fx).Add(curveB).Sqrt()
		if !ok {
			return nil, errors.New("secp256k1: x에 해당하는 점이 곡선 위에 없음")
		}
		if fy.num.Bit(0) != uint(b[0]&1) {
			fy = newFieldElem(new(big.Int).Sub(fieldP, fy.num))
		}
		return &Point{x: fx, y: fy}, nil

	default:
		return nil, errors.New("secp256k1: SEC 형식이 아님(33바이트 02/03 또는 65바이트 04)")
	}
}
