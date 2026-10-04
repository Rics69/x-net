package posts_service

import (
	"context"
	"fmt"

	"github.com/Rics69/x-net/internal/core/domain"
	core_errors "github.com/Rics69/x-net/internal/core/errors"
)

const (
	defaultFeedLimit = 20
	maxFeedLimit     = 100
)

func (s *PostsService) GetFeed(
	ctx context.Context,
	cursor *domain.PostsCursor,
	limit *int,
) (domain.PostsPage, error) {
	pageSize := defaultFeedLimit
	if limit != nil {
		if *limit < 1 || *limit > maxFeedLimit {
			return domain.PostsPage{}, fmt.Errorf(
				"limit must be between 1 and %d: %w",
				maxFeedLimit,
				core_errors.ErrInvalidArgument,
			)
		}

		pageSize = *limit
	}

	// просим на 1 пост больше, чем отдадим: если он пришёл - следующая страница есть.
	// Так не нужен отдельный COUNT(*) (на большой таблице он дорогой),
	// и клиент не делает лишний запрос за пустой страницей
	posts, err := s.postsRepository.GetFeed(ctx, cursor, pageSize+1)
	if err != nil {
		return domain.PostsPage{}, fmt.Errorf("get feed from repository: %w", err)
	}

	page := domain.PostsPage{
		Posts: posts,
	}

	if len(posts) > pageSize {
		page.Posts = posts[:pageSize]

		// курсор = последний ОТДАННЫЙ пост, а не лишний (pageSize+1)-й:
		// следующий запрос возьмёт всё строго старше него, и лишний пост придёт первым
		last := page.Posts[len(page.Posts)-1]
		nextCursor := domain.NewPostsCursor(last.CreatedAt, last.ID)
		page.NextCursor = &nextCursor
	}

	return page, nil
}
