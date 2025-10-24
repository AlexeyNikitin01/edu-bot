package ports

import (
	"bot/internal/middleware"
	"bot/internal/ports/question"
	"bot/internal/ports/task"
	"context"
	"gopkg.in/telebot.v3"
	"strings"

	"bot/internal/domain"
	"bot/internal/repo/edu"
)

func routers(ctx context.Context, b *telebot.Bot, d domain.UseCases) {
	setupCommandHandlers(b)

	setupQuestionHandlers(b, ctx, d)

	setupTagHandlers(b, ctx, d)

	setupEditHandlers(b, ctx, d)

	setupTaskHandlers(ctx, b, d)

	setupContentHandlers(ctx, b, d)

	question.SendQuestion(ctx, b, d)
}

func setupCommandHandlers(b *telebot.Bot) {
	b.Handle(question.CMD_START, func(ctx telebot.Context) error {
		return ctx.Send(question.MSG_GRETING, mainMenu())
	})
}

func setupQuestionHandlers(b *telebot.Bot, ctx context.Context, d domain.UseCases) {
	// Создание вопросов
	b.Handle(telebot.OnText, createQuestionTextHandler(ctx, d))

	// Чтение вопросов
	b.Handle(&telebot.InlineButton{Unique: question.INLINE_NEXT_QUESTION}, question.NextQuestion(ctx, d))
	b.Handle(&telebot.InlineButton{Unique: question.INLINE_BTN_QUESTION_BY_TAG}, func(ctx telebot.Context) error {
		return question.QuestionByTag(ctx.Data())(ctx)
	})

	// Обновление вопросов
	b.Handle(&telebot.InlineButton{Unique: question.INLINE_BTN_REPEAT_QUESTION}, question.HandleToggleRepeat(ctx, d))
	b.Handle(&telebot.InlineButton{Unique: question.INLINE_REMEMBER_HIGH_QUESTION}, question.RememberQuestion(ctx, d))
	b.Handle(&telebot.InlineButton{Unique: question.INLINE_BTN_REPEAT_QUESTION_AFTER_POLL}, question.RepeatQuestionAfterPoll(ctx, d))
	b.Handle(&telebot.InlineButton{Unique: question.INLINE_BTN_REPEAT_QUESTION_AFTER_POLL_HIGH}, question.RepeatQuestionAfterPollHigh(ctx, d))

	// Удаление вопросов
	b.Handle(&telebot.InlineButton{Unique: question.INLINE_BTN_DELETE_QUESTION}, question.DeleteQuestion(ctx, d))
	b.Handle(&telebot.InlineButton{Unique: question.INLINE_BTN_DELETE_QUESTIONS_BY_TAG}, question.DeleteQuestionByTag(ctx, d))
	b.Handle(&telebot.InlineButton{Unique: question.INLINE_BTN_DELETE_QUESTION_AFTER_POLL}, question.DeleteQuestionAfterPoll(ctx, d))
	b.Handle(&telebot.InlineButton{Unique: question.INLINE_BTN_DELETE_QUESTION_AFTER_POLL_HIGH}, question.DeleteQuestionAfterPollHigh(ctx, d))
	b.Handle(&telebot.InlineButton{Unique: question.INLINE_FORGOT_HIGH_QUESTION}, question.ForgotQuestion(ctx, d))

	// Пагинация вопросов
	b.Handle(&telebot.InlineButton{Unique: question.INLINE_BTN_QUESTION_PAGE + "_prev"}, func(ctx telebot.Context) error {
		return question.HandlePageNavigation(ctx, -1)
	})
	b.Handle(&telebot.InlineButton{Unique: question.INLINE_BTN_QUESTION_PAGE + "_next"}, func(ctx telebot.Context) error {
		return question.HandlePageNavigation(ctx, 1)
	})
}

// Блок тегов
func setupTagHandlers(b *telebot.Bot, ctx context.Context, d domain.UseCases) {
	b.Handle(&telebot.InlineButton{Unique: question.INLINE_BTN_TAGS}, func(c telebot.Context) error {
		return question.UpsertUserQuestion(ctx, d)(c)
	})
	b.Handle(&telebot.InlineButton{Unique: question.INLINE_BACK_TAGS}, func(botCtx telebot.Context) error {
		return question.BackTags(ctx, d)(botCtx)
	})
	b.Handle(&telebot.InlineButton{Unique: question.INLINE_PAUSE_TAG}, func(botCtx telebot.Context) error {
		return question.PauseTag(ctx, d)(botCtx)
	})
}

