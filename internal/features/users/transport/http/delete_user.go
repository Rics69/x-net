package users_transport_http

import (
	"net/http"

	core_auth "github.com/Rics69/x-net/internal/core/auth"
	core_logger "github.com/Rics69/x-net/internal/core/logger"
	core_http_response "github.com/Rics69/x-net/internal/core/transport/http/response"
)

// DeleteUser godoc
// @Summary Удаление своего аккаунта
// @Description Удаление авторизованного пользователя. Кука при этом не чистится - после удаления фронт должен вызвать /auth/logout
// @Tags users
// @Success 204 "Успешное удаление пользователя"
// @Failure 401 {object} core_http_response.ErrorResponse "Unauthorized"
// @Failure 404 {object} core_http_response.ErrorResponse "User not found"
// @Failure 500 {object} core_http_response.ErrorResponse "Internal server error"
// @Router /users/me [delete]
func (h *UsersHTTPHandler) DeleteUser(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)

	userID, err := core_auth.UserIDFromContext(ctx)
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to get userID from context",
		)

		return
	}

	if err := h.usersService.DeleteUser(ctx, userID); err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to delete user",
		)

		return
	}

	responseHandler.NoContentResponse()
}
