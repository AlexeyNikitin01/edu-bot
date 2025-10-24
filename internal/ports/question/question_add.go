package question

import (
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

// SetEdit инициализирует черновик редактирования для указанного поля
// field - тип сущности для редактирования (тег, вопрос, ответ)
// domain - слой приложения для работы с данными
// cache - кэш для хранения черновиков
func SetEdit(ctx context.Context, field string, domain domain.UseCases) telebot.HandlerFunc {
	return func(ctxBot telebot.Context) (err error) {
		// Парсим ID сущности из данных callback
		strID := ctxBot.Data()
		id, err := strconv.Atoi(strID)
		if err != nil {
			return err
		}

		// Получаем или создаем черновик для пользователя
		user := GetUserFromContext(ctxBot)
		draft, err := cache.GetDraft(GetContext(ctxBot), user.TGUserID)
		if err != nil {
			return errors.Join(ErrGetDraft, err)
		}

		if draft == nil {
			draft = &dto.QuestionDraft{Step: 1}
		}

		// Устанавливаем соответствующий ID в зависимости от типа редактирования
		switch field {
		case edu.TableNames.Tags:
			draft.TagID = int64(id)
		case edu.QuestionTableColumns.QuestionService:
			draft.QuestionIDByName = int64(id)
		case edu.QuestionTableColumns.TagID:
			draft.QuestionIDByTag = int64(id)
			// Для изменения тега вопроса показываем список доступных тегов
			if err = getTags(ctxBot, user.TGUserID, domain); err != nil {
				return err
			}
			return ctxBot.Send(MSG_EDIT_TAG_BY_QUESTION)
		case edu.AnswerTableColumns.AnswerService:
			draft.AnswerID = int64(id)
		}

		// Сохраняем черновик в кэш
		if err = cache.SaveDraft(GetContext(ctxBot), user.TGUserID, draft); err != nil {
			return errors.Join(ErrSaveDraft, err)
		}

		// Создаем клавиатуру с кнопкой для просмотра текущего значения
		menu := &telebot.ReplyMarkup{}
		btnShowCurrent := menu.Data("👀 Посмотреть текущее значение", INLINE_SHOW_CURRENT_VALUE, strID)
		menu.Inline(menu.Row(btnShowCurrent))

		return ctxBot.Send(MSG_EDIT, menu, telebot.ModeHTML)
	}
}

// UpsertUserQuestion обрабатывает создание или редактирование вопроса пользователя
// Объединяет логику создания нового вопроса и редактирования существующих сущностей
func UpsertUserQuestion(domain domain.Apper, cache domain.DraftCacher) telebot.HandlerFunc {
	return func(ctx telebot.Context) (err error) {
		msg := strings.TrimSpace(ctx.Message().Text)
		u := GetUserFromContext(ctx)

		// Получаем черновик пользователя из кэша
		draft, err := cache.GetDraft(GetContext(ctx), u.TGUserID)
		if err != nil {
			return errors.Join(ErrGetDraft, err)
		}

		if draft == nil {
			return initNewDraft(ctx, u, domain, cache)
		}

		// Обработка отмены действия
		if msg == CMD_CANCEL {
			return cancelDraft(ctx, u, cache)
		}

		// Приоритетная обработка черновиков редактирования
		if draft.TagID != 0 || draft.QuestionIDByName != 0 || draft.AnswerID != 0 || draft.QuestionIDByTag != 0 {
			return updateUserQuestion(ctx, draft, msg, u, domain, cache)
		}

		// Обработка создания нового вопроса
		return createUserQuestion(ctx, draft, msg, u, domain, cache)
	}
}

// initNewDraft инициализирует новый черновик для создания вопроса
// Показывает пользователю список доступных тегов
func initNewDraft(ctx telebot.Context, u *edu.UserService, domain domain.Apper, cache domain.DraftCacher) error {
	draft := &dto.QuestionDraft{Step: 1}
	if err := cache.SaveDraft(GetContext(ctx), u.TGUserID, draft); err != nil {
		return errors.Join(ErrSaveDraft, err)
	}
	if err := ctx.Send(MSG_LIST_TAGS); err != nil {
		return err
	}
	return getTags(ctx, u.TGUserID, domain)
}

// cancelDraft отменяет текущий черновик и очищает состояние
func cancelDraft(ctx telebot.Context, u *edu.UserService, cache domain.DraftCacher) error {
	if err := cache.DeleteDraft(GetContext(ctx), u.TGUserID); err != nil {
		return err
	}
	return ctx.Send(MSG_CANCEL)
}

// updateUserQuestion обрабатывает редактирование существующих сущностей (тегов, вопросов, ответов)
// Определяет тип редактирования и делегирует выполнение соответствующему обработчику
func updateUserQuestion(ctx telebot.Context, draft *dto.QuestionDraft, msg string, u *edu.UserService, domain domain.Apper, cache domain.DraftCacher) error {
	switch {
	case draft.TagID != 0:
		return updateTag(ctx, draft, msg, u, domain, cache)
	case draft.QuestionIDByName != 0:
		return updateQuestionName(ctx, draft, msg, u, domain, cache)
	case draft.AnswerID != 0:
		return updateAnswer(ctx, draft, msg, u, domain, cache)
	case draft.QuestionIDByTag != 0:
		return updateTagByQuestion(ctx, draft, u, domain, cache)
	}
	return nil
}

// updateTag обновляет текст существующего тега
func updateTag(ctx telebot.Context, draft *dto.QuestionDraft, msg string, u *edu.UserService, domain domain.Apper, cache domain.DraftCacher) error {
	if err := domain.UpdateTag(GetContext(ctx), draft.TagID, msg); err != nil {
		return err
	}
	if err := cache.DeleteDraft(GetContext(ctx), u.TGUserID); err != nil {
		return err
	}
	return ctx.Send(MSG_SUCCESS_UPDATE_TAG)
}

// updateQuestionName обновляет текст существующего вопроса
func updateQuestionName(ctx telebot.Context, draft *dto.QuestionDraft, msg string, u *edu.UserService, domain domain.Apper, cache domain.DraftCacher) error {
	if err := domain.UpdateQuestionName(GetContext(ctx), draft.QuestionIDByName, msg); err != nil {
		return err
	}
	if err := cache.DeleteDraft(GetContext(ctx), u.TGUserID); err != nil {
		return err
	}
	return ctx.Send(MSG_SUCCESS_UPDATE_NAME_QUESTION)
}

// updateAnswer обновляет текст существующего ответа
func updateAnswer(ctx telebot.Context, draft *dto.QuestionDraft, msg string, u *edu.UserService, domain domain.Apper, cache domain.DraftCacher) error {
	if err := domain.UpdateAnswer(GetContext(ctx), draft.AnswerID, msg); err != nil {
		return err
	}
	if err := cache.DeleteDraft(GetContext(ctx), u.TGUserID); err != nil {
		return err
	}
	return ctx.Send(MSG_SUCCESS_UPDATE_ANSWER)
}

// updateTagByQuestion обновляет тег для существующего вопроса
// Поддерживает выбор тега из списка или ввод нового
func updateTagByQuestion(ctx telebot.Context, draft *dto.QuestionDraft, u *edu.UserService, domain domain.Apper, cache domain.DraftCacher) error {
	tag := ""

	// Получаем тег из сообщения или callback
	if ctx.Callback() != nil {
		tag = ctx.Callback().Data
	} else if ctx.Message().Text != BTN_ADD_QUESTION && ctx.Message().Text != MSG_ADD_TAG {
		tag = ctx.Message().Text
	}

	// Если тег не выбран, выходим без ошибки
	if tag == "" {
		return nil
	}

	if err := domain.UpdateTagByQuestion(GetContext(ctx), draft.QuestionIDByTag, tag); err != nil {
		return err
	}
	if err := cache.DeleteDraft(GetContext(ctx), u.TGUserID); err != nil {
		return err
	}
	return ctx.Send(MSG_SUCCESS_UPDATE_TAG_BY_QUESTION)
}

// createUserQuestion обрабатывает процесс создания нового вопроса
// Последовательно проходит через шаги: выбор тега → ввод вопроса → ввод ответа
func createUserQuestion(ctx telebot.Context, draft *dto.QuestionDraft, msg string, u *edu.UserService, domain domain.Apper, cache domain.DraftCacher) error {
	switch draft.Step {
	case 1:
		return processTagSelection(ctx, draft, cache)
	case 2:
		return processQuestionInput(ctx, draft, msg, cache)
	case 3:
		return processCorrectAnswerInputAndSaveQuestion(ctx, draft, msg, u, domain, cache)
	}
	return nil
}

// processTagSelection обрабатывает выбор тега для нового вопроса
// Поддерживает выбор из списка или ввод пользовательского тега
func processTagSelection(ctx telebot.Context, draft *dto.QuestionDraft, cache domain.DraftCacher) error {
	tag := ""

	// Получаем тег из сообщения или callback
	if ctx.Callback() != nil {
		tag = ctx.Callback().Data
	} else if ctx.Message().Text != BTN_ADD_QUESTION && ctx.Message().Text != MSG_ADD_TAG {
		tag = ctx.Message().Text
	}

	// Если тег не выбран, выходим без ошибки
	if tag == "" {
		return nil
	}

	draft.Tag = tag
	draft.Step++

	// Сохраняем обновленный черновик в кэш
	if err := cache.SaveDraft(GetContext(ctx), GetUserFromContext(ctx).TGUserID, draft); err != nil {
		return errors.Join(ErrSaveDraft, err)
	}

	return ctx.Send(MSG_ADD_QUESTION)
}

// processQuestionInput обрабатывает ввод текста вопроса
func processQuestionInput(ctx telebot.Context, draft *dto.QuestionDraft, msg string, cache domain.DraftCacher) error {
	draft.QuestionService = msg
	draft.Step++

	// Сохраняем обновленный черновик в кэш
	if err := cache.SaveDraft(GetContext(ctx), GetUserFromContext(ctx).TGUserID, draft); err != nil {
		return errors.Join(ErrSaveDraft, err)
	}

	return ctx.Send(MSG_ADD_CORRECT_ANSWER)
}

// processCorrectAnswerInputAndSaveQuestion обрабатывает ввод правильного ответа и сохраняет вопрос
// Завершает процесс создания вопроса и очищает черновик
func processCorrectAnswerInputAndSaveQuestion(
	ctx telebot.Context, draft *dto.QuestionDraft, msg string, u *edu.UserService, domain domain.Apper, cache domain.DraftCacher,
) error {
	draft.Answers = append(draft.Answers, msg)

	// Удаляем черновик после сохранения вопроса (даже если будет ошибка)
	defer cache.DeleteDraft(GetContext(ctx), u.TGUserID)

	if err := domain.SaveQuestions(
		GetContext(ctx), draft.QuestionService, draft.Tag, draft.Answers, u.TGUserID,
	); err != nil {
		return ctx.Send(errors.Join(ErrSave, err).Error())
	}

	return ctx.Send(MSG_SUCCESS, mainMenu())
}

// getTags получает список уникальных тегов пользователя и отображает их как inline-кнопки
// Если тегов нет, предлагает пользователю добавить новый тег
func getTags(ctx telebot.Context, userID int64, domain domain.Apper) error {
	ts, err := domain.GetUniqueTags(GetContext(ctx), userID)
	if err != nil {
		return ctx.Send(errors.Join(ErrGetTag, err).Error())
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
		if err = ctx.Send(MSG_ADD_TAG, &telebot.ReplyMarkup{
			InlineKeyboard: btns,
		}); err != nil {
			return ctx.Send(errors.Join(ErrGetTag, err).Error())
		}
		return nil
	}

	// Если тегов нет, просим добавить новый
	return ctx.Send(MSG_ADD_TAG)
}
