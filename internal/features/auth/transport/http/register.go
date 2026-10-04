package auth_transport_http

import (
	"net/http"

	"github.com/Rics69/x-net/internal/core/domain"
	core_logger "github.com/Rics69/x-net/internal/core/logger"
	core_http_request "github.com/Rics69/x-net/internal/core/transport/http/request"
	core_http_response "github.com/Rics69/x-net/internal/core/transport/http/response"
)

// max у validator считает символы, а у bcrypt лимит в байтах - точная проверка в сервисе
type RegisterRequest struct {
	Username    string  `json:"username" validate:"required,min=3,max=32" example:"ivan_ivanov"`
	FullName    string  `json:"full_name" validate:"required,min=3,max=100" example:"Ivan Ivanov"`
	PhoneNumber *string `json:"phone_number" validate:"omitempty,min=10,max=15,startswith=+" example:"+79999998877"`
	Password    string  `json:"password" validate:"required,min=8,max=72" example:"supersecret"`
}

type RegisterResponse AuthUserDTOResponse

// Register godoc
// @Summary Регистрация
// @Description Создание нового пользователя. При успехе сразу выставляется httpOnly cookie `access_token`
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Register тело запроса"
// @Success 201 {object} RegisterResponse "Успешно зарегистрированный пользователь"
// @Failure 400 {object} core_http_response.ErrorResponse "Bad request"
// @Failure 409 {object} core_http_response.ErrorResponse "Username already taken"
// @Failure 500 {object} core_http_response.ErrorResponse "Internal server error"
// @Router /auth/register [post]
func (h *AuthHTTPHandler) Register(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)

	var request RegisterRequest
	if err := core_http_request.DecodeAndValidateRequest(r, &request); err != nil {
		responseHandler.ErrorResponse(err, "failed to decode and validate HTTP request")

		return
	}

	userDomain := domain.NewUserUnitialized(request.Username, request.FullName, request.PhoneNumber)

	userDomain, token, err := h.authService.Register(ctx, userDomain, request.Password)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to register user")

		return
	}

	// кука ставится до WriteHeader внутри JSONResponse: после него заголовки уже не изменить
	h.setAccessTokenCookie(rw, token)

	response := RegisterResponse(authUserDTOFromDomain(userDomain))

	responseHandler.JSONResponse(response, http.StatusCreated)
}
