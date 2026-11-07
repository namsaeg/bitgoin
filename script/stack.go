package script

import (
	"bytes"
	"encoding/hex"
	"errors"
	"slices"
	"strings"
)

var ErrStackUnderflow = errors.New("script: 스택 원소 부족")

type Stack struct {
	elems [][]byte
}

func NewStack() *Stack {
	return &Stack{}
}

func (p *Stack) Clone() *Stack {
	elems := make([][]byte, len(p.elems))
	for i, e := range p.elems {
		elems[i] = bytes.Clone(e)
	}
	return &Stack{elems: elems}
}

func (p *Stack) Len() int {
	return len(p.elems)
}

func (p *Stack) Push(e []byte) {
	p.elems = append(p.elems, bytes.Clone(e))
}

func (p *Stack) Insert(i int, e []byte) error {
	if i < 0 || i > len(p.elems) {
		return ErrStackUnderflow
	}
	j := len(p.elems) - i
	p.elems = slices.Insert(p.elems, j, bytes.Clone(e))
	return nil
}

func (p *Stack) Pop() ([]byte, error) {
	l := len(p.elems)
	if l == 0 {
		return nil, ErrStackUnderflow
	}
	top := p.elems[l-1]
	p.elems[l-1] = nil
	p.elems = p.elems[:l-1]
	return top, nil
}

func (p *Stack) Remove(i int) ([]byte, error) {
	j, err := p.pos(i)
	if err != nil {
		return nil, err
	}
	e := p.elems[j]
	p.elems = slices.Delete(p.elems, j, j+1)
	return e, nil
}

func (p *Stack) Peek(i int) ([]byte, error) {
	j, err := p.pos(i)
	if err != nil {
		return nil, err
	}
	return bytes.Clone(p.elems[j]), nil
}

func (p *Stack) Swap(i, j int) error {
	a, err := p.pos(i)
	if err != nil {
		return err
	}
	b, err := p.pos(j)
	if err != nil {
		return err
	}
	p.elems[a], p.elems[b] = p.elems[b], p.elems[a]
	return nil
}

func (p *Stack) pos(i int) (int, error) {
	if i < 0 || i >= len(p.elems) {
		return 0, ErrStackUnderflow
	}
	return len(p.elems) - 1 - i, nil
}

func (p *Stack) String() string {
	var sb strings.Builder
	sb.WriteString("Stack(\n")
	for i := len(p.elems) - 1; i >= 0; i-- {
		s := hex.EncodeToString(p.elems[i])
		if s == "" {
			s = "<empty>"
		}
		sb.WriteString(s)
		sb.WriteByte('\n')
	}
	sb.WriteString(")")
	return sb.String()
}
