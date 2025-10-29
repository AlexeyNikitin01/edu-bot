package task

import (
	"bot/internal/domain"
	"bot/internal/middleware"
	"bot/internal/ports/question"
	"bot/internal/ports/tags"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"gopkg.in/telebot.v3"
	"strconv"
)

func ViewAnswerTask(ctx context.Context, d domain.UseCases, showAnswer bool) telebot.HandlerFunc {
	return func(ctxBot telebot.Context) error {
		data := ctxBot.Data()

		questionID, err := strconv.Atoi(data)
		if err != nil {
			return err
		}

		userID := middleware.GetUserFromContext(ctxBot).TGUserID

		uq, err := d.GetUserQuestion(ctx, userID, int64(questionID))
		if err != nil {
			return err
		}

		q := uq.GetQuestion()
		tagName := uq.R.GetQuestion().R.GetTag().Tag
		answer := uq.R.GetQuestion().R.GetAnswers()[0]

		message := EscapeMarkdown(tagName) + ": " + EscapeMarkdown(q.Question)
		if showAnswer {
			message += "\n\n" + EscapeMarkdown(answer.Answer)
		}

		keyboard := NewTaskButtonsBuilder().
			AddShowAnswer(uq.QuestionID, !showAnswer).
			AddDifficulty(q.ID).
			AddActions(q.ID, uq.IsEdu).
			Build()

		return ctxBot.Edit(message, telebot.ModeMarkdownV2, keyboard)
	}
}

func GetTagsByTask(ctx context.Context, d domain.UseCases) telebot.HandlerFunc {
	return func(ctxBot telebot.Context) error {
		userID := middleware.GetUserFromContext(ctxBot).TGUserID

		ts, err := d.GetUniqueTagsByTask(ctx, userID)
		if err != nil {
			return err
		}

		if len(ts) == 0 {
			return ctxBot.Send(MsgNoTagsAvailable)
		}

		var tagButtons [][]telebot.InlineButton

		for _, tag := range ts {
			tagBtn := telebot.InlineButton{
				Unique: question.INLINE_BTN_TASK_BY_TAG,
				Text:   tag.Tag,
				Data:   tag.Tag,
			}

			tagButtons = append(tagButtons, []telebot.InlineButton{tagBtn})
		}

		return ctxBot.Send(tags.MSG_LIST_TAGS, &telebot.ReplyMarkup{
			InlineKeyboard: tagButtons,
		})
	}
}

func NextTask(ctx context.Context, d domain.UseCases) telebot.HandlerFunc {
	return func(ctxBot telebot.Context) error {
		tag := ctxBot.Data()
		userID := middleware.GetUserFromContext(ctxBot).TGUserID
		uq, err := d.GetTask(ctx, userID, tag)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		} else if errors.Is(err, sql.ErrNoRows) {
			return ctxBot.Send(MsgAllTasksCompleted)
		}

		q := uq.R.GetQuestion()
		tag = question.EscapeMarkdown(q.R.GetTag().Tag)
		questionText := question.EscapeMarkdown(q.Question)

		keyboard := NewTaskButtonsBuilder().
			AddShowAnswer(uq.QuestionID, true).
			AddDifficulty(q.ID).
			AddActions(q.ID, uq.IsEdu).
			Build()

		message := fmt.Sprintf(MsgTagQuestion, tag, questionText)
		return ctxBot.Send(message, telebot.ModeMarkdownV2, keyboard)
	}
}

func SkipTask(ctx context.Context, d domain.UseCases) telebot.HandlerFunc {
	return func(ctxBot telebot.Context) error {
		qData := ctxBot.Data()
		qID, _ := strconv.Atoi(qData)

		userID := middleware.GetUserFromContext(ctxBot).TGUserID

		q, err := d.GetQuestionAnswers(ctx, int64(qID))
		if err != nil {
			return err
		}

		if q == nil {
			return ctxBot.Send(MsgAllTasksCompleted)
		}

		task, err := d.GetTask(ctx, userID, q.R.GetTag().Tag, int64(qID))
		if err != nil {
			return ctxBot.Edit(MsgAllTasksCompleted)
		}

		q = task.R.GetQuestion()
		tag := question.EscapeMarkdown(q.R.GetTag().Tag)
		questionText := question.EscapeMarkdown(q.Question)

		keyboard := NewTaskButtonsBuilder().
			AddShowAnswer(task.QuestionID, true).
			AddDifficulty(q.ID).
			AddActions(q.ID, task.IsEdu).
			Build()

		message := fmt.Sprintf(MsgTagQuestion, tag, questionText)
		return ctxBot.Edit(message, telebot.ModeMarkdownV2, keyboard)
	}
}
