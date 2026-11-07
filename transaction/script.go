package transaction

import (
	"encoding/binary"
	"fmt"
	"io"
	"strings"
)

const (
	maxScriptLen  = 33554432
	opPushBytes1  = 0x01
	opPushBytes75 = 0x4b
	opPushData1   = 0x4c
	opPushData2   = 0x4d
	opPushData4   = 0x4e
)

type Script struct {
	cmds []*ScriptElem
}

func NewScript(cmds []*ScriptElem) (*Script, error) {
	if serializedLen(cmds) > maxScriptLen {
		return nil, fmt.Errorf("cmds의 바이트 길이가 최대 스크립트 길이보다 김")
	}
	if err := validateCmds(cmds); err != nil {
		return nil, fmt.Errorf("cmds가 유효하지 않은 스크립트임: %w", err)
	}
	return &Script{
		cmds: cloneAll(cmds),
	}, nil
}

func (p *Script) Cmds() []*ScriptElem {
	return cloneAll(p.cmds)
}

func (p *Script) SetCmds(cmds []*ScriptElem) error {
	if serializedLen(cmds) > maxScriptLen {
		return fmt.Errorf("cmds의 바이트 길이가 최대 스크립트 길이보다 김")
	}
	if err := validateCmds(cmds); err != nil {
		return fmt.Errorf("cmds가 유효하지 않은 스크립트임: %w", err)
	}
	p.cmds = cloneAll(cmds)
	return nil
}

func (p *Script) Clone() *Script {
	return &Script{
		cmds: cloneAll(p.cmds),
	}
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
	for i := 0; i < len(p.cmds); i++ {
		raw = append(raw, p.cmds[i].Serialize()...)
		op, isOp := p.cmds[i].OpCode()
		if !isOp {
			continue
		}
		lenSize, isPush := pushInfo(op)
		if !isPush || lenSize == 0 {
			continue
		}
		var lb [4]byte
		binary.LittleEndian.PutUint32(lb[:], uint32(p.cmds[i+1].Len()))
		raw = append(raw, lb[:lenSize]...)
	}
	return append(EncodeCVarint(uint64(len(raw))), raw...)
}

func ParseScript(r io.Reader) (*Script, error) {
	scriptLen, err := DecodeCVarint(r)
	if err != nil {
		return nil, fmt.Errorf("스크립트 길이 읽기 실패: %w", err)
	}
	if scriptLen > maxScriptLen {
		return nil, fmt.Errorf("스크립트 길이가 최대 스크립트 길이보다 김")
	}

	b, err := io.ReadAll(io.LimitReader(r, int64(scriptLen)))
	if err != nil {
		return nil, fmt.Errorf("스크립트 본문 읽기 실패: %w", err)
	}

	var cmds []*ScriptElem
	for i := 0; i < len(b); {
		op := b[i]
		i++
		cmds = append(cmds, NewOpElem(op))

		lenSize, isPush := pushInfo(op)
		if !isPush {
			continue
		}
		l := uint32(op)
		if lenSize > 0 {
			if len(b)-i < int(lenSize) {
				return nil, fmt.Errorf("OP_PUSHDATA(offset %d)의 len 필드가 잘려있음", i-1)
			}
			var lb [4]byte
			copy(lb[:], b[i:i+int(lenSize)])
			l = binary.LittleEndian.Uint32(lb[:])
			i += int(lenSize)
		}
		if uint32(len(b)-i) < l {
			return nil, fmt.Errorf("OP_PUSHDATA(offset %d)의 data 필드가 잘려있음", i-1-int(lenSize))
		}
		cmds = append(cmds, NewDataElem(b[i:i+int(l)]))
		i += int(l)
	}

	return &Script{
		cmds: cmds,
	}, nil
}

// serializedLen은 cmds의 모든 요소를 직렬화한 값의 바이트 길이를 반환한다.
func serializedLen(cmds []*ScriptElem) uint64 {
	var len uint64 = 0
	for _, c := range cmds {
		len += uint64(c.Len())
		op, isOp := c.OpCode()
		if !isOp {
			continue
		}
		lenSize, isPush := pushInfo(op)
		if !isPush {
			continue
		}
		len += uint64(lenSize)
	}
	return len
}

// pushInfo는 op가 OP_PUSHDATA인지 확인하고, 맞다면 len 필드의 크기와 true를, 아니면 0과 false를 반환한다.
func pushInfo(op byte) (uint8, bool) {
	switch {
	case opPushBytes1 <= op && op <= opPushBytes75:
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

// validateCmds는 cmds가 유효한 스크립트인지 확인하고, 유효하면 nil을, 아니면 error을 반환한다.
func validateCmds(cmds []*ScriptElem) error {
	for i := 0; i < len(cmds); i++ {
		op, isOp := cmds[i].OpCode()
		if !isOp {
			return fmt.Errorf("데이터(index %d)가 명령어에 의해 추가되고 있지 않음", i)
		}
		lenSize, isPush := pushInfo(op)
		if !isPush {
			continue
		}
		i++
		if i >= len(cmds) || cmds[i].IsOp() {
			return fmt.Errorf("OP_PUSHDATA(index %d) 명령어 뒤에 데이터가 뒤따르고 있지 않음", i-1)
		}
		l := uint64(cmds[i].Len())
		if (lenSize == 0 && uint64(op) != l) || (lenSize > 0 && l >= 1<<(8*lenSize)) {
			return fmt.Errorf("OP_PUSHDATA(index %d)와 데이터가 맞지 않음", i-1)
		}
	}
	return nil
}
