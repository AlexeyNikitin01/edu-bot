package ports

import (
	"context"
	"fmt"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"gopkg.in/telebot.v3"

	"bot/internal/app"
	"bot/internal/repo/edu"
)

const (
	MSG_FORGOT        = "СЛОЖНО"
	MSG_REMEMBER      = "ЛЕГКО"
	MSG_NEXT_QUESTION = "😎"

	BtnShowAnswer = "📝 Показать ответ"
	BtnRepeat     = "🔔"
	BtnRepeatEdu  = "💤"
	BtnDelete     = "🗑️"
	BtnEdit       = "✏️"
)

type QuestionDispatcher struct {
	domain app.Apper
	bot    *telebot.Bot
	ctx    context.Context
	cache  app.UserCacher
	done   chan struct{}
	wg     sync.WaitGroup
}

func NewDispatcher(ctx context.Context, domain app.Apper, bot *telebot.Bot, cache app.UserCacher) *QuestionDispatcher {
	return &QuestionDispatcher{
		domain: domain,
		bot:    bot,
		ctx:    ctx,
		cache:  cache,
		done:   make(chan struct{}),
		wg:     sync.WaitGroup{},
	}
}

func (d *QuestionDispatcher) Stop() {
	close(d.done) // Закрываем канал для уведомления всех воркеров
	d.wg.Wait()   // Ждем завершения всех воркеров
	log.Println("QuestionDispatcher stopped")
}

func (d *QuestionDispatcher) StartPollingLoop() {
	log.Println("QuestionDispatcher start")
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()

		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-d.done:
				return
			case <-d.ctx.Done():
				return
			case <-ticker.C:
				// Получаем пользователей для которых запущены воркеры
				activeWorkers, err := d.cache.GetActiveWorkers(d.ctx)
				if err != nil {
					log.Println("Ошибка получения активных воркеров из Redis:", err)
					continue
				}

				// Забираем всех пользователей для которых не запущен воркер
				users, err := edu.Users(
					qm.Select(edu.UserColumns.TGUserID),
					edu.UserWhere.TGUserID.NIN(activeWorkers),
					edu.UserWhere.Block.EQ(false),
				).All(d.ctx, boil.GetContextDB())
				if err != nil {
					log.Println("Ошибка получения пользователей:", err)
					continue
				}

				for _, user := range users {
					userID := user.TGUserID

					if err = d.cache.AddWorker(d.ctx, userID); err != nil {
						log.Printf("Ошибка добавления воркера %d в Redis: %v", userID, err)
						continue
					}

					d.wg.Add(1)
					go func() {
						defer d.wg.Done()
						d.worker(userID)
					}()
				}
			}
		}
	}()
}

func (d *QuestionDispatcher) worker(userID int64) {
	t := time.NewTicker(time.Second * 2)
	defer t.Stop()
	defer func() {
		if err := d.cache.RemoveWorker(d.ctx, userID); err != nil {
			log.Printf("Ошибка удаления воркера %d из Redis: %v", userID, err)
		}
		log.Printf("Воркер для пользователя %d завершен", userID)
	}()

	for {
		select {
		case <-d.done:
			return
		case <-d.ctx.Done():
			return
		case <-t.C:
			log.Printf("%d пытаемся отправить запрос", userID)
			waiting, err := d.cache.GetUserWaiting(d.ctx, userID)
			if err != nil {
				log.Printf("Ошибка получения статуса waiting из Redis для пользователя %d: %v", userID, err)
				continue
			}

			if waiting {
				log.Printf("%d ждем пока пользователь ответит", userID)
				continue
			}

			if err = d.cache.SetUserWaiting(d.ctx, userID, true); err != nil {
				log.Printf("Ошибка установки статуса waiting в Redis для пользователя %d: %v", userID, err)
				continue
			}

			if err = d.sendRandomQuestionForUser(userID); err != nil {
				log.Printf("Ошибка отправки вопроса пользователю %d: %v", userID, err)
				if err = d.cache.SetUserWaiting(d.ctx, userID, false); err != nil {
					log.Printf("Ошибка сброса статуса waiting в Redis для пользователя %d: %v", userID, err)
				}
			}
			log.Printf("%d отправили вопрос пользователю, ждём пока ответит", userID)
		}
	}
}

func (d *QuestionDispatcher) sendRandomQuestionForUser(userID int64) error {
	uq, err := d.domain.GetRandomNearestQuestionWithAnswer(d.ctx, userID)
	if err != nil {
		return err
	}

	tag := escapeMarkdown(uq.GetQuestion().R.GetTag().Tag)
	questionText := escapeMarkdown(uq.GetQuestion().Question)

	buttons := getQuestionButtons(uq, false)

	rec := &telebot.User{ID: userID}
	_, err = d.bot.Send(
		rec,
		tag+": "+questionText,
		telebot.ModeMarkdownV2,
		&telebot.ReplyMarkup{
			InlineKeyboard: buttons,
		},
	)

	return err
}

