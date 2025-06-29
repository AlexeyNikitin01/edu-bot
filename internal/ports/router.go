package ports

import (
	"context"

	"gopkg.in/telebot.v3"

	"bot/internal/app"
)

const (
	INLINE_BTN_TAGS   = "getTags"
	INLINE_BTN_REPEAT = "toggle_repeat"
	INLINE_BTN_DELETE = "delete_repeat"

	BTN_ADD_QUESTION = "➕ Добавить вопрос"
	BTN_REPEAT       = "📚 Управлять вопросами"
	BTN_ADD_CSV      = "📁 Добавить вопросы через CSV"
	BTN_DEL_QUESTION = "🗑 Удалить вопросы"
	BTN_PAUSE        = "⏸️ Выключить вопросы"
	BTN_RESUME       = "▶️ Включить вопросы"

	MSG_WRONG_BTN = "⚠️ Неизвестная команда. Используйте меню ниже."

	CMD_START         = "/start"
	CMD_DONE   string = "/done"
	CMD_CANCEL string = "/cancel"
)

func routers(ctx context.Context, b *telebot.Bot, domain *app.App) {
	b.Handle(CMD_START, start())

	b.Handle(&telebot.InlineButton{Unique: INLINE_BTN_REPEAT}, handleToggleRepeat())
	b.Handle(&telebot.InlineButton{Unique: INLINE_BTN_DELETE}, deleteRepeat())
	b.Handle(&telebot.InlineButton{Unique: INLINE_BTN_TAGS}, func(c telebot.Context) error {
		return add(domain)(c)
	})

	b.Handle(telebot.OnDocument, setQuestionsByCSV(domain))

	b.Handle(&telebot.InlineButton{Unique: "select_tag"}, func(ctx telebot.Context) error {
		tag := ctx.Data()
		return questionByTag(tag)(ctx)
	})

	b.Handle(telebot.OnText, func(ctx telebot.Context) error {
		userID := ctx.Sender().ID

		// Если пользователь в процессе добавления вопроса
		if draft, ok := drafts[userID]; ok && draft.Step > 0 {
			return add(domain)(ctx)
		}

		// TODO: нужно смотреть если пауза у пользователя, чтобы ничего не ломать
		switch ctx.Text() {
		case BTN_ADD_QUESTION:
			if err := getTags(ctx, GetUserFromContext(ctx).TGUserID, domain); err != nil {
				return err
			}
			drafts[userID] = &QuestionDraft{Step: 1}
			return add(domain)(ctx)
		case BTN_REPEAT:
			return showRepeatTagList(domain)(ctx)
		case BTN_ADD_CSV:
			return ctx.Send("📤 Отправьте CSV файл с вопросами в формате:\n\n"+
				"<code>Вопрос;Тег;Правильный ответ;Неправильный1;Неправильный2</code>\n\n"+
				"Пример:\n"+
				"<code>Что такое GPT?;AI;Generative Pre-trained Transformer;General Purpose Technology</code>",
				telebot.ModeHTML)
		case BTN_DEL_QUESTION:
			return deleteList()(ctx)
		case BTN_PAUSE:
			return pause()(ctx)
		case BTN_RESUME:
			return resume()(ctx)
		default:
			return ctx.Send(MSG_WRONG_BTN, mainMenu())
		}
	})

	dispatcher := NewDispatcher(ctx, domain, b)
	dispatcher.RegisterPollAnswerHandler()
	dispatcher.StartPollingLoop()
}
