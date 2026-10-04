package posts_transport_http

import (
	"net/http"

	core_auth "github.com/Rics69/x-net/internal/core/auth"
	"github.com/Rics69/x-net/internal/core/domain"
	core_logger "github.com/Rics69/x-net/internal/core/logger"
	core_http_request "github.com/Rics69/x-net/internal/core/transport/http/request"
	core_http_response "github.com/Rics69/x-net/internal/core/transport/http/response"
)

type CreatePostRequest struct {
	Content string `json:"content" validate:"required,max=280" example:"Hello, X-Net!"`
}

type CreatePostResponse PostDTOResponse

// CreatePost godoc
// @Summary Создать пост
// @Description Публикация поста от имени авторизованного пользователя
// @Tags posts
// @Accept json
// @Produce json
// @Param request body CreatePostRequest true "CreatePost тело запроса"
// @Success 201 {object} CreatePostResponse "Созданный пост"
// @Failure 400 {object} core_http_response.ErrorResponse "Bad request"
// @Failure 401 {object} core_http_response.ErrorResponse "Unauthorized"
// @Failure 500 {object} core_http_response.ErrorResponse "Internal server error"
// @Router /posts [post]
func (h *PostsHTTPHandler) CreatePost(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)

	userID, err := core_auth.UserIDFromContext(ctx)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get userID from context")

		return
	}

	var request CreatePostRequest
	if err := core_http_request.DecodeAndValidateRequest(r, &request); err != nil {
		responseHandler.ErrorResponse(err, "failed to decode and validate HTTP request")

		return
	}

	// автор всегда из токена, не из тела запроса - иначе можно постить от чужого имени
	postDomain := domain.NewPostUnitialized(request.Content, userID)

	postDomain, err = h.postsService.CreatePost(ctx, postDomain)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to create post")

		return
	}

	response := CreatePostResponse(postDTOFromDomain(postDomain))

	responseHandler.JSONResponse(response, http.StatusCreated)
}
