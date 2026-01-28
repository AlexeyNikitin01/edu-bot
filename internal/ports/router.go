package ports

import (
	"bot/internal/domain"
	"bot/internal/middleware"
	"bot/internal/ports/menu"
	"bot/internal/ports/question"
	"bot/internal/ports/tags"
	"bot/internal/ports/task"
	"bot/internal/repo/edu"
	"context"
	"gopkg.in/telebot.v3"
	"strings"
)

func routers(ctx context.Context, b *telebot.Bot, d domain.UseCases) {
	setupCommandHandlers(b)
	questionHandlerCRUD(b, ctx, d)
	tagHandlersCRUD(b, ctx, d)
	setupEditHandlers(b, ctx, d)
	setupTaskHandlers(ctx, b, d)
	setupContentHandlers(ctx, b, d)

	go question.SendQuestion(ctx, b, d)
}

func setupCommandHandlers(b *telebot.Bot) {
	b.Handle(question.CMD_START, func(ctx telebot.Context) error {
		return ctx.Send(question.MSG_GRETING, menu.BtnsMenu())
	})
}

func questionHandlerCRUD(b *telebot.Bot, ctx context.Context, d domain.UseCases) {
	// Создание вопросов
	b.Handle(telebot.OnText, processBtnsMenu(ctx, d))
	b.Handle(&telebot.InlineButton{Unique: tags.INLINE_SELECT_TAG}, question.HandleTagSelection(ctx, d))

	// Чтение вопросов
	b.Handle(&telebot.InlineButton{Unique: question.INLINE_NEXT_QUESTION}, question.NextQuestion(ctx, d))
	b.Handle(&telebot.InlineButton{Unique: tags.INLINE_BTN_QUESTION_BY_TAG}, func(ctxBot telebot.Context) error {
		return question.ListQuestions(ctx, ctxBot.Data(), d)(ctxBot)
	})
	b.Handle(&telebot.InlineButton{Unique: question.INLINE_SHOW_ANSWER}, question.ViewAnswer(ctx, d, true))
	b.Handle(&telebot.InlineButton{Unique: question.INLINE_TURN_ANSWER}, question.ViewAnswer(ctx, d, false))

	// Обновление вопросов
	b.Handle(&telebot.InlineButton{Unique: question.INLINE_BTN_REPEAT_QUESTION}, question.IsRepeat(ctx, d))
	b.Handle(&telebot.InlineButton{Unique: question.INLINE_REMEMBER_HIGH_QUESTION}, question.RememberQuestion(ctx, d))
	b.Handle(&telebot.InlineButton{Unique: question.INLINE_BTN_REPEAT_QUESTION_AFTER_POLL_HIGH}, question.IsRepeatByPoll(ctx, d))

	// Удаление вопросов
	b.Handle(&telebot.InlineButton{Unique: question.INLINE_BTN_DELETE_QUESTION}, question.DeleteQuestion(ctx, d))
	b.Handle(&telebot.InlineButton{Unique: question.INLINE_BTN_DELETE_QUESTION_AFTER_POLL_HIGH}, question.DeleteQuestionAfterPollHigh(ctx, d))
	b.Handle(&telebot.InlineButton{Unique: question.INLINE_FORGOT_HIGH_QUESTION}, question.ForgotQuestion(ctx, d))

	// Пагинация вопросов
	b.Handle(&telebot.InlineButton{Unique: question.INLINE_BTN_QUESTION_PAGE + "_prev"}, question.HandlePageNavigation(ctx, d))
	b.Handle(&telebot.InlineButton{Unique: question.INLINE_BTN_QUESTION_PAGE + "_next"}, question.HandlePageNavigation(ctx, d))
}

func tagHandlersCRUD(b *telebot.Bot, ctx context.Context, d domain.UseCases) {
	// Основные кнопки тегов
	b.Handle(&telebot.InlineButton{Unique: tags.INLINE_BTN_TAGS}, question.UpsertUserQuestion(ctx, d))
	b.Handle(&telebot.InlineButton{Unique: tags.INLINE_BACK_TAGS}, tags.HandleTagPagination(ctx, d))
	b.Handle(&telebot.InlineButton{Unique: tags.INLINE_PAUSE_TAG}, tags.PauseTag(ctx, d))
	b.Handle(&telebot.InlineButton{Unique: tags.INLINE_BTN_DELETE_QUESTIONS_BY_TAG}, tags.DeleteQuestionByTag(ctx, d))

	// Пагинация тегов
	b.Handle(&telebot.InlineButton{Unique: tags.INLINE_PAGINATION_PREV}, tags.HandleTagPagination(ctx, d))
	b.Handle(&telebot.InlineButton{Unique: tags.INLINE_PAGINATION_NEXT}, tags.HandleTagPagination(ctx, d))

	b.Handle(&telebot.InlineButton{Unique: tags.INLINE_EDIT_TAG}, question.SetEdit(ctx, edu.TableNames.Tags, d))

	// Обработка отсутствия тегов
	b.Handle(&telebot.InlineButton{Unique: tags.INLINE_NO_TAGS}, func(botCtx telebot.Context) error {
		return botCtx.Respond(&telebot.CallbackResponse{
			Text:      task.MsgNoTagsAvailable,
			ShowAlert: true,
		})
	})
}

