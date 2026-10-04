package posts_transport_rabbitmq

import (
	"context"
	"fmt"
	"time"

	"github.com/Rics69/x-net/internal/core/domain"
	posts_transport_ws "github.com/Rics69/x-net/internal/features/posts/transport/ws"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher interface {
	Publish(ctx context.Context, exchange string, routingKey string, msg amqp.Publishing) error
}

// реализация PostsEventsPublisher через брокер. В свой хаб напрямую НЕ пишет:
// свой инстанс тоже получит событие из своей очереди. Один путь доставки - нет дублей
type PostsEventsPublisher struct {
	publisher Publisher
}

func NewPostsEventsPublisher(publisher Publisher) *PostsEventsPublisher {
	return &PostsEventsPublisher{
		publisher: publisher,
	}
}

func (p *PostsEventsPublisher) PublishPostCreated(ctx context.Context, post domain.Post) error {
	body, err := posts_transport_ws.MarshalPostCreatedEvent(post)
	if err != nil {
		return err
	}

	return p.publish(ctx, posts_transport_ws.EventTypePostCreated, body)
}

func (p *PostsEventsPublisher) PublishPostDeleted(ctx context.Context, postID int) error {
	body, err := posts_transport_ws.MarshalPostDeletedEvent(postID)
	if err != nil {
		return err
	}

	return p.publish(ctx, posts_transport_ws.EventTypePostDeleted, body)
}

func (p *PostsEventsPublisher) publish(ctx context.Context, eventType string, body []byte) error {
	msg := amqp.Publishing{
		ContentType: "application/json",
		Type:        eventType,
		MessageId:   uuid.NewString(),
		Timestamp:   time.Now(),

		// Transient: брокер не пишет сообщение на диск. Событие ленты через минуту никому
		// не нужно, а клиенты, пропустившие его, догрузят ленту по HTTP
		DeliveryMode: amqp.Transient,

		Body: body,
	}

	if err := p.publisher.Publish(ctx, PostsEventsExchange, "", msg); err != nil {
		return fmt.Errorf("publish '%s': %w", eventType, err)
	}

	return nil
}
