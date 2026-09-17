package entity

import (
	"strings"

	"github.com/google/uuid"
)

func NewUUID() (string, error) {
	return uuid.NewString(), nil
}

func NormalizeUUID(value string) (string, error) {
	parsed, err := uuid.Parse(strings.TrimSpace(value))
	if err != nil {
		return "", ErrInvalidRequest
	}
	return parsed.String(), nil
}
