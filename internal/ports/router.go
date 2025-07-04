package ports

import (
	"gopkg.in/telebot.v3"

	"bot/internal/app"
	"bot/internal/repo/edu"
)

const (
	INLINE_BTN_TAGS                            = "tags"
	INLINE_BTN_REPEAT_QUESTION                 = "toggle_repeat"
	INLINE_BTN_DELETE_QUESTION                 = "delete_question"
	INLINE_BTN_DELETE_QUESTIONS_BY_TAG         = "delete_tag"
	INLINE_BTN_DELETE_QUESTION_AFTER_POLL      = "delete_question_after_poll"
	INLINE_BTN_DELETE_QUESTION_AFTER_POLL_HIGH = "delete_question_after_poll_high"
	INLINE_BTN_REPEAT_QUESTION_AFTER_POLL      = "repeat_question_after_poll"
	INLINE_BTN_REPEAT_QUESTION_AFTER_POLL_HIGH = "repeat_question_after_poll_high"
	INLINE_BTN_QUESTION_BY_TAG                 = "question_by_tag"
	INLINE_FORGOT_HIGH_QUESTION                = "forgot_high_question"
	INLINE_REMEMBER_HIGH_QUESTION              = "remember_high_question"
	INLINE_COMPLEX_QUESTION                    = "complex"
	INLINE_SIMPLE_QUESTION                     = "simple"
	INLINE_NEXT_QUESTION                       = "next_question"
	INLINE_EDIT_TAG                            = "edit_tag"
	INLINE_NAME_DELETE_AFTER_POLL              = "🗑️ УДАЛЕНИЕ"
	INLINE_NAME_REPEAT_AFTER_POLL              = "️ПОВТОРЕНИЕ"
	INLINE_NAME_DELETE                         = "🗑️"

	BTN_ADD_QUESTION       = "➕ Вопрос"
	BTN_MANAGMENT_QUESTION = "📚 Управление"
	BTN_ADD_CSV            = "➕ Вопросы CSV"
	BTN_NEXT_QUESTION      = "🌀 Дальше"

	MSG_WRONG_BTN = "⚠️ Неизвестная команда. Используйте меню ниже."
	MSG_CSV       = "📤 Отправьте CSV файл с вопросами в формате:\n\n" +
		"<code>Вопрос;Тег;Правильный ответ;Неправильный1;Неправильный2</code>\n\n" +
		"Пример:\n" +
		"<code>Что такое GPT?;AI;Generative Pre-trained Transformer;General Purpose Technology</code>"
	MSG_GRETING = "Добро пожаловать! Выберите действие:"

	CMD_START         = "/start"
	CMD_DONE   string = "/done"
	CMD_CANCEL string = "/cancel"
)

func routers(b *telebot.Bot, domain *app.App, dispatcher *QuestionDispatcher) {
	b.Handle(CMD_START, func(ctx telebot.Context) error {
		return ctx.Send(MSG_GRETING, mainMenu())
	})

	// INLINES BUTTONS
	b.Handle(&telebot.InlineButton{Unique: INLINE_BTN_REPEAT_QUESTION}, handleToggleRepeat(domain))
	b.Handle(&telebot.InlineButton{Unique: INLINE_BTN_DELETE_QUESTION}, deleteQuestion())
	b.Handle(&telebot.InlineButton{Unique: INLINE_BTN_DELETE_QUESTIONS_BY_TAG}, deleteQuestionByTag(domain))
	b.Handle(&telebot.InlineButton{Unique: INLINE_BTN_TAGS}, func(c telebot.Context) error {
		return add(domain)(c)
	})
	b.Handle(&telebot.InlineButton{Unique: INLINE_BTN_QUESTION_BY_TAG}, func(ctx telebot.Context) error {
		return questionByTag(ctx.Data())(ctx)
	})
	b.Handle(&telebot.InlineButton{Unique: INLINE_COMPLEX_QUESTION}, setHigh(true, MSG_CHOOSE_HIGH, domain))
	b.Handle(&telebot.InlineButton{Unique: INLINE_SIMPLE_QUESTION}, setHigh(false, MSG_CHOOSE_SIMPLE, domain))
	b.Handle(&telebot.InlineButton{Unique: INLINE_FORGOT_HIGH_QUESTION}, forgotQuestion(domain, dispatcher))
	b.Handle(&telebot.InlineButton{Unique: INLINE_REMEMBER_HIGH_QUESTION}, rememberQuestion(domain, dispatcher))
	b.Handle(&telebot.InlineButton{Unique: INLINE_BTN_REPEAT_QUESTION_AFTER_POLL}, repeatQuestionAfterPoll(domain))
	b.Handle(&telebot.InlineButton{Unique: INLINE_BTN_REPEAT_QUESTION_AFTER_POLL_HIGH}, repeatQuestionAfterPollHigh(domain))
	b.Handle(&telebot.InlineButton{Unique: INLINE_BTN_DELETE_QUESTION_AFTER_POLL}, deleteQuestionAfterPoll(domain, dispatcher))
	b.Handle(&telebot.InlineButton{Unique: INLINE_BTN_DELETE_QUESTION_AFTER_POLL_HIGH}, deleteQuestionAfterPollHigh(domain, dispatcher))
	b.Handle(&telebot.InlineButton{Unique: INLINE_NEXT_QUESTION}, nextQuestion(dispatcher))
	b.Handle(&telebot.InlineButton{Unique: INLINE_EDIT_TAG}, setEdit(edu.TableNames.Tags))

	// ADD CSV
	b.Handle(telebot.OnDocument, setQuestionsByCSV(domain))

	// WORK WITH MENU
	b.Handle(telebot.OnText, func(ctx telebot.Context) error {
		// Если пользователь в процессе добавления вопроса
		if draft, ok := drafts[GetUserFromContext(ctx).TGUserID]; ok && draft.Step > 0 {
			return add(domain)(ctx)
		}

		switch ctx.Text() {
		case BTN_ADD_QUESTION:
			return add(domain)(ctx)
		case BTN_MANAGMENT_QUESTION:
			return showRepeatTagList(domain)(ctx)
		case BTN_ADD_CSV:
			return ctx.Send(MSG_CSV, telebot.ModeHTML)
		case BTN_NEXT_QUESTION:
			return nextQuestion(dispatcher)(ctx)
		default:
			return ctx.Send(MSG_WRONG_BTN, mainMenu())
		}
	})

	b.Handle(telebot.OnPollAnswer, checkPollAnswer(domain, dispatcher))

	// Воркер для каждого пользователя, каждые 2 секунды рассылка вопросов для пользователей
	dispatcher.StartPollingLoop()
}

func mainMenu() *telebot.ReplyMarkup {
	menu := &telebot.ReplyMarkup{ResizeKeyboard: true}

	btnAdd := menu.Text(BTN_ADD_QUESTION)
	btnMark := menu.Text(BTN_MANAGMENT_QUESTION)
	btnCSV := menu.Text(BTN_ADD_CSV)
	btnNext := menu.Text(BTN_NEXT_QUESTION)

	menu.Reply(
		menu.Row(btnAdd, btnCSV),
		menu.Row(btnMark, btnNext),
	)

	return menu
}
