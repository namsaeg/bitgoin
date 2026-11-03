package transaction

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strings"
)

type Script struct {
	cmds []*ScriptElem
}

func NewScript(cmds []*ScriptElem) (*Script, error) {
	var result Script

	result.cmds = make([]*ScriptElem, len(cmds))
	for i := range cmds {
		result.cmds[i] = cmds[i].Clone()
	}

	return &result, nil
}

func (p *Script) Cmds() []*ScriptElem {
	var result []*ScriptElem = make([]*ScriptElem, len(p.cmds))
	for i := range p.cmds {
		result[i] = p.cmds[i].Clone()
	}

	return result
}
func (p *Script) SetCmds(cmds []*ScriptElem) error {
	p.cmds = make([]*ScriptElem, len(cmds))
	for i := range cmds {
		p.cmds[i] = cmds[i].Clone()
	}

	return nil
}

func (p *Script) Clone() *Script {
	var result Script

	result.cmds = make([]*ScriptElem, len(p.cmds))
	for i := range p.cmds {
		result.cmds[i] = p.cmds[i].Clone()
	}

	return &result
}

func (p *Script) String() string {
	var scriptStr strings.Builder

	for i := range p.cmds {
		scriptStr.WriteString(fmt.Sprintf("%s, ", p.cmds[i].String()))
	}

	return fmt.Sprintf("Script(%s)", scriptStr.String())
}

func (p *Script) Serialize() []byte {
	var (
		buf        bytes.Buffer
		bScriptLen []byte
		bScript    []byte
	)

	for i := range p.cmds {
		if p.cmds[i].IsOp() {
			buf.Write(p.cmds[i].Serialize())

			switch p.cmds[i].GetOpCode() {
			case 0x4c:
				buf.Write([]byte{uint8(p.cmds[i].GetElemLen())})
			case 0x4d:
				binary.Write(&buf, binary.LittleEndian, uint16(p.cmds[i].GetElemLen()))
			case 0xff:
				buf.Write([]byte{uint8(0xff)})
			}
		} else {
			buf.Write(p.cmds[i].Serialize())
		}
	}

	bScriptLen, _ = EncodeCVarint(uint64(buf.Len()))
	bScript = append(bScriptLen, buf.Bytes()...)

	return bScript
}

func ParseScript(buf *bytes.Buffer) (*Script, error) {
	var (
		cmds       []*ScriptElem
		bScriptLen uint64
		bScript    []byte
		scriptBuf  *bytes.Buffer
		bOp        []byte
		op         *ScriptElem
		opCode     uint8
		bData      []byte
		data       *ScriptElem
		bDataLen   []byte
		dataLen    uint16
		n          int
		err        error
	)

	if bScriptLen, err = DecodeCvarint(buf); err != nil {
		return nil, err
	}

	bScript = make([]byte, bScriptLen)
	if n, err = io.ReadFull(buf, bScript); n < len(bScript) || err != nil {
		return nil, err
	} else {
		scriptBuf = bytes.NewBuffer(bScript)
	}

	for scriptBuf.Len() > 0 {
		bOp = make([]byte, 1)
		if n, err = io.ReadFull(scriptBuf, bOp); n < len(bOp) || err != nil {
			return nil, err
		} else {
			if op, err = NewScriptElem(bOp); err != nil {
				return nil, err
			} else {
				cmds = append(cmds, op)
			}
		}

		opCode = op.GetOpCode()
		switch {
		case 1 <= opCode && opCode <= 75:
			bData = make([]byte, opCode)
			if n, err = io.ReadFull(scriptBuf, bData); n < len(bData) || err != nil {
				return nil, err
			} else {
				if data, err = NewScriptElem(bData); err != nil {
					return nil, err
				} else {
					cmds = append(cmds, data)
				}
			}
		case opCode == 0x4c:
			bDataLen = make([]byte, 1)
			if n, err = io.ReadFull(scriptBuf, bDataLen); n < len(bDataLen) || err != nil {
				return nil, err
			} else {
				dataLen = uint16(bDataLen[0])
			}

			bData = make([]byte, dataLen)
			if n, err = io.ReadFull(scriptBuf, bDataLen); n < len(bData) || err != nil {
				return nil, err
			} else {
				if data, err = NewScriptElem(bData); err != nil {
					return nil, err
				} else {
					cmds = append(cmds, data)
				}
			}
		case opCode == 0x4d:
			bDataLen = make([]byte, 2)
			if n, err = io.ReadFull(scriptBuf, bDataLen); n < len(bDataLen) || err != nil {
				return nil, err
			} else {
				dataLen = binary.LittleEndian.Uint16(bDataLen)
			}

			bData = make([]byte, dataLen)
			if n, err = io.ReadFull(scriptBuf, bDataLen); n < len(bData) || err != nil {
				return nil, err
			} else {
				if data, err = NewScriptElem(bData); err != nil {
					return nil, err
				} else {
					cmds = append(cmds, data)
				}
			}
		}
	}

	return &Script{
		cmds: cmds,
	}, nil
}
