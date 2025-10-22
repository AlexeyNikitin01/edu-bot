package ports

import (
	"errors"
	"strconv"
	"strings"

	"gopkg.in/telebot.v3"

	"bot/internal/app"
	"bot/internal/repo/edu"
)

// Ошибки приложения
var (
	ErrGetTag = errors.New("ошибка получения тэгов")
	ErrSave   = errors.New("невозможно сохранить")
)

// Константы сообщений
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

// QuestionDraft представляет черновик вопроса для создания или редактирования
type QuestionDraft struct {
	Step             int      // Текущий шаг в процессе создания
	Question         string   // Текст вопроса
	Tag              string   // Тег вопроса
	Answers          []string // Список ответов
	TagID            int64    // ID тега для редактирования
	QuestionIDByTag  int64    // ID вопроса для изменения тега
	QuestionIDByName int64    // ID вопроса для изменения названия
	AnswerID         int64    // ID ответа для редактирования
}

// drafts хранит активные черновики пользователей по их ID
var drafts = make(map[int64]*QuestionDraft)

// setEdit инициализирует черновик редактирования для указанного поля
// field - тип сущности для редактирования (тег, вопрос, ответ)
// domain - слой приложения для работы с данными
func setEdit(field string, domain app.Apper) telebot.HandlerFunc {
	return func(ctx telebot.Context) (err error) {
		// Парсим ID сущности из данных callback
		strID := ctx.Data()
		id, err := strconv.Atoi(strID)
		if err != nil {
			return err
		}

		// Получаем или создаем черновик для пользователя
		user := GetUserFromContext(ctx)
		draft, exists := drafts[user.TGUserID]
		if !exists {
			drafts[user.TGUserID] = &QuestionDraft{Step: 1}
			draft, _ = drafts[user.TGUserID]
		}

		if draft == nil {
			return nil
		}

		// Устанавливаем соответствующий ID в зависимости от типа редактирования
		switch field {
		case edu.TableNames.Tags:
			draft.TagID = int64(id)
		case edu.QuestionTableColumns.Question:
			draft.QuestionIDByName = int64(id)
		case edu.QuestionTableColumns.TagID:
			draft.QuestionIDByTag = int64(id)
			// Для изменения тега вопроса показываем список доступных тегов
			if err = getTags(ctx, user.TGUserID, domain); err != nil {
				return err
			}
			return ctx.Send(MSG_EDIT_TAG_BY_QUESTION)
		case edu.AnswerTableColumns.Answer:
			draft.AnswerID = int64(id)
		}

		return ctx.Send(MSG_EDIT)
	}
}

// upsertUserQuestion обрабатывает создание или редактирование вопроса пользователя
// Объединяет логику создания нового вопроса и редактирования существующих сущностей
func upsertUserQuestion(domain app.Apper) telebot.HandlerFunc {
	return func(ctx telebot.Context) (err error) {
		msg := strings.TrimSpace(ctx.Message().Text)
		u := GetUserFromContext(ctx)

		// Получаем черновик пользователя
		draft, exists := drafts[u.TGUserID]
		if !exists {
			return initNewDraft(ctx, u, domain)
		}

		// Обработка отмены действия
		if msg == CMD_CANCEL {
			return cancelDraft(ctx, u)
		}

		// Приоритетная обработка черновиков редактирования
		if draft.TagID != 0 || draft.QuestionIDByName != 0 || draft.AnswerID != 0 || draft.QuestionIDByTag != 0 {
			return updateUserQuestion(ctx, draft, msg, u, domain)
		}

		// Обработка создания нового вопроса
		return createUserQuestion(ctx, draft, msg, u, domain)
	}
}

// initNewDraft инициализирует новый черновик для создания вопроса
// Показывает пользователю список доступных тегов
func initNewDraft(ctx telebot.Context, u *edu.User, domain app.Apper) error {
	drafts[u.TGUserID] = &QuestionDraft{Step: 1}
	if err := ctx.Send(MSG_LIST_TAGS); err != nil {
		return err
	}
	return getTags(ctx, u.TGUserID, domain)
}

// cancelDraft отменяет текущий черновик и очищает состояние
func cancelDraft(ctx telebot.Context, u *edu.User) error {
	delete(drafts, u.TGUserID)
	return ctx.Send(MSG_CANCEL)
}

