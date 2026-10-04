package posts_transport_http

import (
	"net/http"

	core_auth "github.com/Rics69/x-net/internal/core/auth"
	core_logger "github.com/Rics69/x-net/internal/core/logger"
	core_http_request "github.com/Rics69/x-net/internal/core/transport/http/request"
	core_http_response "github.com/Rics69/x-net/internal/core/transport/http/response"
)

// DeletePost godoc
// @Summary Удаление поста
// @Description Удалить можно только свой пост
// @Tags posts
// @Param id path int true "ID удаляемого поста"
// @Success 204 "Пост удалён"
// @Failure 400 {object} core_http_response.ErrorResponse "Bad request"
// @Failure 401 {object} core_http_response.ErrorResponse "Unauthorized"
// @Failure 403 {object} core_http_response.ErrorResponse "Post belongs to another user"
// @Failure 404 {object} core_http_response.ErrorResponse "Post not found"
// @Failure 500 {object} core_http_response.ErrorResponse "Internal server error"
// @Router /posts/{id} [delete]
func (h *PostsHTTPHandler) DeletePost(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)

	userID, err := core_auth.UserIDFromContext(ctx)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get userID from context")

		return
	}

	postID, err := core_http_request.GetIntPathValue(r, "id")
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get postID path value")

		return
	}

	if err := h.postsService.DeletePost(ctx, postID, userID); err != nil {
		responseHandler.ErrorResponse(err, "failed to delete post")

		return
	}

	responseHandler.NoContentResponse()
}
