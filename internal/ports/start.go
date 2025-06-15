package ports

import (
	"log"

	"gopkg.in/telebot.v3"

	"bot/internal/app"
)

func start(_ *app.App) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		log.Println("start")
		if err := ctx.Send("hello world"); err != nil {
			return err
		}
		return nil
	}
}
