package transaction

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"math/bits"
	"slices"
	"strings"
)

// const (
// 	minTxInCnt  = 1
// 	maxTxInCnt  = 100
// 	minTxOutCnt = 1
// 	maxTxOutCnt = 100
// )

type Tx struct {
	version  uint32
	txIns    []*TxIn
	txOuts   []*TxOut
	locktime uint32
}

func NewTx(version uint32, txIns []*TxIn, txOuts []*TxOut, locktime uint32) *Tx {
	return &Tx{
		version:  version,
		txIns:    cloneAll(txIns),
		txOuts:   cloneAll(txOuts),
		locktime: locktime,
	}
}

func (p *Tx) Version() uint32 {
	return p.version
}

func (p *Tx) SetVersion(version uint32) {
	p.version = version
}

func (p *Tx) TxIns() []*TxIn {
	return cloneAll(p.txIns)
}

func (p *Tx) SetTxIns(txIns []*TxIn) {
	p.txIns = cloneAll(txIns)
}

func (p *Tx) TxOuts() []*TxOut {
	return cloneAll(p.txOuts)
}

func (p *Tx) SetTxOuts(txOuts []*TxOut) {
	p.txOuts = cloneAll(txOuts)
}

func (p *Tx) Locktime() uint32 {
	return p.locktime
}

func (p *Tx) SetLocktime(locktime uint32) {
	p.locktime = locktime
}

func (p *Tx) Clone() *Tx {
	return &Tx{
		version:  p.version,
		txIns:    cloneAll(p.txIns),
		txOuts:   cloneAll(p.txOuts),
		locktime: p.locktime,
	}
}

func (p *Tx) TxOut(i uint64) *TxOut {
	return p.txOuts[i].Clone()
}

func (p *Tx) Id() [32]byte {
	fir := sha256.Sum256(p.Serialize())
	sec := sha256.Sum256(fir[:])
	slices.Reverse(sec[:])
	return sec
}

func (p *Tx) Fee(f *TxFetcher) (uint64, error) {
	var valOfIns uint64 = 0
	for i, in := range p.txIns {
		v, err := in.Value(f)
		if err != nil {
			return 0, fmt.Errorf("%d번 입력이 가리키는 출력의 금액 가져오기 실패: %w", i, err)
		}
		valOfIns, err = addAmount(valOfIns, v)
		if err != nil {
			return 0, fmt.Errorf("입력 전체의 금액 서로 더하기 실패: %w", err)
		}
	}
	var valOfOuts uint64 = 0
	for _, out := range p.txOuts {
		var err error
		valOfOuts, err = addAmount(valOfOuts, out.Amount())
		if err != nil {
			return 0, fmt.Errorf("출력 전체의 금액 서로 더하기 실패: %w", err)
		}
	}
	fee := valOfIns - valOfOuts
	if valOfIns < valOfOuts {
		return 0, fmt.Errorf("입력 전체의 금액이 출력 전체의 금액보다 작음")
	}
	return fee, nil
}

func (p *Tx) GetSigHash(inIdx uint64, sighash byte, f *TxFetcher) ([]byte, error) {
	if inIdx >= uint64(len(p.txIns)) {
		return nil, fmt.Errorf("inIdx가 전체 입력 슬라이스의 인덱스 범위를 벗어났음")
	}
	tmpTx := p.Clone()
	for i, in := range tmpTx.txIns {
		in.scriptSig = nil
		if i != int(inIdx) {
			continue
		}
		prevTx, err := f.FetchTx(in.prevTxId)
		if err != nil {
			return nil, fmt.Errorf("%d번 입력의 이전 트랜잭션 가져오기 실패", i)
		}
		in.scriptSig = prevTx.TxOut(uint64(in.prevOutIdx)).scriptPubKey
	}
	s := append(tmpTx.Serialize(), sighash)
	fir := sha256.Sum256(s)
	sec := sha256.Sum256(fir[:])
	return sec[:], nil
}

func (p *Tx) String() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Tx(\nversion: %d\ntxIns:\n", p.version)
	for _, in := range p.txIns {
		sb.WriteString(in.String())
		sb.WriteByte('\n')
	}
	sb.WriteString("txOuts:\n")
	for _, out := range p.txOuts {
		sb.WriteString(out.String())
		sb.WriteByte('\n')
	}
	fmt.Fprintf(&sb, "locktime: %d\n)", p.locktime)
	return sb.String()
}

func (p *Tx) Serialize() []byte {
	b := binary.LittleEndian.AppendUint32(nil, p.version)
	b = append(b, EncodeCVarint(uint64(len(p.txIns)))...)
	for _, in := range p.txIns {
		b = append(b, in.Serialize()...)
	}
	b = append(b, EncodeCVarint(uint64(len(p.txOuts)))...)
	for _, out := range p.txOuts {
		b = append(b, out.Serialize()...)
	}
	return binary.LittleEndian.AppendUint32(b, p.locktime)
}

func ParseTx(r io.Reader) (*Tx, error) {
	var tx Tx

	var b [4]byte
	if _, err := io.ReadFull(r, b[:]); err != nil {
		return nil, fmt.Errorf("버전 읽기 실패: %w", err)
	}
	tx.version = binary.LittleEndian.Uint32(b[:])

	inCnt, err := DecodeCVarint(r)
	if err != nil {
		return nil, fmt.Errorf("입력 개수 읽기 실패: %w", err)
	}
	tx.txIns = make([]*TxIn, inCnt)
	for i := uint64(0); i < inCnt; i++ {
		txIn, err := ParseTxIn(r)
		if err != nil {
			return nil, fmt.Errorf("입력(index %d) 읽기 실패: %w", i, err)
		}
		tx.txIns[i] = txIn
	}

	outCnt, err := DecodeCVarint(r)
	if err != nil {
		return nil, fmt.Errorf("출력 개수 읽기 실패: %w", err)
	}
	tx.txOuts = make([]*TxOut, outCnt)
	for i := uint64(0); i < outCnt; i++ {
		txOut, err := ParseTxOut(r)
		if err != nil {
			return nil, fmt.Errorf("출력(index %d) 읽기 실패: %w", i, err)
		}
		tx.txOuts[i] = txOut
	}

	if _, err := io.ReadFull(r, b[:]); err != nil {
		return nil, fmt.Errorf("록타임 읽기 실패: %w", err)
	}
	tx.locktime = binary.LittleEndian.Uint32(b[:])

	return &tx, nil
}

func addAmount(a, b uint64) (uint64, error) {
	sum, carry := bits.Add64(a, b, 0)
	if carry != 0 {
		return 0, fmt.Errorf("오버플로우 발생")
	}
	return sum, nil
}
