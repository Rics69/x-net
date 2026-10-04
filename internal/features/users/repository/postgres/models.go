package users_postgres_repository

import "github.com/Rics69/x-net/internal/core/domain"

type UserModel struct {
	ID int `gorm:"primaryKey"`

	// default:1 - при Create с нулевым Version GORM подставит 1 из тега (а не 0).
	// Значение должно совпадать с DEFAULT в миграции
	Version int `gorm:"default:1"`

	FullName    string
	PhoneNumber *string
}

// схема не public, поэтому имя таблицы задаём явно,
// иначе GORM по naming strategy пойдёт в "users"
func (UserModel) TableName() string {
	return "xchat.users"
}

func userDomainFromModel(user UserModel) domain.User {
	return domain.NewUser(
		user.ID,
		user.Version,
		user.FullName,
		user.PhoneNumber,
	)
}

func userDomainsFromModels(users []UserModel) []domain.User {
	userDomains := make([]domain.User, len(users))
	for i, user := range users {
		userDomains[i] = userDomainFromModel(user)
	}

	return userDomains
}
