package ports

import (
	"context"
	"log"

	"gopkg.in/telebot.v3"

	"bot/internal/app"
)

func StartBot(ctx context.Context, bot *telebot.Bot, domain *app.App) {
	bot.Use(ContextMiddleware(ctx))

	routers(bot, domain)

	dispatcher := NewDispatcher(ctx, domain, bot)
	dispatcher.RegisterPollAnswerHandler()
	dispatcher.StartPollingLoop()

	log.Println("Bot is now running.  Press CTRL-C to exit.")
	bot.Start()
}
