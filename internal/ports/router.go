package ports

import (
	"gopkg.in/telebot.v3"

	"bot/internal/app"
)

func routers(b *telebot.Bot, domain *app.App) {
	b.Handle("/start", start())
	b.Handle("/add", add())
	b.Handle(telebot.OnText, add())
}
