package users_postgres_repository

import (
	"context"
	"fmt"

	"github.com/Rics69/x-net/internal/core/domain"
	core_errors "github.com/Rics69/x-net/internal/core/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (r *UsersRepository) PatchUser(ctx context.Context, id int, user domain.User) (domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.db.OpTimeout())
	defer cancel()

	var userModel UserModel

	// Updates через map, а не через структуру: со структурой GORM пропускает нулевые поля,
	// и phone_number=nil просто не попал бы в UPDATE (нельзя было бы очистить телефон).
	// Optimistic lock: WHERE version = старая версия, если 0 строк - кто-то обновил раньше нас
	result := r.db.WithContext(ctx).
		Model(&userModel).
		Clauses(clause.Returning{}).
		Where("id = ? AND version = ?", id, user.Version).
		Updates(map[string]any{
			"full_name":    user.FullName,
			"phone_number": user.PhoneNumber,
			"version":      gorm.Expr("version + 1"),
		})
	if result.Error != nil {
		return domain.User{}, fmt.Errorf("update user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return domain.User{}, fmt.Errorf(
			"user with id='%d' concurrently accessed: %w",
			id,
			core_errors.ErrConflict,
		)
	}

	return userDomainFromModel(userModel), nil
}
