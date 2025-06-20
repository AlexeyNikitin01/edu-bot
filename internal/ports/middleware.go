package ports

import (
	"context"

	"gopkg.in/telebot.v3"
)

// ContextMiddleware добавляет стандартный context.Context к telebot.Context
func ContextMiddleware(baseCtx context.Context) telebot.MiddlewareFunc {
	return func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(c telebot.Context) error {
			c.Set("ctx", baseCtx)

			return next(c)
		}
	}
}

// GetContext извлекает context.Context из telebot.Context
func GetContext(c telebot.Context) context.Context {
	if ctx, ok := c.Get("ctx").(context.Context); ok {
		return ctx
	}
	return context.Background()
}
