package ports

import (
	"strconv"

	"github.com/volatiletech/sqlboiler/v4/boil"
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
			return ctx.Respond(&telebot.CallbackResponse{Text: "Некорректные данные"})
		}

		q, err := edu.Questions(
			edu.QuestionWhere.ID.EQ(int64(questionID))).One(GetContext(ctx), boil.GetContextDB())
		if err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: "Некорректные данные"})
		}

		_, err = edu.UsersQuestions(
			edu.UsersQuestionWhere.UserID.EQ(GetUserFromContext(ctx).TGUserID),
			edu.UsersQuestionWhere.QuestionID.EQ(int64(questionID)),
		).DeleteAll(GetContext(ctx), boil.GetContextDB(), false)
		if err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: "Ошибка удаления"})
		}

		return ctx.Edit(&telebot.ReplyMarkup{
			InlineKeyboard: getQuestionBtns(ctx, q.Tag),
		})
	}
}

// deleteQuestionByTag Удаление категории вопросов
func deleteQuestionByTag(domain app.Apper) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		tag := ctx.Data()

		_, err := edu.Questions(
			edu.QuestionWhere.Tag.EQ(tag)).DeleteAll(GetContext(ctx), boil.GetContextDB(), false)
		if err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: "Некорректные данные"})
		}

		return getTags(ctx, GetUserFromContext(ctx).TGUserID, domain)
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
