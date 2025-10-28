package question

import (
	"bot/internal/middleware"
	"context"
	"errors"
	"gopkg.in/telebot.v3"
	"log"
	"strconv"
	"strings"

	"bot/internal/domain"
)

func DeleteQuestion(ctx context.Context, d domain.UseCases) telebot.HandlerFunc {
	return func(ctxBot telebot.Context) error {
		// Разбираем данные callback: "questionID_page_tag"
		parts := strings.Split(ctxBot.Data(), "_")
		if len(parts) < 3 {
			return errors.New("invalid question")
		}

		questionID, err := strconv.Atoi(parts[0])
		if err != nil {
			return err
		}

		page, err := strconv.Atoi(parts[1])
		if err != nil {
			return err
		}

		tag := strings.Join(parts[2:], "_")
		userID := middleware.GetUserFromContext(ctxBot).TGUserID

		// Удаляем вопрос
		if err = d.DeleteQuestion(ctx, int64(questionID)); err != nil {
			return err
		}

		// Показываем обновленный список вопросов
		return showQuestionsPage(ctx, ctxBot, tag, page, userID, d, 0)
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
