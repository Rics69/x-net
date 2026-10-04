package auth_transport_http

import (
	"net/http"

	core_logger "github.com/Rics69/x-net/internal/core/logger"
	core_http_request "github.com/Rics69/x-net/internal/core/transport/http/request"
	core_http_response "github.com/Rics69/x-net/internal/core/transport/http/response"
)

type LoginRequest struct {
	Username string `json:"username" validate:"required" example:"ivan_ivanov"`
	Password string `json:"password" validate:"required" example:"supersecret"`
}

type LoginResponse AuthUserDTOResponse

// Login godoc
// @Summary Вход
// @Description Проверка username/пароля. При успехе выставляется httpOnly cookie `access_token`,
// @Description после чего защищённые ручки (в т.ч. из этого Swagger UI) работают автоматически
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login тело запроса"
// @Success 200 {object} LoginResponse "Успешный вход"
// @Failure 400 {object} core_http_response.ErrorResponse "Bad request"
// @Failure 401 {object} core_http_response.ErrorResponse "Invalid username or password"
// @Failure 500 {object} core_http_response.ErrorResponse "Internal server error"
// @Router /auth/login [post]
func (h *AuthHTTPHandler) Login(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)

	var request LoginRequest
	if err := core_http_request.DecodeAndValidateRequest(r, &request); err != nil {
		responseHandler.ErrorResponse(err, "failed to decode and validate HTTP request")

		return
	}

	userDomain, token, err := h.authService.Login(ctx, request.Username, request.Password)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to login")

		return
	}

	h.setAccessTokenCookie(rw, token)

	response := LoginResponse(authUserDTOFromDomain(userDomain))

	responseHandler.JSONResponse(response, http.StatusOK)
}
