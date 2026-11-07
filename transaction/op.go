package transaction

import (
	"crypto/ecdsa"
	"fmt"
)

func encodeNum(n int) []byte {
	if n == 0 {
		return []byte{}
	}

	var absNum int
	if n < 0 {
		absNum = -n
	} else {
		absNum = n
	}
	var b []byte
	for absNum > 0 {
		b = append(b, byte(absNum&0xff))
		absNum >>= 8
	}

	if b[len(b)-1]&0x80 == 1 {
		if n > 0 {
			b = append(b, byte(0x00))
		} else {
			b = append(b, byte(0x80))
		}
	} else if n < 0 {
		b[len(b)-1] |= 0x80
	}

	return b
}

func decodeNum(b []byte) int {
	if len(b) == 0 {
		return 0
	}

	var isNegative bool
	var n int
	if b[len(b)-1]&0x80 == 1 {
		isNegative = true
		n = int(b[len(b)-1]) & 0x7f
	} else {
		isNegative = false
		n = int(b[len(b)-1])
	}
	for i := 1; i < len(b); i++ {
		n <<= 8
		n += int(b[i])
	}

	if isNegative {
		return -n
	} else {
		return n
	}
}

func OpCheckSig(stack *ScriptStack, tx *Tx, inIdx uint64, f *TxFetcher) error {
	if stack.Len() < 2 {
		return fmt.Errorf("스택에 원소가 2개 이상 있지 않음")
	}

	pubKeyData, err := stack.Pop()
	if err != nil {
		return fmt.Errorf("스택에서 잠금 스크립트 데이터 뽑기 실패: %w", err)
	}
	pubKey, err := ParseSECPubKey(pubKeyData.Data())
	if err != nil {
		return fmt.Errorf("SEC 형식에서 공개키 추출 실패: %w", err)
	}
	sigData, err := stack.Pop()
	if err != nil {
		return fmt.Errorf("스택에서 서명 데이터 뽑기 실패: %w", err)
	}
	r, s, err := ParseDERSignature(sigData.Data()[:len(sigData.Data())-1])
	if err != nil {
		return fmt.Errorf("DER 형식에서 서명 추출 실패: %w", err)
	}
	sigHash, err := tx.GetSigHash(inIdx, sigData.Data()[len(sigData.Data())-1], f)
	if err != nil {
		return fmt.Errorf("서명해시 계산 실패: %w", err)
	}
	fmt.Println(ecdsa.Verify(pubKey, sigHash[:], r, s))
	return nil
}
