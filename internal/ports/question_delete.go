package ports

import (
	"fmt"
	"strconv"

	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"gopkg.in/telebot.v3"

	"bot/internal/app"
	"bot/internal/repo/edu"
)

const (
	MSG_SUCESS_DELETE_QUESTION = "🤫Вопрос удален👁"
)

// deleteQuestion Обрабатывает нажатие на кнопку удаления
func deleteQuestion() telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		qidStr := ctx.Data()
		questionID, err := strconv.Atoi(qidStr)
		if err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		q, err := edu.Questions(
			edu.QuestionWhere.ID.EQ(int64(questionID)),
			qm.Load(edu.QuestionRels.Tag)).One(GetContext(ctx), boil.GetContextDB())
		if err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		_, err = edu.UsersQuestions(
			edu.UsersQuestionWhere.UserID.EQ(GetUserFromContext(ctx).TGUserID),
			edu.UsersQuestionWhere.QuestionID.EQ(int64(questionID)),
		).DeleteAll(GetContext(ctx), boil.GetContextDB(), false)
		if err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		return ctx.Edit(&telebot.ReplyMarkup{
			InlineKeyboard: getQuestionBtns(ctx, q.R.GetTag().Tag),
		})
	}
}

// deleteQuestionByTag Удаление категории вопросов
func deleteQuestionByTag(domain app.Apper) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		tag := ctx.Data()

		_, err := edu.Questions(
			qm.InnerJoin(
				fmt.Sprintf("%s ON %s = %s",
					edu.TableNames.Tags,
					edu.QuestionTableColumns.TagID,
					edu.TagTableColumns.ID)),
			edu.TagWhere.Tag.EQ(tag)).DeleteAll(GetContext(ctx), boil.GetContextDB(), false)
		if err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		return showRepeatTagList(domain)(ctx)
	}
}

func deleteQuestionAfterPoll(_ *app.App, dispatcher *QuestionDispatcher) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		qidStr := ctx.Data()
		questionID, err := strconv.Atoi(qidStr)
		if err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		_, err = edu.UsersQuestions(
			edu.UsersQuestionWhere.UserID.EQ(GetUserFromContext(ctx).TGUserID),
			edu.UsersQuestionWhere.QuestionID.EQ(int64(questionID)),
		).DeleteAll(GetContext(ctx), boil.GetContextDB(), false)
		if err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		if err = ctx.Delete(); err != nil {
			return ctx.Send(err.Error())
		}

		if err = ctx.Send(MSG_SUCESS_DELETE_QUESTION); err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		dispatcher.mu.Lock()
		dispatcher.waitingForAnswer[GetUserFromContext(ctx).TGUserID] = false
		dispatcher.mu.Unlock()

		return nil
	}
}

func deleteQuestionAfterPollHigh(_ *app.App, dispatcher *QuestionDispatcher) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		qidStr := ctx.Data()
		questionID, err := strconv.Atoi(qidStr)
		if err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		_, err = edu.UsersQuestions(
			edu.UsersQuestionWhere.UserID.EQ(GetUserFromContext(ctx).TGUserID),
			edu.UsersQuestionWhere.QuestionID.EQ(int64(questionID)),
		).DeleteAll(GetContext(ctx), boil.GetContextDB(), false)
		if err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		if err = ctx.Delete(); err != nil {
			return ctx.Send(err.Error())
		}

		if err = ctx.Send(MSG_SUCESS_DELETE_QUESTION); err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		dispatcher.mu.Lock()
		dispatcher.waitingForAnswer[GetUserFromContext(ctx).TGUserID] = false
		dispatcher.mu.Unlock()

		return nil
	}
}
