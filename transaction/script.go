package transaction

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"strings"
)

const (
	opPushBytes75 = 0x4b
	opPushData1   = 0x4c
	opPushData2   = 0x4d
	opPushData4   = 0x4e
)

var ErrMalformedScript = errors.New("malformed script")

type Script struct {
	cmds []*ScriptElem
}

func NewScript(cmds []*ScriptElem) (*Script, error) {
	if err := validateCmds(cmds); err != nil {
		return nil, err
	}
	return &Script{cmds: cloneScriptElems(cmds)}, nil
}

func (p *Script) Cmds() []*ScriptElem {
	return cloneScriptElems(p.cmds)
}

func (p *Script) SetCmds(cmds []*ScriptElem) error {
	if err := validateCmds(cmds); err != nil {
		return err
	}
	p.cmds = cloneScriptElems(cmds)
	return nil
}

func (p *Script) Clone() *Script {
	return &Script{cmds: cloneScriptElems(p.cmds)}
}

func (p *Script) String() string {
	parts := make([]string, len(p.cmds))
	for i, c := range p.cmds {
		parts[i] = c.String()
	}
	return "Script(" + strings.Join(parts, ", ") + ")"
}

func (p *Script) Serialize() []byte {
	var raw []byte
	for i, c := range p.cmds {
		raw = append(raw, c.Serialize()...)
		op, isOp := c.OpCode()
		if lenSize, _ := pushInfo(op); isOp && lenSize > 0 {
			var lb [4]byte
			binary.LittleEndian.PutUint32(lb[:], uint32(p.cmds[i+1].Len()))
			raw = append(raw, lb[:lenSize]...)
		}
	}
	return append(EncodeCVarint(uint64(len(raw))), raw...)
}

func ParseScript(r io.Reader) (*Script, error) {
	n, err := DecodeCVarint(r)
	if err != nil {
		return nil, fmt.Errorf("script length: %w", err)
	}

	if n > math.MaxInt64 {
		return nil, io.ErrUnexpectedEOF
	}
	b, err := io.ReadAll(io.LimitReader(r, int64(n)))
	if err != nil {
		return nil, err
	}
	if uint64(len(b)) != n {
		return nil, io.ErrUnexpectedEOF
	}

	var cmds []*ScriptElem
	for i := 0; i < len(b); {
		op := b[i]
		cmds = append(cmds, NewOpElem(op))
		i++

		lenSize, isPush := pushInfo(op)
		if !isPush {
			continue
		}

		l := uint32(op)
		if lenSize > 0 {
			if len(b)-i < lenSize {
				return nil, fmt.Errorf("%w: truncated length field of opcode %#x at offset %d", ErrMalformedScript, op, i-1)
			}
			var lb [4]byte
			copy(lb[:], b[i:i+lenSize])
			l = binary.LittleEndian.Uint32(lb[:])
			i += lenSize
		}
		if uint32(len(b)-i) < l {
			return nil, fmt.Errorf("%w: push of %d bytes at offset %d runs past end of script", ErrMalformedScript, l, i)
		}
		cmds = append(cmds, NewDataElem(b[i:i+int(l)]))
		i += int(l)
	}

	return &Script{cmds: cmds}, nil
}

func pushInfo(op byte) (lenSize int, isPush bool) {
	switch {
	case op >= 0x01 && op <= opPushBytes75:
		return 0, true
	case op == opPushData1:
		return 1, true
	case op == opPushData2:
		return 2, true
	case op == opPushData4:
		return 4, true
	default:
		return 0, false
	}
}

func validateCmds(cmds []*ScriptElem) error {
	for i := 0; i < len(cmds); i++ {
		op, isOp := cmds[i].OpCode()
		if !isOp {
			return fmt.Errorf("%w: data at cmd %d is not preceded by a push opcode", ErrMalformedScript, i)
		}
		lenSize, isPush := pushInfo(op)
		if !isPush {
			continue
		}

		i++
		if i >= len(cmds) || cmds[i].IsOp() {
			return fmt.Errorf("%w: push opcode %#x at cmd %d is not followed by data", ErrMalformedScript, op, i-1)
		}
		l := uint32(cmds[i].Len())
		if (lenSize == 0 && l != uint32(op)) || (lenSize > 0 && l >= 1<<(8*lenSize)) {
			return fmt.Errorf("%w: %d bytes of data do not match push opcode %#x", ErrMalformedScript, l, op)
		}
	}
	return nil
}
