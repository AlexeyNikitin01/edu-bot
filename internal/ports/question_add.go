package ports

import (
	"encoding/csv"
	"errors"
	"io"
	"strconv"
	"strings"

	"gopkg.in/telebot.v3"

	"bot/internal/app"
)

var (
	ErrGetTag       = errors.New("ошибка получения тэгов")
	ErrLengthAnswer = errors.New("ответ должен быть меншьше 100 символов")
	ErrSave         = errors.New("невозможно сохранить")
)

const (
	MSG_ADD_TAG            string = "🏷 Добавьте тэг или /cancel: "
	MSG_ADD_QUESTION       string = "✍️ Напишите вопрос или /cancel"
	MSG_ADD_CORRECT_ANSWER string = "✍✅ Введите правильный ответ или /cancel: "
	MSG_ADD_WRONG_ANSWER   string = "❌ Введите неправильный ответ (или /done, чтобы завершить, /cancel):"
	MSG_CHOOSE_ACTION      string = "ℹ️ Выберите действие."
	MSG_CANCEL             string = "Вопрос не добавлен!"
	MSG_SUCCESS            string = "✅ Вопрос успешно добавлен!"
	MSG_CSV_SUCCESS        string = "✅ Вопросы из CSV успешно добавлены!"
	MSG_CSV_ERROR          string = "❌ Ошибка при обработке CSV файла: "
)

type QuestionDraft struct {
	Step     int
	Question string
	Tag      string
	Answers  []string
	High     bool
}

var drafts = make(map[int64]*QuestionDraft)

const (
	MSG_TEST          = "Вопрос с тестом"
	MSG_HIGH_QUESTION = "Вопрос с развернутым ответом"
	MSG_TYPE_QUESTION = "Выберите тип вопроса"
	MSG_CHOOSE_HIGH   = "Выбран вопрос с развернутым ответом"
	MSG_CHOOSE_SIMPLE = "Выбран вопрос с вариантами ответа"
)

func setHigh(b bool, msg string, a app.Apper) telebot.HandlerFunc {
	return func(ctx telebot.Context) (err error) {
		draft, exists := drafts[GetUserFromContext(ctx).TGUserID]
		if !exists {
			drafts[GetUserFromContext(ctx).TGUserID] = &QuestionDraft{Step: 1}
		}

		if draft == nil {
			return nil
		}

		draft.High = b

		if err = ctx.Send(msg); err != nil {
			return err
		}

		if err = getTags(ctx, GetUserFromContext(ctx).TGUserID, a); err != nil {
			return err
		}

		return nil
	}
}

func add(domain app.Apper) telebot.HandlerFunc {
	return func(ctx telebot.Context) (err error) {
		msg := strings.TrimSpace(ctx.Message().Text)
		u := GetUserFromContext(ctx)

		draft, exists := drafts[u.TGUserID]
		if !exists {
			drafts[u.TGUserID] = &QuestionDraft{Step: 1}
			selector := &telebot.ReplyMarkup{}
			btnSimple := selector.Data(MSG_TEST, INLINE_SIMPLE_QUESTION)
			btnComplex := selector.Data(MSG_HIGH_QUESTION, INLINE_COMPLEX_QUESTION)
			selector.Inline(selector.Row(btnSimple), selector.Row(btnComplex))

			return ctx.Send(MSG_TYPE_QUESTION, selector)
		}

		if msg == CMD_CANCEL {
			delete(drafts, u.TGUserID)
			return ctx.Send(MSG_CANCEL)
		}

		switch draft.Step {
		case 1:
			draft.Tag, err = setTags(ctx)
			if err != nil {
				return err
			} else if draft.Tag == "" {
				return nil
			}
			draft.Step++
			return ctx.Send(MSG_ADD_QUESTION)
		case 2:
			draft.Question = msg
			draft.Step++
			return ctx.Send(MSG_ADD_CORRECT_ANSWER)
		case 3:
			if len(draft.Answers) >= 100 && !draft.High {
				return ctx.Send(ErrLengthAnswer.Error())
			}
			draft.Answers = append(draft.Answers, msg) // правильный
			draft.Step++
			if draft.High {
				goto Save
			}
			return ctx.Send(MSG_ADD_WRONG_ANSWER)
		case 4:
			if len(draft.Answers) >= 100 {
				return ctx.Send(ErrLengthAnswer.Error())
			}
			if msg == CMD_DONE {
				goto Save
			}
			draft.Answers = append(draft.Answers, msg)
			return ctx.Send(MSG_ADD_WRONG_ANSWER)
		}

	Save:
		delete(drafts, u.TGUserID)
		if err = domain.SaveQuestions(
			GetContext(ctx), draft.Question, draft.Tag, draft.Answers, u.TGUserID,
		); err != nil {
			return ctx.Send(errors.Join(ErrSave, err).Error())
		}
		return ctx.Send(MSG_SUCCESS, mainMenu())
	}
}

