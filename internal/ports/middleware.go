package ports

import (
	"context"

	"gopkg.in/telebot.v3"

	"bot/internal/app"
	"bot/internal/repo/edu"
)

// AuthMiddleware создает middleware для авторизации пользователей
func AuthMiddleware(ctx context.Context, domain app.Apper) telebot.MiddlewareFunc {
	return func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(c telebot.Context) error {
			if c.Message() != nil && c.Message().Text == "/start" {
				return next(c)
			}

			sender := c.Sender()
			chat := c.Chat()

			user, err := domain.GetOrCreate(ctx, sender.ID, chat.ID, sender.FirstName)
			if err != nil {
				return c.Reply("Произошла ошибка при авторизации")
			}

			c.Set("user", user)

			return next(c)
		}
	}
}

// GetUserFromContext извлекает пользователя из контекста
func GetUserFromContext(c telebot.Context) *edu.User {
	user, ok := c.Get("user").(*edu.User)
	if !ok {
		return nil
	}
	return user
}

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
