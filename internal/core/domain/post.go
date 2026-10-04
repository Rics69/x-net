package domain

import (
	"fmt"
	"strings"
	"time"

	core_errors "github.com/Rics69/x-net/internal/core/errors"
)

const maxPostContentLen = 280

// посты не редактируются (как в твиттере), поэтому без Version и без PostPatch
type Post struct {
	ID int

	Content   string
	CreatedAt time.Time

	AuthorUserID int

	// заполняется при чтении (JOIN users), при создании пустой
	Author *PostAuthor
}

// не User целиком: ленте нужны только публичные поля автора
type PostAuthor struct {
	ID       int
	Username string
	FullName string
}

func NewPost(
	id int,
	content string,
	createdAt time.Time,
	authorUserID int,
	author *PostAuthor,
) Post {
	return Post{
		ID:           id,
		Content:      content,
		CreatedAt:    createdAt,
		AuthorUserID: authorUserID,
		Author:       author,
	}
}

func NewPostUnitialized(content string, authorUserID int) Post {
	return NewPost(
		UnitializedID,
		content,
		time.Now(),
		authorUserID,
		nil,
	)
}

func NewPostAuthor(id int, username string, fullName string) PostAuthor {
	return PostAuthor{
		ID:       id,
		Username: username,
		FullName: fullName,
	}
}

func (p *Post) Validate() error {
	if strings.TrimSpace(p.Content) == "" {
		return fmt.Errorf("'Content' can't be empty: %w", core_errors.ErrInvalidArgument)
	}

	contentLen := len([]rune(p.Content))
	if contentLen > maxPostContentLen {
		return fmt.Errorf(
			"invalid 'Content' len: %d, max %d: %w",
			contentLen,
			maxPostContentLen,
			core_errors.ErrInvalidArgument,
		)
	}

	return nil
}

// позиция в ленте, с которой отдавать следующую страницу (keyset pagination).
// Пара (CreatedAt, ID), а не только CreatedAt: два поста могут быть созданы
// в одну микросекунду, ID разруливает ничью и делает порядок строгим
type PostsCursor struct {
	CreatedAt time.Time
	ID        int
}

func NewPostsCursor(createdAt time.Time, id int) PostsCursor {
	return PostsCursor{
		CreatedAt: createdAt,
		ID:        id,
	}
}

type PostsPage struct {
	Posts []Post

	// nil - дальше постов нет
	NextCursor *PostsCursor
}
