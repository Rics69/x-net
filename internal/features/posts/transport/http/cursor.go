package posts_transport_http

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Rics69/x-net/internal/core/domain"
	core_errors "github.com/Rics69/x-net/internal/core/errors"
)

// курсор для клиента - непрозрачная строка: он её не парсит и не собирает сам,
// только возвращает то, что получил в next_cursor.
// Поэтому формат внутри можно менять, не ломая фронт
type cursorPayload struct {
	CreatedAt time.Time `json:"t"`
	ID        int       `json:"id"`
}

// RawURLEncoding - без '+', '/' и '=', чтобы курсор можно было
// положить в query string без экранирования
func encodeCursor(cursor domain.PostsCursor) string {
	payload, _ := json.Marshal(cursorPayload{
		CreatedAt: cursor.CreatedAt,
		ID:        cursor.ID,
	})

	return base64.RawURLEncoding.EncodeToString(payload)
}

func decodeCursor(raw string) (*domain.PostsCursor, error) {
	if raw == "" {
		return nil, nil
	}

	payload, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("decode cursor base64: %v: %w", err, core_errors.ErrInvalidArgument)
	}

	var p cursorPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, fmt.Errorf("decode cursor json: %v: %w", err, core_errors.ErrInvalidArgument)
	}

	if p.ID <= 0 || p.CreatedAt.IsZero() {
		return nil, fmt.Errorf("cursor has empty fields: %w", core_errors.ErrInvalidArgument)
	}

	cursor := domain.NewPostsCursor(p.CreatedAt, p.ID)

	return &cursor, nil
}
