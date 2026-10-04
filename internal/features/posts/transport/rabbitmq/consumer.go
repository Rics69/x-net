package posts_transport_rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	core_logger "github.com/Rics69/x-net/internal/core/logger"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

// сколько неподтверждённых сообщений брокер отдаёт консьюмеру за раз.
// Без лимита брокер вывалит в память приложения всю очередь
const prefetchCount = 50

type Broadcaster interface {
	Broadcast(msg []byte)
}

type PostsEventsConsumer struct {
	broadcaster Broadcaster
	log         *core_logger.Logger
}

func NewPostsEventsConsumer(broadcaster Broadcaster, log *core_logger.Logger) *PostsEventsConsumer {
	return &PostsEventsConsumer{
		broadcaster: broadcaster,
		log:         log.With(zap.String("component", "posts_events_consumer")),
	}
}

// регистрируется в core_rabbitmq.Client.OnConnect - вызывается после каждого переподключения
func (c *PostsEventsConsumer) Setup(ctx context.Context, conn *amqp.Connection) error {
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("open consumer channel: %w", err)
	}

	// объявление идемпотентно: если exchange уже есть с теми же параметрами - ничего не происходит.
	// Объявляют все инстансы, поэтому неважно, кто стартанул первым
	err = ch.ExchangeDeclare(
		PostsEventsExchange,
		amqp.ExchangeFanout,
		true,  // durable: exchange переживёт рестарт брокера
		false, // autoDelete
		false, // internal
		false, // noWait
		nil,
	)
	if err != nil {
		return fmt.Errorf("declare exchange='%s': %w", PostsEventsExchange, err)
	}

	// очередь своя у каждого инстанса и живёт ровно столько, сколько его соединение:
	// exclusive - только наше соединение, autoDelete - удалится, когда инстанс отключится.
	// Упал инстанс - очередь не копит события, которые потом некому раздать
	queue, err := ch.QueueDeclare(
		instanceQueueName(),
		false, // durable
		true,  // autoDelete
		true,  // exclusive
		false, // noWait
		nil,
	)
	if err != nil {
		return fmt.Errorf("declare queue: %w", err)
	}

	if err := ch.QueueBind(queue.Name, "", PostsEventsExchange, false, nil); err != nil {
		return fmt.Errorf("bind queue='%s': %w", queue.Name, err)
	}

	if err := ch.Qos(prefetchCount, 0, false); err != nil {
		return fmt.Errorf("set qos: %w", err)
	}

	deliveries, err := ch.ConsumeWithContext(
		ctx,
		queue.Name,
		"",    // consumer tag: сгенерит библиотека
		false, // autoAck: подтверждаем сами после обработки
		true,  // exclusive
		false, // noLocal (RabbitMQ его не поддерживает)
		false, // noWait
		nil,
	)
	if err != nil {
		return fmt.Errorf("consume queue='%s': %w", queue.Name, err)
	}

	c.log.Info("consuming posts events", zap.String("queue", queue.Name))

	// канал deliveries закроется при обрыве соединения - горутина выйдет сама,
	// а после переподключения Setup запустит новую
	go c.consume(deliveries)

	return nil
}

func (c *PostsEventsConsumer) consume(deliveries <-chan amqp.Delivery) {
	for delivery := range deliveries {
		// "ядовитое" сообщение (кривой JSON) бессмысленно возвращать в очередь:
		// оно будет падать снова и снова по кругу. Nack без requeue - брокер его выбросит
		// (или отправит в dead letter exchange, если он настроен)
		if !json.Valid(delivery.Body) {
			c.log.Warn(
				"invalid posts event, dropping",
				zap.String("message_id", delivery.MessageId),
				zap.String("type", delivery.Type),
			)

			_ = delivery.Nack(false, false)

			continue
		}

		c.broadcaster.Broadcast(delivery.Body)

		if err := delivery.Ack(false); err != nil {
			c.log.Warn("ack posts event", zap.Error(err))
		}
	}

	c.log.Debug("deliveries channel closed")
}

// имя задаём сами (а не "" -> amq.gen-... от брокера): в Management UI сразу видно, чья очередь.
// uuid новый на каждое подключение: после обрыва брокер может ещё не заметить смерть
// старого соединения, и exclusive-очередь со старым именем будет занята (RESOURCE_LOCKED)
func instanceQueueName() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	return fmt.Sprintf("%s.%s.%s", PostsEventsExchange, hostname, uuid.NewString()[:8])
}
