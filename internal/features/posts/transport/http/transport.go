package posts_transport_http

import (
	"context"
	"net/http"

	"github.com/Rics69/x-net/internal/core/domain"
	core_http_middleware "github.com/Rics69/x-net/internal/core/transport/http/middleware"
	core_http_server "github.com/Rics69/x-net/internal/core/transport/http/server"
)

type PostsHTTPHandler struct {
	postsService   PostsService
	authMiddleware core_http_middleware.Middleware
}

type PostsService interface {
	CreatePost(
		ctx context.Context,
		post domain.Post,
	) (domain.Post, error)
	GetPost(
		ctx context.Context,
		id int,
	) (domain.Post, error)
	GetFeed(
		ctx context.Context,
		cursor *domain.PostsCursor,
		limit *int,
	) (domain.PostsPage, error)
	DeletePost(
		ctx context.Context,
		id int,
		userID int,
	) error
}

func NewPostsHTTPHandler(
	postsService PostsService,
	authMiddleware core_http_middleware.Middleware,
) *PostsHTTPHandler {
	return &PostsHTTPHandler{
		postsService:   postsService,
		authMiddleware: authMiddleware,
	}
}

func (h *PostsHTTPHandler) Routes() []core_http_server.Route {
	authOnly := []core_http_middleware.Middleware{h.authMiddleware}

	return []core_http_server.Route{
		{
			Method:     http.MethodPost,
			Path:       "/posts",
			Handler:    h.CreatePost,
			Middleware: authOnly,
		},
		{
			Method:     http.MethodGet,
			Path:       "/posts",
			Handler:    h.GetFeed,
			Middleware: authOnly,
		},
		{
			Method:     http.MethodGet,
			Path:       "/posts/{id}",
			Handler:    h.GetPost,
			Middleware: authOnly,
		},
		{
			Method:     http.MethodDelete,
			Path:       "/posts/{id}",
			Handler:    h.DeletePost,
			Middleware: authOnly,
		},
	}
}