func setTags(ctx telebot.Context) (string, error) {
	if ctx.Callback() != nil {
		return ctx.Callback().Data, nil
	}

	if ctx.Message().Text != BTN_ADD_QUESTION && ctx.Message().Text != MSG_ADD_TAG { // Ввели свой тэг
		return ctx.Message().Text, nil
	}

	return "", nil
}

// getTags todo: дублирование логики
func getTags(ctx telebot.Context, userID int64, domain app.Apper) error {
	ts, err := domain.GetUniqueTags(GetContext(ctx), userID)
	if err != nil {
		return ctx.Send(errors.Join(ErrGetTag, err).Error())
	}

	var btns [][]telebot.InlineButton

	for _, t := range ts {
		btn := telebot.InlineButton{
			Unique: INLINE_BTN_TAGS,
			Text:   t,
			Data:   t,
		}
		btns = append(btns, []telebot.InlineButton{btn})
	}

	if len(btns) != 0 {
		if err = ctx.Send(MSG_ADD_TAG, &telebot.ReplyMarkup{
			InlineKeyboard: btns,
		}); err != nil {
			return ctx.Send(errors.Join(ErrGetTag, err).Error())
		}
		return nil
	}

	// Просим добавить тэг, если их нет
	return ctx.Send(MSG_ADD_TAG)
}

func setQuestionsByCSV(domain app.Apper) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		if !strings.HasSuffix(ctx.Message().Document.FileName, ".csv") {
			return ctx.Send("Пожалуйста, отправьте CSV файл")
		}

		file, err := ctx.Bot().File(&ctx.Message().Document.File)
		if err != nil {
			return ctx.Send(MSG_CSV_ERROR + err.Error())
		}
		defer file.Close()

		reader := csv.NewReader(file)
		reader.Comma = ';' // Указываем разделитель
		reader.TrimLeadingSpace = true

		userID := ctx.Sender().ID
		var successCount, errorCount int

		for {
			record, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				errorCount++
				continue
			}

			// Проверяем формат записи: вопрос, тег, правильный ответ, неправильные ответы...
			if len(record) < 3 {
				errorCount++
				continue
			}

			question := strings.TrimSpace(record[0])
			tag := strings.TrimSpace(record[1])
			correctAnswer := strings.TrimSpace(record[2])
			var wrongAnswers []string

			for i := 3; i < len(record); i++ {
				if ans := strings.TrimSpace(record[i]); ans != "" {
					wrongAnswers = append(wrongAnswers, ans)
				}
			}

			allAnswers := append([]string{correctAnswer}, wrongAnswers...)

			if err = domain.SaveQuestions(
				GetContext(ctx), question, tag, allAnswers, userID,
			); err != nil {
				errorCount++
				continue
			}

			successCount++
		}

		return ctx.Send(
			MSG_CSV_SUCCESS +
				"\nУспешно: " + strconv.Itoa(successCount) +
				"\nОшибок: " + strconv.Itoa(errorCount),
		)
	}
}
