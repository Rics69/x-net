package posts_transport_ws

import (
	"time"

	"github.com/Rics69/x-net/internal/core/domain"
)

const (
	EventTypePostCreated = "post.created"
	EventTypePostDeleted = "post.deleted"
)

// конверт для всех сообщений в сокете: фронт смотрит на type и решает, что делать с data.
// Новые типы событий (лайки, комменты) добавляются без смены протокола
type Event struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

// та же форма, что PostDTOResponse в HTTP - фронт рендерит пост из ленты
// и пост из сокета одной функцией
type PostEventDTO struct {
	ID        int                `json:"id"`
	Content   string             `json:"content"`
	CreatedAt time.Time          `json:"created_at"`
	Author    PostAuthorEventDTO `json:"author"`
}

type PostAuthorEventDTO struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	FullName string `json:"full_name"`
}

type PostDeletedEventDTO struct {
	ID int `json:"id"`
}

func postEventDTOFromDomain(post domain.Post) PostEventDTO {
	dto := PostEventDTO{
		ID:        post.ID,
		Content:   post.Content,
		CreatedAt: post.CreatedAt,
		Author: PostAuthorEventDTO{
			ID: post.AuthorUserID,
		},
	}

	if post.Author != nil {
		dto.Author.Username = post.Author.Username
		dto.Author.FullName = post.Author.FullName
	}

	return dto
}
