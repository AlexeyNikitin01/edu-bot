package ports

import (
	"fmt"
	"strconv"

	"github.com/volatiletech/sqlboiler/v4/boil"
	"gopkg.in/telebot.v3"

	"bot/internal/app"
	"bot/internal/repo/edu"
)

func showRepeatTagList(domain app.Apper) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		u := GetUserFromContext(ctx)

		uniqueTags, err := domain.GetUniqueTags(GetContext(ctx), u.TGUserID)
		if err != nil {
			return ctx.Send("Ошибка при получении тегов.")
		}

		if len(uniqueTags) == 0 {
			return nil
		}

		var tagButtons [][]telebot.InlineButton

		for _, tag := range uniqueTags {
			btn := telebot.InlineButton{
				Unique: "select_tag",
				Text:   tag,
				Data:   tag,
			}
			tagButtons = append(tagButtons, []telebot.InlineButton{btn})
		}

		allBtn := telebot.InlineButton{
			Unique: "select_tag",
			Text:   "Все вопросы",
			Data:   "all",
		}
		tagButtons = append(tagButtons, []telebot.InlineButton{allBtn})

		return ctx.Send("Выберите тег для просмотра вопросов:", &telebot.ReplyMarkup{
			InlineKeyboard: tagButtons,
		})
	}
}

func questionByTag(tag string) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		uqs, err := edu.UsersQuestions(edu.UsersQuestionWhere.UserID.EQ(GetUserFromContext(ctx).TGUserID)).
			All(GetContext(ctx), boil.GetContextDB())
		if err != nil || len(uqs) == 0 {
			return ctx.Send("У вас нет вопросов.")
		}

		var btns [][]telebot.InlineButton

		for _, uq := range uqs {
			q, err := edu.Questions(
				edu.QuestionWhere.Tag.EQ(tag),
				edu.QuestionWhere.ID.EQ(uq.QuestionID),
			).One(GetContext(ctx), boil.GetContextDB())
			if err != nil {
				continue
			}

			label := "☑️"
			if uq.IsEdu {
				label = "✅"
			}

			btn := telebot.InlineButton{
				Unique: INLINE_BTN_REPEAT,
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

// handleToggleRepeat выбор учить или не учить вопрос.
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
				Unique: INLINE_BTN_REPEAT,
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
