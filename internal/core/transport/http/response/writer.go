package core_http_response

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
)

var (
	StatusCodeUnitialized = -1
)

type ResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter: w,
		statusCode:     StatusCodeUnitialized,
	}
}

func (rw *ResponseWriter) WriteHeader(statusCode int) {
	rw.ResponseWriter.WriteHeader(statusCode)
	rw.statusCode = statusCode
}

func (rw *ResponseWriter) GetStatusCode() int {
	if rw.statusCode == StatusCodeUnitialized {
		return http.StatusOK
	}
	return rw.statusCode
}

// встраивание http.ResponseWriter даёт только его 3 метода. Опциональные интерфейсы
// оригинального writer'а (Hijacker, Flusher) через обёртку теряются.
// gorilla/websocket делает w.(http.Hijacker) - без этого метода upgrade падал бы с 500,
// т.к. middleware Trace оборачивает writer для всех запросов
func (rw *ResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("underlying ResponseWriter does not implement http.Hijacker")
	}

	// после hijack WriteHeader не вызывается - gorilla сама пишет "101 Switching Protocols"
	// в сырое соединение. Проставляем, чтобы Trace залогировал реальный статус
	rw.statusCode = http.StatusSwitchingProtocols

	return hijacker.Hijack()
}

// для http.ResponseController (Go 1.20+): он ищет нужные методы через цепочку Unwrap
func (rw *ResponseWriter) Unwrap() http.ResponseWriter {
	return rw.ResponseWriter
}
