package transaction

import (
	"bytes"
	"encoding/hex"
)

type ScriptElem struct {
	op   byte
	data []byte
	isOp bool
}

func NewOpElem(op byte) *ScriptElem {
	return &ScriptElem{
		op:   op,
		isOp: true,
	}
}

func NewDataElem(data []byte) *ScriptElem {
	return &ScriptElem{
		data: bytes.Clone(data),
		isOp: false,
	}
}

func (p *ScriptElem) Op() byte {
	return p.op
}

func (p *ScriptElem) Data() []byte {
	return bytes.Clone(p.data)
}

func (p *ScriptElem) IsOp() bool {
	return p.isOp
}

func (p *ScriptElem) Clone() *ScriptElem {
	return &ScriptElem{
		op:   p.op,
		data: bytes.Clone(p.data),
		isOp: p.isOp,
	}
}

func (p *ScriptElem) Len() uint32 {
	if p.isOp {
		return 1
	}
	return uint32(len(p.data))
}

func (p *ScriptElem) OpCode() (byte, bool) {
	return p.op, p.isOp
}

func (p *ScriptElem) String() string {
	return hex.EncodeToString(p.Serialize())
}

func (p *ScriptElem) Serialize() []byte {
	if p.isOp {
		return []byte{p.op}
	}
	return bytes.Clone(p.data)
}
