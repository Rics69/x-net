package posts_service

import (
	"context"
	"fmt"

	core_errors "github.com/Rics69/x-net/internal/core/errors"
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

	return nil
}
