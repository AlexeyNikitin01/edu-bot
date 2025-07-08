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
			// ищем пользователя
			sender := c.Sender()

			user, err := domain.GetUser(ctx, sender.ID)
			if err != nil {
				return c.Reply("Произошла ошибка при авторизации: невозможно получить пользователя")
			}

			if user != nil {
				c.Set("user", user)
				return next(c)
			}

			// Пытаемся создать пользователя
			chatUser := c.Chat()

			if chatUser == nil {
				return c.Reply("Произошла ошибка при авторизации: пустой чат")
			}

			user, err = domain.CreateUser(ctx, sender.ID, chatUser.ID, sender.FirstName)
			if err != nil {
				return c.Reply("Произошла ошибка при авторизации: невозможно создать или обновить пользователя")
			}

			if user == nil {
				return c.Reply("Произошла ошибка при авторизации")
			}

			c.Set("user", user)

			if err = c.Send(MSG_GRETING, mainMenu()); err != nil {
				return c.Reply("Произошла ошибка при отправке приветствия")
			}

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
