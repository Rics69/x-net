package auth_service

import (
	"fmt"

	core_errors "github.com/Rics69/x-net/internal/core/errors"
	"golang.org/x/crypto/bcrypt"
)

const (
	minPasswordLen = 8

	// bcrypt работает только с первыми 72 БАЙТАМИ пароля (не символами: кириллица = 2 байта).
	// Старые версии молча обрезали хвост, новые x/crypto возвращают ErrPasswordTooLong.
	// Проверяем сами, чтобы отдать 400, а не 500
	maxPasswordBytes = 72
)

// хеш "пустышки" для Login, когда юзер не найден (см. login.go)
var dummyPasswordHash, _ = bcrypt.GenerateFromPassword([]byte("dummy-password"), bcrypt.DefaultCost)

func validatePassword(password string) error {
	if len([]rune(password)) < minPasswordLen {
		return fmt.Errorf(
			"password must be at least %d symbols: %w",
			minPasswordLen,
			core_errors.ErrInvalidArgument,
		)
	}

	if len(password) > maxPasswordBytes {
		return fmt.Errorf(
			"password must be at most %d bytes: %w",
			maxPasswordBytes,
			core_errors.ErrInvalidArgument,
		)
	}

	return nil
}

// bcrypt сам генерит соль и кладёт её (и cost) внутрь хеша,
// поэтому отдельная колонка под соль не нужна.
// DefaultCost=10 - ~50-100мс на хеш: дорого для перебора, терпимо для логина
func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("bcrypt generate: %w", err)
	}

	return string(hash), nil
}
