package auth_service

import (
	"context"
	"errors"
	"fmt"

	core_auth "github.com/Rics69/x-net/internal/core/auth"
	"github.com/Rics69/x-net/internal/core/domain"
	core_errors "github.com/Rics69/x-net/internal/core/errors"
	"golang.org/x/crypto/bcrypt"
)

// и на "нет такого юзера", и на "неверный пароль" отвечаем одинаково,
// иначе по ответу можно перебором узнать, какие username зарегистрированы
var errInvalidCredentials = fmt.Errorf("invalid username or password: %w", core_errors.ErrUnauthorized)

func (s *AuthService) Login(
	ctx context.Context,
	username string,
	password string,
) (domain.User, core_auth.AccessToken, error) {
	user, err := s.usersRepository.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, core_errors.ErrNotFound) {
			// всё равно гоняем bcrypt: без этого "нет юзера" отвечает за ~1мс,
			// а "неверный пароль" за ~70мс, и по времени ответа снова видно, есть ли username
			_ = bcrypt.CompareHashAndPassword(dummyPasswordHash, []byte(password))

			return domain.User{}, core_auth.AccessToken{}, errInvalidCredentials
		}

		return domain.User{}, core_auth.AccessToken{}, fmt.Errorf("get user by username: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return domain.User{}, core_auth.AccessToken{}, errInvalidCredentials
	}

	token, err := s.tokenManager.Issue(user.ID)
	if err != nil {
		return domain.User{}, core_auth.AccessToken{}, fmt.Errorf("issue access token: %w", err)
	}

	return user, token, nil
}
