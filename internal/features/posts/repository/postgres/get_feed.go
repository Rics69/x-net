package posts_postgres_repository

import (
	"context"
	"fmt"

	"github.com/Rics69/x-net/internal/core/domain"
)

func (r *PostsRepository) GetFeed(
	ctx context.Context,
	cursor *domain.PostsCursor,
	limit int,
) ([]domain.Post, error) {
	ctx, cancel := context.WithTimeout(ctx, r.db.OpTimeout())
	defer cancel()

	query := r.db.WithContext(ctx).
		Joins("Author").
		Order("posts.created_at DESC, posts.id DESC").
		Limit(limit)

	// keyset: "всё, что строго старше последнего поста предыдущей страницы".
	// Row comparison (a, b) < (x, y) = a < x OR (a = x AND b < y),
	// и постгрес умеет искать такое условие по составному индексу (created_at, id)
	if cursor != nil {
		query = query.Where(
			"(posts.created_at, posts.id) < (?, ?)",
			cursor.CreatedAt,
			cursor.ID,
		)
	}

	var postModels []PostModel
	if err := query.Find(&postModels).Error; err != nil {
		return nil, fmt.Errorf("select feed: %w", err)
	}

	return postDomainsFromModels(postModels), nil
}
