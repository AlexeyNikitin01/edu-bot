package ports

import (
	"bot/internal/app"
	"fmt"
	"gopkg.in/telebot.v3"
)

func getTagsByTask(domain app.Apper) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		u := GetUserFromContext(ctx)

		tags, err := domain.GetUniqueTagsByTask(GetContext(ctx), u.TGUserID)
		if err != nil {
			return sendErrorResponse(ctx, err.Error())
		}

		if len(tags) == 0 {
			return sendErrorResponse(ctx, MSG_EMPTY)
		}

		var tagButtons [][]telebot.InlineButton

		for _, tag := range tags {
			tagBtn := telebot.InlineButton{
				Unique: INLINE_BTN_TASK_BY_TAG,
				Text:   tag.Tag,
				Data:   tag.Tag,
			}

			tagButtons = append(tagButtons, []telebot.InlineButton{tagBtn})
		}

		return ctx.Send(MSG_LIST_TAGS, &telebot.ReplyMarkup{
			InlineKeyboard: tagButtons,
		})
	}
}

func nextTask(domain app.Apper) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		tag := ctx.Data()

		uq, err := domain.GetTask(GetContext(ctx), GetUserFromContext(ctx).TGUserID, tag)
		if err != nil {
			return sendErrorResponse(ctx, err.Error())
		}

		q := uq.R.GetQuestion()

		tag = escapeMarkdown(q.R.GetTag().Tag)
		questionText := escapeMarkdown(q.Question)

		forgot := telebot.InlineButton{
			Unique: INLINE_FORGOT_HIGH_QUESTION,
			Text:   MSG_FORGOT,
			Data:   fmt.Sprintf("%d", q.ID),
		}

		easy := telebot.InlineButton{
			Unique: INLINE_REMEMBER_HIGH_QUESTION,
			Text:   MSG_REMEMBER,
			Data:   fmt.Sprintf("%d", q.ID),
		}

		label := "🔔"
		if uq.IsEdu {
			label = "💤"
		}

		repeatBtn := telebot.InlineButton{
			Unique: INLINE_BTN_REPEAT_QUESTION_AFTER_POLL_HIGH,
			Text:   label,
			Data:   fmt.Sprintf("%d", uq.QuestionID),
		}

		deleteBtn := telebot.InlineButton{
			Unique: INLINE_BTN_DELETE_QUESTION_AFTER_POLL_HIGH,
			Text:   INLINE_NAME_DELETE_AFTER_POLL,
			Data:   fmt.Sprintf("%d", uq.QuestionID),
		}

		editBtn := telebot.InlineButton{
			Unique: INLINE_EDIT_QUESTION,
			Text:   "✏️",
			Data:   fmt.Sprintf("%d", uq.QuestionID),
		}

		showAnswerBtn := telebot.InlineButton{
			Unique: INLINE_SHOW_ANSWER,
			Text:   "📝 Показать ответ",
			Data:   fmt.Sprintf("%d", uq.QuestionID),
		}

		return ctx.Send(
			tag+": "+questionText, telebot.ModeMarkdownV2, &telebot.ReplyMarkup{
				InlineKeyboard: [][]telebot.InlineButton{
					{showAnswerBtn},
					{easy, forgot},
					{repeatBtn, deleteBtn, editBtn},
				},
			})
	}
}
