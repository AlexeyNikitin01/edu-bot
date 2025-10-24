package question

import (
	"bot/internal/middleware"
	"bot/internal/repo/dto"
	"context"
	"errors"
	"strconv"
	"strings"

	"gopkg.in/telebot.v3"

	"bot/internal/domain"
	"bot/internal/repo/edu"
)

// Ошибки приложения
var (
	ErrGetTag    = errors.New("ошибка получения тэгов")
	ErrSave      = errors.New("невозможно сохранить")
	ErrSaveDraft = errors.New("ошибка сохранения черновика")
	ErrGetDraft  = errors.New("ошибка получения черновика")
)

// Константы сообщений
const (
	MSG_ADD_TAG                        = "🏷 Введите свой тэг или выберите из списка, или нажмите /cancel для отмены: "
	MSG_ADD_QUESTION                   = "✍️ Напишите вопрос или нажмите /cancel для отмены"
	MSG_ADD_CORRECT_ANSWER             = "✍✅ Введите правильный ответ или нажмите /cancel для отмены: "
	MSG_CANCEL                         = "Вы отменили действие👊!"
	MSG_SUCCESS                        = "✅ Успех!"
	MSG_EDIT                           = "<b>Введите новое значение для или нажмите /cancel для отмены:</b>\n\n "
	MSG_SUCCESS_UPDATE_TAG             = "Тэг обновлен"
	MSG_SUCCESS_UPDATE_NAME_QUESTION   = "Вопрос обновлен"
	MSG_SUCCESS_UPDATE_ANSWER          = "Ответ обновлен"
	MSG_EDIT_TAG_BY_QUESTION           = "Выберите или введите свой тэг или нажмите /cancel для отмены: "
	MSG_SUCCESS_UPDATE_TAG_BY_QUESTION = "Тэг для вопроса обновлен"
)

func SetEdit(ctx context.Context, field string, d domain.UseCases) telebot.HandlerFunc {
	return func(ctxBot telebot.Context) (err error) {
		strID := ctxBot.Data()
		id, err := strconv.Atoi(strID)
		if err != nil {
			return err
		}

		userID := middleware.GetUserFromContext(ctxBot).TGUserID

		draft, err := d.GetDraftQuestion(ctx, userID)
		if err != nil {
			return err
		}

		if draft == nil {
			draft = &dto.QuestionDraft{Step: 1}
		}

		switch field {
		case edu.TableNames.Tags:
			draft.TagID = int64(id)
		case edu.QuestionTableColumns.Question:
			draft.QuestionIDByName = int64(id)
		case edu.QuestionTableColumns.TagID:
			draft.QuestionIDByTag = int64(id)
			if err = getTags(ctx, ctxBot, userID, d); err != nil {
				return err
			}
			return ctxBot.Send(MSG_EDIT_TAG_BY_QUESTION)
		case edu.AnswerTableColumns.Answer:
			draft.AnswerID = int64(id)
		}

		// Сохраняем черновик в кэш
		if err = d.SetDraftQuestion(ctx, userID, draft); err != nil {
			return err
		}

		menu := &telebot.ReplyMarkup{}
		btnShowCurrent := menu.Data("👀 Посмотреть текущее значение", INLINE_SHOW_CURRENT_VALUE, strID)
		menu.Inline(menu.Row(btnShowCurrent))

		return ctxBot.Send(MSG_EDIT, menu, telebot.ModeHTML)
	}
}

// UpsertUserQuestion обрабатывает создание или редактирование вопроса пользователя
// Объединяет логику создания нового вопроса и редактирования существующих сущностей
func UpsertUserQuestion(ctx context.Context, d domain.UseCases) telebot.HandlerFunc {
	return func(ctxBot telebot.Context) (err error) {
		msg := strings.TrimSpace(ctxBot.Message().Text)
		userID := middleware.GetUserFromContext(ctxBot).TGUserID

		draft, err := d.GetDraftQuestion(ctx, userID)
		if err != nil {
			return err
		}

		if draft == nil {
			return initNewDraft(ctx, ctxBot, userID, d)
		}

		// Обработка отмены действия
		if msg == CMD_CANCEL {
			return cancelDraft(ctx, ctxBot, userID, d)
		}

		// Приоритетная обработка черновиков редактирования
		if draft.TagID != 0 || draft.QuestionIDByName != 0 || draft.AnswerID != 0 || draft.QuestionIDByTag != 0 {
			return updateUserQuestion(ctx, ctxBot, draft, msg, userID, d)
		}

		// Обработка создания нового вопроса
		return createUserQuestion(ctx, ctxBot, draft, msg, userID, d)
	}
}

