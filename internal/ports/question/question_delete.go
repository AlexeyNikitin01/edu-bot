package question

import (
	"bot/internal/middleware"
	"context"
	"gopkg.in/telebot.v3"
	"log"
	"strconv"

	"bot/internal/domain"
)

func DeleteQuestion(ctx context.Context, domain domain.UseCases) telebot.HandlerFunc {
	return func(ctxBot telebot.Context) error {
		return nil
	}
}

func DeleteQuestionByTag(ctx context.Context, d domain.UseCases) telebot.HandlerFunc {
	return func(ctxBot telebot.Context) error {
		tag := ctxBot.Data()
		userID := middleware.GetUserFromContext(ctxBot).TGUserID

		if err := d.DeleteQuestionsByTag(ctx, userID, tag); err != nil {
			return err
		}

		return nil
	}
}

func DeleteQuestionAfterPoll(
	ctx context.Context, d domain.UseCases,
) telebot.HandlerFunc {
	return func(ctxBot telebot.Context) error {
		userID := middleware.GetUserFromContext(ctxBot).TGUserID

		qidStr := ctxBot.Data()
		questionID, err := strconv.Atoi(qidStr)
		if err != nil {
			return err
		}

		if err = d.DeleteQuestionUser(ctx, userID, int64(questionID)); err != nil {
			return ctxBot.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		if err = ctxBot.Delete(); err != nil {
			return err
		}

		if err = ctxBot.Send(MSG_SUCESS_DELETE_QUESTION); err != nil {
			return err
		}

		if err = d.SetUserWaiting(ctx, userID, false); err != nil {
			log.Printf("Ошибка сброса статуса waiting в Redis для пользователя %d: %v", userID, err)
		}

		return nil
	}
}

func DeleteQuestionAfterPollHigh(
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

		if err = ctxBot.Send(MSG_SUCESS_DELETE_QUESTION); err != nil {
			return err
		}

		if err = d.SetUserWaiting(ctx, userID, false); err != nil {
			log.Printf("Ошибка сброса статуса waiting в Redis для пользователя %d: %v", userID, err)
		}

		return nil
	}
}
