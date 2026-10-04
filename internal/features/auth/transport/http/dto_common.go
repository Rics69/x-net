package auth_transport_http

import (
	"time"

	"github.com/Rics69/x-net/internal/core/domain"
)

// свой DTO, а не импорт из users transport: фичи не зависят от транспорта друг друга
type AuthUserDTOResponse struct {
	ID        int       `json:"id" example:"1"`
	Username  string    `json:"username" example:"ivan_ivanov"`
	FullName  string    `json:"full_name" example:"Ivan Ivanov"`
	CreatedAt time.Time `json:"created_at" example:"2026-10-04T12:00:00Z"`
}

func authUserDTOFromDomain(user domain.User) AuthUserDTOResponse {
	return AuthUserDTOResponse{
		ID:        user.ID,
		Username:  user.Username,
		FullName:  user.FullName,
		CreatedAt: user.CreatedAt,
	}
}
