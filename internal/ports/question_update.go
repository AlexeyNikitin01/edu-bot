package ports

import (
	"fmt"
	"log"
	"strconv"

	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"gopkg.in/telebot.v3"

	"bot/internal/app"
	"bot/internal/repo/edu"
)

func forgotQuestion(domain *app.App, dispatcher *QuestionDispatcher) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		qidStr := ctx.Data()
		questionID, err := strconv.Atoi(qidStr)
		if err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		uq, err := edu.UsersQuestions(
			edu.UsersQuestionWhere.UserID.EQ(GetUserFromContext(ctx).TGUserID),
			edu.UsersQuestionWhere.QuestionID.EQ(int64(questionID)),
		).One(GetContext(ctx), boil.GetContextDB())
		if err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		if err = domain.UpdateRepeatTime(GetContext(ctx), uq, false); err != nil {
			return err
		}

		forgot := telebot.InlineButton{
			Unique: INLINE_FORGOT_HIGH_QUESTION,
			Text:   "🔴 " + MSG_FORGOT,
			Data:   fmt.Sprintf("%d", questionID),
		}

		easy := telebot.InlineButton{
			Unique: INLINE_REMEMBER_HIGH_QUESTION,
			Text:   MSG_REMEMBER,
			Data:   fmt.Sprintf("%d", questionID),
		}

		if err = ctx.Edit(&telebot.ReplyMarkup{
			InlineKeyboard: [][]telebot.InlineButton{{easy, forgot}},
		}); err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		dispatcher.mu.Lock()
		dispatcher.waitingForAnswer[GetUserFromContext(ctx).TGUserID] = false
		dispatcher.mu.Unlock()

		return ctx.Send(MSG_RESET_QUESTION)
	}
}

func rememberQuestion(domain *app.App, dispatcher *QuestionDispatcher) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		qidStr := ctx.Data()
		questionID, err := strconv.Atoi(qidStr)
		if err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		uq, err := edu.UsersQuestions(
			edu.UsersQuestionWhere.UserID.EQ(GetUserFromContext(ctx).TGUserID),
			edu.UsersQuestionWhere.QuestionID.EQ(int64(questionID)),
		).One(GetContext(ctx), boil.GetContextDB())
		if err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		if err = domain.UpdateRepeatTime(GetContext(ctx), uq, true); err != nil {
			return err
		}

		forgot := telebot.InlineButton{
			Unique: INLINE_FORGOT_HIGH_QUESTION,
			Text:   MSG_FORGOT,
			Data:   fmt.Sprintf("%d", questionID),
		}

		easy := telebot.InlineButton{
			Unique: INLINE_REMEMBER_HIGH_QUESTION,
			Text:   "✅ " + MSG_REMEMBER,
			Data:   fmt.Sprintf("%d", questionID),
		}

		if err = ctx.Edit(&telebot.ReplyMarkup{
			InlineKeyboard: [][]telebot.InlineButton{{easy, forgot}},
		}); err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		if err = ctx.Send(MSG_INC_SERIAL_QUESTION); err != nil {
			return err
		}

		dispatcher.mu.Lock()
		dispatcher.waitingForAnswer[GetUserFromContext(ctx).TGUserID] = false
		dispatcher.mu.Unlock()

		return nil
	}
}

func repeatQuestionAfterPoll(domain *app.App) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		qidStr := ctx.Data() // получаем questionID из callback data
		questionID, err := strconv.Atoi(qidStr)
		if err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		if err = domain.UpdateIsEduUserQuestion(GetContext(ctx), GetUserFromContext(ctx).TGUserID, int64(questionID)); err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		if err = ctx.Edit(&telebot.ReplyMarkup{
			InlineKeyboard: [][]telebot.InlineButton{getQuestionBtn(
				ctx,
				int64(questionID),
				INLINE_BTN_REPEAT_QUESTION_AFTER_POLL,
				INLINE_NAME_REPEAT_AFTER_POLL,
				INLINE_NAME_DELETE_AFTER_POLL,
				INLINE_BTN_DELETE_QUESTION_AFTER_POLL,
			)},
		}); err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		return nil
	}
}

func repeatQuestionAfterPollHigh(domain *app.App) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		qidStr := ctx.Data() // получаем questionID из callback data
		questionID, err := strconv.Atoi(qidStr)
		if err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		if err = domain.UpdateIsEduUserQuestion(GetContext(ctx), GetUserFromContext(ctx).TGUserID, int64(questionID)); err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		forgot := telebot.InlineButton{
			Unique: INLINE_FORGOT_HIGH_QUESTION,
			Text:   MSG_FORGOT,
			Data:   fmt.Sprintf("%d", questionID),
		}

		easy := telebot.InlineButton{
			Unique: INLINE_REMEMBER_HIGH_QUESTION,
			Text:   MSG_REMEMBER,
			Data:   fmt.Sprintf("%d", questionID),
		}

		if err = ctx.Edit(&telebot.ReplyMarkup{
			InlineKeyboard: [][]telebot.InlineButton{{easy, forgot}, getQuestionBtn(
				ctx,
				int64(questionID),
				INLINE_BTN_REPEAT_QUESTION_AFTER_POLL_HIGH,
				INLINE_NAME_REPEAT_AFTER_POLL,
				INLINE_NAME_DELETE_AFTER_POLL,
				INLINE_BTN_DELETE_QUESTION_AFTER_POLL_HIGH,
			)},
		}); err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		return nil
	}
}

func checkPollAnswer(domain *app.App, dispatcher *QuestionDispatcher) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		poll := c.PollAnswer()
		userID := poll.Sender.ID

		log.Printf("Ответ от пользователя %d получен", userID)

		uq, err := edu.UsersQuestions(edu.UsersQuestionWhere.PollID.EQ(null.StringFrom(poll.PollID))).
			One(GetContext(c), boil.GetContextDB())
		if err != nil {
			return err
		}

		correct := int(uq.CorrectAnswer.Int64) == poll.Options[0]

		if err = domain.UpdateRepeatTime(GetContext(c), uq, correct); err != nil {
			return err
		}

		dispatcher.mu.Lock()
		dispatcher.waitingForAnswer[userID] = false
		dispatcher.mu.Unlock()

		return nil
	}
}
