package posts_transport_ws

import (
	"context"

	"github.com/Rics69/x-net/internal/core/domain"
)

type Broadcaster interface {
	Broadcast(msg []byte)
}

// реализация PostsEventsPublisher "в лоб": сразу в WebSocket-хаб этого процесса.
// Работает, пока инстанс приложения один (RABBITMQ_ENABLED=false)
type PostsEventsPublisher struct {
	broadcaster Broadcaster
}

func NewPostsEventsPublisher(broadcaster Broadcaster) *PostsEventsPublisher {
	return &PostsEventsPublisher{
		broadcaster: broadcaster,
	}
}

// JSON собираем один раз и раздаём всем клиентам одинаковые байты,
// а не маршалим заново под каждое соединение
func (p *PostsEventsPublisher) PublishPostCreated(_ context.Context, post domain.Post) error {
	msg, err := MarshalPostCreatedEvent(post)
	if err != nil {
		return err
	}

	p.broadcaster.Broadcast(msg)

	return nil
}

func (p *PostsEventsPublisher) PublishPostDeleted(_ context.Context, postID int) error {
	msg, err := MarshalPostDeletedEvent(postID)
	if err != nil {
		return err
	}

	p.broadcaster.Broadcast(msg)

	return nil
}
