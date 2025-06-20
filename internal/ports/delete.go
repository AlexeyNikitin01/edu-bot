package ports

import (
	"fmt"
	"strconv"

	"github.com/volatiletech/sqlboiler/v4/boil"
	"gopkg.in/telebot.v3"

	"bot/internal/repo/edu"
)

// deleteList Показывает список вопросов для повторения
func deleteList() telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		tgUser := ctx.Sender()
		userID := tgUser.ID

		u, err := edu.Users(edu.UserWhere.TGUserID.EQ(userID)).One(GetContext(ctx), boil.GetContextDB())
		if err != nil {
			return ctx.Send("Вы не зарегистрированы.")
		}

		uqs, err := edu.UsersQuestions(edu.UsersQuestionWhere.UserID.EQ(u.TGUserID)).
			All(GetContext(ctx), boil.GetContextDB())
		if err != nil || len(uqs) == 0 {
			return ctx.Send("У вас нет вопросов.")
		}

		var btns [][]telebot.InlineButton

		for _, uq := range uqs {
			q, err := edu.Questions(edu.QuestionWhere.ID.EQ(uq.QuestionID)).
				One(GetContext(ctx), boil.GetContextDB())
			if err != nil {
				continue
			}

			label := "☑️"
			if uq.IsEdu {
				label = "✅"
			}

			// Кнопка удаления (можно добавить toggle отдельно при необходимости)
			btn := telebot.InlineButton{
				Unique: "delete_repeat",
				Text:   label + " " + q.Question,
				Data:   fmt.Sprintf("%d", uq.QuestionID),
			}

			btns = append(btns, []telebot.InlineButton{btn})
		}

		return ctx.Send("Выберите вопрос для удаления:", &telebot.ReplyMarkup{
			InlineKeyboard: btns,
		})
	}
}

// deleteRepeat Обрабатывает нажатие на кнопку удаления
func deleteRepeat() telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		qidStr := ctx.Data()
		questionID, err := strconv.Atoi(qidStr)
		if err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: "Некорректные данные"})
		}

		tgUser := ctx.Sender()
		userID := tgUser.ID

		_, err = edu.Users(edu.UserWhere.TGUserID.EQ(userID)).One(GetContext(ctx), boil.GetContextDB())
		if err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: "Вы не зарегистрированы"})
		}

		_, err = edu.UsersQuestions(
			edu.UsersQuestionWhere.UserID.EQ(userID),
			edu.UsersQuestionWhere.QuestionID.EQ(int64(questionID)),
		).DeleteAll(GetContext(ctx), boil.GetContextDB())
		if err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: "Ошибка удаления"})
		}

		// Повторная генерация оставшихся кнопок
		uqs, err := edu.UsersQuestions(edu.UsersQuestionWhere.UserID.EQ(userID)).
			All(GetContext(ctx), boil.GetContextDB())
		if err != nil || len(uqs) == 0 {
			return ctx.Edit("У вас больше нет вопросов.")
		}

		var btns [][]telebot.InlineButton
		for _, uq := range uqs {
			q, err := edu.Questions(edu.QuestionWhere.ID.EQ(uq.QuestionID)).
				One(GetContext(ctx), boil.GetContextDB())
			if err != nil {
				continue
			}

			label := "☑️"
			if uq.IsEdu {
				label = "✅"
			}

			btn := telebot.InlineButton{
				Unique: "delete_repeat",
				Text:   label + " " + q.Question,
				Data:   fmt.Sprintf("%d", uq.QuestionID),
			}

			btns = append(btns, []telebot.InlineButton{btn})
		}

		return ctx.Edit(&telebot.ReplyMarkup{
			InlineKeyboard: btns,
		})
	}
}
