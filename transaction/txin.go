package transaction

import (
	"encoding/binary"
	"fmt"
	"io"
	"slices"
)

type TxIn struct {
	prevTxId   [32]byte
	prevOutIdx uint32
	scriptSig  *Script
	sequence   uint32
}

func NewTxIn(prevTxId [32]byte, prevOutIdx uint32, scriptSig *Script, sequence uint32) *TxIn {
	return &TxIn{
		prevTxId:   prevTxId,
		prevOutIdx: prevOutIdx,
		scriptSig:  scriptSig.Clone(),
		sequence:   sequence,
	}
}

func (p *TxIn) PrevTxId() [32]byte {
	return p.prevTxId
}

func (p *TxIn) SetPrevTxId(prevTx [32]byte) {
	p.prevTxId = prevTx
}

func (p *TxIn) PrevOutIdx() uint32 {
	return p.prevOutIdx
}

func (p *TxIn) SetPrevOutIdx(prevOutIdx uint32) {
	p.prevOutIdx = prevOutIdx
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
		prevTxId:   p.prevTxId,
		prevOutIdx: p.prevOutIdx,
		scriptSig:  p.scriptSig.Clone(),
		sequence:   p.sequence,
	}
}

func (p *TxIn) Value(f *TxFetcher) (uint64, error) {
	tx, err := f.FetchTx(p.prevTxId)
	if err != nil {
		return 0, fmt.Errorf("트랜잭션 가져오기 실패: %w", err)
	}
	return tx.TxOuts()[p.prevOutIdx].Amount(), nil
}

func (p *TxIn) String() string {
	return fmt.Sprintf("TxIn(\nprevTxId: %x\nprevOutIdx: %x\nscriptSig: %s\nsequence: %x\n)", p.prevTxId, p.prevOutIdx, p.scriptSig, p.sequence)
}

func (p *TxIn) Serialize() []byte {
	id := p.prevTxId
	slices.Reverse(id[:])
	b := append([]byte(nil), id[:]...)
	b = binary.LittleEndian.AppendUint32(b, p.prevOutIdx)
	b = append(b, p.scriptSig.Serialize()...)
	return binary.LittleEndian.AppendUint32(b, p.sequence)
}

func ParseTxIn(r io.Reader) (*TxIn, error) {
	var in TxIn
	if _, err := io.ReadFull(r, in.prevTxId[:]); err != nil {
		return nil, fmt.Errorf("이전 트랜잭션 아이디 읽기 실패: %w", err)
	}
	slices.Reverse(in.prevTxId[:])
	var b [4]byte
	if _, err := io.ReadFull(r, b[:]); err != nil {
		return nil, fmt.Errorf("이전 트랜잭션 출력 인덱스 읽기 실패: %w", err)
	}
	in.prevOutIdx = binary.LittleEndian.Uint32(b[:])
	scriptSig, err := ParseScript(r)
	if err != nil {
		return nil, fmt.Errorf("해제 스크립트 읽기 실패: %w", err)
	}
	in.scriptSig = scriptSig
	if _, err := io.ReadFull(r, b[:]); err != nil {
		return nil, fmt.Errorf("이전 트랜잭션 출력 인덱스 읽기 실패: %w", err)
	}
	in.sequence = binary.LittleEndian.Uint32(b[:])
	return &in, nil
}
