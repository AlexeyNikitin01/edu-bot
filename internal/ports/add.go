package ports

import (
	"errors"
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

	DONE   string = "/done"
	CANCEL string = "/cancel"
)

type QuestionDraft struct {
	Step     int
	Question string
	Tag      string
	Answers  []string
}

var drafts = make(map[int64]*QuestionDraft)

func add(domain app.Apper) telebot.HandlerFunc {
	return func(ctx telebot.Context) (err error) {
		msg := strings.TrimSpace(ctx.Message().Text)
		u := GetUserFromContext(ctx)

		draft, exists := drafts[u.TGUserID]
		if !exists {
			return ctx.Send(MSG_CHOOSE_ACTION)
		}

		if msg == CANCEL {
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
			if len(draft.Answers) >= 100 {
				return ctx.Send(ErrLengthAnswer.Error())
			}
			draft.Answers = append(draft.Answers, msg) // правильный
			draft.Step++
			return ctx.Send(MSG_ADD_WRONG_ANSWER)
		case 4:
			if len(draft.Answers) >= 100 {
				return ctx.Send(ErrLengthAnswer.Error())
			}
			if msg == DONE {
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

	if ctx.Message().Text != ADD_QUESTION && ctx.Message().Text != MSG_ADD_TAG { // Ввели свой тэг
		return ctx.Message().Text, nil
	}

	return "", nil
}

func getTags(ctx telebot.Context, userID int64, domain app.Apper) error {
	ts, err := domain.GetUniqueTags(GetContext(ctx), userID)
	if err != nil {
		return ctx.Send(errors.Join(ErrGetTag, err).Error())
	}

	var btns [][]telebot.InlineButton

	for _, t := range ts {
		btn := telebot.InlineButton{
			Unique: TAGS,
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
