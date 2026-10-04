package posts_postgres_repository

import (
	"context"
	"fmt"

	core_errors "github.com/Rics69/x-net/internal/core/errors"
)

func (r *PostsRepository) DeletePost(ctx context.Context, id int) error {
	ctx, cancel := context.WithTimeout(ctx, r.db.OpTimeout())
	defer cancel()

	result := r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&PostModel{})
	if result.Error != nil {
		return fmt.Errorf("delete post: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("post with id='%d': %w", id, core_errors.ErrNotFound)
	}

	return nil
}
