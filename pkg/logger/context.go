package logger

import (
	"context"
)

type contextKey struct{}

var ctxLoggerKey = contextKey{}

type logCtx struct {
	sessionID string
	name      string
}

func WithSessionID(ctx context.Context, sessionID string) context.Context {
	if c, ok := ctx.Value(ctxLoggerKey).(logCtx); ok {
		c.sessionID = sessionID
		return context.WithValue(ctx, ctxLoggerKey, c)
	}

	return context.WithValue(ctx, ctxLoggerKey, logCtx{
		sessionID: sessionID,
	})
}

func WithName(ctx context.Context, name string) context.Context {
	if c, ok := ctx.Value(ctxLoggerKey).(logCtx); ok {
		c.name = name
		return context.WithValue(ctx, ctxLoggerKey, c)
	}
	return context.WithValue(ctx, ctxLoggerKey, logCtx{
		name: name,
	})
}
