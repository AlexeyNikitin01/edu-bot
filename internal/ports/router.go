package ports

import (
	"gopkg.in/telebot.v3"

	"bot/internal/app"
)

func routers(b *telebot.Bot, domain *app.App) {
	b.Handle("/start", start())
	b.Handle("/add", add())
	b.Handle(telebot.OnText, add())
	b.Handle("/edu", showRepeatList())
	b.Handle(&telebot.InlineButton{Unique: "toggle_repeat"}, handleToggleRepeat())
	b.Handle("/delete", deleteList())
	b.Handle(&telebot.InlineButton{Unique: "delete_repeat"}, deleteRepeat())

	b.Handle(telebot.OnText, func(ctx telebot.Context) error {
		switch ctx.Text() {
		case "➕ Добавить вопрос":
			return add()(ctx)
		case "📚 Повторение":
			return showRepeatList()(ctx)
		case "🗑 Удалить вопрос":
			return deleteList()(ctx)
		default:
			return ctx.Send("⚠️ Неизвестная команда. Используйте меню ниже.", mainMenu())
		}
	})
}
