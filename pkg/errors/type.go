package errors

import (
	"sync/atomic"
)

var (
	seq uint32 = 0
)

type Type struct {
	TypeId	uint32
	Info	string
}

func NewType(info string) *Type {
	return &Type{
		Info:	info,
		TypeId:	atomic.AddUint32(&seq, 1) - 1,
	}
}

func (t *Type) New(info string) *Instance {
	return &Instance{
		TypeId: t.TypeId,
		Info: info,
		Err: nil,
	}
}

func (t *Type) NewDefault() *Instance {
	return &Instance{
		TypeId: t.TypeId,
		Info: t.Info,
		Err: nil,
	}
}
