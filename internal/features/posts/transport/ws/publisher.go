package posts_transport_ws

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Rics69/x-net/internal/core/domain"
)

type Broadcaster interface {
	Broadcast(msg []byte)
}

// реализация PostsEventsPublisher "в лоб": сразу в WebSocket-хаб этого процесса.
// Работает, пока инстанс приложения один
type PostsEventsPublisher struct {
	broadcaster Broadcaster
}

func NewPostsEventsPublisher(broadcaster Broadcaster) *PostsEventsPublisher {
	return &PostsEventsPublisher{
		broadcaster: broadcaster,
	}
}

func (p *PostsEventsPublisher) PublishPostCreated(_ context.Context, post domain.Post) error {
	return p.publish(Event{
		Type: EventTypePostCreated,
		Data: postEventDTOFromDomain(post),
	})
}

func (p *PostsEventsPublisher) PublishPostDeleted(_ context.Context, postID int) error {
	return p.publish(Event{
		Type: EventTypePostDeleted,
		Data: PostDeletedEventDTO{ID: postID},
	})
}

// JSON собираем один раз и раздаём всем клиентам одинаковые байты,
// а не маршалим заново под каждое соединение
func (p *PostsEventsPublisher) publish(event Event) error {
	msg, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal '%s' event: %w", event.Type, err)
	}

	p.broadcaster.Broadcast(msg)

	return nil
}
