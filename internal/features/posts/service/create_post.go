package posts_service

import (
	"context"
	"fmt"

	"github.com/Rics69/x-net/internal/core/domain"
)

func (s *PostsService) CreatePost(ctx context.Context, post domain.Post) (domain.Post, error) {
	if err := post.Validate(); err != nil {
		return domain.Post{}, fmt.Errorf("validate post domain: %w", err)
	}

	post, err := s.postsRepository.CreatePost(ctx, post)
	if err != nil {
		return domain.Post{}, fmt.Errorf("create post: %w", err)
	}

	return post, nil
}
