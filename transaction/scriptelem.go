package transaction

import (
	"encoding/hex"
)

type ScriptElem struct {
	elem []byte
}

func NewScriptElem(elem []byte) (*ScriptElem, error) {
	var result ScriptElem

	result.elem = make([]byte, len(elem))
	copy(result.elem, elem)

	return &result, nil
}

func (p *ScriptElem) Elem() []byte {
	var result []byte

	result = make([]byte, len(p.elem))
	copy(result, p.elem)

	return result
}
func (p *ScriptElem) SetElem(elem []byte) error {
	p.elem = make([]byte, len(elem))
	copy(p.elem, elem)

	return nil
}

func (p *ScriptElem) Clone() *ScriptElem {
	var result ScriptElem

	result.elem = make([]byte, len(p.elem))
	copy(result.elem, p.elem)

	return &result
}

func (p *ScriptElem) String() string {
	return hex.EncodeToString(p.elem)
}

func (p *ScriptElem) Serialize() []byte {
	return p.elem
}

func (p *ScriptElem) GetElemLen() uint16 {
	return uint16(len(p.elem))
}

func (p *ScriptElem) IsOp() bool {
	return len(p.elem) == 1
}

func (p *ScriptElem) GetOpCode() uint8 {
	if len(p.elem) == 1 {
		return p.elem[0]
	} else {
		return 0xff
	}
}
