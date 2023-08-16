package errors

type Instance struct {
	TypeId	uint32
	Err		error

	Info	string
}

func (i *Instance) Error() string {
	return i.Info
}

func (i *Instance) Wrap(err error) *Instance {
	i.Err = err
	return i
}

func (i *Instance) Unwrap() error {
	return i.Err
}

func (i *Instance) TypeIs(t *Type) bool {
	return i.TypeId == t.TypeId
}
