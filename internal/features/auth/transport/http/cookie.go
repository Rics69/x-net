package auth_transport_http

import (
	"net/http"

	core_auth "github.com/Rics69/x-net/internal/core/auth"
)

// HttpOnly - JS не может прочитать куку, токен не утащить через XSS.
// SameSite=Lax - кука не уходит в POST с чужих сайтов (базовая защита от CSRF),
// но уходит при обычном переходе по ссылке на наш сайт.
// Path=/ - чтобы кука ехала и на /api/v1/..., и на /ws
func (h *AuthHTTPHandler) setAccessTokenCookie(rw http.ResponseWriter, token core_auth.AccessToken) {
	http.SetCookie(rw, &http.Cookie{
		Name:     core_auth.AccessTokenCookieName,
		Value:    token.Value,
		Path:     "/",
		Expires:  token.ExpiresAt,
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}

// куку нельзя "удалить" с сервера - только перезаписать на просроченную.
// Сам JWT при этом остаётся валидным до exp: stateless-токен не отозвать
// без blacklist/сессий в БД. Для учебного проекта ок, TTL держим небольшим
func (h *AuthHTTPHandler) clearAccessTokenCookie(rw http.ResponseWriter) {
	http.SetCookie(rw, &http.Cookie{
		Name:     core_auth.AccessTokenCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}
