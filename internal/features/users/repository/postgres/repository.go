package users_postgres_repository

import core_postgres_gorm "github.com/Rics69/x-net/internal/core/repository/postgres/gorm"

type UsersRepository struct {
	db *core_postgres_gorm.DB
}

func NewUsersRepository(db *core_postgres_gorm.DB) *UsersRepository {
	return &UsersRepository{
		db: db,
	}
}
