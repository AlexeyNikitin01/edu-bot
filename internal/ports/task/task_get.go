package task

import (
	"bot/internal/domain"
	"bot/internal/middleware"
	"bot/internal/ports/question"
	"context"
	"gopkg.in/telebot.v3"
	"strconv"
)

func GetTagsByTask(ctx context.Context, d domain.UseCases) telebot.HandlerFunc {
	return func(ctxBot telebot.Context) error {
		userID := middleware.GetUserFromContext(ctxBot).TGUserID

		tags, err := d.GetUniqueTagsByTask(ctx, userID)
		if err != nil {
			return err
		}

		if len(tags) == 0 {
			return err
		}

		var tagButtons [][]telebot.InlineButton

		for _, tag := range tags {
			tagBtn := telebot.InlineButton{
				Unique: question.INLINE_BTN_TASK_BY_TAG,
				Text:   tag.Tag,
				Data:   tag.Tag,
			}

			tagButtons = append(tagButtons, []telebot.InlineButton{tagBtn})
		}

		return ctxBot.Send(question.MSG_LIST_TAGS, &telebot.ReplyMarkup{
			InlineKeyboard: tagButtons,
		})
	}
}

func NextTask(ctx context.Context, d domain.UseCases) telebot.HandlerFunc {
	return func(ctxBot telebot.Context) error {
		tag := ctxBot.Data()
		userID := middleware.GetUserFromContext(ctxBot).TGUserID
		uq, err := d.GetTask(ctx, userID, tag)
		if err != nil {
			return err
		}

		q := uq.R.GetQuestion()
		tag = question.EscapeMarkdown(q.R.GetTag().Tag)
		questionText := question.EscapeMarkdown(q.Question)

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
func showQuestionAfterChoice(ctx context.Context, d domain.UseCases) telebot.HandlerFunc {
	return func(ctxBot telebot.Context) error {
		questionID, err := strconv.Atoi(ctxBot.Data())
		if err != nil {
			return err
		}

		q, err := d.GetQuestionAnswers(ctx, int64(questionID))
		if err != nil {
			return err
		}

		tag := question.EscapeMarkdown(q.R.GetTag().Tag)
		questionText := question.EscapeMarkdown(q.Question)

		// Используем билдер для создания кнопок навигации
		keyboard := NewTaskButtonsBuilder().
			AddNavigation(int64(questionID)).
			Build()

		return ctxBot.Send(
			"Выбор сохранён!\n\n"+tag+": "+questionText,
			telebot.ModeMarkdownV2,
			keyboard,
		)
	}
}

func skipTask(ctx context.Context, domain domain.UseCases) telebot.HandlerFunc {
	return func(ctxBot telebot.Context) error {
		tagData := ctxBot.Data()

		userID := middleware.GetUserFromContext(ctxBot).TGUserID

		uq, err := domain.GetTask(ctx, userID, tagData)
		if err != nil {
			return err
		}

		if uq == nil {
			return ctxBot.Send("🎉 Все вопросы завершены! Вы великолепны!")
		}

		q := uq.R.GetQuestion()
		tag := question.EscapeMarkdown(q.R.GetTag().Tag)
		questionText := question.EscapeMarkdown(q.Question)

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

		return ctxBot.Send(
			"⏩ Вопрос пропущен!\n\n"+tag+": "+questionText,
			telebot.ModeMarkdownV2,
			keyboard,
		)
	}
}
