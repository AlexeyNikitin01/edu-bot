package ports

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"gopkg.in/telebot.v3"

	"bot/internal/app"
	"bot/internal/repo/edu"
)

const (
	MSG_FORGOT              = "СЛОЖНО"
	MSG_REMEMBER            = "ЛЕГКО"
	MSG_INC_SERIAL_QUESTION = "Отлично, вопрос будет реже вам попадаться🤗🤗🤗"
	MSG_RESET_QUESTION      = "Ничего страшного, вопрос снова повториться в скором времени👈🤝🕕"
)

type QuestionDispatcher struct {
	mu               sync.Mutex
	workers          map[int64]chan *edu.UsersQuestion
	waitingForAnswer map[int64]bool
	domain           app.Apper
	bot              *telebot.Bot
	ctx              context.Context
}

func NewDispatcher(ctx context.Context, domain app.Apper, bot *telebot.Bot) *QuestionDispatcher {
	return &QuestionDispatcher{
		mu:               sync.Mutex{},
		workers:          make(map[int64]chan *edu.UsersQuestion),
		waitingForAnswer: make(map[int64]bool),
		domain:           domain,
		bot:              bot,
		ctx:              ctx,
	}
}

func (d *QuestionDispatcher) StartPollingLoop() {
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-d.ctx.Done():
				return
			case <-ticker.C:
				d.checkAndDispatch()
			}
		}
	}()
}

func (d *QuestionDispatcher) checkAndDispatch() {
	users, err := edu.Users().All(d.ctx, boil.GetContextDB())
	if err != nil {
		log.Println("Ошибка получения пользователей:", err)
		return
	}

	for _, user := range users {
		userID := user.TGUserID

		d.mu.Lock()
		ch, exists := d.workers[userID]
		if !exists {
			ch = make(chan *edu.UsersQuestion, 1) // буферизированный канал
			d.workers[userID] = ch
			go d.worker(userID, ch)
		}
		d.mu.Unlock()

		uqs, err := d.domain.GetQuestionsAnswers(d.ctx, userID)
		if err != nil || len(uqs) == 0 {
			continue
		}

		for _, q := range uqs {
			select {
			case ch <- q:
			default:
				log.Printf("Очередь переполнена для пользователя %d", userID)
			}
		}
	}
}

func (d *QuestionDispatcher) worker(userID int64, ch chan *edu.UsersQuestion) {
	for {
		select {
		case <-d.ctx.Done():
			return
		case uq := <-ch:
			d.mu.Lock()
			if d.waitingForAnswer[userID] {
				// ждем ответ
				d.mu.Unlock()
				time.Sleep(2 * time.Second)
				continue
			}
			d.waitingForAnswer[userID] = true
			d.mu.Unlock()

			if err := d.sendQuestion(userID, uq); err != nil {
				log.Printf("Ошибка отправки вопроса пользователю %d: %v", userID, err)

				d.mu.Lock()
				d.waitingForAnswer[userID] = false
				d.mu.Unlock()
			}
		}
	}
}

func (d *QuestionDispatcher) sendQuestion(userID int64, uq *edu.UsersQuestion) error {
	answers := uq.R.GetQuestion().R.GetAnswers()

	if len(answers) == 1 || uq.TotalSerial > 4 {
		return d.questionWithHigh(userID, uq.R.GetQuestion(), answers[0])
	}

	return d.questionWithTest(userID, uq)
}

func (d *QuestionDispatcher) questionWithHigh(id int64, q *edu.Question, answer *edu.Answer) error {
	forgot := telebot.InlineButton{
		Unique: INLINE_FORGOT_HIGH_QUESTION,
		Text:   MSG_FORGOT,
		Data:   fmt.Sprintf("%d", q.ID),
	}

	easy := telebot.InlineButton{
		Unique: INLINE_REMEMBER_HIGH_QUESTION,
		Text:   MSG_REMEMBER,
		Data:   fmt.Sprintf("%d", q.ID),
	}

	rec := &telebot.User{ID: id}
	_, err := d.bot.Send(
		rec,
		fmt.Sprintf("%s \n\n || %s ||", q.Question, answer.Answer),
		telebot.ModeMarkdownV2,
		&telebot.ReplyMarkup{
			InlineKeyboard: [][]telebot.InlineButton{{easy, forgot}},
		},
	)
	return err
}

