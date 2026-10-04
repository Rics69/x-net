package posts_postgres_repository

import core_postgres_gorm "github.com/Rics69/x-net/internal/core/repository/postgres/gorm"

type PostsRepository struct {
	db *core_postgres_gorm.DB
}

func NewPostsRepository(db *core_postgres_gorm.DB) *PostsRepository {
	return &PostsRepository{
		db: db,
	}
}
