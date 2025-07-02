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
	btnMark := menu.Text(BTN_MANAGMENT_QUESTION)
	btnCSV := menu.Text(BTN_ADD_CSV)

	menu.Reply(
		menu.Row(btnAdd, btnCSV),
		menu.Row(btnMark),
	)

	return menu
}
