package ports

import (
	"context"
	"log"

	"gopkg.in/telebot.v3"

	"bot/internal/app"
)

func StartBot(ctx context.Context, bot *telebot.Bot, domain *app.App) {
	bot.Use(ContextMiddleware(ctx))
	bot.Use(AuthMiddleware(ctx, domain))

	dispatcher := NewDispatcher(ctx, domain, bot)

	routers(bot, domain, dispatcher)

	log.Println("Bot is now running. Press CTRL+C to exit")
	bot.Start()
}
