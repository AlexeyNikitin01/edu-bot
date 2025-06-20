package ports

import (
	"github.com/volatiletech/sqlboiler/v4/boil"
	"gopkg.in/telebot.v3"

	"bot/internal/repo/edu"
)

func pause() telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		tgUser := ctx.Sender()
		userID := tgUser.ID

		u, err := edu.Users(edu.UserWhere.TGUserID.EQ(userID)).One(GetContext(ctx), boil.GetContextDB())
		if err != nil {
			return ctx.Send("Вы не зарегистрированы.")
		}

		_, err = edu.UsersQuestions(edu.UsersQuestionWhere.UserID.EQ(u.TGUserID)).
			UpdateAll(GetContext(ctx), boil.GetContextDB(), edu.M{
				edu.UsersQuestionColumns.IsEdu:   false,
				edu.UsersQuestionColumns.IsPause: true,
			})
		if err != nil {
			return ctx.Send("У вас нет вопросов.")
		}

		return nil
	}
}

func resume() telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		tgUser := ctx.Sender()
		userID := tgUser.ID

		u, err := edu.Users(edu.UserWhere.TGUserID.EQ(userID)).One(GetContext(ctx), boil.GetContextDB())
		if err != nil {
			return ctx.Send("Вы не зарегистрированы.")
		}

		_, err = edu.UsersQuestions(
			edu.UsersQuestionWhere.UserID.EQ(u.TGUserID),
			edu.UsersQuestionWhere.IsPause.EQ(true),
		).UpdateAll(GetContext(ctx), boil.GetContextDB(), edu.M{
			edu.UsersQuestionColumns.IsEdu:   true,
			edu.UsersQuestionColumns.IsPause: false,
		})
		if err != nil {
			return ctx.Send("У вас нет вопросов.")
		}

		return nil
	}
}
