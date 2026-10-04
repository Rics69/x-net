package auth_transport_http

import (
	"net/http"

	core_logger "github.com/Rics69/x-net/internal/core/logger"
	core_http_response "github.com/Rics69/x-net/internal/core/transport/http/response"
)

// Logout godoc
// @Summary Выход
// @Description Сбрасывает cookie `access_token`. Авторизация не требуется: выйти можно и с протухшим токеном
// @Tags auth
// @Success 204 "Успешный выход"
// @Router /auth/logout [post]
func (h *AuthHTTPHandler) Logout(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)

	h.clearAccessTokenCookie(rw)

	responseHandler.NoContentResponse()
}
