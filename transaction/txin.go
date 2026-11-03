package transaction

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"slices"
)

type TxIn struct {
	prevTxId     []byte
	prevTxOutIdx uint32
	scriptSig    *Script
	sequence     uint32
}

func NewTxIn(prevTxId []byte, prevTxOutIdx uint32, scriptSig *Script, sequence uint32) (*TxIn, error) {
	var result TxIn

	if len(prevTxId) != 32 {
		return nil, fmt.Errorf("wrong prevTxId")
	}

	result.prevTxId = make([]byte, len(prevTxId))
	copy(result.prevTxId, prevTxId)

	result.prevTxOutIdx = prevTxOutIdx

	result.scriptSig = scriptSig.Clone()

	result.sequence = sequence

	return &result, nil
}

func (p *TxIn) PrevTxID() []byte {
	var prevTxId []byte = make([]byte, len(p.prevTxId))
	copy(prevTxId, p.prevTxId)

	return prevTxId
}
func (p *TxIn) SetPrevTxID(prevTxId []byte) error {
	if len(prevTxId) != 32 {
		return fmt.Errorf("wrong prevTxId")
	}

	p.prevTxId = make([]byte, len(prevTxId))
	copy(p.prevTxId, prevTxId)

	return nil
}

func (p *TxIn) PrevTxOutIdx() uint32 {
	return p.prevTxOutIdx
}
func (p *TxIn) SetPrevTxOutIdx(prevTxOutIdx uint32) error {
	p.prevTxOutIdx = prevTxOutIdx

	return nil
}

func (p *TxIn) ScriptSig() *Script {
	return p.scriptSig.Clone()
}
func (p *TxIn) SetScriptSIg(scriptSig *Script) error {
	p.scriptSig = scriptSig.Clone()

	return nil
}

func (p *TxIn) Sequence() uint32 {
	return p.sequence
}
func (p *TxIn) SetSequence(sequence uint32) error {
	p.sequence = sequence

	return nil
}

func (p *TxIn) Clone() *TxIn {
	var result TxIn

	result.prevTxId = make([]byte, len(p.prevTxId))
	copy(result.prevTxId, p.prevTxId)

	result.prevTxOutIdx = p.prevTxOutIdx

	result.scriptSig = p.scriptSig.Clone()

	result.sequence = p.sequence

	return &result
}

func (p *TxIn) String() string {
	return fmt.Sprintf("TxIn(\nprevTxId: %064x\nprevTxOutIdx: %08x\nscriptSig: %s\nsequence: %08x\n)", p.prevTxId, p.prevTxOutIdx, p.scriptSig.String(), p.sequence)
}

func (p *TxIn) Value(txFetcher *TxFetcher) (uint64, error) {
	var (
		prevTx *Tx
		err    error
	)

	if prevTx, err = txFetcher.FetchTx(hex.EncodeToString(p.prevTxId)); err != nil {
		return 0xffffffffffffffff, err
	} else {
		return prevTx.txOuts[p.prevTxOutIdx].Amount(), nil
	}
}

func (p *TxIn) Serialize() []byte {
	var buf bytes.Buffer

	slices.Reverse(p.prevTxId)
	buf.Write(p.prevTxId)
	slices.Reverse(p.prevTxId)

	binary.Write(&buf, binary.LittleEndian, p.prevTxOutIdx)

	buf.Write(p.scriptSig.Serialize())

	binary.Write(&buf, binary.LittleEndian, p.sequence)

	return buf.Bytes()
}

func ParseTxIn(buf *bytes.Buffer) (*TxIn, error) {
	var (
		bPrevTxId     []byte
		prevTxId      []byte
		bPrevTxOutIdx []byte
		prevTxOutIdx  uint32
		scriptSig     *Script
		bSequence     []byte
		sequence      uint32
		n             int
		err           error
	)

	bPrevTxId = make([]byte, 32)
	if n, err = io.ReadFull(buf, bPrevTxId); n < len(bPrevTxId) || err != nil {
		return nil, err
	} else {
		slices.Reverse(bPrevTxId)
		prevTxId = bPrevTxId
	}

	bPrevTxOutIdx = make([]byte, 4)
	if n, err = io.ReadFull(buf, bPrevTxOutIdx); n < len(bPrevTxOutIdx) || err != nil {
		return nil, err
	} else {
		prevTxOutIdx = binary.LittleEndian.Uint32(bPrevTxOutIdx)
	}

	if scriptSig, err = ParseScript(buf); err != nil {
		return nil, err
	}

	bSequence = make([]byte, 4)
	if n, err = io.ReadFull(buf, bSequence); n < len(bSequence) || err != nil {
		return nil, err
	} else {
		sequence = binary.LittleEndian.Uint32(bSequence)
	}

	return &TxIn{
		prevTxId:     prevTxId,
		prevTxOutIdx: prevTxOutIdx,
		scriptSig:    scriptSig,
		sequence:     sequence,
	}, nil
}
