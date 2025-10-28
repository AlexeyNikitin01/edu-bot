package question

import (
	"bot/internal/middleware"
	"bot/internal/ports/tags"
	"bot/internal/repo/dto"
	"context"
	"strconv"
	"strings"

	"gopkg.in/telebot.v3"

	"bot/internal/domain"
	"bot/internal/repo/edu"
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
			// Используем существующую функцию для показа тегов с дополнительным сообщением
			return tags.ShowEditTagList(ctx, d)(ctxBot)
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

	// Используем существующую функцию для показа списка тегов
	return tags.ShowEditTagList(ctx, d)(ctxBot)
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

	// Получаем тег из сообщения (текстовый ввод)
	if ctxBot.Message() != nil && ctxBot.Message().Text != "" {
		tag = strings.TrimSpace(ctxBot.Message().Text)
	}

	// Если тег не выбран, показываем список тегов снова
	if tag == "" {
		return tags.ShowEditTagList(ctx, d)(ctxBot)
	}

	draft.Tag = tag
	draft.Step++

	// Сохраняем обновленный черновик в кэш
	if err := d.SetDraftQuestion(ctx, userID, draft); err != nil {
		return err
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
		return err
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

// HandleTagSelection обрабатывает выбор тега при создании вопроса
func HandleTagSelection(ctx context.Context, d domain.UseCases) telebot.HandlerFunc {
	return func(ctxBot telebot.Context) error {
		tagName := ctxBot.Data()
		userID := middleware.GetUserFromContext(ctxBot).TGUserID

		// Получаем черновик
		draft, err := d.GetDraftQuestion(ctx, userID)
		if err != nil {
			return ctxBot.Send("❌ Ошибка при получении черновика: " + err.Error())
		}

		if draft == nil {
			return ctxBot.Send("❌ Черновик не найден. Начните создание вопроса заново.")
		}

		// Сохраняем выбранный тег в черновик
		draft.Tag = tagName
		draft.Step++

		// Сохраняем обновленный черновик
		if err = d.SetDraftQuestion(ctx, userID, draft); err != nil {
			return ctxBot.Send("❌ Ошибка при сохранении тега: " + err.Error())
		}

		// Удаляем сообщение со списком тегов
		if err = ctxBot.Delete(); err != nil {
			// Если не удалось удалить, продолжаем
		}

		// Переходим к следующему шагу - вводу вопроса
		return ctxBot.Send("Вы выбрали: " + tagName + "\n" + MSG_ADD_QUESTION)
	}
}
