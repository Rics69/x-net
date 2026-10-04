package users_postgres_repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Rics69/x-net/internal/core/domain"
	core_errors "github.com/Rics69/x-net/internal/core/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (r *UsersRepository) CreateUser(ctx context.Context, user domain.User) (domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.db.OpTimeout())
	defer cancel()

	// ID и Version из домена не переносим (там Unitialized = -1),
	// иначе GORM вставит id=-1. Их проставит БД
	userModel := UserModel{
		Username:     user.Username,
		FullName:     user.FullName,
		PhoneNumber:  user.PhoneNumber,
		CreatedAt:    user.CreatedAt,
		PasswordHash: user.PasswordHash,
	}

	// RETURNING * - модель заполнится тем, что реально легло в таблицу
	err := r.db.WithContext(ctx).
		Clauses(clause.Returning{}).
		Create(&userModel).
		Error
	if err != nil {
		// проверку "username свободен?" не делаем отдельным SELECT до INSERT:
		// между ними другой запрос может успеть занять username (race).
		// Уникальный индекс в БД - единственная надёжная проверка
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return domain.User{}, fmt.Errorf(
				"username='%s' already taken: %w",
				user.Username,
				core_errors.ErrConflict,
			)
		}

		return domain.User{}, fmt.Errorf("insert user: %w", err)
	}

	return userDomainFromModel(userModel), nil
}
