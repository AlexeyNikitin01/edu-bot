package ports

import (
	"gopkg.in/telebot.v3"

	"bot/internal/app"
)

func routers(b *telebot.Bot, domain *app.App) {
	b.Handle("/start", start(domain))
	b.Handle("/add", add(domain))
	b.Handle(telebot.OnText, add(domain))
}
