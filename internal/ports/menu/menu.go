package menu

import (
	"gopkg.in/telebot.v3"
)

func BtnsMenu() *telebot.ReplyMarkup {
	m := &telebot.ReplyMarkup{ResizeKeyboard: true}

	btnAdd := m.Text(BTN_ADD_QUESTION)
	btnMark := m.Text(BTN_MANAGMENT_QUESTION)
	btnCSV := m.Text(BTN_ADD_CSV)
	btnNext := m.Text(BTN_NEXT_QUESTION)
	btnNextTask := m.Text(BTN_NEXT_TASK)

	m.Reply(
		m.Row(btnAdd, btnCSV),
		m.Row(btnMark, btnNext),
		m.Row(btnNextTask),
	)

	return m
}
