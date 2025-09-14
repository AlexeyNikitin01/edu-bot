package ports

import (
	"fmt"
	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"log"
	"strconv"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
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
			Text: "🔴 " + MSG_FORGOT,
			Data: fmt.Sprintf("%d", questionID),
		}

		if err = ctx.Edit(&telebot.ReplyMarkup{
			InlineKeyboard: [][]telebot.InlineButton{{forgot}},
		}); err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		dispatcher.mu.Lock()
		dispatcher.waitingForAnswer[GetUserFromContext(ctx).TGUserID] = false
		dispatcher.mu.Unlock()

		return nil
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
			qm.Load(qm.Rels(edu.UsersQuestionRels.Question)),
		).One(GetContext(ctx), boil.GetContextDB())
		if err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		if err = domain.UpdateRepeatTime(GetContext(ctx), uq, true); err != nil {
			return err
		}

		easy := telebot.InlineButton{
			Text: "✅ " + MSG_REMEMBER + " " + timeLeftMsg(uq.TimeRepeat.Sub(time.Now().UTC())),
			Data: fmt.Sprintf("%d", questionID),
		}

		if err = ctx.Edit(&telebot.ReplyMarkup{
			InlineKeyboard: [][]telebot.InlineButton{{easy}},
		}); err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		// смотрим через сколько будет следующий вопрос, если не будет ближайшие 10 мин, то выведем, через сколько
		user := GetUserFromContext(ctx)

		t, err := dispatcher.domain.GetNearestTimeRepeat(GetContext(ctx), user.TGUserID)
		if err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		now := time.Now().UTC()
		if !now.Add(time.Minute*10).After(t) && !uq.R.GetQuestion().IsTask {
			duration := t.Sub(now)

			msg := fmt.Sprintf("⏳ Следующий вопрос будет доступен через: %s", timeLeftMsg(duration))

			if err = ctx.Send(msg, telebot.ModeMarkdown); err != nil {
				return ctx.Respond(&telebot.CallbackResponse{Text: err.Error()})
			}
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
				0,
				"",
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
				"",
				INLINE_NAME_DELETE_AFTER_POLL,
				INLINE_BTN_DELETE_QUESTION_AFTER_POLL_HIGH,
				-1,
				"",
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

func pauseTag(domain app.Apper) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		tagIDStr := ctx.Data()
		tagID, err := strconv.Atoi(tagIDStr)
		if err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		tag, err := edu.Tags(
			edu.TagWhere.ID.EQ(int64(tagID)),
		).One(GetContext(ctx), boil.GetContextDB())
		if err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		tag.IsPause = !tag.IsPause
		if _, err = tag.Update(GetContext(ctx), boil.GetContextDB(), boil.Whitelist(
			edu.TagColumns.IsPause,
		)); err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		tagButtons, err := getButtonsTags(ctx, domain)
		if err != nil {
			return err
		}

		return ctx.Edit(MSG_LIST_TAGS, &telebot.ReplyMarkup{
			InlineKeyboard: tagButtons,
		})
	}
}
