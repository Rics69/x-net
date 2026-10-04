package core_rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"os"
	"sync"
	"time"

	core_logger "github.com/Rics69/x-net/internal/core/logger"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

const (
	dialTimeout = 5 * time.Second

	// heartbeat - пинг между клиентом и брокером. Без него оборванное TCP-соединение
	// (умер брокер, отвалилась сеть) мы бы заметили только при следующей попытке записи
	heartbeat = 10 * time.Second
)

var ErrNotConnected = errors.New("rabbitmq: not connected")

// вызывается после КАЖДОГО (пере)подключения. После обрыва на брокере исчезает всё,
// что было привязано к соединению (exclusive-очереди, консьюмеры), поэтому
// объявлять топологию и подписываться нужно заново, а не один раз на старте
type SetupFunc func(ctx context.Context, conn *amqp.Connection) error

// amqp091-go сам не переподключается: при обрыве Connection умирает насовсем.
// Client держит соединение живым: Run в цикле подключается, ждёт обрыва и подключается снова
type Client struct {
	config Config
	log    *core_logger.Logger

	setups []SetupFunc

	// канал для публикации. AMQP-канал не рассчитан на одновременную запись
	// из разных горутин, а публикуют сразу все HTTP-хендлеры - поэтому под мьютексом.
	// nil - соединения сейчас нет
	mu        sync.Mutex
	publishCh *amqp.Channel
}

func NewClient(config Config, log *core_logger.Logger) *Client {
	return &Client{
		config: config,
		log:    log.With(zap.String("component", "rabbitmq")),
	}
}

// регистрировать до Run
func (c *Client) OnConnect(setup SetupFunc) {
	c.setups = append(c.setups, setup)
}

func (c *Client) Run(ctx context.Context) {
	attempt := 0

	for {
		conn, publishCh, err := c.connect(ctx)
		if err != nil {
			delay := c.backoff(attempt)
			attempt++

			c.log.Warn("connect failed, retrying", zap.Duration("in", delay), zap.Error(err))

			select {
			case <-time.After(delay):
				continue
			case <-ctx.Done():
				return
			}
		}

		attempt = 0
		c.log.Info("connected", zap.String("host", c.config.Host), zap.Int("port", c.config.Port))

		// библиотека шлёт в эти каналы ровно один раз и закрывает их.
		// Буфер 1 обязателен: без читателя библиотека заблокируется на отправке
		connClosed := conn.NotifyClose(make(chan *amqp.Error, 1))
		publishChClosed := publishCh.NotifyClose(make(chan *amqp.Error, 1))

		c.setPublishChannel(publishCh)

		select {
		case <-ctx.Done():
			c.setPublishChannel(nil)

			// штатное закрытие: каналы доставки у консьюмеров закроются, их горутины выйдут
			if err := conn.Close(); err != nil {
				c.log.Warn("close connection", zap.Error(err))
			}

			c.log.Info("disconnected")

			return

		case amqpErr := <-connClosed:
			c.setPublishChannel(nil)
			c.log.Warn("connection lost", zap.Any("reason", amqpErr))

		// канал закрывается брокером при ошибке на нём (например publish в несуществующий exchange),
		// а соединение при этом живо. Чинить по частям сложнее, чем пересоздать всё целиком
		case amqpErr := <-publishChClosed:
			c.setPublishChannel(nil)
			c.log.Warn("publish channel closed, reconnecting", zap.Any("reason", amqpErr))
			_ = conn.Close()
		}
	}
}

func (c *Client) Publish(ctx context.Context, exchange string, routingKey string, msg amqp.Publishing) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.publishCh == nil {
		return ErrNotConnected
	}

	// без publisher confirms: успех тут = "отдали в сокет", а не "брокер принял".
	// Для живой ленты ок (потерю закроет догрузка ленты на фронте),
	// для денег/заказов нужен ch.Confirm + ожидание подтверждения
	if err := c.publishCh.PublishWithContext(ctx, exchange, routingKey, false, false, msg); err != nil {
		return fmt.Errorf("publish to exchange='%s': %w", exchange, err)
	}

	return nil
}

func (c *Client) connect(ctx context.Context) (*amqp.Connection, *amqp.Channel, error) {
	properties := amqp.NewConnectionProperties()
	// имя видно в Management UI во вкладке Connections - сразу понятно, какой это инстанс
	hostname, _ := os.Hostname()
	properties.SetClientConnectionName("x-net@" + hostname)

	conn, err := amqp.DialConfig(c.url(), amqp.Config{
		Heartbeat:  heartbeat,
		Properties: properties,
		Dial:       amqp.DefaultDial(dialTimeout),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("dial: %w", err)
	}

	publishCh, err := conn.Channel()
	if err != nil {
		_ = conn.Close()

		return nil, nil, fmt.Errorf("open publish channel: %w", err)
	}

	for _, setup := range c.setups {
		if err := setup(ctx, conn); err != nil {
			_ = conn.Close()

			return nil, nil, fmt.Errorf("setup: %w", err)
		}
	}

	return conn, publishCh, nil
}

func (c *Client) setPublishChannel(ch *amqp.Channel) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.publishCh = ch
}

// amqp.URI, а не fmt.Sprintf("amqp://%s:%s@..."): пароль с '@' или '/' сломал бы строку.
// И не url.URL руками: vhost "/" надо закодировать как "%2F" ровно один раз,
// url.URL с уже экранированным Path экранирует его повторно ("%252F") -
// брокер увидит vhost с именем "%2F" и ответит 403 no access to this vhost
func (c *Client) url() string {
	uri := amqp.URI{
		Scheme:   "amqp",
		Host:     c.config.Host,
		Port:     c.config.Port,
		Username: c.config.User,
		Password: c.config.Password,
		Vhost:    c.config.VHost,
	}

	return uri.String()
}

// та же идея, что на фронте: 1с, 2с, 4с ... до максимума, со случайным разбросом,
// чтобы все инстансы не долбили поднявшийся брокер синхронно
func (c *Client) backoff(attempt int) time.Duration {
	base := time.Second << min(attempt, 10)
	if base > c.config.ReconnectMaxDelay {
		base = c.config.ReconnectMaxDelay
	}

	return base/2 + rand.N(base/2+1)
}