func (d *QuestionDispatcher) questionWithTest(userID int64, uq *edu.UsersQuestion) error {
	answers := uq.R.GetQuestion().R.GetAnswers()

	shuffled := make([]*edu.Answer, len(answers))
	copy(shuffled, answers)

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	options := make([]telebot.PollOption, len(shuffled))
	correctIndex := -1

	for i, ans := range shuffled {
		options[i] = telebot.PollOption{Text: ans.Answer}
		if ans.IsCorrect {
			correctIndex = i
		}
	}

	poll := &telebot.Poll{
		Question:        uq.R.GetQuestion().Question,
		Options:         options,
		Type:            telebot.PollQuiz,
		CorrectOption:   correctIndex,
		Anonymous:       false,
		Explanation:     "Правильный ответ будет показан после выбора",
		MultipleAnswers: false,
	}

	label := "☑️"
	if uq.IsEdu {
		label = "✅"
	}

	repeatBtn := telebot.InlineButton{
		Unique: INLINE_BTN_REPEAT_QUESTION_AFTER_POLL,
		Text:   label + INLINE_NAME_REPEAT_AFTER_POLL,
		Data:   fmt.Sprintf("%d", uq.QuestionID),
	}

	deleteBtn := telebot.InlineButton{
		Unique: INLINE_BTN_DELETE_QUESTION_AFTER_POLL,
		Text:   INLINE_NAME_DELETE_AFTER_POLL,
		Data:   fmt.Sprintf("%d", uq.QuestionID),
	}

	recipient := &telebot.User{ID: userID}
	msg, err := d.bot.Send(recipient, poll, &telebot.ReplyMarkup{
		InlineKeyboard: [][]telebot.InlineButton{{repeatBtn, deleteBtn}},
	})
	if err != nil {
		return err
	}

	uq.PollID = null.StringFrom(msg.Poll.ID)
	uq.CorrectAnswer = null.Int64From(int64(correctIndex))
	if _, err = uq.Update(d.ctx, boil.GetContextDB(),
		boil.Whitelist(edu.UsersQuestionColumns.PollID, edu.UsersQuestionColumns.CorrectAnswer)); err != nil {
		return err
	}

	return nil
}

// RegisterPollAnswerHandler todo: отрефакторить
func (d *QuestionDispatcher) RegisterPollAnswerHandler() {
	d.bot.Handle(telebot.OnPollAnswer, func(c telebot.Context) error {
		poll := c.PollAnswer()
		userID := poll.Sender.ID

		log.Printf("Ответ от пользователя %d получен", userID)

		uq, err := edu.UsersQuestions(edu.UsersQuestionWhere.PollID.EQ(null.StringFrom(poll.PollID))).
			One(d.ctx, boil.GetContextDB())
		if err != nil {
			return err
		}

		correct := int(uq.CorrectAnswer.Int64) == poll.Options[0]

		if err = d.domain.UpdateRepeatTime(d.ctx, uq, correct); err != nil {
			return err
		}

		d.mu.Lock()
		d.waitingForAnswer[userID] = false
		d.mu.Unlock()

		return nil
	})

	d.bot.Handle(&telebot.InlineButton{Unique: INLINE_FORGOT_HIGH_QUESTION}, func(ctx telebot.Context) error {
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

		if err = d.domain.UpdateRepeatTime(GetContext(ctx), uq, false); err != nil {
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

		d.mu.Lock()
		d.waitingForAnswer[GetUserFromContext(ctx).TGUserID] = false
		d.mu.Unlock()

		return ctx.Send(MSG_RESET_QUESTION)
	})

	d.bot.Handle(&telebot.InlineButton{Unique: INLINE_REMEMBER_HIGH_QUESTION}, func(ctx telebot.Context) error {
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

		if err = d.domain.UpdateRepeatTime(GetContext(ctx), uq, true); err != nil {
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

		d.mu.Lock()
		d.waitingForAnswer[GetUserFromContext(ctx).TGUserID] = false
		d.mu.Unlock()

		return nil
	})

	d.bot.Handle(&telebot.InlineButton{Unique: INLINE_BTN_REPEAT_QUESTION_AFTER_POLL}, func(ctx telebot.Context) error {
		qidStr := ctx.Data() // получаем questionID из callback data
		questionID, err := strconv.Atoi(qidStr)
		if err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		if err = d.domain.UpdateIsEduUserQuestion(GetContext(ctx), GetUserFromContext(ctx).TGUserID, int64(questionID)); err != nil {
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
	})
	d.bot.Handle(&telebot.InlineButton{Unique: INLINE_BTN_DELETE_QUESTION_AFTER_POLL}, func(ctx telebot.Context) error {
		qidStr := ctx.Data()
		questionID, err := strconv.Atoi(qidStr)
		if err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		_, err = edu.UsersQuestions(
			edu.UsersQuestionWhere.UserID.EQ(GetUserFromContext(ctx).TGUserID),
			edu.UsersQuestionWhere.QuestionID.EQ(int64(questionID)),
		).DeleteAll(GetContext(ctx), boil.GetContextDB(), false)
		if err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		if err = ctx.Delete(); err != nil {
			return ctx.Send(err.Error())
		}

		if err = ctx.Send(MSG_SUCESS_DELETE_QUESTION); err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		d.mu.Lock()
		d.waitingForAnswer[GetUserFromContext(ctx).TGUserID] = false
		d.mu.Unlock()

		return nil
	})
}
