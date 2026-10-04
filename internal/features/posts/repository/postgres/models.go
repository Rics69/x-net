package posts_postgres_repository

import (
	"time"

	"github.com/Rics69/x-net/internal/core/domain"
)

type PostModel struct {
	ID           int `gorm:"primaryKey"`
	Content      string
	CreatedAt    time.Time
	AuthorUserID int

	// belongs-to: GORM по имени поля Author ищет внешний ключ AuthorID,
	// а у нас author_user_id - поэтому foreignKey указываем явно
	Author *PostAuthorModel `gorm:"foreignKey:AuthorUserID"`
}

func (PostModel) TableName() string {
	return "xchat.posts"
}

// своя урезанная модель над xchat.users, а не UserModel из фичи users:
// 1) фичи не импортируют репозитории друг друга
// 2) в SELECT попадут только эти 3 колонки, password_hash в ленту даже не читается
type PostAuthorModel struct {
	ID       int
	Username string
	FullName string
}

func (PostAuthorModel) TableName() string {
	return "xchat.users"
}

func postDomainFromModel(post PostModel) domain.Post {
	var author *domain.PostAuthor
	if post.Author != nil {
		a := domain.NewPostAuthor(post.Author.ID, post.Author.Username, post.Author.FullName)
		author = &a
	}

	return domain.NewPost(
		post.ID,
		post.Content,
		post.CreatedAt,
		post.AuthorUserID,
		author,
	)
}

func postDomainsFromModels(posts []PostModel) []domain.Post {
	domains := make([]domain.Post, len(posts))

	for i, post := range posts {
		domains[i] = postDomainFromModel(post)
	}

	return domains
}
