package core_auth

import (
	"fmt"
	"strconv"
	"time"

	core_errors "github.com/Rics69/x-net/internal/core/errors"
	"github.com/golang-jwt/jwt/v5"
)

const (
	AccessTokenCookieName = "access_token"

	// для HS256 ключ должен быть не короче размера хеша (256 бит),
	// короткий секрет подбирается брутфорсом по любому выданному токену
	minJWTSecretLen = 32
)

type AccessToken struct {
	Value     string
	ExpiresAt time.Time
}

type TokenManager struct {
	secret []byte
	ttl    time.Duration
}

func NewTokenManager(config Config) (*TokenManager, error) {
	if len(config.JWTSecret) < minJWTSecretLen {
		return nil, fmt.Errorf(
			"JWT secret must be at least %d bytes, got %d",
			minJWTSecretLen,
			len(config.JWTSecret),
		)
	}

	return &TokenManager{
		secret: []byte(config.JWTSecret),
		ttl:    config.TokenTTL,
	}, nil
}

func (m *TokenManager) Issue(userID int) (AccessToken, error) {
	now := time.Now()
	expiresAt := now.Add(m.ttl)

	claims := jwt.RegisteredClaims{
		Subject:   strconv.Itoa(userID),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
	if err != nil {
		return AccessToken{}, fmt.Errorf("sign jwt: %w", err)
	}

	return AccessToken{
		Value:     token,
		ExpiresAt: expiresAt,
	}, nil
}

func (m *TokenManager) Parse(token string) (int, error) {
	var claims jwt.RegisteredClaims

	// WithValidMethods обязателен: без него можно подсунуть токен с alg=none
	// или с другим алгоритмом (alg confusion), и библиотека попробует его проверить.
	// exp проверяется библиотекой автоматически
	_, err := jwt.ParseWithClaims(
		token,
		&claims,
		func(*jwt.Token) (any, error) {
			return m.secret, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return 0, fmt.Errorf("parse jwt: %v: %w", err, core_errors.ErrUnauthorized)
	}

	userID, err := strconv.Atoi(claims.Subject)
	if err != nil {
		return 0, fmt.Errorf("invalid jwt subject='%s': %w", claims.Subject, core_errors.ErrUnauthorized)
	}

	return userID, nil
}
