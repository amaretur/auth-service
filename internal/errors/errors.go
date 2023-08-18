package errors

import (
	errutil "github.com/amaretur/auth-service/pkg/errors"
)

var (
	// Используется для обозначения непредвиденных ошибок,
	// которые не должны появляться
	Internal = errutil.NewType("internal error")

	InvalidToken = errutil.NewType("parse token error")
	NotFound = errutil.NewType("not found")
)
