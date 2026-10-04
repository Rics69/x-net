package users_postgres_repository

import (
	"context"
	"fmt"

	"github.com/Rics69/x-net/internal/core/domain"
)

func (r *UsersRepository) GetUsers(ctx context.Context, limit *int, offset *int) ([]domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.db.OpTimeout())
	defer cancel()

	query := r.db.WithContext(ctx).Order("id ASC")

	// раньше nil уходил в LIMIT $1 как NULL (= без лимита),
	// в GORM просто не навешиваем Limit/Offset
	if limit != nil {
		query = query.Limit(*limit)
	}

	if offset != nil {
		query = query.Offset(*offset)
	}

	var usersModels []UserModel
	if err := query.Find(&usersModels).Error; err != nil {
		return nil, fmt.Errorf("select users: %w", err)
	}

	return userDomainsFromModels(usersModels), nil
}
