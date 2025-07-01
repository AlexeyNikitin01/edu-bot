package ports

import (
	"gopkg.in/telebot.v3"
)

const MSG_GRETING = "Добро пожаловать! Выберите действие:"

func start() telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
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
