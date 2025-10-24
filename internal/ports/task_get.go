package ports

import (
	"bot/internal/app"
	"gopkg.in/telebot.v3"
	"strconv"
)

func getTagsByTask(domain app.Apper) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		u := GetUserFromContext(ctx)

		tags, err := domain.GetUniqueTagsByTask(GetContext(ctx), u.TGUserID)
		if err != nil {
			return err
		}

		if len(tags) == 0 {
			return err
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
			return err
		}

		q := uq.R.GetQuestion()
		tag = escapeMarkdown(q.R.GetTag().Tag)
		questionText := escapeMarkdown(q.Question)

		label := "🔔"
		if uq.IsEdu {
			label = "💤"
		}

		keyboard := NewTaskButtonsBuilder().
			AddShowAnswer(uq.QuestionID).
			AddDifficulty(q.ID).
			AddActions(q.ID, label).
			Build()

		return ctx.Send(
			tag+": "+questionText,
			telebot.ModeMarkdownV2,
			keyboard,
		)
	}
}

// Новая функция для отображения вопроса после выбора "легко" или "сложно"
func showQuestionAfterChoice(domain app.Apper) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		questionID, err := strconv.Atoi(ctx.Data())
		if err != nil {
			return err
		}

		q, err := domain.GetQuestionAnswers(GetContext(ctx), int64(questionID))
		if err != nil {
			return err
		}

		tag := escapeMarkdown(q.R.GetTag().Tag)
		questionText := escapeMarkdown(q.Question)

		// Используем билдер для создания кнопок навигации
		keyboard := NewTaskButtonsBuilder().
			AddNavigation(int64(questionID)).
			Build()

		return ctx.Send(
			"Выбор сохранён!\n\n"+tag+": "+questionText,
			telebot.ModeMarkdownV2,
			keyboard,
		)
	}
}

func skipTask(domain app.Apper) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		tagData := ctx.Data()

		uq, err := domain.GetTask(GetContext(ctx), GetUserFromContext(ctx).TGUserID, tagData)
		if err != nil {
			return err
		}

		if uq == nil {
			return ctx.Send("🎉 Все вопросы завершены! Вы великолепны!")
		}

		q := uq.R.GetQuestion()
		tag := escapeMarkdown(q.R.GetTag().Tag)
		questionText := escapeMarkdown(q.Question)

		label := "🔔"
		if uq.IsEdu {
			label = "💤"
		}

		// Используем билдер
		keyboard := NewTaskButtonsBuilder().
			AddShowAnswer(uq.QuestionID).
			AddDifficulty(q.ID).
			AddActions(q.ID, label).
			Build()

		return ctx.Send(
			"⏩ Вопрос пропущен!\n\n"+tag+": "+questionText,
			telebot.ModeMarkdownV2,
			keyboard,
		)
	}
}
