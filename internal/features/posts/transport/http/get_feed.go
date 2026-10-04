package posts_transport_http

import (
	"net/http"

	core_logger "github.com/Rics69/x-net/internal/core/logger"
	core_http_request "github.com/Rics69/x-net/internal/core/transport/http/request"
	core_http_response "github.com/Rics69/x-net/internal/core/transport/http/response"
)

type GetFeedResponse struct {
	Posts []PostDTOResponse `json:"posts"`

	// null - это последняя страница
	NextCursor *string `json:"next_cursor" example:"eyJ0IjoiMjAyNi0xMC0wNFQxMjowMDowMFoiLCJpZCI6NDJ9"`
}

// GetFeed godoc
// @Summary Общая лента
// @Description Посты всех пользователей, от новых к старым, с курсорной пагинацией.
// @Description Первая страница - без `cursor`. Следующая - `cursor` = `next_cursor` из предыдущего ответа.
// @Description `next_cursor: null` - дальше постов нет
// @Tags posts
// @Produce json
// @Param limit query int false "Размер страницы (1-100, по умолчанию 20)"
// @Param cursor query string false "Курсор из next_cursor предыдущей страницы"
// @Success 200 {object} GetFeedResponse "Страница ленты"
// @Failure 400 {object} core_http_response.ErrorResponse "Bad request"
// @Failure 401 {object} core_http_response.ErrorResponse "Unauthorized"
// @Failure 500 {object} core_http_response.ErrorResponse "Internal server error"
// @Router /posts [get]
func (h *PostsHTTPHandler) GetFeed(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)

	limit, err := core_http_request.GetIntQueryParam(r, "limit")
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get 'limit' query param")

		return
	}

	cursor, err := decodeCursor(r.URL.Query().Get("cursor"))
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to decode 'cursor' query param")

		return
	}

	page, err := h.postsService.GetFeed(ctx, cursor, limit)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get feed")

		return
	}

	response := GetFeedResponse{
		Posts: postDTOsFromDomains(page.Posts),
	}

	if page.NextCursor != nil {
		nextCursor := encodeCursor(*page.NextCursor)
		response.NextCursor = &nextCursor
	}

	responseHandler.JSONResponse(response, http.StatusOK)
}
