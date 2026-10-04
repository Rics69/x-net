package core_errors

import "errors"

var (
	ErrNotFound        = errors.New("not found")
	ErrInvalidArgument = errors.New("invalid argument")
	ErrConflict        = errors.New("conflict")

	// 401 - "не знаю кто ты" (нет/битый токен), 403 - "знаю кто ты, но тебе нельзя"
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
)
