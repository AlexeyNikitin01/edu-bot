package ports

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"gopkg.in/telebot.v3"

	"bot/internal/app"
	"bot/internal/repo/edu"
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
		ticker := time.NewTicker(10 * time.Second)
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
			ch = make(chan *edu.UsersQuestion, 10) // буферизированный канал
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
				// ждем ответ — откладываем вопрос обратно
				d.mu.Unlock()
				time.Sleep(2 * time.Second)
				//go func() { ch <- uq }()
				continue
			}
			d.waitingForAnswer[userID] = true
			d.mu.Unlock()

			if err := d.sendPoll(userID, uq); err != nil {
				log.Printf("Ошибка отправки вопроса пользователю %d: %v", userID, err)

				d.mu.Lock()
				d.waitingForAnswer[userID] = false
				d.mu.Unlock()
			}
		}
	}
}

func (d *QuestionDispatcher) sendPoll(userID int64, uq *edu.UsersQuestion) error {
	options := make([]telebot.PollOption, len(uq.R.GetQuestion().R.GetAnswers()))
	correctIndex := -1

	for i, ans := range uq.R.GetQuestion().R.GetAnswers() {
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

	recipient := &telebot.User{ID: userID}
	msg, err := d.bot.Send(recipient, poll)
	if err != nil {
		return err
	}

	uq.PollID = null.StringFrom(msg.Poll.ID)
	uq.CorrectAnswer = null.Int64From(int64(correctIndex))
	if _, err = uq.Update(d.ctx, boil.GetContextDB(), boil.Whitelist(edu.UsersQuestionColumns.PollID)); err != nil {
		return err
	}

	return nil
}

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
}
