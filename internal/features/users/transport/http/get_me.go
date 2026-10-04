package users_transport_http

import (
	"net/http"

	core_auth "github.com/Rics69/x-net/internal/core/auth"
	core_logger "github.com/Rics69/x-net/internal/core/logger"
	core_http_response "github.com/Rics69/x-net/internal/core/transport/http/response"
)

type GetMeResponse UserDTOResponse

// GetMe godoc
// @Summary Текущий пользователь
// @Description Получение профиля авторизованного пользователя (по access_token из cookie)
// @Tags users
// @Produce json
// @Success 200 {object} GetMeResponse "Текущий пользователь"
// @Failure 401 {object} core_http_response.ErrorResponse "Unauthorized"
// @Failure 404 {object} core_http_response.ErrorResponse "User not found"
// @Failure 500 {object} core_http_response.ErrorResponse "Internal server error"
// @Router /users/me [get]
func (h *UsersHTTPHandler) GetMe(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)

	userID, err := core_auth.UserIDFromContext(ctx)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get userID from context")

		return
	}

	// токен ещё может быть жив, а юзер уже удалён - тогда тут честный 404
	user, err := h.usersService.GetUser(ctx, userID)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get user")

		return
	}

	response := GetMeResponse(userDTOFromDomain(user))

	responseHandler.JSONResponse(response, http.StatusOK)
}
