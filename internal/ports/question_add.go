package ports

import (
	"errors"
	"strconv"
	"strings"

	"gopkg.in/telebot.v3"

	"bot/internal/app"
	"bot/internal/repo/edu"
)

var (
	ErrGetTag       = errors.New("ошибка получения тэгов")
	ErrLengthAnswer = errors.New("ответ должен быть меншьше 100 символов")
	ErrSave         = errors.New("невозможно сохранить")
)

const (
	MSG_ADD_TAG                        = "🏷 Введите свой тэг или выберите из списка, или нажмите /cancel для отмены: "
	MSG_ADD_QUESTION                   = "✍️ Напишите вопрос или нажмите /cancel для отмены"
	MSG_ADD_CORRECT_ANSWER             = "✍✅ Введите правильный ответ или нажмите /cancel для отмены: "
	MSG_CANCEL                         = "Вы отменили действие👊!"
	MSG_SUCCESS                        = "✅ Успех!"
	MSG_EDIT                           = "Введите новое значение для или нажмите /cancel для отмены: "
	MSG_SUCCESS_UPDATE_TAG             = "Тэг обновлен"
	MSG_SUCCESS_UPDATE_NAME_QUESTION   = "Вопрос обновлен"
	MSG_SUCCESS_UPDATE_ANSWER          = "Ответ обновлен"
	MSG_EDIT_TAG_BY_QUESTION           = "Выберите или введите свой тэг или нажмите /cancel для отмены: "
	MSG_SUCCESS_UPDATE_TAG_BY_QUESTION = "Тэг для вопроса обновлен"
)

type QuestionDraft struct {
	Step             int
	Question         string
	Tag              string
	Answers          []string
	TagID            int64
	QuestionIDByTag  int64
	QuestionIDByName int64
	AnswerID         int64
}

var drafts = make(map[int64]*QuestionDraft)

func setEdit(field string, domain app.Apper) telebot.HandlerFunc {
	return func(ctx telebot.Context) (err error) {
		strID := ctx.Data()
		id, err := strconv.Atoi(strID)
		if err != nil {
			return err
		}

		draft, exists := drafts[GetUserFromContext(ctx).TGUserID]
		if !exists {
			drafts[GetUserFromContext(ctx).TGUserID] = &QuestionDraft{Step: 1}
			draft, _ = drafts[GetUserFromContext(ctx).TGUserID]
		}

		if draft == nil {
			return nil
		}

		switch field {
		case edu.TableNames.Tags:
			draft.TagID = int64(id)
		case edu.QuestionTableColumns.Question:
			draft.QuestionIDByName = int64(id)
		case edu.QuestionTableColumns.TagID:
			draft.QuestionIDByTag = int64(id)
			if err = getTags(ctx, GetUserFromContext(ctx).TGUserID, domain); err != nil {
				return err
			}
			return ctx.Send(MSG_EDIT_TAG_BY_QUESTION)
		case edu.AnswerTableColumns.Answer:
			draft.AnswerID = int64(id)
		}

		return ctx.Send(MSG_EDIT)
	}
}

func add(domain app.Apper) telebot.HandlerFunc {
	return func(ctx telebot.Context) (err error) {
		msg := strings.TrimSpace(ctx.Message().Text)
		u := GetUserFromContext(ctx)

		draft, exists := drafts[u.TGUserID]
		if !exists {
			drafts[u.TGUserID] = &QuestionDraft{Step: 1}
			if err = ctx.Send(msg); err != nil {
				return err
			}

			return getTags(ctx, GetUserFromContext(ctx).TGUserID, domain)
		}

		if msg == CMD_CANCEL {
			delete(drafts, u.TGUserID)
			return ctx.Send(MSG_CANCEL)
		}

		//todo: править
		if draft.TagID != 0 {
			if err = domain.UpdateTag(GetContext(ctx), draft.TagID, msg); err != nil {
				return err
			}
			delete(drafts, u.TGUserID)
			return ctx.Send(MSG_SUCCESS_UPDATE_TAG)
		} else if draft.QuestionIDByName != 0 {
			if err = domain.UpdateQuestionName(GetContext(ctx), draft.QuestionIDByName, msg); err != nil {
				return err
			}
			delete(drafts, u.TGUserID)
			return ctx.Send(MSG_SUCCESS_UPDATE_NAME_QUESTION)
		} else if draft.AnswerID != 0 {
			if err = domain.UpdateAnswer(GetContext(ctx), draft.AnswerID, msg); err != nil {
				return err
			}
			delete(drafts, u.TGUserID)
			return ctx.Send(MSG_SUCCESS_UPDATE_ANSWER)
		} else if draft.QuestionIDByTag != 0 {
			newTag, err := setTags(ctx)
			if err != nil {
				return err
			}
			if err = domain.UpdateTagByQuestion(GetContext(ctx), draft.QuestionIDByTag, newTag); err != nil {
				return err
			}
			delete(drafts, u.TGUserID)
			return ctx.Send(MSG_SUCCESS_UPDATE_TAG_BY_QUESTION)
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
			draft.Answers = append(draft.Answers, msg) // правильный
			goto Save
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
			Text:   t.Tag,
			Data:   t.Tag,
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
