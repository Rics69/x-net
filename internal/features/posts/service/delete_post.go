package posts_service

import (
	"context"
	"fmt"

	core_errors "github.com/Rics69/x-net/internal/core/errors"
	core_logger "github.com/Rics69/x-net/internal/core/logger"
	"go.uber.org/zap"
)

func (s *PostsService) DeletePost(ctx context.Context, id int, userID int) error {
	post, err := s.postsRepository.GetPost(ctx, id)
	if err != nil {
		return fmt.Errorf("get post: %w", err)
	}

	// пост публичный, поэтому честный 403, а не 404 "делаем вид, что поста нет".
	// Гонки тут нет: автор поста не меняется
	if post.AuthorUserID != userID {
		return fmt.Errorf(
			"post with id='%d' belongs to another user: %w",
			id,
			core_errors.ErrForbidden,
		)
	}

	if err := s.postsRepository.DeletePost(ctx, id); err != nil {
		return fmt.Errorf("delete post: %w", err)
	}

	// чтобы пост пропал из открытых лент без перезагрузки
	if err := s.eventsPublisher.PublishPostDeleted(ctx, id); err != nil {
		core_logger.FromContext(ctx).Warn(
			"failed to publish post deleted event",
			zap.Int("post_id", id),
			zap.Error(err),
		)
	}

	return nil
}
