package script

import "errors"

const (
	MaxScriptSize         = 10000
	MaxElementSize        = 520
	MaxOpsPerScript       = 201
	MaxStackSize          = 1000
	MaxPubKeysPerMultisig = 20
	maxNumSize            = 4
)

var (
	ErrScriptSize  = errors.New("script: 스크립트가 10,000바이트를 넘음")
	ErrPushSize    = errors.New("script: push가 520바이트를 넘음")
	ErrOpCount     = errors.New("script: opcode가 201개를 넘음")
	ErrStackSize   = errors.New("script: 스택 원소가 1,000개를 넘음")
	ErrDisabledOp  = errors.New("script: 비활성화된 opcode")
	ErrNumOverflow = errors.New("script: 숫자가 4바이트를 넘음")
)
