package users_postgres_repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Rics69/x-net/internal/core/domain"
	core_errors "github.com/Rics69/x-net/internal/core/errors"
	"gorm.io/gorm"
)

func (r *UsersRepository) GetUserByUsername(ctx context.Context, username string) (domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.db.OpTimeout())
	defer cancel()

	var userModel UserModel

	// lower() с обеих сторон - чтобы попасть в индекс users_username_lower_uidx
	err := r.db.WithContext(ctx).
		Where("lower(username) = lower(?)", username).
		Take(&userModel).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.User{}, fmt.Errorf(
				"user with username='%s': %w",
				username,
				core_errors.ErrNotFound,
			)
		}

		return domain.User{}, fmt.Errorf("select user by username: %w", err)
	}

	return userDomainFromModel(userModel), nil
}
