package httpx

import (
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-chi/chi/v5"
)

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)

func ReadRequiredHeader(r *http.Request, name string) (string, error) {
	value := strings.TrimSpace(r.Header.Get(name))
	if value == "" {
		return "", errors.New("missing required header")
	}
	return value, nil
}

func ReadRequiredUUIDHeader(r *http.Request, name string) (string, error) {
	header, err := ReadRequiredHeader(r, name)
	if err != nil {
		return "", err
	}
	return validateUUID(header)
}

func ReadRequiredUUIDParam(r *http.Request, name string) (string, error) {
	return validateUUID(chi.URLParam(r, name))
}

func ReadRequiredStringParam(r *http.Request, name string) (string, error) {
	value := strings.TrimSpace(chi.URLParam(r, name))
	if value == "" {
		return "", errors.New("missing required path param")
	}
	return value, nil
}

func validateUUID(value string) (string, error) {
	trimmed := strings.ToLower(strings.TrimSpace(value))
	if trimmed == "" {
		return "", errors.New("missing uuid")
	}
	if !uuidPattern.MatchString(trimmed) {
		return "", errors.New("invalid uuid")
	}
	return trimmed, nil
}
