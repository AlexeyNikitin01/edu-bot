package ports

import (
	"log"

	"gopkg.in/telebot.v3"

	"bot/internal/app"
)

func StartBot(bot *telebot.Bot, domain *app.App) {
	routers(bot, domain)

	log.Println("Bot is now running.  Press CTRL-C to exit.")
	bot.Start()
}
