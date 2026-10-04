package posts_service

import (
	"context"
	"fmt"

	"github.com/Rics69/x-net/internal/core/domain"
)

func (s *PostsService) GetPost(ctx context.Context, id int) (domain.Post, error) {
	post, err := s.postsRepository.GetPost(ctx, id)
	if err != nil {
		return domain.Post{}, fmt.Errorf("get post from repository: %w", err)
	}

	return post, nil
}