func setupEditHandlers(b *telebot.Bot, ctx context.Context, d domain.UseCases) {
	b.Handle(&telebot.InlineButton{Unique: question.INLINE_EDIT_QUESTION}, question.GetForUpdate(ctx, d))
	b.Handle(&telebot.InlineButton{Unique: question.INLINE_EDIT_NAME_QUESTION}, question.SetEdit(ctx, edu.QuestionTableColumns.Question, d))
	b.Handle(&telebot.InlineButton{Unique: question.INLINE_EDIT_NAME_TAG_QUESTION}, question.SetEdit(ctx, edu.QuestionTableColumns.TagID, d))
	b.Handle(&telebot.InlineButton{Unique: question.INLINE_EDIT_ANSWER_QUESTION}, question.SetEdit(ctx, edu.AnswerTableColumns.Answer, d))
	b.Handle(&telebot.InlineButton{Unique: question.INLINE_SHOW_CURRENT_VALUE}, question.ShowCurrentValue(ctx, d))
	b.Handle(&telebot.InlineButton{Unique: question.INLINE_COLLAPSE_VALUE}, question.CollapseValue(ctx, d))
}

func setupTaskHandlers(ctx context.Context, b *telebot.Bot, d domain.UseCases) {
	b.Handle(&telebot.InlineButton{Unique: question.INLINE_BTN_TASK_BY_TAG}, task.NextTask(ctx, d))
	b.Handle(&telebot.InlineButton{Unique: task.INLINE_REMEMBER_HIGH_TASK}, task.UpdateTaskTotal(ctx, d, true))
	b.Handle(&telebot.InlineButton{Unique: task.INLINE_FORGOT_HIGH_TASK}, task.UpdateTaskTotal(ctx, d, false))
	b.Handle(&telebot.InlineButton{Unique: task.INLINE_NEXT_TASK}, task.NextTask(ctx, d))
	b.Handle(&telebot.InlineButton{Unique: task.INLINE_SKIP_TASK}, task.SkipTask(ctx, d))
	b.Handle(&telebot.InlineButton{Unique: task.INLINE_SHOW_ANSWER_TASK}, task.ViewAnswerTask(ctx, d, true))
	b.Handle(&telebot.InlineButton{Unique: task.INLINE_TURN_ANSWER_TASK}, task.ViewAnswerTask(ctx, d, false))
	b.Handle(&telebot.InlineButton{Unique: task.INLINE_BTN_REPEAT_TASK_AFTER_POLL}, task.IsRepeatTask(ctx, d))
	b.Handle(&telebot.InlineButton{Unique: task.INLINE_BTN_DELETE_TASK_AFTER_POLL}, task.DeleteTask(ctx, d))
}

func setupContentHandlers(ctx context.Context, b *telebot.Bot, d domain.UseCases) {
	b.Handle(telebot.OnDocument, question.SetQuestionsByCSV(ctx, d))
	b.Handle(telebot.OnPollAnswer, question.CheckPollAnswer(ctx, d))
}

func processBtnsMenu(ctx context.Context, d domain.UseCases) func(telebot.Context) error {
	return func(ctxBot telebot.Context) error {
		user := middleware.GetUserFromContext(ctxBot)
		if user == nil {
			return ctxBot.Send(menu.MSG_WRONG_BTN, menu.BtnsMenu())
		}

		draft, err := d.GetDraftQuestion(ctx, user.TGUserID)
		if err != nil {
			return err
		}

		if draft != nil && draft.Step > 0 {
			return question.UpsertUserQuestion(ctx, d)(ctxBot)
		}

		text := ctxBot.Text()

		if strings.Contains(text, ";") && len(strings.Split(text, ";")) >= 3 {
			return question.SetQuestionsByCSV(ctx, d)(ctxBot)
		}

		switch ctxBot.Text() {
		case menu.BTN_ADD_QUESTION:
			return question.UpsertUserQuestion(ctx, d)(ctxBot)
		case menu.BTN_MANAGMENT_QUESTION:
			return tags.ShowRepeatTagList(ctx, d)(ctxBot)
		case menu.BTN_ADD_CSV:
			return ctxBot.Send(question.MSG_CSV, telebot.ModeHTML)
		case menu.BTN_NEXT_TASK:
			return task.GetTagsByTask(ctx, d)(ctxBot)
		case menu.BTN_NEXT_QUESTION:
			return question.NextQuestion(ctx, d)(ctxBot)
		default:
			return ctxBot.Send(menu.MSG_WRONG_BTN, menu.BtnsMenu())
		}
	}
}
