package web

import (
	"embed"
	"io/fs"
	"mime"
	"net/http"
)

// go:embed вшивает файлы в бинарник на этапе компиляции: один exe без папки со статикой рядом,
// в Docker ничего докопировать не надо. Минус - после правки html/js нужен перезапуск go run
//
//go:embed static
var staticFiles embed.FS

// CSP - вторая линия обороны от XSS: даже если где-то вставим чужой текст как HTML,
// браузер не выполнит inline-скрипты и не пойдёт на чужие домены.
// connect-src 'self' покрывает и fetch, и ws:// на тот же хост
const contentSecurityPolicy = "default-src 'self'; " +
	"script-src 'self'; " +
	"style-src 'self'; " +
	"font-src 'self'; " +
	"img-src 'self' data:; " +
	"connect-src 'self'; " +
	"base-uri 'none'; " +
	"form-action 'self'; " +
	"frame-ancestors 'none'"

func init() {
	// во встроенной таблице mime у Go нет .woff2 - без этого шрифт уйдёт как octet-stream
	_ = mime.AddExtensionType(".woff2", "font/woff2")
}

func Handler() http.Handler {
	// embed хранит пути вместе с папкой (static/index.html), а раздавать нужно от корня (/index.html)
	sub, err := fs.Sub(staticFiles, "static")
	if err != nil {
		panic("web: static dir not embedded: " + err.Error())
	}

	fileServer := http.FileServerFS(sub)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", contentSecurityPolicy)
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// у embed-файлов нет даты изменения, браузер не может проверить свежесть сам.
		// no-cache = "можно хранить, но перед использованием спроси сервер"
		w.Header().Set("Cache-Control", "no-cache")

		fileServer.ServeHTTP(w, r)
	})
}
