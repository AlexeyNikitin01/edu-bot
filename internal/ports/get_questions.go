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
	INLINE_NAME_DELETE_AFTER_POLL = "🗑️ УДАЛЕНИЕ"
	INLINE_NAME_REPEAT_AFTER_POLL = "️ПОВТОРЕНИЕ"
	INLINE_NAME_DELETE            = "🗑️"

	MSG_LIST_QUESTION = "ВОПРОСЫ: "
	MSG_LIST_TAGS     = "ТЭГИ: "
)

func showRepeatTagList(domain app.Apper, action string) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		u := GetUserFromContext(ctx)

		uniqueTags, err := domain.GetUniqueTags(GetContext(ctx), u.TGUserID)
		if err != nil {
			return ctx.Send(err.Error())
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
				Unique: INLINE_BTN_DELETE_QUESTIONS_BY_TAG,
				Text:   INLINE_NAME_DELETE,
				Data:   tag,
			}
			tagButtons = append(tagButtons, []telebot.InlineButton{tagBtn, deleteBtn})
		}

		return ctx.Send(MSG_LIST_TAGS, &telebot.ReplyMarkup{
			InlineKeyboard: tagButtons,
		})
	}
}

func questionByTag(tag string) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		return ctx.Send(MSG_LIST_QUESTION, &telebot.ReplyMarkup{
			InlineKeyboard: getQuestionBtns(ctx, tag),
		})
	}
}

// handleToggleRepeat выбор учить или не учить вопрос.
func handleToggleRepeat(domain app.Apper) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		qidStr := ctx.Data() // получаем questionID из callback data
		questionID, err := strconv.Atoi(qidStr)
		if err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		if err = domain.UpdateIsEduUserQuestion(GetContext(ctx), GetUserFromContext(ctx).TGUserID, int64(questionID)); err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		t, err := edu.FindQuestion(GetContext(ctx), boil.GetContextDB(), int64(questionID), edu.QuestionColumns.Tag)
		if err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		return ctx.Edit(&telebot.ReplyMarkup{
			InlineKeyboard: getQuestionBtns(ctx, t.Tag),
		})
	}
}

func getQuestionBtns(ctx telebot.Context, tag string) [][]telebot.InlineButton {
	qs, err := edu.Questions(
		qm.InnerJoin(fmt.Sprintf("%s ON %s = %s", edu.TableNames.UsersQuestions,
			edu.QuestionTableColumns.ID,
			edu.UsersQuestionTableColumns.QuestionID,
		)),
		edu.UsersQuestionWhere.UserID.EQ(GetUserFromContext(ctx).TGUserID),
		edu.QuestionWhere.Tag.EQ(tag),
		edu.UsersQuestionWhere.DeletedAt.IsNull(),
	).All(GetContext(ctx), boil.GetContextDB())
	if err != nil || len(qs) == 0 {
		return nil
	}

	var btns [][]telebot.InlineButton

	for _, q := range qs {
		btns = append(btns, getQuestionBtn(
			ctx,
			q.ID,
			INLINE_BTN_REPEAT_QUESTION,
			q.Question,
			INLINE_NAME_DELETE,
			INLINE_BTN_DELETE_QUESTION,
		))
	}

	return btns
}

// handleToggleRepeatAfterPoll выбор учить или не учить вопрос рядом с опросом.
func handleToggleRepeatAfterPoll(domain app.Apper) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		qidStr := ctx.Data() // получаем questionID из callback data
		questionID, err := strconv.Atoi(qidStr)
		if err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		if err = domain.UpdateIsEduUserQuestion(GetContext(ctx), GetUserFromContext(ctx).TGUserID, int64(questionID)); err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		return ctx.Edit(&telebot.ReplyMarkup{
			InlineKeyboard: [][]telebot.InlineButton{getQuestionBtn(
				ctx,
				int64(questionID),
				INLINE_BTN_REPEAT_QUESTION_AFTER_POLL,
				INLINE_NAME_REPEAT_AFTER_POLL,
				INLINE_NAME_DELETE_AFTER_POLL,
				INLINE_BTN_DELETE_QUESTION_AFTER_POLL,
			)},
		})
	}
}

func getQuestionBtn(
	ctx telebot.Context, qID int64, repeat, repeatMSG, deleteMSG, delete string,
) []telebot.InlineButton {
	uq, err := edu.UsersQuestions(
		edu.UsersQuestionWhere.UserID.EQ(GetUserFromContext(ctx).TGUserID),
		edu.UsersQuestionWhere.QuestionID.EQ(qID),
		edu.UsersQuestionWhere.DeletedAt.IsNull(),
	).One(GetContext(ctx), boil.GetContextDB())
	if err != nil {
		return nil
	}

	label := "☑️"
	if uq.IsEdu {
		label = "✅"
	}

	repeatBtn := telebot.InlineButton{
		Unique: repeat,
		Text:   label + repeatMSG,
		Data:   fmt.Sprintf("%d", qID),
	}

	deleteBtn := telebot.InlineButton{
		Unique: delete,
		Text:   deleteMSG,
		Data:   fmt.Sprintf("%d", qID),
	}

	return []telebot.InlineButton{repeatBtn, deleteBtn}
}