// Блок редактирования
func setupEditHandlers(b *telebot.Bot, ctx context.Context, d domain.UseCases) {
	b.Handle(&telebot.InlineButton{Unique: question.INLINE_EDIT_TAG}, question.SetEdit(ctx, edu.TableNames.Tags, d))
	b.Handle(&telebot.InlineButton{Unique: question.INLINE_EDIT_QUESTION}, question.GetForUpdate(ctx, d))
	b.Handle(&telebot.InlineButton{Unique: question.INLINE_EDIT_NAME_QUESTION}, question.SetEdit(ctx, edu.QuestionTableColumns.Question, d))
	b.Handle(&telebot.InlineButton{Unique: question.INLINE_EDIT_NAME_TAG_QUESTION}, question.SetEdit(ctx, edu.QuestionTableColumns.TagID, d))
	b.Handle(&telebot.InlineButton{Unique: question.INLINE_EDIT_ANSWER_QUESTION}, question.SetEdit(ctx, edu.AnswerTableColumns.Answer, d))
	b.Handle(&telebot.InlineButton{Unique: question.INLINE_SHOW_CURRENT_VALUE}, question.ShowCurrentValue(ctx, d))
	b.Handle(&telebot.InlineButton{Unique: question.INLINE_COLLAPSE_VALUE}, question.CollapseValue(ctx, d))
}

// Блок задач
func setupTaskHandlers(ctx context.Context, b *telebot.Bot, d domain.UseCases) {
	b.Handle(&telebot.InlineButton{Unique: question.INLINE_BTN_TASK_BY_TAG}, task.NextTask(ctx, d))
}

// Блок обработки контента
func setupContentHandlers(ctx context.Context, b *telebot.Bot, d domain.UseCases) {
	b.Handle(telebot.OnDocument, question.SetQuestionsByCSV(ctx, d))
	b.Handle(telebot.OnPollAnswer, question.CheckPollAnswer(ctx, d))
}

// Обработчик текстовых команд для создания вопросов
func createQuestionTextHandler(ctx context.Context, d domain.UseCases) func(telebot.Context) error {
	return func(ctxBot telebot.Context) error {
		user := middleware.GetUserFromContext(ctxBot)
		if user == nil {
			return ctxBot.Send(question.MSG_WRONG_BTN, mainMenu())
		}

		// Проверяем, есть ли активный черновик у пользователя в Redis
		//draft, err := cache.GetDraft(ctxBot, user.TGUserID)
		//if err != nil {
		//	return err
		//}

		// Если есть активный черновик, обрабатываем его
		//if draft != nil && draft.Step > 0 {
		//	return question.UpsertUserQuestion(d, cache)(ctxBot)
		//}

		text := ctxBot.Text()

		// Проверяем, может ли текст быть CSV (содержит хотя бы один разделитель)
		if strings.Contains(text, ";") && len(strings.Split(text, ";")) >= 3 {
			return question.SetQuestionsByCSV(ctx, d)(ctxBot)
		}

		switch ctxBot.Text() {
		case question.BTN_ADD_QUESTION:
			return question.UpsertUserQuestion(ctx, d)(ctxBot)
		case question.BTN_MANAGMENT_QUESTION:
			return question.ShowRepeatTagList(ctx, d)(ctxBot)
		case question.BTN_ADD_CSV:
			return ctxBot.Send(question.MSG_CSV, telebot.ModeHTML)
		case question.BTN_NEXT_TASK:
			return task.GetTagsByTask(ctx, d)(ctxBot)
		default:
			return ctxBot.Send(question.MSG_WRONG_BTN, mainMenu())
		}
	}
}

func mainMenu() *telebot.ReplyMarkup {
	menu := &telebot.ReplyMarkup{ResizeKeyboard: true}

	btnAdd := menu.Text(question.BTN_ADD_QUESTION)
	btnMark := menu.Text(question.BTN_MANAGMENT_QUESTION)
	btnCSV := menu.Text(question.BTN_ADD_CSV)
	btnNext := menu.Text(question.BTN_NEXT_QUESTION)
	btnNextTask := menu.Text(question.BTN_NEXT_TASK)

	menu.Reply(
		menu.Row(btnAdd, btnCSV),
		menu.Row(btnMark, btnNext),
		menu.Row(btnNextTask),
	)

	return menu
}
