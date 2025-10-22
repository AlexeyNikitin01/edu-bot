package ports

import (
	"strings"

	"gopkg.in/telebot.v3"

	"bot/internal/app"
	"bot/internal/repo/edu"
)

func routers(b *telebot.Bot, domain app.Apper, dispatcher *QuestionDispatcher) {
	b.Handle(CMD_START, func(ctx telebot.Context) error {
		return ctx.Send(MSG_GRETING, mainMenu())
	})

	// INLINES BUTTONS
	b.Handle(&telebot.InlineButton{Unique: INLINE_BTN_REPEAT_QUESTION}, handleToggleRepeat(domain))
	b.Handle(&telebot.InlineButton{Unique: INLINE_BTN_DELETE_QUESTION}, deleteQuestion(domain))
	b.Handle(&telebot.InlineButton{Unique: INLINE_BTN_DELETE_QUESTIONS_BY_TAG}, deleteQuestionByTag(domain))
	b.Handle(&telebot.InlineButton{Unique: INLINE_BTN_TAGS}, func(c telebot.Context) error {
		return upsertUserQuestion(domain)(c)
	})
	b.Handle(&telebot.InlineButton{Unique: INLINE_BTN_QUESTION_BY_TAG}, func(ctx telebot.Context) error {
		return questionByTag(ctx.Data())(ctx)
	})
	b.Handle(&telebot.InlineButton{Unique: INLINE_BACK_TAGS}, func(ctx telebot.Context) error {
		return backTags(domain)(ctx)
	})
	b.Handle(&telebot.InlineButton{Unique: INLINE_PAUSE_TAG}, func(ctx telebot.Context) error {
		return pauseTag(domain)(ctx)
	})
	b.Handle(&telebot.InlineButton{Unique: INLINE_FORGOT_HIGH_QUESTION}, forgotQuestion(domain, dispatcher))
	b.Handle(&telebot.InlineButton{Unique: INLINE_REMEMBER_HIGH_QUESTION}, rememberQuestion(domain, dispatcher))
	b.Handle(&telebot.InlineButton{Unique: INLINE_BTN_REPEAT_QUESTION_AFTER_POLL}, repeatQuestionAfterPoll(domain))
	b.Handle(&telebot.InlineButton{Unique: INLINE_BTN_REPEAT_QUESTION_AFTER_POLL_HIGH}, repeatQuestionAfterPollHigh(domain))
	b.Handle(&telebot.InlineButton{Unique: INLINE_BTN_DELETE_QUESTION_AFTER_POLL}, deleteQuestionAfterPoll(domain, dispatcher))
	b.Handle(&telebot.InlineButton{Unique: INLINE_BTN_DELETE_QUESTION_AFTER_POLL_HIGH}, deleteQuestionAfterPollHigh(domain, dispatcher))
	b.Handle(&telebot.InlineButton{Unique: INLINE_NEXT_QUESTION}, nextQuestion(dispatcher))
	b.Handle(&telebot.InlineButton{Unique: INLINE_EDIT_TAG}, setEdit(edu.TableNames.Tags, domain))
	b.Handle(&telebot.InlineButton{Unique: INLINE_EDIT_QUESTION}, getForUpdate(domain))
	b.Handle(&telebot.InlineButton{Unique: INLINE_EDIT_NAME_QUESTION}, setEdit(edu.QuestionTableColumns.Question, domain))
	b.Handle(&telebot.InlineButton{Unique: INLINE_EDIT_NAME_TAG_QUESTION}, setEdit(edu.QuestionTableColumns.TagID, domain))
	b.Handle(&telebot.InlineButton{Unique: INLINE_EDIT_ANSWER_QUESTION}, setEdit(edu.AnswerTableColumns.Answer, domain))
	b.Handle(&telebot.InlineButton{Unique: INLINE_SHOW_ANSWER}, viewAnswer(domain, true))
	b.Handle(&telebot.InlineButton{Unique: INLINE_TURN_ANSWER}, viewAnswer(domain, false))
	b.Handle(&telebot.InlineButton{Unique: INLINE_BTN_TASK_BY_TAG}, nextTask(domain))
	b.Handle(&telebot.InlineButton{Unique: INLINE_BTN_QUESTION_PAGE + "_prev"}, func(ctx telebot.Context) error {
		return handlePageNavigation(ctx, -1)
	})
	b.Handle(&telebot.InlineButton{Unique: INLINE_BTN_QUESTION_PAGE + "_next"}, func(ctx telebot.Context) error {
		return handlePageNavigation(ctx, 1)
	})

	b.Handle(telebot.OnDocument, setQuestionsByCSV(domain))

	b.Handle(telebot.OnText, func(ctx telebot.Context) error {
		// Если пользователь в процессе добавления вопроса
		if draft, ok := drafts[GetUserFromContext(ctx).TGUserID]; ok && draft.Step > 0 {
			return upsertUserQuestion(domain)(ctx)
		}

		text := ctx.Text()

		// Проверяем, может ли текст быть CSV (содержит хотя бы один разделитель)
		if strings.Contains(text, ";") && len(strings.Split(text, ";")) >= 3 {
			return setQuestionsByCSV(domain)(ctx)
		}

		switch ctx.Text() {
		case BTN_ADD_QUESTION:
			return upsertUserQuestion(domain)(ctx)
		case BTN_MANAGMENT_QUESTION:
			return showRepeatTagList(domain)(ctx)
		case BTN_ADD_CSV:
			return ctx.Send(MSG_CSV, telebot.ModeHTML)
		case BTN_NEXT_QUESTION:
			return nextQuestion(dispatcher)(ctx)
		case BTN_NEXT_TASK:
			return getTagsByTask(domain)(ctx)
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
	btnNextTask := menu.Text(BTN_NEXT_TASK)

	menu.Reply(
		menu.Row(btnAdd, btnCSV),
		menu.Row(btnMark, btnNext),
		menu.Row(btnNextTask),
	)

	return menu
}
