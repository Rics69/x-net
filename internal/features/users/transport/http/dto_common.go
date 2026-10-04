package users_transport_http

import (
	"time"

	"github.com/Rics69/x-net/internal/core/domain"
)

type UserDTOResponse struct {
	ID          int       `json:"id" example:"1"`
	Version     int       `json:"version" example:"3"`
	Username    string    `json:"username" example:"ivan_ivanov"`
	FullName    string    `json:"full_name" example:"Ivan Ivanov"`
	PhoneNumber *string   `json:"phone_number" example:"+79999998877"`
	CreatedAt   time.Time `json:"created_at" example:"2026-10-04T12:00:00Z"`
}

func userDTOFromDomain(user domain.User) UserDTOResponse {
	return UserDTOResponse{
		ID:          user.ID,
		Version:     user.Version,
		Username:    user.Username,
		FullName:    user.FullName,
		PhoneNumber: user.PhoneNumber,
		CreatedAt:   user.CreatedAt,
	}
}

func usersDTOFromDomain(users []domain.User) []UserDTOResponse {
	usersDTO := make([]UserDTOResponse, len(users))

	for i, user := range users {
		usersDTO[i] = userDTOFromDomain(user)
	}

	return usersDTO
}
