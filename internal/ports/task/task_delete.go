package task

import (
	"bot/internal/domain"
	"bot/internal/middleware"
	"context"
	"gopkg.in/telebot.v3"
	"strconv"
)

func DeleteTask(
	ctx context.Context, d domain.UseCases,
) telebot.HandlerFunc {
	return func(ctxBot telebot.Context) error {
		userID := middleware.GetUserFromContext(ctxBot).TGUserID

		qidStr := ctxBot.Data()
		questionID, err := strconv.Atoi(qidStr)
		if err != nil {
			return ctxBot.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		if err = d.DeleteQuestionUser(ctx, userID, int64(questionID)); err != nil {
			return ctxBot.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		if err = ctxBot.Delete(); err != nil {
			return err
		}

		keyboard := NewTaskButtonsBuilder().
			AddNavigation(int64(questionID)).
			Build()

		message := MSG_SUCESS_DELETE_QUESTION + "\n\n"
		return ctxBot.Send(message, telebot.ModeMarkdownV2, keyboard)
	}
}
