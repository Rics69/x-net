package posts_service

import (
	"context"

	"github.com/Rics69/x-net/internal/core/domain"
)

type PostsService struct {
	postsRepository PostsRepository
	eventsPublisher PostsEventsPublisher
}

type PostsRepository interface {
	CreatePost(ctx context.Context, post domain.Post) (domain.Post, error)
	GetPost(ctx context.Context, id int) (domain.Post, error)
	GetFeed(ctx context.Context, cursor *domain.PostsCursor, limit int) ([]domain.Post, error)
	DeletePost(ctx context.Context, id int) error
}

// сервис не знает, КАК доставляются события: сейчас это WebSocket-хаб в этом же процессе,
// потом будет RabbitMQ - поменяется только реализация, сервис не трогаем
type PostsEventsPublisher interface {
	PublishPostCreated(ctx context.Context, post domain.Post) error
	PublishPostDeleted(ctx context.Context, postID int) error
}

func NewPostsService(postsRepository PostsRepository, eventsPublisher PostsEventsPublisher) *PostsService {
	return &PostsService{
		postsRepository: postsRepository,
		eventsPublisher: eventsPublisher,
	}
}
