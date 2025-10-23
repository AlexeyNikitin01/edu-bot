package ports

import (
	"context"
	"log"

	"gopkg.in/telebot.v3"

	"bot/internal/app"
)

func StartBot(ctx context.Context, bot *telebot.Bot, domain app.Apper, cache app.UserCacher, d *QuestionDispatcher) {
	bot.Use(ContextMiddleware(ctx))
	bot.Use(AuthMiddleware(ctx, domain))

	routers(bot, domain, d, cache)

	log.Println("Bot is now running. Press CTRL+C to exit")

	go func() {
		bot.Start()
	}()
}
