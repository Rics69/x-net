package core_auth

import (
	"context"
	"fmt"

	core_errors "github.com/Rics69/x-net/internal/core/errors"
)

type userIDContextKey struct{}

var (
	userIDKey = userIDContextKey{}
)

func UserIDToContext(ctx context.Context, userID int) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// не паникуем как logger.FromContext: если хендлер случайно повесили без Auth middleware,
// лучше честный 401, чем 500 от паники
func UserIDFromContext(ctx context.Context) (int, error) {
	userID, ok := ctx.Value(userIDKey).(int)
	if !ok {
		return 0, fmt.Errorf("no user id in context: %w", core_errors.ErrUnauthorized)
	}

	return userID, nil
}
