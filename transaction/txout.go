package transaction

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

type TxOut struct {
	amount       uint64
	scriptPubKey *Script
}

func NewTxOut(amount uint64, scriptPubKey *Script) (*TxOut, error) {
	var result TxOut

	result.amount = amount

	result.scriptPubKey = scriptPubKey.Clone()

	return &result, nil
}

func (p *TxOut) Amount() uint64 {
	return p.amount
}
func (p *TxOut) SetAmount(amount uint64) error {
	p.amount = amount

	return nil
}

func (p *TxOut) ScriptPubKey() *Script {
	return p.scriptPubKey.Clone()
}
func (p *TxOut) SetScriptPubKey(scriptPubKey *Script) error {
	p.scriptPubKey = scriptPubKey.Clone()

	return nil
}

func (p *TxOut) Clone() *TxOut {
	var result TxOut

	result.amount = p.amount

	result.scriptPubKey = p.scriptPubKey.Clone()

	return &result
}

func (p *TxOut) String() string {
	return fmt.Sprintf("TxOut(\namount: %016x\nscriptPubKey: %s\n)", p.amount, p.scriptPubKey.String())
}

func (p *TxOut) Serialize() []byte {
	var buf bytes.Buffer

	binary.Write(&buf, binary.LittleEndian, p.amount)

	buf.Write(p.scriptPubKey.Serialize())

	return buf.Bytes()
}

func ParseTxOut(buf *bytes.Buffer) (*TxOut, error) {
	var (
		bAmount      []byte
		amount       uint64
		scriptPubKey *Script
		n            int
		err          error
	)

	bAmount = make([]byte, 8)
	if n, err = io.ReadFull(buf, bAmount); n < len(bAmount) || err != nil {
		return nil, err
	} else {
		amount = binary.LittleEndian.Uint64(bAmount)
	}

	if scriptPubKey, err = ParseScript(buf); err != nil {
		return nil, err
	}

	return &TxOut{
		amount:       amount,
		scriptPubKey: scriptPubKey,
	}, nil
}
