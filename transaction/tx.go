package transaction

// 1. 함수에 입력이 하나라도 있다면 error 반환하기
// 2. 구조체 변수를 복사할 때는 무조건 Clone 사용하기
// 3. (p Tx) 대신 (p *Tx) 사용하기
// 4. 함수의 매개변수나 반환값으로 구조체 변수를 줄 때는 무조건 포인터 쓰기
// 5. getter의 반환값으로 슬라이스를 줄 때 무조건 copy한 복사본을 반환하기

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"slices"
	"strings"
)

type Tx struct {
	version  uint32
	txIns    []*TxIn
	txOuts   []*TxOut
	locktime uint32
}

func NewTx(version uint32, txIns []*TxIn, txOuts []*TxOut, locktime uint32) (*Tx, error) {
	var result Tx

	result.version = version

	result.txIns = make([]*TxIn, len(txIns))
	for i := range txIns {
		result.txIns[i] = txIns[i].Clone()
	}

	result.txOuts = make([]*TxOut, len(txOuts))
	for i := range txOuts {
		result.txOuts[i] = txOuts[i].Clone()
	}

	result.locktime = locktime

	return &result, nil
}

func (p *Tx) Version() uint32 {
	return p.version
}
func (p *Tx) SetVersion(version uint32) error {
	p.version = version

	return nil
}

func (p *Tx) TxIns() []*TxIn {
	var txIns []*TxIn = make([]*TxIn, len(p.txIns))
	for i := range p.txIns {
		txIns[i] = p.txIns[i].Clone()
	}

	return txIns
}
func (p *Tx) SetTxIns(txIns []*TxIn) error {
	p.txIns = make([]*TxIn, len(txIns))
	for i := range txIns {
		p.txIns[i] = txIns[i].Clone()
	}

	return nil
}

func (p *Tx) TxOuts() []*TxOut {
	var txOuts []*TxOut = make([]*TxOut, len(p.txOuts))
	for i := range p.txOuts {
		txOuts[i] = p.txOuts[i].Clone()
	}

	return txOuts
}
func (p *Tx) SetTxOuts(txOuts []*TxOut) error {
	p.txOuts = make([]*TxOut, len(txOuts))
	for i := range txOuts {
		p.txOuts[i] = txOuts[i].Clone()
	}

	return nil
}

func (p *Tx) Locktime() uint32 {
	return p.locktime
}
func (p *Tx) SetLocktime(locktime uint32) error {
	p.locktime = locktime

	return nil
}

func (p *Tx) Clone() *Tx {
	var result Tx

	result.version = p.version

	result.txIns = make([]*TxIn, len(p.txIns))
	for i := range p.txIns {
		result.txIns[i] = p.txIns[i].Clone()
	}

	result.txOuts = make([]*TxOut, len(p.txOuts))
	for i := range p.txOuts {
		result.txOuts[i] = p.txOuts[i].Clone()
	}

	result.locktime = p.locktime

	return &result
}

func (p *Tx) String() string {
	var (
		txInsStr  strings.Builder
		txOutsStr strings.Builder
	)

	for i := range p.txIns {
		txInsStr.WriteString(p.txIns[i].String())
	}

	for i := range p.txOuts {
		txOutsStr.WriteString(p.txOuts[i].String())
	}

	return fmt.Sprintf("Tx(\nversion: %08x\ntxIns: \n%s\ntxOuts: \n%s\nlocktime: %08x\n)", p.version, txInsStr.String(), txOutsStr.String(), p.locktime)
}

func (p *Tx) Id() []byte {
	var (
		firHash [32]byte
		secHash [32]byte
	)

	firHash = sha256.Sum256(p.Serialize())
	secHash = sha256.Sum256(firHash[:])

	slices.Reverse(secHash[:])

	return secHash[:]
}

func (p *Tx) Fee(txFetcher *TxFetcher) (uint64, error) {
	var (
		totalValueOfTxIns  uint64
		valueOfTxIn        uint64
		totalValueOfTxOuts uint64
		err                error
	)

	for i := range p.txIns {
		if valueOfTxIn, err = p.txIns[i].Value(txFetcher); err != nil {
			return 0xffffffffffffffff, err
		} else {
			totalValueOfTxIns += valueOfTxIn
		}
	}

	for i := range p.txOuts {
		totalValueOfTxOuts += p.txOuts[i].Amount()
	}

	return totalValueOfTxIns - totalValueOfTxOuts, nil
}

// func (p *Tx) GetMessage(txFetcher *TxFetcher) ([]byte, error) {
// 	tmp := p.Clone()

// 	tt := []*Script
// 	for i := range tmp.TxIns() {
// 		if prevTx, err := txFetcher.FetchTx(hex.EncodeToString(tmp.txIns[i].prevTxId)); err != nil {
// 			return nil, err
// 		} else {
// 			tmp.SetTxIns() prevTx.TxOuts()[tmp.TxIns()[i].PrevTxOutIdx()].ScriptPubKey()
// 		}
// 	}
// }

// func (p *Tx) GetMessage() []byte {

// }

func (p *Tx) Serialize() []byte {
	var (
		buf       bytes.Buffer
		txInsLen  []byte
		txOutsLen []byte
	)

	binary.Write(&buf, binary.LittleEndian, p.version)

	txInsLen, _ = EncodeCVarint(uint64(len(p.txIns)))
	buf.Write(txInsLen)
	for i := range p.txIns {
		buf.Write(p.txIns[i].Serialize())
	}

	txOutsLen, _ = EncodeCVarint(uint64(len(p.txOuts)))
	buf.Write(txOutsLen)
	for i := range p.txOuts {
		buf.Write(p.txOuts[i].Serialize())
	}

	binary.Write(&buf, binary.LittleEndian, p.locktime)

	return buf.Bytes()
}

func ParseTx(buf *bytes.Buffer) (*Tx, error) {
	var (
		bVersion  []byte
		version   uint32
		txInsLen  uint64
		txIns     []*TxIn
		txOutsLen uint64
		txOuts    []*TxOut
		bLocktime []byte
		locktime  uint32
		n         int
		err       error
	)

	bVersion = make([]byte, 4)
	if n, err = io.ReadFull(buf, bVersion); n < len(bVersion) || err != nil {
		return nil, err
	} else {
		version = binary.LittleEndian.Uint32(bVersion)
	}

	if txInsLen, err = DecodeCvarint(buf); err != nil {
		return nil, err
	}

	txIns = make([]*TxIn, txInsLen)
	for i := range txIns {
		if txIns[i], err = ParseTxIn(buf); err != nil {
			return nil, err
		}
	}

	if txOutsLen, err = DecodeCvarint(buf); err != nil {
		return nil, err
	}

	txOuts = make([]*TxOut, txOutsLen)
	for i := range txOuts {
		if txOuts[i], err = ParseTxOut(buf); err != nil {
			return nil, err
		}
	}

	bLocktime = make([]byte, 4)
	if n, err = io.ReadFull(buf, bLocktime); n < len(bLocktime) || err != nil {
		return nil, err
	} else {
		locktime = binary.LittleEndian.Uint32(bLocktime)
	}

	return &Tx{
		version:  version,
		txIns:    txIns,
		txOuts:   txOuts,
		locktime: locktime,
	}, nil
}
