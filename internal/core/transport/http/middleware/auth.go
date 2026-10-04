package core_http_middleware

import (
	"fmt"
	"net/http"

	core_auth "github.com/Rics69/x-net/internal/core/auth"
	core_errors "github.com/Rics69/x-net/internal/core/errors"
	core_logger "github.com/Rics69/x-net/internal/core/logger"
	core_http_response "github.com/Rics69/x-net/internal/core/transport/http/response"
	"go.uber.org/zap"
)

type TokenParser interface {
	Parse(token string) (int, error)
}

// вешается на конкретные роуты через Route.Middleware, а не на весь роутер,
// т.к. /auth/login и /auth/register должны быть доступны без токена.
// Токен берём из httpOnly-куки: JS её не видит (защита от XSS),
// и браузер сам шлёт её при WebSocket-handshake, куда заголовок Authorization не передать
func Auth(tokens TokenParser) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			log := core_logger.FromContext(ctx)
			responseHandler := core_http_response.NewHTTPResponseHandler(log, w)

			cookie, err := r.Cookie(core_auth.AccessTokenCookieName)
			if err != nil {
				responseHandler.ErrorResponse(
					fmt.Errorf("no '%s' cookie: %w", core_auth.AccessTokenCookieName, core_errors.ErrUnauthorized),
					"unauthorized",
				)

				return
			}

			userID, err := tokens.Parse(cookie.Value)
			if err != nil {
				responseHandler.ErrorResponse(err, "invalid access token")

				return
			}

			ctx = core_auth.UserIDToContext(ctx, userID)
			ctx = core_logger.ToContext(ctx, log.With(zap.Int("user_id", userID)))

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
