package posts_transport_ws

import (
	"encoding/json"
	"fmt"
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

// одни и те же байты уходят и в WebSocket, и в RabbitMQ: консьюмер на другом инстансе
// просто отдаёт тело сообщения в свой хаб, не перекладывая JSON
func MarshalPostCreatedEvent(post domain.Post) ([]byte, error) {
	return marshalEvent(Event{
		Type: EventTypePostCreated,
		Data: postEventDTOFromDomain(post),
	})
}

func MarshalPostDeletedEvent(postID int) ([]byte, error) {
	return marshalEvent(Event{
		Type: EventTypePostDeleted,
		Data: PostDeletedEventDTO{ID: postID},
	})
}

func marshalEvent(event Event) ([]byte, error) {
	msg, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("marshal '%s' event: %w", event.Type, err)
	}

	return msg, nil
}
