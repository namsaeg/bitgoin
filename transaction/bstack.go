package transaction

import "fmt"

type BStack struct {
	stack [][]byte
}

func (p *BStack) NewBStack(stack [][]byte) (*BStack, error) {
	var s BStack

	s.stack = make([][]byte, len(stack))
	for i, e := range stack {
		s.stack[i] = make([]byte, len(e))
		copy(s.stack[i], e)
	}

	return &s, nil
}

func (p *BStack) Stack() [][]byte {
	var stack [][]byte = make([][]byte, len(p.stack))
	for i, e := range p.stack {
		stack[i] = make([]byte, len(e))
		copy(stack[i], e)
	}

	return stack
}
func (p *BStack) SetStack(stack [][]byte) error {
	p.stack = make([][]byte, len(stack))
	for i, e := range stack {
		p.stack[i] = make([]byte, len(e))
		copy(p.stack[i], e)
	}

	return nil
}

func (p *BStack) Clone() *BStack {
	var s BStack

	s.stack = make([][]byte, len(p.stack))
	for i, e := range p.stack {
		s.stack[i] = make([]byte, len(e))
		copy(s.stack[i], e)
	}

	return &s
}

func (p *BStack) String() string {
	var bStackStr string = ""
	for _, e := range p.stack {
		bStackStr = fmt.Sprintf("%x", e) + "\n" + bStackStr
	}

	return fmt.Sprintf("BStack(\n%s\n)", bStackStr)
}