func initNewDraft(ctx context.Context, ctxBot telebot.Context, userID int64, d domain.UseCases) error {
	draft := &dto.QuestionDraft{Step: 1}
	if err := d.SetDraftQuestion(ctx, userID, draft); err != nil {
		return err
	}
	if err := ctxBot.Send(MSG_LIST_TAGS); err != nil {
		return err
	}
	return getTags(ctx, ctxBot, userID, d)
}

func cancelDraft(ctx context.Context, ctxBot telebot.Context, userID int64, d domain.UseCases) error {
	if err := d.DeleteDraftQuestion(ctx, userID); err != nil {
		return err
	}
	return ctxBot.Send(MSG_CANCEL)
}

func updateUserQuestion(
	ctx context.Context, ctxBot telebot.Context, draft *dto.QuestionDraft, msg string, userID int64, d domain.UseCases,
) error {
	switch {
	case draft.TagID != 0:
		return updateTag(ctx, ctxBot, draft, msg, userID, d)
	case draft.QuestionIDByName != 0:
		return updateQuestionName(ctx, ctxBot, draft, msg, userID, d)
	case draft.AnswerID != 0:
		return updateAnswer(ctx, ctxBot, draft, msg, userID, d)
	case draft.QuestionIDByTag != 0:
		return updateTagByQuestion(ctx, ctxBot, draft, msg, userID, d)
	}
	return nil
}

// updateTag обновляет текст существующего тега
func updateTag(
	ctx context.Context, ctxBot telebot.Context, draft *dto.QuestionDraft, msg string, userID int64, d domain.UseCases,
) error {
	if err := d.UpdateTag(ctx, draft.TagID, msg); err != nil {
		return err
	}
	if err := d.DeleteDraftQuestion(ctx, userID); err != nil {
		return err
	}
	return ctxBot.Send(MSG_SUCCESS_UPDATE_TAG)
}

// updateQuestionName обновляет текст существующего вопроса
func updateQuestionName(
	ctx context.Context, ctxBot telebot.Context, draft *dto.QuestionDraft, msg string, userID int64, d domain.UseCases,
) error {
	if err := d.UpdateQuestionName(ctx, draft.QuestionIDByName, msg); err != nil {
		return err
	}
	if err := d.DeleteDraftQuestion(ctx, userID); err != nil {
		return err
	}
	return ctxBot.Send(MSG_SUCCESS_UPDATE_NAME_QUESTION)
}

// updateAnswer обновляет текст существующего ответа
func updateAnswer(
	ctx context.Context, ctxBot telebot.Context, draft *dto.QuestionDraft, msg string, userID int64, d domain.UseCases,
) error {
	if err := d.UpdateAnswer(ctx, draft.AnswerID, msg); err != nil {
		return err
	}
	if err := d.DeleteDraftQuestion(ctx, userID); err != nil {
		return err
	}
	return ctxBot.Send(MSG_SUCCESS_UPDATE_ANSWER)
}

// updateTagByQuestion обновляет тег для существующего вопроса
// Поддерживает выбор тега из списка или ввод нового
func updateTagByQuestion(
	ctx context.Context, ctxBot telebot.Context, draft *dto.QuestionDraft, msg string, userID int64, d domain.UseCases,
) error {
	tag := ""

	if ctxBot.Callback() != nil {
		tag = ctxBot.Callback().Data
	} else if ctxBot.Message().Text != BTN_ADD_QUESTION && ctxBot.Message().Text != MSG_ADD_TAG {
		tag = ctxBot.Message().Text
	}

	if tag == "" {
		return nil
	}

	if err := d.UpdateTagByQuestion(ctx, draft.QuestionIDByTag, tag); err != nil {
		return err
	}
	if err := d.DeleteDraftQuestion(ctx, userID); err != nil {
		return err
	}
	return ctxBot.Send(MSG_SUCCESS_UPDATE_TAG_BY_QUESTION)
}

