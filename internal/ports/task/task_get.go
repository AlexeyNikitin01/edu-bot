package task

import (
	"bot/internal/domain"
	"bot/internal/middleware"
	"context"
	"gopkg.in/telebot.v3"
	"strconv"
)

func GetTagsByTask(ctx context.Context, d domain.UseCases) telebot.HandlerFunc {
	return func(ctxBot telebot.Context) error {
		u := middleware.GetUserFromContext(ctxBot)

		tags, err := d.GetUniqueTagsByTask(GetContext(ctxBot), u.TGUserID)
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
				Text:   tag.TagService,
				Data:   tag.TagService,
			}

			tagButtons = append(tagButtons, []telebot.InlineButton{tagBtn})
		}

		return ctxBot.Send(MSG_LIST_TAGS, &telebot.ReplyMarkup{
			InlineKeyboard: tagButtons,
		})
	}
}

func NextTask(ctx context.Context, d domain.UseCases) telebot.HandlerFunc {
	return func(ctxBot telebot.Context) error {
		tag := ctxBot.Data()

		uq, err := d.GetTask(GetContext(ctxBot), GetUserFromContext(ctxBot).TGUserID, tag)
		if err != nil {
			return err
		}

		q := uq.R.GetQuestion()
		tag = domain.escapeMarkdown(q.R.GetTag().TagService)
		questionText := domain.escapeMarkdown(q.QuestionService)

		label := "🔔"
		if uq.IsEdu {
			label = "💤"
		}

		keyboard := NewTaskButtonsBuilder().
			AddShowAnswer(uq.QuestionID).
			AddDifficulty(q.ID).
			AddActions(q.ID, label).
			Build()

		return ctxBot.Send(
			tag+": "+questionText,
			telebot.ModeMarkdownV2,
			keyboard,
		)
	}
}

// Новая функция для отображения вопроса после выбора "легко" или "сложно"
func showQuestionAfterChoice(domain domain.Apper) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		questionID, err := strconv.Atoi(ctx.Data())
		if err != nil {
			return err
		}

		q, err := domain.GetQuestionAnswers(GetContext(ctx), int64(questionID))
		if err != nil {
			return err
		}

		tag := domain.escapeMarkdown(q.R.GetTag().TagService)
		questionText := domain.escapeMarkdown(q.QuestionService)

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

func skipTask(domain domain.Apper) telebot.HandlerFunc {
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
		tag := domain.escapeMarkdown(q.R.GetTag().TagService)
		questionText := domain.escapeMarkdown(q.QuestionService)

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
