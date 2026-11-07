package transaction

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/asn1"
	"errors"
	"fmt"
	"math/big"
)

// secp256k1Curve는 비트코인 곡선 y² = x³ + 7을 elliptic.Curve로 구현한다.
// elliptic.CurveParams의 기본 연산은 a = -3인 곡선만 맞으므로, 연산 메서드를 모두 직접 정의한다.
type secp256k1Curve struct {
	*elliptic.CurveParams
}

var secp256k1 = &secp256k1Curve{&elliptic.CurveParams{
	P:       hexToBig("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEFFFFFC2F"),
	N:       hexToBig("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141"),
	B:       big.NewInt(7),
	Gx:      hexToBig("79BE667EF9DCBBAC55A06295CE870B07029BFCDB2DCE28D959F2815B16F81798"),
	Gy:      hexToBig("483ADA7726A3C4655DA4FBFC0E1108A8FD17B448A68554199C47D08FFB10D4B8"),
	BitSize: 256,
	Name:    "secp256k1",
}}

// S256은 secp256k1 곡선을 반환한다.
func S256() elliptic.Curve { return secp256k1 }

func hexToBig(s string) *big.Int {
	n, ok := new(big.Int).SetString(s, 16)
	if !ok {
		panic("잘못된 16진수: " + s)
	}
	return n
}

// IsOnCurve는 (x, y)가 y² = x³ + 7 (mod p)를 만족하는지 확인한다.
func (c *secp256k1Curve) IsOnCurve(x, y *big.Int) bool {
	if x.Sign() < 0 || x.Cmp(c.P) >= 0 || y.Sign() < 0 || y.Cmp(c.P) >= 0 {
		return false
	}
	lhs := new(big.Int).Mul(y, y)
	lhs.Mod(lhs, c.P)
	return lhs.Cmp(c.rhs(x)) == 0
}

// rhs는 x³ + 7 (mod p)를 계산한다.
func (c *secp256k1Curve) rhs(x *big.Int) *big.Int {
	r := new(big.Int).Exp(x, big.NewInt(3), c.P)
	r.Add(r, c.B)
	return r.Mod(r, c.P)
}

// 무한원점은 elliptic 패키지 관례대로 (0, 0)으로 나타낸다.
func isInf(x, y *big.Int) bool { return x.Sign() == 0 && y.Sign() == 0 }

// Add는 두 점의 합을 반환한다.
func (c *secp256k1Curve) Add(x1, y1, x2, y2 *big.Int) (*big.Int, *big.Int) {
	switch {
	case isInf(x1, y1):
		return new(big.Int).Set(x2), new(big.Int).Set(y2)
	case isInf(x2, y2):
		return new(big.Int).Set(x1), new(big.Int).Set(y1)
	case x1.Cmp(x2) == 0:
		if y1.Cmp(y2) == 0 {
			return c.Double(x1, y1)
		}
		return new(big.Int), new(big.Int) // P + (-P) = 무한원점
	}
	// λ = (y2 - y1) / (x2 - x1)
	num := new(big.Int).Sub(y2, y1)
	den := new(big.Int).Sub(x2, x1)
	den.ModInverse(den.Mod(den, c.P), c.P)
	lambda := num.Mul(num, den)
	lambda.Mod(lambda, c.P)
	return c.fromLambda(lambda, x1, y1, x2)
}

// Double은 점을 두 배 한다.
func (c *secp256k1Curve) Double(x, y *big.Int) (*big.Int, *big.Int) {
	if isInf(x, y) || y.Sign() == 0 {
		return new(big.Int), new(big.Int)
	}
	// λ = 3x² / 2y   (a = 0이라 분자에 a가 더해지지 않음)
	num := new(big.Int).Mul(x, x)
	num.Mul(num, big.NewInt(3))
	den := new(big.Int).Lsh(y, 1)
	den.ModInverse(den.Mod(den, c.P), c.P)
	lambda := num.Mul(num, den)
	lambda.Mod(lambda, c.P)
	return c.fromLambda(lambda, x, y, x)
}

