package script

import (
	"bytes"
	"encoding/hex"
)

type Elem struct {
	op   byte
	data []byte
	isOp bool
}

func NewOp(op byte) *Elem {
	return &Elem{
		op:   op,
		isOp: true,
	}
}

func NewData(data []byte) *Elem {
	return &Elem{
		data: bytes.Clone(data),
		isOp: false,
	}
}

func (p *Elem) Data() []byte {
	return bytes.Clone(p.data)
}

func (p *Elem) IsOp() bool {
	return p.isOp
}

func (p *Elem) Clone() *Elem {
	return &Elem{
		op:   p.op,
		data: bytes.Clone(p.data),
		isOp: p.isOp,
	}
}

func (p *Elem) Equal(q *Elem) bool {
	return p.isOp == q.isOp && p.op == q.op && bytes.Equal(p.data, q.data)
}

func (p *Elem) Len() int {
	if p.isOp {
		return 1
	}
	return len(p.data)
}

func (p *Elem) OpCode() (byte, bool) {
	return p.op, p.isOp
}

func (p *Elem) String() string {
	return hex.EncodeToString(p.Serialize())
}

func (p *Elem) Serialize() []byte {
	if p.isOp {
		return []byte{p.op}
	}
	return bytes.Clone(p.data)
}
