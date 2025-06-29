package ports

import (
	"fmt"

	"github.com/volatiletech/sqlboiler/v4/boil"
	"gopkg.in/telebot.v3"

	"bot/internal/repo/edu"
)

const MSG_GRETING = "Добро пожаловать! Выберите действие:"

func start() telebot.HandlerFunc {
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

		return ctx.Send(MSG_GRETING, mainMenu())
	}
}

func mainMenu() *telebot.ReplyMarkup {
	menu := &telebot.ReplyMarkup{ResizeKeyboard: true}

	btnAdd := menu.Text(BTN_ADD_QUESTION)
	btnMark := menu.Text(BTN_REPEAT)
	btnDelete := menu.Text(BTN_DEL_QUESTION)
	btnPause := menu.Text(BTN_PAUSE)
	btnResume := menu.Text(BTN_RESUME)
	btnCSV := menu.Text(BTN_ADD_CSV)

	menu.Reply(
		menu.Row(btnAdd, btnCSV),
		menu.Row(btnMark, btnDelete),
		menu.Row(btnPause, btnResume),
	)

	return menu
}
