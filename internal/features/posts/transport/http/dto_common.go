package posts_transport_http

import (
	"time"

	"github.com/Rics69/x-net/internal/core/domain"
)

type PostAuthorDTOResponse struct {
	ID       int    `json:"id" example:"1"`
	Username string `json:"username" example:"ivan_ivanov"`
	FullName string `json:"full_name" example:"Ivan Ivanov"`
}

type PostDTOResponse struct {
	ID        int                   `json:"id" example:"42"`
	Content   string                `json:"content" example:"Hello, X-Net!"`
	CreatedAt time.Time             `json:"created_at" example:"2026-10-04T12:00:00Z"`
	Author    PostAuthorDTOResponse `json:"author"`
}

func postDTOFromDomain(post domain.Post) PostDTOResponse {
	dto := PostDTOResponse{
		ID:        post.ID,
		Content:   post.Content,
		CreatedAt: post.CreatedAt,
		Author: PostAuthorDTOResponse{
			ID: post.AuthorUserID,
		},
	}

	if post.Author != nil {
		dto.Author.Username = post.Author.Username
		dto.Author.FullName = post.Author.FullName
	}

	return dto
}

func postDTOsFromDomains(posts []domain.Post) []PostDTOResponse {
	dtos := make([]PostDTOResponse, len(posts))

	for i, post := range posts {
		dtos[i] = postDTOFromDomain(post)
	}

	return dtos
}
