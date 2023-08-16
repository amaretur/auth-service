package errors

import (
	"errors"
)

const (
	UnTypedErr uint32 = 0
)

func Is(err, target error) bool {
	return errors.Is(err, target)
}

func Wrap(err error, wrapper *Instance) *Instance {

	wrapper.Err = err

	return wrapper
}

// Возвращает вложенную ошибку, если она есть
func Unwrap(err error) error {
	switch e := err.(type) {
		case *Instance:
			return e.Err
		default:
			return nil
	}
}

// Проверяет, принадлежит ли ошибка указанному типу
func TypeIs(err error, t *Type) bool {
	switch e := err.(type) {
		case *Instance:
			return e.TypeId == t.TypeId
		default:
			return false
	}
}

// Проверяет, есть ли в стеке ошибок, ошибка указанного типа
func Has(err error, t *Type) bool {

	for {
		if err == nil {
			return false
		}

		if TypeIs(err, t) {
			return true
		}

		err = Unwrap(err)
	}
}

func TypeId(err error) uint32 {

	if e, ok := err.(*Instance); ok {
		return e.TypeId
	}

	return UnTypedErr
}
