package core_ws

import (
	"net/http"

	core_auth "github.com/Rics69/x-net/internal/core/auth"
	core_logger "github.com/Rics69/x-net/internal/core/logger"
	core_http_middleware "github.com/Rics69/x-net/internal/core/transport/http/middleware"
	core_http_response "github.com/Rics69/x-net/internal/core/transport/http/response"
	core_http_server "github.com/Rics69/x-net/internal/core/transport/http/server"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type WSHTTPHandler struct {
	hub            *Hub
	upgrader       websocket.Upgrader
	authMiddleware core_http_middleware.Middleware
}

func NewWSHTTPHandler(hub *Hub, authMiddleware core_http_middleware.Middleware) *WSHTTPHandler {
	return &WSHTTPHandler{
		hub: hub,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,

			// CheckOrigin не переопределяем: по умолчанию gorilla пускает только
			// Origin == Host. Для WS это важно - браузер НЕ применяет к нему CORS,
			// и без проверки чужой сайт мог бы открыть сокет с куками нашего юзера
			// (Cross-Site WebSocket Hijacking)
		},
		authMiddleware: authMiddleware,
	}
}

func (h *WSHTTPHandler) Routes() []core_http_server.Route {
	return []core_http_server.Route{
		{
			Method:     http.MethodGet,
			Path:       "/ws",
			Handler:    h.ServeWS,
			Middleware: []core_http_middleware.Middleware{h.authMiddleware},
		},
	}
}

// авторизация обычной кукой через Auth middleware: при handshake браузер шлёт куки сам.
// Хендлер живёт, пока живёт соединение (readPump блокирует), поэтому в Trace
// latency = длительность WS-сессии, а логгер из ctx с request_id/user_id остаётся валидным
func (h *WSHTTPHandler) ServeWS(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)

	userID, err := core_auth.UserIDFromContext(ctx)
	if err != nil {
		core_http_response.NewHTTPResponseHandler(log, rw).ErrorResponse(err, "failed to get userID from context")

		return
	}

	// при ошибке upgrade gorilla сама отвечает клиенту (400/403), писать ответ нельзя
	conn, err := h.upgrader.Upgrade(rw, r, nil)
	if err != nil {
		log.Debug("ws upgrade failed", zap.Error(err))

		return
	}

	c := &client{
		hub:    h.hub,
		conn:   conn,
		send:   make(chan []byte, sendBufferSize),
		userID: userID,
	}

	if !h.hub.registerClient(c) {
		_ = conn.Close()

		return
	}

	go c.writePump()
	c.readPump(log)
}
