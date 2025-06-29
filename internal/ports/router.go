package ports

import (
	"context"

	"gopkg.in/telebot.v3"

	"bot/internal/app"
)

const (
	TAGS = "getTags"

	ADD_QUESTION = "➕ Добавить вопрос"
)

func routers(ctx context.Context, b *telebot.Bot, domain *app.App) {
	b.Handle("/start", start())

	b.Handle(&telebot.InlineButton{Unique: "toggle_repeat"}, handleToggleRepeat())
	b.Handle(&telebot.InlineButton{Unique: "delete_repeat"}, deleteRepeat())
	b.Handle(&telebot.InlineButton{Unique: TAGS}, func(c telebot.Context) error {
		return add(domain)(c)
	})

	b.Handle(telebot.OnText, func(ctx telebot.Context) error {
		userID := ctx.Sender().ID

		// Если пользователь в процессе добавления вопроса
		if draft, ok := drafts[userID]; ok && draft.Step > 0 {
			return add(domain)(ctx)
		}

		// TODO: нужно смотреть если пауза у пользователя, чтобы ничего не ломать
		switch ctx.Text() {
		case ADD_QUESTION:
			if err := getTags(ctx, GetUserFromContext(ctx).TGUserID, domain); err != nil {
				return err
			}
			drafts[userID] = &QuestionDraft{Step: 1}
			return add(domain)(ctx)
		case "📚 Повторение":
			return showRepeatList()(ctx)
		case "🗑 Удалить вопрос":
			return deleteList()(ctx)
		case "⏸️ Пауза":
			return pause()(ctx)
		case "▶️ Старт":
			return resume()(ctx)
		default:
			return ctx.Send("⚠️ Неизвестная команда. Используйте меню ниже.", mainMenu())
		}
	})

	dispatcher := NewDispatcher(ctx, domain, b)
	dispatcher.RegisterPollAnswerHandler()
	dispatcher.StartPollingLoop()
}
