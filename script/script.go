package script

import (
	"bitgoin_v1/compactsize"
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"slices"
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

var (
	ErrMalformed     = errors.New("script: 잘못된 형식")
	ErrInvalidPubKey = errors.New("script: 공개키 형식이 아님")
)

type Script struct {
	elems []*Elem
}

func New(elems []*Elem) (*Script, error) {
	if err := validate(elems); err != nil {
		return nil, err
	}
	return &Script{elems: elems}, nil
}

func (p *Script) Clone() *Script {
	return &Script{elems: slices.Clone(p.elems)}
}

func (p *Script) Equal(q *Script) bool {
	return slices.EqualFunc(p.elems, q.elems, (*Elem).Equal)
}

func (p *Script) Len() int {
	return len(p.elems)
}

func (p *Script) Elems() []*Elem {
	return slices.Clone(p.elems)
}

func (p *Script) Elem(i int) (e *Elem, ok bool) {
	if i < 0 || i >= len(p.elems) {
		return nil, false
	}
	return p.elems[i], true
}

func (p *Script) IsPushOnly() bool {
	for _, e := range p.elems {
		if op, isOp := e.OpCode(); isOp && op > op16 {
			return false
		}
	}
	return true
}

func (p *Script) IsP2PK() (pub []byte, ok bool) {
	if len(p.elems) != 3 || !p.isOpAt(2, opCheckSig) {
		return nil, false
	}
	op, isOp := p.elems[0].OpCode()
	if !isOp {
		return nil, false
	}
	data := p.elems[1].data
	switch {
	case op == opPushData33 && (data[0] == 0x02 || data[0] == 0x03):
	case op == opPushData65 && (data[0] == 0x04 || data[0] == 0x06 || data[0] == 0x07):
	default:
		return nil, false
	}
	return bytes.Clone(data), true
}

func (p *Script) IsP2PKH() (h [20]byte, ok bool) {
	if !p.matchOps(opDup, opHash160, opPushData20) || len(p.elems) != 6 || !p.isOpAt(4, opEqualVerify) || !p.isOpAt(5, opCheckSig) {
		return h, false
	}
	copy(h[:], p.elems[3].data)
	return h, true
}

func (p *Script) IsP2SH() (h [20]byte, ok bool) {
	if !p.matchOps(opHash160, opPushData20) || len(p.elems) != 4 || !p.isOpAt(3, opEqual) {
		return h, false
	}
	copy(h[:], p.elems[2].data)
	return h, true
}

func (p *Script) isOpAt(i int, want byte) bool {
	if i >= len(p.elems) {
		return false
	}
	op, isOp := p.elems[i].OpCode()
	return isOp && op == want
}

func (s *Script) WithoutCodeSeparators() *Script {
	elems := make([]*Elem, 0, len(s.elems))
	for _, e := range s.elems {
		if op, isOp := e.OpCode(); isOp && op == opCodeSeparator {
			continue
		}
		elems = append(elems, e)
	}
	return &Script{elems: elems}
}

func (p *Script) matchOps(ops ...byte) bool {
	if len(p.elems) < len(ops) {
		return false
	}
	for i, want := range ops {
		if !p.isOpAt(i, want) {
			return false
		}
	}
	return true
}

func (p *Script) String() string {
	parts := make([]string, 0, len(p.elems))
	for i := 0; i < len(p.elems); i++ {
		op, _ := p.elems[i].OpCode()
		lenSize, isPush := pushInfo(op)
		if !isPush {
			parts = append(parts, opName(op))
			continue
		}
		i++
		data := hex.EncodeToString(p.elems[i].data)
		if data == "" {
			data = "<empty>"
		}
		if lenSize == 0 {
			parts = append(parts, data)
		} else {
			parts = append(parts, opName(op), data)
		}
	}
	return strings.Join(parts, " ")
}

func (p *Script) appendRaw(b []byte) []byte {
	for i, e := range p.elems {
		if !e.isOp {
			b = append(b, e.data...)
			continue
		}
		b = append(b, e.op)
		lenSize, isPush := pushInfo(e.op)
		if !isPush || lenSize == 0 {
			continue
		}
		var lb [4]byte
		binary.LittleEndian.PutUint32(lb[:], uint32(len(p.elems[i+1].data)))
		b = append(b, lb[:lenSize]...)
	}
	return b
}

func (p *Script) RawLen() int {
	n := 0
	for _, e := range p.elems {
		n += e.Len()
		if op, isOp := e.OpCode(); isOp {
			lenSize, _ := pushInfo(op)
			n += int(lenSize)
		}
	}
	return n
}

func (p *Script) Raw() []byte {
	return p.appendRaw(make([]byte, 0, p.RawLen()))
}

func (p *Script) Serialize() []byte {
	n := p.RawLen()
	b := make([]byte, 0, compactsize.Size(uint64(n))+n)
	b = compactsize.Append(b, uint64(n))
	return p.appendRaw(b)
}

func Parse(r io.Reader) (*Script, error) {
	n, err := compactsize.ReadMax(r, maxScriptLen)
	if err != nil {
		return nil, fmt.Errorf("script: 스크립트 길이 읽기: %w", err)
	}
	b, err := io.ReadAll(io.LimitReader(r, int64(n)))
	if err != nil {
		return nil, fmt.Errorf("script: 스크립트 본문 읽기: %w", err)
	}
	if uint64(len(b)) != n {
		return nil, fmt.Errorf("script: 본문 %d바이트 중 %d바이트만 있음: %w", n, len(b), io.ErrUnexpectedEOF)
	}
	return ParseRaw(b)
}

func ParseRaw(b []byte) (*Script, error) {
	if uint64(len(b)) > maxScriptLen {
		return nil, fmt.Errorf("%w: 길이 %d가 상한 %d를 넘음", ErrMalformed, len(b), maxScriptLen)
	}
	var elems []*Elem
	for i := 0; i < len(b); {
		start := i
		op := b[i]
		i++
		elems = append(elems, NewOp(op))

		lenSize, isPush := pushInfo(op)
		if !isPush {
			continue
		}
		n := uint64(op)
		if lenSize > 0 {
			if len(b)-i < int(lenSize) {
				return nil, fmt.Errorf("%w: offset %d의 opcode %#02x 길이 필드가 잘림", ErrMalformed, start, op)
			}
			var lb [4]byte
			copy(lb[:], b[i:i+int(lenSize)])
			n = uint64(binary.LittleEndian.Uint32(lb[:]))
			i += int(lenSize)
		}
		if uint64(len(b)-i) < n {
			return nil, fmt.Errorf("%w: offset %d의 opcode %#02x가 %d바이트를 push하는데 %d바이트만 남음", ErrMalformed, start, op, n, len(b)-i)
		}
		elems = append(elems, NewData(b[i:i+int(n)]))
		i += int(n)
	}
	return &Script{elems: elems}, nil
}

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

func validate(elems []*Elem) error {
	var rawLen uint64
	for i := 0; i < len(elems); i++ {
		e := elems[i]
		if e == nil {
			return fmt.Errorf("%w: index %d가 nil", ErrMalformed, i)
		}
		op, isOp := e.OpCode()
		if !isOp {
			return fmt.Errorf("%w: index %d의 데이터의 앞에 push opcode가 없음", ErrMalformed, i)
		}
		rawLen++

		lenSize, isPush := pushInfo(op)
		if !isPush {
			continue
		}
		if i+1 >= len(elems) || elems[i+1] == nil || elems[i+1].IsOp() {
			return fmt.Errorf("%w: index %d의 push opcode %#02x 뒤에 데이터가 없음", ErrMalformed, i, op)
		}
		i++
		n := uint64(elems[i].Len())
		if lenSize == 0 && n != uint64(op) {
			return fmt.Errorf("%w: index %d의 opcode %#02x는 %d바이트를 push하는데 데이터는 %d바이트", ErrMalformed, i-1, op, op, n)
		}
		if lenSize > 0 && n >= 1<<(8*uint64(lenSize)) {
			return fmt.Errorf("%w: index %d의 opcode %#02x로 %d바이트를 나타낼 수 없음", ErrMalformed, i-1, op, n)
		}
		rawLen += uint64(lenSize) + n
	}
	if rawLen > maxScriptLen {
		return fmt.Errorf("%w: 직렬화 길이 %d가 상한 %d를 넘음", ErrMalformed, rawLen, maxScriptLen)
	}
	return nil
}

func P2PK(pub []byte) (*Script, error) {
	switch {
	case len(pub) == 33 && (pub[0] == 0x02 || pub[0] == 0x03):
	case len(pub) == 65 && pub[0] == 0x04:
	default:
		return nil, fmt.Errorf("%w: %d바이트, 접두 %#02x", ErrInvalidPubKey, len(pub), firstByte(pub))
	}
	return &Script{elems: []*Elem{
		NewOp(byte(len(pub))), NewData(pub),
		NewOp(opCheckSig),
	}}, nil
}

func P2PKH(h [20]byte) *Script {
	return &Script{elems: []*Elem{
		NewOp(opDup),
		NewOp(opHash160),
		NewOp(opPushData20), NewData(h[:]),
		NewOp(opEqualVerify),
		NewOp(opCheckSig),
	}}
}

func P2SH(h [20]byte) *Script {
	return &Script{elems: []*Elem{
		NewOp(opHash160),
		NewOp(opPushData20), NewData(h[:]),
		NewOp(opEqual),
	}}
}

func firstByte(b []byte) byte {
	if len(b) == 0 {
		return 0
	}
	return b[0]
}

func opName(op byte) string {
	if name := opNames[op]; name != "" {
		return name
	}
	return fmt.Sprintf("OP_UNKNOWN(%#02x)", op)
}
