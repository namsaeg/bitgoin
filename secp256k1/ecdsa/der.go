package ecdsa

import (
	"errors"
	"fmt"
	"math/big"

	"bitgoin_v1/secp256k1"
)

var halfN = new(big.Int).Rsh(secp256k1.N(), 1)

func (p *Signature) IsLowS() bool {
	return p.s.Cmp(halfN) <= 0
}

func (p *Signature) DER() []byte {
	r, s := derInt(p.r), derInt(p.s)
	b := make([]byte, 0, 6+len(r)+len(s))
	b = append(b, 0x30, byte(4+len(r)+len(s)))
	b = append(b, 0x02, byte(len(r)))
	b = append(b, r...)
	b = append(b, 0x02, byte(len(s)))
	return append(b, s...)
}

func derInt(x *big.Int) []byte {
	b := x.Bytes()
	if b[0]&0x80 != 0 {
		b = append([]byte{0x00}, b...)
	}
	return b
}

func ParseDER(b []byte) (*Signature, error) {
	if len(b) < 8 || len(b) > 72 {
		return nil, fmt.Errorf("ecdsa: DER 서명 길이 %d가 8~72바이트가 아님", len(b))
	}
	if b[0] != 0x30 {
		return nil, errors.New("ecdsa: DER 서명이 SEQUENCE(0x30)로 시작하지 않음")
	}
	if int(b[1]) != len(b)-2 {
		return nil, errors.New("ecdsa: DER SEQUENCE 길이가 실제 길이와 다름")
	}
	r, rest, err := parseDERInt(b[2:])
	if err != nil {
		return nil, fmt.Errorf("ecdsa: DER r: %w", err)
	}
	s, rest, err := parseDERInt(rest)
	if err != nil {
		return nil, fmt.Errorf("ecdsa: DER s: %w", err)
	}
	if len(rest) != 0 {
		return nil, errors.New("ecdsa: DER 서명 뒤에 남은 바이트가 있음")
	}
	return NewSignature(r, s)
}

func parseDERInt(b []byte) (*big.Int, []byte, error) {
	if len(b) < 2 || b[0] != 0x02 {
		return nil, nil, errors.New("INTEGER 태그(0x02)가 아님")
	}
	l := int(b[1])
	if l == 0 {
		return nil, nil, errors.New("길이가 0")
	}
	if len(b)-2 < l {
		return nil, nil, errors.New("값이 길이보다 짧음")
	}
	v := b[2 : 2+l]
	if v[0]&0x80 != 0 {
		return nil, nil, errors.New("음수")
	}
	if l > 1 && v[0] == 0x00 && v[1]&0x80 == 0 {
		return nil, nil, errors.New("불필요한 0x00 패딩")
	}
	return new(big.Int).SetBytes(v), b[2+l:], nil
}
