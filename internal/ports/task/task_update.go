package task

import (
	"bot/internal/domain"
	"bot/internal/middleware"
	"context"
	"fmt"
	"gopkg.in/telebot.v3"
	"strconv"
)

func UpdateTaskTotal(ctx context.Context, d domain.UseCases, correct bool) telebot.HandlerFunc {
	return func(ctxBot telebot.Context) error {
		userID := middleware.GetUserFromContext(ctxBot).TGUserID

		qidStr := ctxBot.Data()
		questionID, err := strconv.Atoi(qidStr)
		if err != nil {
			return ctxBot.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		uq, err := d.GetUserQuestion(ctx, userID, int64(questionID))
		if err != nil {
			return err
		}

		if err = d.UpdateRepeatTime(ctx, uq, correct); err != nil {
			return err
		}

		keyboard := NewTaskButtonsBuilder().
			AddNavigation(uq.R.GetQuestion().R.GetTag().Tag).
			Build()

		return ctxBot.Edit(EscapeMarkdown(
			uq.R.GetQuestion().R.GetTag().Tag+
				uq.R.GetQuestion().Question+": "+
				uq.R.GetQuestion().R.GetAnswers()[0].Answer), keyboard)
	}
}

func IsRepeatTask(ctx context.Context, d domain.UseCases) telebot.HandlerFunc {
	return func(ctxBot telebot.Context) error {
		qidStr := ctxBot.Data()
		questionID, err := strconv.Atoi(qidStr)
		if err != nil {
			return err
		}

		userID := middleware.GetUserFromContext(ctxBot).TGUserID

		if err = d.UpdateIsEduUserQuestion(ctx, userID, int64(questionID)); err != nil {
			return err
		}

		task, err := d.GetUserQuestion(ctx, userID, int64(questionID))
		if err != nil {
			return err
		}

		q := task.R.GetQuestion()
		tag := EscapeMarkdown(q.R.GetTag().Tag)
		questionText := EscapeMarkdown(q.Question)

		keyboard := NewTaskButtonsBuilder().
			AddShowAnswer(task.QuestionID, true).
			AddDifficulty(q.ID).
			AddActions(q.ID, task.IsEdu).
			Build()

		message := fmt.Sprintf(MsgTagQuestion, tag, questionText)
		return ctxBot.Edit(message, telebot.ModeMarkdownV2, keyboard)
	}
}
