package users_postgres_repository

import (
	"context"
	"fmt"

	"github.com/Rics69/x-net/internal/core/domain"
	"gorm.io/gorm/clause"
)

func (r *UsersRepository) CreateUser(ctx context.Context, user domain.User) (domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.db.OpTimeout())
	defer cancel()

	// ID и Version из домена не переносим (там Unitialized = -1),
	// иначе GORM вставит id=-1. Их проставит БД
	userModel := UserModel{
		FullName:    user.FullName,
		PhoneNumber: user.PhoneNumber,
	}

	// RETURNING * - модель заполнится тем, что реально легло в таблицу
	err := r.db.WithContext(ctx).
		Clauses(clause.Returning{}).
		Create(&userModel).
		Error
	if err != nil {
		return domain.User{}, fmt.Errorf("insert user: %w", err)
	}

	return userDomainFromModel(userModel), nil
}