func createUserQuestion(
	ctx context.Context, ctxBot telebot.Context, draft *dto.QuestionDraft, msg string, userID int64, d domain.UseCases,
) error {
	switch draft.Step {
	case 1:
		return processTagSelection(ctx, ctxBot, draft, userID, d)
	case 2:
		return processQuestionInput(ctx, ctxBot, draft, userID, msg, d)
	case 3:
		return processCorrectAnswerInputAndSaveQuestion(ctx, ctxBot, draft, userID, msg, d)
	}
	return nil
}

func processTagSelection(
	ctx context.Context, ctxBot telebot.Context, draft *dto.QuestionDraft, userID int64, d domain.UseCases,
) error {
	tag := ""

	// Получаем тег из сообщения или callback
	if ctxBot.Callback() != nil {
		tag = ctxBot.Callback().Data
	} else if ctxBot.Message().Text != BTN_ADD_QUESTION && ctxBot.Message().Text != MSG_ADD_TAG {
		tag = ctxBot.Message().Text
	}

	// Если тег не выбран, выходим без ошибки
	if tag == "" {
		return nil
	}

	draft.Tag = tag
	draft.Step++

	// Сохраняем обновленный черновик в кэш
	if err := d.SetDraftQuestion(ctx, userID, draft); err != nil {
		return errors.Join(ErrSaveDraft, err)
	}

	return ctxBot.Send(MSG_ADD_QUESTION)
}

// processQuestionInput обрабатывает ввод текста вопроса
func processQuestionInput(
	ctx context.Context, ctxBot telebot.Context, draft *dto.QuestionDraft, userID int64, msg string, d domain.UseCases,
) error {
	draft.Question = msg
	draft.Step++

	// Сохраняем обновленный черновик в кэш
	if err := d.SetDraftQuestion(ctx, userID, draft); err != nil {
		return errors.Join(ErrSaveDraft, err)
	}

	return ctxBot.Send(MSG_ADD_CORRECT_ANSWER)
}

func processCorrectAnswerInputAndSaveQuestion(
	ctx context.Context, ctxBot telebot.Context, draft *dto.QuestionDraft, userID int64, msg string, d domain.UseCases,
) error {
	draft.Answers = append(draft.Answers, msg)

	// Удаляем черновик после сохранения вопроса (даже если будет ошибка)
	defer d.DeleteDraftQuestion(ctx, userID)

	if err := d.SaveQuestions(ctx, draft.Question, draft.Tag, draft.Answers, userID); err != nil {
		return err
	}

	return ctxBot.Send(MSG_SUCCESS)
}

func getTags(
	ctx context.Context, ctxBot telebot.Context, userID int64, d domain.UseCases) error {
	ts, err := d.GetUniqueTags(ctx, userID)
	if err != nil {
		return err
	}

	var btns [][]telebot.InlineButton

	// Создаем кнопки для каждого тега
	for _, t := range ts {
		btn := telebot.InlineButton{
			Unique: INLINE_BTN_TAGS,
			Text:   t.Tag,
			Data:   t.Tag,
		}
		btns = append(btns, []telebot.InlineButton{btn})
	}

	// Если есть теги, показываем их списком
	if len(btns) != 0 {
		if err = ctxBot.Send(MSG_ADD_TAG, &telebot.ReplyMarkup{
			InlineKeyboard: btns,
		}); err != nil {
			return ctxBot.Send(errors.Join(ErrGetTag, err).Error())
		}
		return nil
	}

	// Если тегов нет, просим добавить новый
	return ctxBot.Send(MSG_ADD_TAG)
}
