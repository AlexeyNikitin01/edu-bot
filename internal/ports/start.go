package ports

import (
	"fmt"

	"github.com/volatiletech/sqlboiler/v4/boil"
	"gopkg.in/telebot.v3"

	"bot/internal/app"
	"bot/internal/repo/edu"
)

func start(_ *app.App) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		tgUser := ctx.Sender()
		chatUser := ctx.Chat()

		if err := ctx.Send(fmt.Sprintf("Привет %s!", tgUser.FirstName)); err != nil {
			return err
		}

		u := &edu.User{
			TGUserID:  tgUser.ID,
			ChatID:    chatUser.ID,
			FirstName: tgUser.FirstName,
		}

		if err := u.Upsert(
			GetContext(ctx),
			boil.GetContextDB(),
			true,
			[]string{edu.UserColumns.TGUserID},
			boil.Infer(),
			boil.Infer(),
		); err != nil {
			return err
		}

		return nil
	}
}
