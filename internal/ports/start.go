package ports

import (
	"fmt"

	"gopkg.in/telebot.v3"

	"bot/internal/app"
)

func start(_ *app.App) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {

		if err := ctx.Send(fmt.Printf("Привет %s!", ctx.Sender().FirstName)); err != nil {
			return err
		}

		return nil
	}
}
