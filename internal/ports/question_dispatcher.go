package ports

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"
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
	MSG_NEXT_QUESTION       = "😎"
	MSG_NEXT_TIME_QUESTION  = "⏳ Следующий вопрос будет доступен "
	MSG_WRONG               = "Нет правильного ответа для вопроса"
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
	users, err := edu.Users(
		edu.UserWhere.Block.EQ(false),
	).All(d.ctx, boil.GetContextDB())
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
		for _, answer := range answers {
			if answer.IsCorrect {
				return d.questionWithHigh(userID, uq, uq.R.GetQuestion(), answers[0])
			}
		}
		_, err := d.bot.Send(&telebot.User{ID: userID}, MSG_WRONG)
		return err
	}

	return d.questionWithTest(userID, uq)
}
func (d *QuestionDispatcher) questionWithHigh(
	id int64, uq *edu.UsersQuestion, q *edu.Question, answer *edu.Answer,
) error {
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

	label := "☑️"
	if uq.IsEdu {
		label = "✅"
	}

	repeatBtn := telebot.InlineButton{
		Unique: INLINE_BTN_REPEAT_QUESTION_AFTER_POLL_HIGH,
		Text:   label,
		Data:   fmt.Sprintf("%d", uq.QuestionID),
	}

	deleteBtn := telebot.InlineButton{
		Unique: INLINE_BTN_DELETE_QUESTION_AFTER_POLL_HIGH,
		Text:   INLINE_NAME_DELETE_AFTER_POLL,
		Data:   fmt.Sprintf("%d", uq.QuestionID),
	}

	editBtn := telebot.InlineButton{
		Unique: INLINE_EDIT_QUESTION,
		Text:   "✏️",
		Data:   fmt.Sprintf("%d", uq.QuestionID),
	}

	// Функция для экранирования специальных символов MarkdownV2
	escapeMarkdown := func(text string) string {
		specialChars := []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}
		for _, char := range specialChars {
			text = strings.ReplaceAll(text, char, "\\"+char)
		}
		return text
	}

	questionText := escapeMarkdown(q.Question)
	answerText := escapeMarkdown(answer.Answer)

	rec := &telebot.User{ID: id}
	_, err := d.bot.Send(
		rec,
		questionText+"\n\n||"+answerText+"||",
		telebot.ModeMarkdownV2,
		&telebot.ReplyMarkup{
			InlineKeyboard: [][]telebot.InlineButton{{easy, forgot}, {repeatBtn, deleteBtn, editBtn}},
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
		MultipleAnswers: false,
	}

	label := "☑️"
	if uq.IsEdu {
		label = "✅"
	}

	repeatBtn := telebot.InlineButton{
		Unique: INLINE_BTN_REPEAT_QUESTION_AFTER_POLL,
		Text:   label,
		Data:   fmt.Sprintf("%d", uq.QuestionID),
	}

	deleteBtn := telebot.InlineButton{
		Unique: INLINE_BTN_DELETE_QUESTION_AFTER_POLL,
		Text:   INLINE_NAME_DELETE_AFTER_POLL,
		Data:   fmt.Sprintf("%d", uq.QuestionID),
	}

	editBtn := telebot.InlineButton{
		Unique: INLINE_EDIT_QUESTION,
		Text:   "✏️",
		Data:   fmt.Sprintf("%d", uq.QuestionID),
	}

	recipient := &telebot.User{ID: userID}
	msg, err := d.bot.Send(recipient, poll, &telebot.ReplyMarkup{
		InlineKeyboard: [][]telebot.InlineButton{{repeatBtn, deleteBtn, editBtn}},
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

func nextQuestion(dispatcher *QuestionDispatcher) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		if err := ctx.Send(MSG_NEXT_QUESTION); err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		t, err := dispatcher.domain.GetNearestTimeRepeat(GetContext(ctx), GetUserFromContext(ctx).TGUserID)
		if err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		if !time.Now().UTC().After(t) {
			nextTimeStr := t.Local().Format("02.01.2006 в 15:04")

			msg := fmt.Sprintf("%s *%s*", MSG_NEXT_TIME_QUESTION, nextTimeStr)

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
