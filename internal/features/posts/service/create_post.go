package posts_service

import (
	"context"
	"fmt"

	"github.com/Rics69/x-net/internal/core/domain"
	core_logger "github.com/Rics69/x-net/internal/core/logger"
	"go.uber.org/zap"
)

func (s *PostsService) CreatePost(ctx context.Context, post domain.Post) (domain.Post, error) {
	if err := post.Validate(); err != nil {
		return domain.Post{}, fmt.Errorf("validate post domain: %w", err)
	}

	post, err := s.postsRepository.CreatePost(ctx, post)
	if err != nil {
		return domain.Post{}, fmt.Errorf("create post: %w", err)
	}

	// пост уже в БД, поэтому ошибка рассылки - не ошибка запроса (не отдаём 500 на сохранённый пост).
	// Минус: если рассылка упала, онлайн-клиенты пост не увидят до перезагрузки ленты.
	// Гарантированная доставка "сохранили => точно отправили" - это outbox, к нему вернёмся с RabbitMQ
	if err := s.eventsPublisher.PublishPostCreated(ctx, post); err != nil {
		core_logger.FromContext(ctx).Warn(
			"failed to publish post created event",
			zap.Int("post_id", post.ID),
			zap.Error(err),
		)
	}

	return post, nil
}
