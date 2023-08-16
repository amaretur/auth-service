package validator

import (
	"github.com/google/uuid"
)

func ValidateUuid(data string) error {
	_, err := uuid.Parse(data)
	return err
}
