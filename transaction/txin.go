package transaction

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"slices"
)

var ErrPrevOutNotFound = errors.New("previous output index out of range")

type TxIn struct {
	prevTxID     [32]byte
	prevTxOutIdx uint32
	scriptSig    *Script
	sequence     uint32
}

func NewTxIn(prevTxID [32]byte, prevTxOutIdx uint32, scriptSig *Script, sequence uint32) *TxIn {
	return &TxIn{
		prevTxID:     prevTxID,
		prevTxOutIdx: prevTxOutIdx,
		scriptSig:    scriptSig.Clone(),
		sequence:     sequence,
	}
}

func (p *TxIn) PrevTxID() [32]byte {
	return p.prevTxID
}

func (p *TxIn) SetPrevTxID(prevTxID [32]byte) {
	p.prevTxID = prevTxID
}

func (p *TxIn) PrevTxOutIdx() uint32 {
	return p.prevTxOutIdx
}

func (p *TxIn) SetPrevTxOutIdx(prevTxOutIdx uint32) {
	p.prevTxOutIdx = prevTxOutIdx
}

func (p *TxIn) ScriptSig() *Script {
	return p.scriptSig.Clone()
}

func (p *TxIn) SetScriptSig(scriptSig *Script) {
	p.scriptSig = scriptSig.Clone()
}

func (p *TxIn) Sequence() uint32 {
	return p.sequence
}

func (p *TxIn) SetSequence(sequence uint32) {
	p.sequence = sequence
}

func (p *TxIn) Clone() *TxIn {
	return &TxIn{
		prevTxID:     p.prevTxID,
		prevTxOutIdx: p.prevTxOutIdx,
		scriptSig:    p.scriptSig.Clone(),
		sequence:     p.sequence,
	}
}

func (p *TxIn) Value(f TxFetcher) (uint64, error) {
	prevTx, err := f.FetchTx(p.prevTxID)
	if err != nil {
		return 0, fmt.Errorf("fetch prev tx %x: %w", p.prevTxID, err)
	}
	if uint64(p.prevTxOutIdx) >= uint64(len(prevTx.txOuts)) {
		return 0, fmt.Errorf("%w: tx %x has %d outputs, want index %d",
			ErrPrevOutNotFound, p.prevTxID, len(prevTx.txOuts), p.prevTxOutIdx)
	}
	return prevTx.txOuts[p.prevTxOutIdx].Amount(), nil
}

func (p *TxIn) String() string {
	return fmt.Sprintf("TxIn(\nprevTxId: %x\nprevTxOutIdx: %d\nscriptSig: %s\nsequence: %08x\n)", p.prevTxID, p.prevTxOutIdx, p.scriptSig, p.sequence)
}

func (p *TxIn) Serialize() []byte {
	id := p.prevTxID
	slices.Reverse(id[:])
	b := append([]byte(nil), id[:]...)
	b = binary.LittleEndian.AppendUint32(b, p.prevTxOutIdx)
	b = append(b, p.scriptSig.Serialize()...)
	return binary.LittleEndian.AppendUint32(b, p.sequence)
}

func ParseTxIn(r io.Reader) (*TxIn, error) {
	var in TxIn
	if _, err := io.ReadFull(r, in.prevTxID[:]); err != nil {
		return nil, fmt.Errorf("txin prev txid: %w", err)
	}
	slices.Reverse(in.prevTxID[:])
	var b [4]byte
	if _, err := io.ReadFull(r, b[:]); err != nil {
		return nil, fmt.Errorf("txin prev index: %w", noEOF(err))
	}
	in.prevTxOutIdx = binary.LittleEndian.Uint32(b[:])
	scriptSig, err := ParseScript(r)
	if err != nil {
		return nil, fmt.Errorf("txin scriptSig: %w", noEOF(err))
	}
	in.scriptSig = scriptSig
	if _, err := io.ReadFull(r, b[:]); err != nil {
		return nil, fmt.Errorf("txin sequence: %w", noEOF(err))
	}
	in.sequence = binary.LittleEndian.Uint32(b[:])
	return &in, nil
}

func noEOF(err error) error {
	if errors.Is(err, io.EOF) {
		return io.ErrUnexpectedEOF
	}
	return err
}
