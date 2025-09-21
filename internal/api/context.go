package api

import "context"

type externalContextKey string

const externalUserIDKey externalContextKey = "externalUserID"

func withExternalUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, externalUserIDKey, userID)
}

func externalUserIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if v, ok := ctx.Value(externalUserIDKey).(string); ok {
		return v
	}
	return ""
}
