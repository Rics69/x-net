package users_postgres_repository

import (
	"context"
	"fmt"

	core_errors "github.com/Rics69/x-net/internal/core/errors"
)

func (r *UsersRepository) DeleteUser(ctx context.Context, id int) error {
	ctx, cancel := context.WithTimeout(ctx, r.db.OpTimeout())
	defer cancel()

	// UserModel не встраивает gorm.Model (нет DeletedAt), поэтому удаление настоящее, не soft delete
	result := r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&UserModel{})
	if result.Error != nil {
		return fmt.Errorf("delete user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user with id='%d': %w", id, core_errors.ErrNotFound)
	}

	return nil
}
