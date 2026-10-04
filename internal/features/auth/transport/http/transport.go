package auth_transport_http

import (
	"context"
	"net/http"

	core_auth "github.com/Rics69/x-net/internal/core/auth"
	"github.com/Rics69/x-net/internal/core/domain"
	core_http_server "github.com/Rics69/x-net/internal/core/transport/http/server"
)

type AuthHTTPHandler struct {
	authService  AuthService
	cookieSecure bool
}

type AuthService interface {
	Register(
		ctx context.Context,
		user domain.User,
		password string,
	) (domain.User, core_auth.AccessToken, error)
	Login(
		ctx context.Context,
		username string,
		password string,
	) (domain.User, core_auth.AccessToken, error)
}

func NewAuthHTTPHandler(authService AuthService, cookieSecure bool) *AuthHTTPHandler {
	return &AuthHTTPHandler{
		authService:  authService,
		cookieSecure: cookieSecure,
	}
}

func (h *AuthHTTPHandler) Routes() []core_http_server.Route {
	return []core_http_server.Route{
		{
			Method:  http.MethodPost,
			Path:    "/auth/register",
			Handler: h.Register,
		},
		{
			Method:  http.MethodPost,
			Path:    "/auth/login",
			Handler: h.Login,
		},
		{
			Method:  http.MethodPost,
			Path:    "/auth/logout",
			Handler: h.Logout,
		},
	}
}
