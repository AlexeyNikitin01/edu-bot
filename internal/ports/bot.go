package ports

import (
	"context"
	"log"

	"gopkg.in/telebot.v3"

	"bot/internal/app"
)

func StartBot(ctx context.Context, bot *telebot.Bot, domain app.Apper, cache app.UserCacher) {
	bot.Use(ContextMiddleware(ctx))
	bot.Use(AuthMiddleware(ctx, domain))

	dispatcher := NewDispatcher(ctx, domain, bot, cache)

	routers(bot, domain, dispatcher)

	log.Println("Bot is now running. Press CTRL+C to exit")
	bot.Start()
}
