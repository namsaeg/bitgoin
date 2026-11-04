package transaction

import (
	"errors"
	"strings"
)

var ErrStackEmpty = errors.New("script stack is empty")

type ScriptStack struct {
	elems []*ScriptElem
}

func NewScriptStack(elems []*ScriptElem) *ScriptStack {
	return &ScriptStack{elems: cloneScriptElems(elems)}
}

func (p *ScriptStack) Stack() []*ScriptElem {
	return cloneScriptElems(p.elems)
}

func (p *ScriptStack) SetStack(elems []*ScriptElem) {
	p.elems = cloneScriptElems(elems)
}

func (p *ScriptStack) Clone() *ScriptStack {
	return &ScriptStack{elems: cloneScriptElems(p.elems)}
}

func (p *ScriptStack) Len() int {
	return len(p.elems)
}

func (p *ScriptStack) Push(e *ScriptElem) {
	p.elems = append(p.elems, e.Clone())
}

func (p *ScriptStack) Pop() (*ScriptElem, error) {
	n := len(p.elems)
	if n == 0 {
		return nil, ErrStackEmpty
	}
	top := p.elems[n-1]
	p.elems[n-1] = nil
	p.elems = p.elems[:n-1]
	return top, nil
}

func (p *ScriptStack) String() string {
	var sb strings.Builder
	sb.WriteString("ScriptStack(\n")
	for i := len(p.elems) - 1; i >= 0; i-- {
		sb.WriteString(p.elems[i].String())
		sb.WriteByte('\n')
	}
	sb.WriteString(")")
	return sb.String()
}
