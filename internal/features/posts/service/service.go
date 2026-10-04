package posts_service

import (
	"context"

	"github.com/Rics69/x-net/internal/core/domain"
)

type PostsService struct {
	postsRepository PostsRepository
}

type PostsRepository interface {
	CreatePost(ctx context.Context, post domain.Post) (domain.Post, error)
	GetPost(ctx context.Context, id int) (domain.Post, error)
	GetFeed(ctx context.Context, cursor *domain.PostsCursor, limit int) ([]domain.Post, error)
	DeletePost(ctx context.Context, id int) error
}

func NewPostsService(postsRepository PostsRepository) *PostsService {
	return &PostsService{
		postsRepository: postsRepository,
	}
}
