package posts_postgres_repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Rics69/x-net/internal/core/domain"
	core_errors "github.com/Rics69/x-net/internal/core/errors"
	"gorm.io/gorm"
)

func (r *PostsRepository) GetPost(ctx context.Context, id int) (domain.Post, error) {
	ctx, cancel := context.WithTimeout(ctx, r.db.OpTimeout())
	defer cancel()

	var postModel PostModel

	// Joins("Author") - один запрос с LEFT JOIN users.
	// Альтернатива Preload("Author") - второй запрос SELECT ... FROM users WHERE id IN (...).
	// Колонки обеих таблиц пересекаются (id), поэтому в Where имя таблицы указываем явно
	err := r.db.WithContext(ctx).
		Joins("Author").
		Where("posts.id = ?", id).
		Take(&postModel).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Post{}, fmt.Errorf("post with id='%d': %w", id, core_errors.ErrNotFound)
		}

		return domain.Post{}, fmt.Errorf("select post: %w", err)
	}

	return postDomainFromModel(postModel), nil
}
