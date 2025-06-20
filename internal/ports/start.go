package ports

import (
	"fmt"

	"github.com/volatiletech/sqlboiler/v4/boil"
	"gopkg.in/telebot.v3"

	"bot/internal/repo/edu"
)

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

		return ctx.Send("Добро пожаловать! Выберите действие:", mainMenu())
	}
}

func mainMenu() *telebot.ReplyMarkup {
	menu := &telebot.ReplyMarkup{ResizeKeyboard: true}

	btnAdd := menu.Text("➕ Добавить вопрос")
	btnMark := menu.Text("📚 Повторение")
	btnDelete := menu.Text("🗑 Удалить вопрос")
	btnPause := menu.Text("⏸️ Пауза")
	btnResume := menu.Text("▶️ Старт")

	menu.Reply(
		menu.Row(btnAdd),
		menu.Row(btnMark, btnDelete),
		menu.Row(btnPause, btnResume),
	)

	return menu
}
