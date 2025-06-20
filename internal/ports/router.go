package ports

import (
	"gopkg.in/telebot.v3"

	"bot/internal/app"
)

func routers(b *telebot.Bot, domain *app.App) {
	b.Handle("/start", start())

	b.Handle(&telebot.InlineButton{Unique: "toggle_repeat"}, handleToggleRepeat())
	b.Handle(&telebot.InlineButton{Unique: "delete_repeat"}, deleteRepeat())

	b.Handle(telebot.OnText, func(ctx telebot.Context) error {
		userID := ctx.Sender().ID

		// Если пользователь в процессе добавления вопроса
		if draft, ok := drafts[userID]; ok && draft.Step > 0 {
			return add()(ctx)
		}

		// TODO: нужно смотреть если пауза у пользователя, чтобы ничего не ломать
		switch ctx.Text() {
		case "/add", "➕ Добавить вопрос":
			drafts[userID] = &QuestionDraft{Step: 1}
			return ctx.Send("✍️ Введите текст вопроса:")
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
}
