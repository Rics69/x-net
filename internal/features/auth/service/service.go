package auth_service

import (
	"context"

	core_auth "github.com/Rics69/x-net/internal/core/auth"
	"github.com/Rics69/x-net/internal/core/domain"
)

type AuthService struct {
	usersRepository UsersRepository
	tokenManager    TokenManager
}

// своей таблицы у auth нет - работаем с users через их репозиторий.
// Интерфейс объявлен тут (у потребителя), users про auth ничего не знает
type UsersRepository interface {
	CreateUser(ctx context.Context, user domain.User) (domain.User, error)
	GetUserByUsername(ctx context.Context, username string) (domain.User, error)
}

type TokenManager interface {
	Issue(userID int) (core_auth.AccessToken, error)
}

func NewAuthService(usersRepository UsersRepository, tokenManager TokenManager) *AuthService {
	return &AuthService{
		usersRepository: usersRepository,
		tokenManager:    tokenManager,
	}
}
