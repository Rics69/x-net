package auth_service

import (
	"context"
	"fmt"

	core_auth "github.com/Rics69/x-net/internal/core/auth"
	"github.com/Rics69/x-net/internal/core/domain"
)

// после регистрации сразу выдаём токен, чтобы не гонять юзера на отдельный логин
func (s *AuthService) Register(
	ctx context.Context,
	user domain.User,
	password string,
) (domain.User, core_auth.AccessToken, error) {
	// сначала дешёвые проверки, потом дорогой bcrypt
	if err := user.Validate(); err != nil {
		return domain.User{}, core_auth.AccessToken{}, fmt.Errorf("validate user domain: %w", err)
	}

	if err := validatePassword(password); err != nil {
		return domain.User{}, core_auth.AccessToken{}, fmt.Errorf("validate password: %w", err)
	}

	passwordHash, err := hashPassword(password)
	if err != nil {
		return domain.User{}, core_auth.AccessToken{}, fmt.Errorf("hash password: %w", err)
	}

	user.PasswordHash = passwordHash

	user, err = s.usersRepository.CreateUser(ctx, user)
	if err != nil {
		return domain.User{}, core_auth.AccessToken{}, fmt.Errorf("create user: %w", err)
	}

	token, err := s.tokenManager.Issue(user.ID)
	if err != nil {
		return domain.User{}, core_auth.AccessToken{}, fmt.Errorf("issue access token: %w", err)
	}

	return user, token, nil
}
