package tags

import (
	"bot/internal/domain"
	"bot/internal/middleware"
	"context"
	"fmt"
	"gopkg.in/telebot.v3"
	"strconv"
)

// ShowEditTagList показывает список тегов для выбора при создании/редактировании вопроса
func ShowEditTagList(ctx context.Context, d domain.UseCases) telebot.HandlerFunc {
	return func(ctxBot telebot.Context) error {
		userID := middleware.GetUserFromContext(ctxBot).TGUserID

		// Получаем теги с пагинацией
		page := 1
		pageSize := DEFAULT_PAGE_SIZE
		tags, totalCount, err := d.GetUniqueTags(ctx, userID, page, pageSize)
		if err != nil {
			return err
		}

		// Создаем билдер
		builder := NewTagButtonsBuilder(tags, totalCount).
			WithCurrentPage(page)

		// Создаем сообщение с текстовым списком тегов
		message := fmt.Sprintf("%s\n\n%s", MSG_EDIT_TAG_BY_QUESTION, builder.BuildTextTags())

		// Отправляем сообщение с кнопками пагинации
		return ctxBot.Send(message, &telebot.ReplyMarkup{
			InlineKeyboard: builder.BuildTextRows(),
		})
	}
}

// ShowRepeatTagList показывает список тегов с кнопками управления
func ShowRepeatTagList(ctx context.Context, d domain.UseCases) telebot.HandlerFunc {
	return func(ctxBot telebot.Context) error {
		userID := middleware.GetUserFromContext(ctxBot).TGUserID

		tags, totalCount, err := d.GetUniqueTags(ctx, userID, 1, DEFAULT_PAGE_SIZE)
		if err != nil {
			return err
		}

		builder := NewTagButtonsBuilder(tags, totalCount).
			WithCurrentPage(1)

		message := fmt.Sprintf("%s\n\n%s", MSG_LIST_TAGS, builder.GetPaginationInfo())

		return ctxBot.Send(message, &telebot.ReplyMarkup{
			InlineKeyboard: builder.BuildPageRows(), // Используем полные кнопки управления
		})
	}
}

// HandleTagPagination обрабатывает пагинацию для тегов
func HandleTagPagination(ctx context.Context, d domain.UseCases) telebot.HandlerFunc {
	return func(ctxBot telebot.Context) error {
		userID := middleware.GetUserFromContext(ctxBot).TGUserID

		// Получаем номер страницы из callback data
		pageStr := ctxBot.Data()
		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			page = 1
		}

		// Получаем теги для запрошенной страницы
		tags, totalCount, err := d.GetUniqueTags(ctx, userID, page, DEFAULT_PAGE_SIZE)
		if err != nil {
			return err
		}

		// Создаем билдер
		builder := NewTagButtonsBuilder(tags, totalCount).
			WithCurrentPage(page)

		// Если запрошенная страница больше общей, корректируем
		if page > builder.totalPages {
			builder.WithCurrentPage(builder.totalPages)
		}

		// Определяем контекст: если это создание вопроса, используем текстовый режим
		draft, err := d.GetDraftQuestion(ctx, userID)
		isEditMode := err == nil && draft != nil && (draft.Step == 1 || draft.QuestionIDByTag != 0)

		var message string
		var keyboard [][]telebot.InlineButton

		if isEditMode {
			// Текстовый режим для выбора тегов
			message = fmt.Sprintf("%s\n\n%s", MSG_EDIT_TAG_BY_QUESTION, builder.BuildTextTags())
			keyboard = builder.BuildTextRows()
		} else {
			// Полный режим с кнопками управления
			message = fmt.Sprintf("%s\n\n%s", MSG_LIST_TAGS, builder.GetPaginationInfo())
			keyboard = builder.BuildPageRows()
		}

		return ctxBot.Edit(message, &telebot.ReplyMarkup{
			InlineKeyboard: keyboard,
		})
	}
}
