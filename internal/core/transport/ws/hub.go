package core_ws

import (
	"context"

	core_logger "github.com/Rics69/x-net/internal/core/logger"
	"go.uber.org/zap"
)

const broadcastBufferSize = 256

// Hub знает всех подключённых клиентов и рассылает им сообщения.
// Мапой clients владеет ОДНА горутина (Run), остальные общаются с ней через каналы -
// поэтому мьютекс не нужен ("share memory by communicating").
// Hub ничего не знает про посты: на вход []byte, на выход - всем клиентам
type Hub struct {
	log *core_logger.Logger

	clients    map[*client]struct{}
	register   chan *client
	unregister chan *client
	broadcast  chan []byte

	// закрывается, когда Run завершился: после этого отправки в каналы выше
	// заблокировались бы навсегда, а через select с done - сразу выходят
	done chan struct{}
}

func NewHub(log *core_logger.Logger) *Hub {
	return &Hub{
		log:        log,
		clients:    make(map[*client]struct{}),
		register:   make(chan *client),
		unregister: make(chan *client),
		// буфер, чтобы HTTP-хендлер CreatePost не ждал, пока хаб разошлёт предыдущее
		broadcast: make(chan []byte, broadcastBufferSize),
		done:      make(chan struct{}),
	}
}

func (h *Hub) Run(ctx context.Context) {
	defer close(h.done)

	for {
		select {
		case c := <-h.register:
			h.clients[c] = struct{}{}
			h.log.Debug("ws client connected", zap.Int("user_id", c.userID), zap.Int("clients", len(h.clients)))

		case c := <-h.unregister:
			h.removeClient(c)

		case msg := <-h.broadcast:
			for c := range h.clients {
				select {
				case c.send <- msg:
				default:
					// буфер клиента забит - он не успевает читать (плохая сеть, вкладка в фоне).
					// Ждать его нельзя: из-за одного медленного клиента встала бы рассылка всем.
					// Отключаем, фронт переподключится и догрузит ленту через HTTP
					h.log.Warn("ws client too slow, disconnecting", zap.Int("user_id", c.userID))
					h.removeClient(c)
				}
			}

		case <-ctx.Done():
			// http.Server.Shutdown НЕ ждёт и не закрывает hijacked-соединения (а WS - это они),
			// поэтому закрываем клиентов сами: writePump отправит close frame
			for c := range h.clients {
				h.removeClient(c)
			}

			h.log.Warn("ws hub stopped")

			return
		}
	}
}

func (h *Hub) Broadcast(msg []byte) {
	select {
	case h.broadcast <- msg:
	case <-h.done:
	}
}

func (h *Hub) registerClient(c *client) bool {
	select {
	case h.register <- c:
		return true
	case <-h.done:
		return false
	}
}

func (h *Hub) unregisterClient(c *client) {
	select {
	case h.unregister <- c:
	case <-h.done:
	}
}

// клиент может уйти двумя путями сразу (хаб отключил медленного + readPump увидел обрыв),
// проверка в мапе защищает от двойного close(c.send) - это паника
func (h *Hub) removeClient(c *client) {
	if _, ok := h.clients[c]; !ok {
		return
	}

	delete(h.clients, c)
	close(c.send)

	h.log.Debug("ws client disconnected", zap.Int("user_id", c.userID), zap.Int("clients", len(h.clients)))
}
