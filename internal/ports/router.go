package ports

import (
	"context"
	"strings"

	"gopkg.in/telebot.v3"

	"bot/internal/app"
)

const (
	INLINE_BTN_TAGS            = "tags"
	INLINE_BTN_REPEAT          = "toggle_repeat"
	INLINE_BTN_DELETE          = "delete_repeat"
	INLINE_BTN_QUESTION_BY_TAG = "question_by_tag"

	BTN_ADD_QUESTION = "➕ Добавить вопрос"
	BTN_REPEAT       = "📚 Управлять вопросами"
	BTN_ADD_CSV      = "📁 Добавить вопросы через CSV"
	BTN_DEL_QUESTION = "🗑 Удалить вопросы"
	BTN_PAUSE        = "⏸️ Выключить вопросы"
	BTN_RESUME       = "▶️ Включить вопросы"

	MSG_WRONG_BTN = "⚠️ Неизвестная команда. Используйте меню ниже."
	MSG_CSV       = "📤 Отправьте CSV файл с вопросами в формате:\n\n" +
		"<code>Вопрос;Тег;Правильный ответ;Неправильный1;Неправильный2</code>\n\n" +
		"Пример:\n" +
		"<code>Что такое GPT?;AI;Generative Pre-trained Transformer;General Purpose Technology</code>"

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

	b.Handle(&telebot.InlineButton{Unique: INLINE_BTN_QUESTION_BY_TAG}, func(ctx telebot.Context) error {
		datas := strings.Split(ctx.Data(), ";")
		return questionByTag(datas[0], datas[1])(ctx)
	})

	b.Handle(telebot.OnText, func(ctx telebot.Context) error {
		// Если пользователь в процессе добавления вопроса
		if draft, ok := drafts[GetUserFromContext(ctx).TGUserID]; ok && draft.Step > 0 {
			return add(domain)(ctx)
		}

		// TODO: нужно смотреть если пауза у пользователя, чтобы ничего не ломать
		switch ctx.Text() {
		case BTN_ADD_QUESTION:
			if err := getTags(ctx, GetUserFromContext(ctx).TGUserID, domain); err != nil {
				return err
			}
			drafts[GetUserFromContext(ctx).TGUserID] = &QuestionDraft{Step: 1}
			return add(domain)(ctx)
		case BTN_REPEAT:
			return showRepeatTagList(domain, INLINE_BTN_REPEAT)(ctx)
		case BTN_ADD_CSV:
			return ctx.Send(MSG_CSV, telebot.ModeHTML)
		case BTN_DEL_QUESTION:
			return showRepeatTagList(domain, INLINE_BTN_DELETE)(ctx)
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
