package users_transport_http

import (
	"context"
	"net/http"

	"github.com/Rics69/x-net/internal/core/domain"
	core_http_middleware "github.com/Rics69/x-net/internal/core/transport/http/middleware"
	core_http_server "github.com/Rics69/x-net/internal/core/transport/http/server"
)

type UsersHTTPHandler struct {
	usersService   UsersService
	authMiddleware core_http_middleware.Middleware
}

type UsersService interface {
	GetUsers(
		ctx context.Context,
		limit *int,
		offset *int,
	) ([]domain.User, error)
	GetUser(
		ctx context.Context,
		id int,
	) (domain.User, error)
	DeleteUser(
		ctx context.Context,
		id int,
	) error
	PatchUser(
		ctx context.Context,
		id int,
		patch domain.UserPatch,
	) (domain.User, error)
}

func NewUsersHTTPHandler(
	usersService UsersService,
	authMiddleware core_http_middleware.Middleware,
) *UsersHTTPHandler {
	return &UsersHTTPHandler{
		usersService:   usersService,
		authMiddleware: authMiddleware,
	}
}

// создание юзера переехало в POST /auth/register.
// Менять/удалять можно только себя, поэтому PATCH/DELETE на /users/me, а id берётся из токена.
// /users/me и /users/{id} не конфликтуют: ServeMux выбирает более конкретный паттерн
func (h *UsersHTTPHandler) Routes() []core_http_server.Route {
	authOnly := []core_http_middleware.Middleware{h.authMiddleware}

	return []core_http_server.Route{
		{
			Method:     http.MethodGet,
			Path:       "/users",
			Handler:    h.GetUsers,
			Middleware: authOnly,
		},
		{
			Method:     http.MethodGet,
			Path:       "/users/me",
			Handler:    h.GetMe,
			Middleware: authOnly,
		},
		{
			Method:     http.MethodGet,
			Path:       "/users/{id}",
			Handler:    h.GetUser,
			Middleware: authOnly,
		},
		{
			Method:     http.MethodPatch,
			Path:       "/users/me",
			Handler:    h.PatchUser,
			Middleware: authOnly,
		},
		{
			Method:     http.MethodDelete,
			Path:       "/users/me",
			Handler:    h.DeleteUser,
			Middleware: authOnly,
		},
	}
}