// updateUserQuestion обрабатывает редактирование существующих сущностей (тегов, вопросов, ответов)
// Определяет тип редактирования и делегирует выполнение соответствующему обработчику
func updateUserQuestion(ctx telebot.Context, draft *QuestionDraft, msg string, u *edu.User, domain app.Apper) error {
	switch {
	case draft.TagID != 0:
		return updateTag(ctx, draft, msg, u, domain)
	case draft.QuestionIDByName != 0:
		return updateQuestionName(ctx, draft, msg, u, domain)
	case draft.AnswerID != 0:
		return updateAnswer(ctx, draft, msg, u, domain)
	case draft.QuestionIDByTag != 0:
		return updateTagByQuestion(ctx, draft, u, domain)
	}
	return nil
}

// updateTag обновляет текст существующего тега
func updateTag(ctx telebot.Context, draft *QuestionDraft, msg string, u *edu.User, domain app.Apper) error {
	if err := domain.UpdateTag(GetContext(ctx), draft.TagID, msg); err != nil {
		return err
	}
	delete(drafts, u.TGUserID)
	return ctx.Send(MSG_SUCCESS_UPDATE_TAG)
}

// updateQuestionName обновляет текст существующего вопроса
func updateQuestionName(ctx telebot.Context, draft *QuestionDraft, msg string, u *edu.User, domain app.Apper) error {
	if err := domain.UpdateQuestionName(GetContext(ctx), draft.QuestionIDByName, msg); err != nil {
		return err
	}
	delete(drafts, u.TGUserID)
	return ctx.Send(MSG_SUCCESS_UPDATE_NAME_QUESTION)
}

// updateAnswer обновляет текст существующего ответа
func updateAnswer(ctx telebot.Context, draft *QuestionDraft, msg string, u *edu.User, domain app.Apper) error {
	if err := domain.UpdateAnswer(GetContext(ctx), draft.AnswerID, msg); err != nil {
		return err
	}
	delete(drafts, u.TGUserID)
	return ctx.Send(MSG_SUCCESS_UPDATE_ANSWER)
}

// updateTagByQuestion обновляет тег для существующего вопроса
// Поддерживает выбор тега из списка или ввод нового
func updateTagByQuestion(ctx telebot.Context, draft *QuestionDraft, u *edu.User, domain app.Apper) error {
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
	delete(drafts, u.TGUserID)
	return ctx.Send(MSG_SUCCESS_UPDATE_TAG_BY_QUESTION)
}

// createUserQuestion обрабатывает процесс создания нового вопроса
// Последовательно проходит через шаги: выбор тега → ввод вопроса → ввод ответа
func createUserQuestion(ctx telebot.Context, draft *QuestionDraft, msg string, u *edu.User, domain app.Apper) error {
	switch draft.Step {
	case 1:
		return processTagSelection(ctx, draft)
	case 2:
		return processQuestionInput(ctx, draft, msg)
	case 3:
		return processCorrectAnswerInputAndSaveQuestion(ctx, draft, msg, u, domain)
	}
	return nil
}

// processTagSelection обрабатывает выбор тега для нового вопроса
// Поддерживает выбор из списка или ввод пользовательского тега
func processTagSelection(ctx telebot.Context, draft *QuestionDraft) error {
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
	return ctx.Send(MSG_ADD_QUESTION)
}

// processQuestionInput обрабатывает ввод текста вопроса
func processQuestionInput(ctx telebot.Context, draft *QuestionDraft, msg string) error {
	draft.Question = msg
	draft.Step++
	return ctx.Send(MSG_ADD_CORRECT_ANSWER)
}

// processCorrectAnswerInputAndSaveQuestion обрабатывает ввод правильного ответа и сохраняет вопрос
// Завершает процесс создания вопроса и очищает черновик
func processCorrectAnswerInputAndSaveQuestion(ctx telebot.Context, draft *QuestionDraft, msg string, u *edu.User, domain app.Apper) error {
	draft.Answers = append(draft.Answers, msg)
	defer delete(drafts, u.TGUserID)

	if err := domain.SaveQuestions(
		GetContext(ctx), draft.Question, draft.Tag, draft.Answers, u.TGUserID,
	); err != nil {
		return ctx.Send(errors.Join(ErrSave, err).Error())
	}

	return ctx.Send(MSG_SUCCESS, mainMenu())
}

// getTags получает список уникальных тегов пользователя и отображает их как inline-кнопки
// Если тегов нет, предлагает пользователю добавить новый тег
func getTags(ctx telebot.Context, userID int64, domain app.Apper) error {
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
