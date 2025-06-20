package ports

import (
	"fmt"
	"strconv"

	"github.com/volatiletech/sqlboiler/v4/boil"
	"gopkg.in/telebot.v3"

	"bot/internal/repo/edu"
)

func showRepeatList() telebot.HandlerFunc {
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
			q, err := edu.Questions(edu.QuestionWhere.ID.EQ(uq.QuestionID)).One(GetContext(ctx), boil.GetContextDB())
			if err != nil {
				continue
			}

			label := "☑️"
			if uq.IsEdu {
				label = "✅"
			}

			btn := telebot.InlineButton{
				Unique: "toggle_repeat",
				Text:   label + " " + q.Question,
				Data:   fmt.Sprintf("%d", uq.QuestionID),
			}

			btns = append(btns, []telebot.InlineButton{btn})
		}

		return ctx.Send("Выберите вопросы для повторения:", &telebot.ReplyMarkup{
			InlineKeyboard: btns,
		})
	}
}

func handleToggleRepeat() telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		qidStr := ctx.Data() // получаем questionID из callback data
		questionID, err := strconv.Atoi(qidStr)
		if err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: "Ошибка данных."})
		}

		tgUser := ctx.Sender()
		userID := tgUser.ID

		u, err := edu.Users(edu.UserWhere.TGUserID.EQ(userID)).One(GetContext(ctx), boil.GetContextDB())
		if err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: "Вы не зарегистрированы."})
		}

		uq, err := edu.UsersQuestions(
			edu.UsersQuestionWhere.UserID.EQ(u.TGUserID),
			edu.UsersQuestionWhere.QuestionID.EQ(int64(questionID)),
		).One(GetContext(ctx), boil.GetContextDB())
		if err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: "Вопрос не найден."})
		}

		uq.IsEdu = !uq.IsEdu
		_, err = uq.Update(GetContext(ctx), boil.GetContextDB(), boil.Infer())
		if err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: "Не удалось обновить."})
		}

		// Получаем все вопросы заново, чтобы обновить inline-клавиатуру
		uqs, err := edu.UsersQuestions(edu.UsersQuestionWhere.UserID.EQ(u.TGUserID)).
			All(GetContext(ctx), boil.GetContextDB())
		if err != nil || len(uqs) == 0 {
			return ctx.Edit("У вас нет вопросов.")
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
				Unique: "toggle_repeat",
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
