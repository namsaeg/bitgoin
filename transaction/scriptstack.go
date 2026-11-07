package transaction

import (
	"fmt"
	"strings"
)

type ScriptStack struct {
	stack []*ScriptElem
}

func NewScriptStack(stack []*ScriptElem) *ScriptStack {
	return &ScriptStack{
		stack: cloneAll(stack),
	}
}

func (p *ScriptStack) Stack() []*ScriptElem {
	return cloneAll(p.stack)
}

func (p *ScriptStack) SetStack(stack []*ScriptElem) {
	p.stack = cloneAll(stack)
}

func (p *ScriptStack) Clone() *ScriptStack {
	return &ScriptStack{
		stack: cloneAll(p.stack),
	}
}

func (p *ScriptStack) Len() int {
	return len(p.stack)
}

func (p *ScriptStack) Push(e *ScriptElem) {
	p.stack = append(p.stack, e.Clone())
}

func (p *ScriptStack) Pop() (*ScriptElem, error) {
	l := len(p.stack)
	if l == 0 {
		return nil, fmt.Errorf("스택이 비어있음")
	}
	top := p.stack[l-1]
	p.stack[l-1] = nil
	p.stack = p.stack[:l-1]
	return top, nil
}

func (p *ScriptStack) String() string {
	var sb strings.Builder
	sb.WriteString("ScriptStack(\n")
	for i := len(p.stack) - 1; i >= 0; i-- {
		sb.WriteString(p.stack[i].String())
		sb.WriteByte('\n')
	}
	sb.WriteString(")")
	return sb.String()
}
