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

func showRepeatTagList(domain app.Apper, action string) telebot.HandlerFunc {
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
			tagBtn := telebot.InlineButton{
				Unique: INLINE_BTN_QUESTION_BY_TAG,
				Text:   tag,
				Data:   tag,
			}
			deleteBtn := telebot.InlineButton{
				Unique: INLINE_BTN_DELETE_QUESTION_BY_TAG,
				Text:   "🗑️",
				Data:   tag,
			}
			tagButtons = append(tagButtons, []telebot.InlineButton{tagBtn, deleteBtn})
		}

		return ctx.Send("ТЭГИ: ", &telebot.ReplyMarkup{
			InlineKeyboard: tagButtons,
		})
	}
}

func questionByTag(tag string) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		return ctx.Send("ВОПРОСЫ: ", &telebot.ReplyMarkup{
			InlineKeyboard: getQuestionBtns(ctx, tag),
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

		uq, err := edu.UsersQuestions(
			edu.UsersQuestionWhere.UserID.EQ(GetUserFromContext(ctx).TGUserID),
			edu.UsersQuestionWhere.QuestionID.EQ(int64(questionID)),
			qm.Load(edu.UsersQuestionRels.Question),
		).One(GetContext(ctx), boil.GetContextDB())
		if err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: "Вопрос не найден."})
		}

		uq.IsEdu = !uq.IsEdu
		_, err = uq.Update(GetContext(ctx), boil.GetContextDB(), boil.Infer())
		if err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: "Не удалось обновить."})
		}

		return ctx.Edit(&telebot.ReplyMarkup{
			InlineKeyboard: getQuestionBtns(ctx, uq.R.GetQuestion().Tag),
		})
	}
}

func getQuestionBtns(ctx telebot.Context, tag string) [][]telebot.InlineButton {
	uqs, err := edu.UsersQuestions(edu.UsersQuestionWhere.UserID.EQ(GetUserFromContext(ctx).TGUserID)).
		All(GetContext(ctx), boil.GetContextDB())
	if err != nil || len(uqs) == 0 {
		return nil
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

		repeatBtn := telebot.InlineButton{
			Unique: INLINE_BTN_REPEAT_QUESTION,
			Text:   label + " " + q.Question,
			Data:   fmt.Sprintf("%d", uq.QuestionID),
		}

		deleteBtn := telebot.InlineButton{
			Unique: INLINE_BTN_DELETE_QUESTION,
			Text:   "🗑️",
			Data:   fmt.Sprintf("%d", uq.QuestionID),
		}

		btns = append(btns, []telebot.InlineButton{repeatBtn, deleteBtn})
	}

	return btns
}