// viewAnswer обработчик для отображения вопросов после взаимодействия "показать" или "спрятать"
func viewAnswer(domain app.Apper, showAnswer bool) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		data := ctx.Data()
		qID, err := strconv.Atoi(data)
		if err != nil {
			return err
		}

		uq, err := domain.GetUserQuestion(GetContext(ctx), GetUserFromContext(ctx).TGUserID, int64(qID))
		if err != nil {
			return err
		}

		question := uq.GetQuestion().Question
		tag := uq.R.GetQuestion().R.GetTag().Tag
		answer := uq.R.GetQuestion().R.GetAnswers()[0]

		result := escapeMarkdown(tag) + ": " + escapeMarkdown(question)
		if showAnswer {
			result += "\n\n" + escapeMarkdown(answer.Answer)
		}

		return ctx.Edit(
			result,
			telebot.ModeMarkdownV2,
			&telebot.ReplyMarkup{
				InlineKeyboard: getQuestionButtons(uq, showAnswer),
			},
		)
	}
}

func escapeMarkdown(text string) string {
	specialChars := []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}
	for _, char := range specialChars {
		text = strings.ReplaceAll(text, char, "\\"+char)
	}
	return text
}

// nextQuestion кнопка дальше
func nextQuestion(d *QuestionDispatcher) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		if err := ctx.Send(MSG_NEXT_QUESTION); err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		user := GetUserFromContext(ctx)
		t, err := d.domain.GetNearestTimeRepeat(GetContext(ctx), user.TGUserID)
		if err != nil {
			return ctx.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		now := time.Now().UTC()
		if !now.After(t) {
			duration := t.Sub(now)
			msg := fmt.Sprintf("⏳ Следующий вопрос будет доступен через: %s", timeLeftMsg(duration))

			if err = ctx.Send(msg, telebot.ModeMarkdown); err != nil {
				return ctx.Respond(&telebot.CallbackResponse{Text: err.Error()})
			}
		}

		if err = d.cache.SetUserWaiting(d.ctx, user.TGUserID, false); err != nil {
			log.Printf("Ошибка сброса статуса waiting в Redis для пользователя %d: %v", user.TGUserID, err)
		}

		return nil
	}
}

func timeLeftMsg(duration time.Duration) string {
	var timeParts []string

	// Функция для правильного склонения
	pluralize := func(n int, forms []string) string {
		n = n % 100
		if n > 10 && n < 20 {
			return forms[2]
		}
		n = n % 10
		if n == 1 {
			return forms[0]
		}
		if n >= 2 && n <= 4 {
			return forms[1]
		}
		return forms[2]
	}

	// Разбиваем duration на дни, часы и минуты
	days := int(duration.Hours() / 24)
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60

	if days > 0 {
		timeParts = append(timeParts, fmt.Sprintf("%d %s", days, pluralize(days, []string{"день", "дня", "дней"})))
	}
	if hours > 0 {
		timeParts = append(timeParts, fmt.Sprintf("%d %s", hours, pluralize(hours, []string{"час", "часа", "часов"})))
	}
	if minutes > 0 && days == 0 { // Минуты показываем только если нет дней
		timeParts = append(timeParts, fmt.Sprintf("%d %s", minutes, pluralize(minutes, []string{"минуту", "минуты", "минут"})))
	}

	t := strings.Join(timeParts, " ")
	if t == "" {
		t = "менее минуты"
	}

	return t
}

// getQuestionButtons создает клавиатуру для сообщения с вопросом
func getQuestionButtons(uq *edu.UsersQuestion, showAnswer bool) [][]telebot.InlineButton {
	forgot := telebot.InlineButton{
		Unique: INLINE_FORGOT_HIGH_QUESTION,
		Text:   MSG_FORGOT,
		Data:   fmt.Sprintf("%d", uq.QuestionID),
	}

	easy := telebot.InlineButton{
		Unique: INLINE_REMEMBER_HIGH_QUESTION,
		Text:   MSG_REMEMBER,
		Data:   fmt.Sprintf("%d", uq.QuestionID),
	}

	label := BtnRepeat
	if uq.IsEdu {
		label = BtnRepeatEdu
	}

	repeatBtn := telebot.InlineButton{
		Unique: INLINE_BTN_REPEAT_QUESTION_AFTER_POLL_HIGH,
		Text:   label,
		Data:   fmt.Sprintf("%d", uq.QuestionID),
	}

	deleteBtn := telebot.InlineButton{
		Unique: INLINE_BTN_DELETE_QUESTION_AFTER_POLL_HIGH,
		Text:   BtnDelete,
		Data:   fmt.Sprintf("%d", uq.QuestionID),
	}

	editBtn := telebot.InlineButton{
		Unique: INLINE_EDIT_QUESTION,
		Text:   BtnEdit,
		Data:   fmt.Sprintf("%d", uq.QuestionID),
	}

	// Определяем кнопку показа/скрытия ответа
	var answerBtn telebot.InlineButton
	if showAnswer {
		answerBtn = telebot.InlineButton{
			Unique: INLINE_TURN_ANSWER,
			Text:   "📝 Свернуть ответ",
			Data:   fmt.Sprintf("%d", uq.QuestionID),
		}
	} else {
		answerBtn = telebot.InlineButton{
			Unique: INLINE_SHOW_ANSWER,
			Text:   BtnShowAnswer,
			Data:   fmt.Sprintf("%d", uq.QuestionID),
		}
	}

	return [][]telebot.InlineButton{
		{answerBtn},
		{easy, forgot},
		{repeatBtn, deleteBtn, editBtn},
	}
}
