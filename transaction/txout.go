package transaction

import (
	"encoding/binary"
	"fmt"
	"io"
)

type TxOut struct {
	amount       uint64
	scriptPubKey *Script
}

func NewTxOut(amount uint64, scriptPubKey *Script) *TxOut {
	return &TxOut{
		amount:       amount,
		scriptPubKey: scriptPubKey.Clone(),
	}
}

func (p *TxOut) Amount() uint64 {
	return p.amount
}

func (p *TxOut) SetAmount(amount uint64) {
	p.amount = amount
}

func (p *TxOut) ScriptPubKey() *Script {
	return p.scriptPubKey.Clone()
}

func (p *TxOut) SetScriptPubKey(scriptPubKey *Script) {
	p.scriptPubKey = scriptPubKey.Clone()
}

func (p *TxOut) Clone() *TxOut {
	return &TxOut{
		amount:       p.amount,
		scriptPubKey: p.scriptPubKey.Clone(),
	}
}

func (p *TxOut) String() string {
	return fmt.Sprintf("TxOut(\namount: %x\nscriptPubKey: %s\n)", p.amount, p.scriptPubKey)
}

func (p *TxOut) Serialize() []byte {
	var b []byte
	b = binary.LittleEndian.AppendUint64(b, p.amount)
	return append(b, p.scriptPubKey.Serialize()...)
}

func ParseTxOut(r io.Reader) (*TxOut, error) {
	var b [8]byte
	if _, err := io.ReadFull(r, b[:]); err != nil {
		return nil, fmt.Errorf("금액 읽기 실패: %w", err)
	}
	amount := binary.LittleEndian.Uint64(b[:])
	scriptPubKey, err := ParseScript(r)
	if err != nil {
		return nil, fmt.Errorf("잠금 스크립트 읽기 실패: %w", err)
	}
	return &TxOut{
		amount:       amount,
		scriptPubKey: scriptPubKey,
	}, nil
}
