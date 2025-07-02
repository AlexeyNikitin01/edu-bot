package ports

import (
	"strconv"

	"github.com/volatiletech/sqlboiler/v4/boil"
	"gopkg.in/telebot.v3"

	"bot/internal/repo/edu"
)

// deleteRepeat Обрабатывает нажатие на кнопку удаления
func deleteRepeat() telebot.HandlerFunc {
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
		).DeleteAll(GetContext(ctx), boil.GetContextDB())
		if err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: "Ошибка удаления"})
		}

		return ctx.Edit(&telebot.ReplyMarkup{
			InlineKeyboard: getQuestionBtns(ctx, q.Tag, INLINE_BTN_DELETE),
		})
	}
}
