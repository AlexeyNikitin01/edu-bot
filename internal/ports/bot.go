package ports

import (
	"bot/internal/middleware"
	"context"
	"log"

	"gopkg.in/telebot.v3"

	"bot/internal/domain"
)

func StartBot(ctx context.Context, bot *telebot.Bot, d domain.UseCases) {
	bot.Use(middleware.ContextMiddleware(ctx))
	bot.Use(middleware.AuthMiddleware(ctx, d))

	routers(ctx, bot, d)

	log.Println("Bot is now running. Press CTRL+C to exit")

	go func() {
		bot.Start()
	}()
}
