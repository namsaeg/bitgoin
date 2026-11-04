package transaction

import (
	"bytes"
	"encoding/hex"
)

type ScriptElem struct {
	data []byte
	op   byte
	isOp bool
}

func NewDataElem(data []byte) *ScriptElem {
	return &ScriptElem{data: bytes.Clone(data)}
}

func (p *ScriptElem) Data() []byte {
	return bytes.Clone(p.data)
}

func NewOpElem(op byte) *ScriptElem {
	return &ScriptElem{op: op, isOp: true}
}

func (p *ScriptElem) Clone() *ScriptElem {
	return &ScriptElem{data: bytes.Clone(p.data), op: p.op, isOp: p.isOp}
}

func (p *ScriptElem) IsOp() bool {
	return p.isOp
}

func (p *ScriptElem) OpCode() (op byte, ok bool) {
	return p.op, p.isOp
}

func (p *ScriptElem) Len() int {
	return len(p.data)
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

func cloneScriptElems(src []*ScriptElem) []*ScriptElem {
	dst := make([]*ScriptElem, len(src))
	for i, e := range src {
		dst[i] = e.Clone()
	}
	return dst
}
