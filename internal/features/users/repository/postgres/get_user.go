package users_postgres_repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Rics69/x-net/internal/core/domain"
	core_errors "github.com/Rics69/x-net/internal/core/errors"
	"gorm.io/gorm"
)

func (r *UsersRepository) GetUser(ctx context.Context, id int) (domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.db.OpTimeout())
	defer cancel()

	var userModel UserModel

	// Take, а не First: First добавляет ORDER BY id, для поиска по PK это лишнее.
	// ErrRecordNotFound кидают только First/Take/Last, Find на пустой выборке ошибку не вернёт
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		Take(&userModel).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.User{}, fmt.Errorf(
				"user with id='%d': %w",
				id,
				core_errors.ErrNotFound,
			)
		}

		return domain.User{}, fmt.Errorf("select user: %w", err)
	}

	return userDomainFromModel(userModel), nil
}
