package posts_postgres_repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Rics69/x-net/internal/core/domain"
	core_errors "github.com/Rics69/x-net/internal/core/errors"
	"gorm.io/gorm"
)

func (r *PostsRepository) CreatePost(ctx context.Context, post domain.Post) (domain.Post, error) {
	ctx, cancel := context.WithTimeout(ctx, r.db.OpTimeout())
	defer cancel()

	postModel := PostModel{
		Content:      post.Content,
		CreatedAt:    post.CreatedAt,
		AuthorUserID: post.AuthorUserID,
	}

	// Omit("Author"): иначе GORM при Create попробует сохранить и связанную модель (upsert в users)
	err := r.db.WithContext(ctx).
		Omit("Author").
		Create(&postModel).
		Error
	if err != nil {
		// юзер удалил аккаунт, а токен ещё жив
		if errors.Is(err, gorm.ErrForeignKeyViolated) {
			return domain.Post{}, fmt.Errorf(
				"author with id='%d': %w",
				post.AuthorUserID,
				core_errors.ErrNotFound,
			)
		}

		return domain.Post{}, fmt.Errorf("insert post: %w", err)
	}

	// перечитываем вместе с автором: ответу (и рассылке по WebSocket потом)
	// нужен username автора, а при INSERT его нет
	return r.GetPost(ctx, postModel.ID)
}
