package core_ws

import (
	"errors"
	"time"

	core_logger "github.com/Rics69/x-net/internal/core/logger"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const (
	// сколько ждём, пока сообщение уйдёт в сокет
	writeWait = 10 * time.Second

	// если за это время от клиента не пришло ничего (даже pong) - считаем соединение мёртвым.
	// Без этого оборванное соединение (выдернули wifi) висело бы в хабе вечно:
	// TCP сам по себе обрыв не замечает, пока не попробуешь записать
	pongWait = 60 * time.Second

	// пингуем чаще, чем pongWait, чтобы pong успел прийти до дедлайна
	pingPeriod = pongWait * 9 / 10

	// клиент нам ничего не шлёт (лента только на чтение), ограничиваем,
	// чтобы нельзя было забить память огромным фреймом
	maxMessageSize = 512

	sendBufferSize = 64
)

type client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	userID int
}

// gorilla разрешает только ОДНОГО писателя и ОДНОГО читателя на соединение одновременно.
// Поэтому вся запись (сообщения, ping, close) идёт только из writePump,
// а всё чтение - только из readPump
func (c *client) writePump() {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			// канал закрыл хаб: шатдаун или клиент слишком медленный
			if !ok {
				_ = c.conn.WriteMessage(
					websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseGoingAway, ""),
				)

				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// читаем, даже если клиент ничего не присылает: только чтение обрабатывает
// control frames (pong, close) и замечает, что соединение закрылось
func (c *client) readPump(log *core_logger.Logger) {
	defer func() {
		c.hub.unregisterClient(c)
		_ = c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			// закрыли вкладку / ушли со страницы - норма, остальное логируем
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) &&
				!errors.Is(err, websocket.ErrCloseSent) {
				log.Debug("ws read error", zap.Error(err))
			}

			return
		}
	}
}
