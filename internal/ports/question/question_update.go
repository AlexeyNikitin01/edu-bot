package question

import (
	"bot/internal/domain"
	"bot/internal/middleware"
	"bot/internal/repo/edu"
	"context"
	"fmt"
	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"gopkg.in/telebot.v3"
	"log"
	"strconv"
	"time"
)

func ForgotQuestion(ctx context.Context, d domain.UseCases) telebot.HandlerFunc {
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

		if err = d.UpdateRepeatTime(ctx, uq, false); err != nil {
			return err
		}

		if err = ctxBot.Edit(&telebot.ReplyMarkup{
			InlineKeyboard: [][]telebot.InlineButton{{NewQuestionButtonBuilder().BuildForgotButton(uq)}},
		}); err != nil {
			return ctxBot.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		if err = d.SetUserWaiting(ctx, userID, false); err != nil {
			log.Printf("Ошибка сброса статуса waiting в Redis для пользователя %d: %v", userID, err)
		}

		return nil
	}
}

func RememberQuestion(ctx context.Context, d domain.UseCases) telebot.HandlerFunc {
	return func(ctxBot telebot.Context) error {
		qidStr := ctxBot.Data()
		questionID, err := strconv.Atoi(qidStr)
		if err != nil {
			return err
		}

		userID := middleware.GetUserFromContext(ctxBot).TGUserID

		uq, err := d.GetUserQuestion(ctx, userID, int64(questionID))
		if err != nil {
			return err
		}

		if err = d.UpdateRepeatTime(ctx, uq, true); err != nil {
			return err
		}

		easy := telebot.InlineButton{
			Text: "✅ " + MSG_REMEMBER + " " + timeLeftMsg(uq.TimeRepeat.Sub(time.Now().UTC())),
			Data: fmt.Sprintf("%d", questionID),
		}

		if err = ctxBot.Edit(&telebot.ReplyMarkup{
			InlineKeyboard: [][]telebot.InlineButton{{easy}},
		}); err != nil {
			return ctxBot.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		t, err := d.GetNearestTimeRepeat(ctx, userID)
		if err != nil {
			return ctxBot.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		now := time.Now().UTC()
		if !now.Add(time.Minute*10).After(t) && !uq.R.GetQuestion().IsTask {
			duration := t.Sub(now)

			msg := fmt.Sprintf("⏳ Следующий вопрос будет доступен через: %s", timeLeftMsg(duration))

			if err = ctxBot.Send(msg, telebot.ModeMarkdown); err != nil {
				return ctxBot.Respond(&telebot.CallbackResponse{Text: err.Error()})
			}
		}

		// Сбрасываем флаг ожидания ответа в Redis
		if err = d.SetUserWaiting(ctx, userID, false); err != nil {
			log.Printf("Ошибка сброса статуса waiting в Redis для пользователя %d: %v", userID, err)
		}

		return nil
	}
}

func RepeatQuestionAfterPoll(ctx context.Context, d domain.UseCases) telebot.HandlerFunc {
	return func(ctxBot telebot.Context) error {
		qidStr := ctxBot.Data() // получаем questionID из callback data
		questionID, err := strconv.Atoi(qidStr)
		if err != nil {
			return ctxBot.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		userID := middleware.GetUserFromContext(ctxBot).TGUserID

		if err = d.UpdateIsEduUserQuestion(ctx, userID, int64(questionID)); err != nil {
			return ctxBot.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		if err = ctxBot.Edit(&telebot.ReplyMarkup{
			InlineKeyboard: [][]telebot.InlineButton{
				// todo кнопки
			},
		}); err != nil {
			return ctxBot.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		return nil
	}
}

func RepeatQuestionAfterPollHigh(ctx context.Context, d domain.UseCases) telebot.HandlerFunc {
	return func(ctxBot telebot.Context) error {
		qidStr := ctxBot.Data()
		questionID, err := strconv.Atoi(qidStr)
		if err != nil {
			return ctxBot.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		userID := middleware.GetUserFromContext(ctxBot).TGUserID

		if err = d.UpdateIsEduUserQuestion(ctx, userID, int64(questionID)); err != nil {
			return ctxBot.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		if err = ctxBot.Edit(&telebot.ReplyMarkup{ // todo поправить кнопки
		},
		); err != nil {
			return ctxBot.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		return nil
	}
}

func CheckPollAnswer(ctx context.Context, d domain.UseCases) telebot.HandlerFunc {
	return func(ctxBot telebot.Context) error {
		poll := ctxBot.PollAnswer()
		userID := poll.Sender.ID

		log.Printf("Ответ от пользователя %d получен", userID)

		uq, err := edu.UsersQuestions(
			edu.UsersQuestionWhere.PollID.EQ(null.StringFrom(poll.PollID)),
		).One(ctx, boil.GetContextDB())
		if err != nil {
			return err
		}

		correct := int(uq.CorrectAnswer.Int64) == poll.Options[0]

		if err = d.UpdateRepeatTime(ctx, uq, correct); err != nil {
			return err
		}

		if err = d.SetUserWaiting(ctx, userID, false); err != nil {
			log.Printf("Ошибка сброса статуса waiting в Redis для пользователя %d: %v", userID, err)
		}

		return nil
	}
}

func PauseTag(ctx context.Context, d domain.UseCases) telebot.HandlerFunc {
	return func(ctxBot telebot.Context) error {
		tagIDStr := ctxBot.Data()
		tagID, err := strconv.Atoi(tagIDStr)
		if err != nil {
			return ctxBot.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		tag, err := edu.Tags(
			edu.TagWhere.ID.EQ(int64(tagID)),
		).One(ctx, boil.GetContextDB())
		if err != nil {
			return ctxBot.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		tag.IsPause = !tag.IsPause
		if _, err = tag.Update(ctx, boil.GetContextDB(), boil.Whitelist(
			edu.TagColumns.IsPause,
		)); err != nil {
			return ctxBot.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		tagButtons, err := getButtonsTags(ctxBot, d)
		if err != nil {
			return err
		}

		return ctxBot.Edit(MSG_LIST_TAGS, &telebot.ReplyMarkup{
			InlineKeyboard: tagButtons,
		})
	}
}
