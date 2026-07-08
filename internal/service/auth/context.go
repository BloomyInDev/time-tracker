package auth

import "context"

type contextKey int

const userIDKey contextKey = iota

func UserIDFromContext(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(userIDKey).(int64)
	return id, ok
}
