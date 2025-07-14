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
	MSG_LIST_QUESTION = "ВОПРОСЫ: "
	MSG_LIST_TAGS     = "ТЭГИ: "
	MSG_EMPTY         = "У вас нет тэгов с вопросами"
)

func showRepeatTagList(domain app.Apper) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		u := GetUserFromContext(ctx)

		tags, err := domain.GetUniqueTags(GetContext(ctx), u.TGUserID)
		if err != nil {
			return ctx.Send(err.Error())
		}

		if len(tags) == 0 {
			return ctx.Send(MSG_EMPTY)
		}

		var tagButtons [][]telebot.InlineButton

		for _, tag := range tags {
			tagBtn := telebot.InlineButton{
				Unique: INLINE_BTN_QUESTION_BY_TAG,
				Text:   tag.Tag,
				Data:   tag.Tag,
			}
			deleteBtn := telebot.InlineButton{
				Unique: INLINE_BTN_DELETE_QUESTIONS_BY_TAG,
				Text:   INLINE_NAME_DELETE,
				Data:   tag.Tag,
			}
			editBtn := telebot.InlineButton{
				Unique: INLINE_EDIT_TAG,
				Text:   "✏️",
				Data:   fmt.Sprintf("%d", tag.ID),
			}
			tagButtons = append(tagButtons, []telebot.InlineButton{tagBtn, deleteBtn, editBtn})
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

		t, err := edu.Questions(edu.QuestionWhere.ID.EQ(int64(questionID)),
			qm.Load(edu.QuestionRels.Tag)).One(GetContext(ctx), boil.GetContextDB())
		if err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		return ctx.Edit(&telebot.ReplyMarkup{
			InlineKeyboard: getQuestionBtns(ctx, t.R.GetTag().Tag),
		})
	}
}

func getQuestionBtns(ctx telebot.Context, tag string) [][]telebot.InlineButton {
	qs, err := edu.Questions(
		qm.InnerJoin(fmt.Sprintf("%s ON %s = %s", edu.TableNames.UsersQuestions,
			edu.QuestionTableColumns.ID,
			edu.UsersQuestionTableColumns.QuestionID,
		)),
		qm.InnerJoin(fmt.Sprintf("%s ON %s = %s", edu.TableNames.Tags,
			edu.TagTableColumns.ID,
			edu.QuestionTableColumns.TagID,
		)),
		edu.UsersQuestionWhere.UserID.EQ(GetUserFromContext(ctx).TGUserID),
		edu.TagWhere.Tag.EQ(tag),
		edu.UsersQuestionWhere.DeletedAt.IsNull(),
	).All(GetContext(ctx), boil.GetContextDB())
	if err != nil || len(qs) == 0 {
		return nil
	}

	var btns [][]telebot.InlineButton

	for _, q := range qs {
		questionButtons := getQuestionBtn(
			ctx,
			q.ID,
			INLINE_BTN_REPEAT_QUESTION,
			q.Question,
			INLINE_NAME_DELETE,
			INLINE_BTN_DELETE_QUESTION,
		)
		btns = append(btns, []telebot.InlineButton{questionButtons[0]},
			[]telebot.InlineButton{questionButtons[1], questionButtons[2]})
	}

	return btns
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

	editBtn := telebot.InlineButton{
		Unique: INLINE_EDIT_QUESTION,
		Text:   "✏️",
		Data:   fmt.Sprintf("%d", qID),
	}

	return []telebot.InlineButton{repeatBtn, deleteBtn, editBtn}
}

func getForUpdate(domain app.Apper) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		qID := ctx.Data()
		id, err := strconv.Atoi(qID)
		if err != nil {
			return err
		}
		q, err := domain.GetQuestionAnswers(GetContext(ctx), int64(id))
		if err != nil {
			return err
		}

		var btns [][]telebot.InlineButton

		editQuestion := telebot.InlineButton{
			Unique: INLINE_EDIT_NAME_QUESTION,
			Text:   "вопрос: " + q.Question,
			Data:   fmt.Sprintf("%d", id),
		}

		editTag := telebot.InlineButton{
			Unique: INLINE_EDIT_NAME_TAG_QUESTION,
			Text:   "тэг: " + q.R.GetTag().Tag,
			Data:   fmt.Sprintf("%d", id),
		}

		btns = append(btns, []telebot.InlineButton{editQuestion})
		btns = append(btns, []telebot.InlineButton{editTag})

		for _, a := range q.R.GetAnswers() {
			answer := telebot.InlineButton{
				Unique: INLINE_EDIT_ANSWER_QUESTION,
				Text:   "ответ: " + a.Answer,
				Data:   fmt.Sprintf("%d", a.ID),
			}
			btns = append(btns, []telebot.InlineButton{answer})
		}

		return ctx.Send("Выберите поле: ", &telebot.ReplyMarkup{
			InlineKeyboard: btns,
		})
	}
}
