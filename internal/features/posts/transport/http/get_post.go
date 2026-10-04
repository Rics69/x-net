package posts_transport_http

import (
	"net/http"

	core_logger "github.com/Rics69/x-net/internal/core/logger"
	core_http_request "github.com/Rics69/x-net/internal/core/transport/http/request"
	core_http_response "github.com/Rics69/x-net/internal/core/transport/http/response"
)

type GetPostResponse PostDTOResponse

// GetPost godoc
// @Summary Получение поста
// @Description Получение конкретного поста по его ID
// @Tags posts
// @Param id path int true "ID поста"
// @Produce json
// @Success 200 {object} GetPostResponse "Пост найден"
// @Failure 400 {object} core_http_response.ErrorResponse "Bad request"
// @Failure 401 {object} core_http_response.ErrorResponse "Unauthorized"
// @Failure 404 {object} core_http_response.ErrorResponse "Post not found"
// @Failure 500 {object} core_http_response.ErrorResponse "Internal server error"
// @Router /posts/{id} [get]
func (h *PostsHTTPHandler) GetPost(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)

	postID, err := core_http_request.GetIntPathValue(r, "id")
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get postID path value")

		return
	}

	post, err := h.postsService.GetPost(ctx, postID)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get post")

		return
	}

	response := GetPostResponse(postDTOFromDomain(post))

	responseHandler.JSONResponse(response, http.StatusOK)
}