// fromLambda는 기울기 λ로 x3 = λ² - x1 - x2, y3 = λ(x1 - x3) - y1을 계산한다.
func (c *secp256k1Curve) fromLambda(lambda, x1, y1, x2 *big.Int) (*big.Int, *big.Int) {
	x3 := new(big.Int).Mul(lambda, lambda)
	x3.Sub(x3, x1).Sub(x3, x2).Mod(x3, c.P)
	y3 := new(big.Int).Sub(x1, x3)
	y3.Mul(y3, lambda).Sub(y3, y1).Mod(y3, c.P)
	return x3, y3
}

// ScalarMult는 k·(bx, by)를 double-and-add로 계산한다. k는 빅엔디언 바이트다.
// 공부용 구현이라 상수 시간이 아니다. 공개된 값으로 서명을 검증할 때만 쓴다.
func (c *secp256k1Curve) ScalarMult(bx, by *big.Int, k []byte) (*big.Int, *big.Int) {
	x, y := new(big.Int), new(big.Int)
	for _, b := range k {
		for i := 7; i >= 0; i-- {
			x, y = c.Double(x, y)
			if b>>uint(i)&1 == 1 {
				x, y = c.Add(x, y, bx, by)
			}
		}
	}
	return x, y
}

// ScalarBaseMult는 k·G를 계산한다.
func (c *secp256k1Curve) ScalarBaseMult(k []byte) (*big.Int, *big.Int) {
	return c.ScalarMult(c.Gx, c.Gy, k)
}

// ParseSECPubKey는 SEC 형식으로 직렬화된 secp256k1 공개키를 읽어 ecdsa.PublicKey로 반환한다.
// 압축(02/03 + x, 33바이트)과 비압축(04 + x + y, 65바이트) 형식을 받으며,
// 점이 곡선 위에 있는지 확인한다.
func ParseSECPubKey(b []byte) (*ecdsa.PublicKey, error) {
	c := secp256k1
	switch {
	case len(b) == 65 && b[0] == 0x04:
		x := new(big.Int).SetBytes(b[1:33])
		y := new(big.Int).SetBytes(b[33:])
		if !c.IsOnCurve(x, y) {
			return nil, errors.New("공개키가 곡선 위의 점이 아님")
		}
		return &ecdsa.PublicKey{Curve: c, X: x, Y: y}, nil

	case len(b) == 33 && (b[0] == 0x02 || b[0] == 0x03):
		x := new(big.Int).SetBytes(b[1:])
		if x.Cmp(c.P) >= 0 {
			return nil, errors.New("x 좌표가 p 이상")
		}
		y := new(big.Int).ModSqrt(c.rhs(x), c.P) // y² = x³ + 7의 제곱근
		if y == nil {
			return nil, errors.New("공개키가 곡선 위의 점이 아님")
		}
		if y.Bit(0) != uint(b[0]&1) { // 02 → 짝수 y, 03 → 홀수 y
			y.Sub(c.P, y)
		}
		return &ecdsa.PublicKey{Curve: c, X: x, Y: y}, nil

	default:
		return nil, errors.New("SEC 공개키 형식이 아님(33바이트 02/03 또는 65바이트 04)")
	}
}

// ParseDERSignature는 DER로 인코딩된 ECDSA 서명(sighash 바이트 제외)에서 r, s를 꺼낸다.
func ParseDERSignature(der []byte) (r, s *big.Int, err error) {
	var sig struct{ R, S *big.Int }
	rest, err := asn1.Unmarshal(der, &sig)
	if err != nil {
		return nil, nil, fmt.Errorf("DER 파싱 실패: %w", err)
	}
	if len(rest) != 0 { // Unmarshal은 뒤에 남은 바이트를 에러 없이 돌려준다
		return nil, nil, fmt.Errorf("서명 뒤에 %d바이트가 남음", len(rest))
	}
	if sig.R.Sign() <= 0 || sig.S.Sign() <= 0 { // 음수와 0도 에러 없이 통과한다
		return nil, nil, errors.New("r, s는 양수여야 함")
	}
	return sig.R, sig.S, nil
}
