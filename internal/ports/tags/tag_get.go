package tags

import (
	"bot/internal/domain"
	"bot/internal/middleware"
	"context"
	"fmt"
	"gopkg.in/telebot.v3"
	"strconv"
)

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
			InlineKeyboard: builder.BuildPageRows(),
		})
	}
}
func HandleTagPagination(ctx context.Context, d domain.UseCases) telebot.HandlerFunc {
	return func(ctxBot telebot.Context) error {
		userID := middleware.GetUserFromContext(ctxBot).TGUserID

		// Получаем номер страницы из callback data
		pageStr := ctxBot.Data()
		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			page = 1
		}

		fmt.Println(page)

		tags, totalCount, err := d.GetUniqueTags(ctx, userID, page, DEFAULT_PAGE_SIZE)
		if err != nil {
			return err
		}

		builder := NewTagButtonsBuilder(tags, totalCount).
			WithCurrentPage(page)

		// Если запрошенная страница больше общей, корректируем
		if page > builder.totalPages {
			builder.WithCurrentPage(builder.totalPages)
		}

		message := fmt.Sprintf("%s\n\n%s", MSG_LIST_TAGS, builder.GetPaginationInfo())

		return ctxBot.Edit(message, &telebot.ReplyMarkup{
			InlineKeyboard: builder.BuildPageRows(),
		})
	}
}

// ShowEditTagList показывает список тегов для редактирования специальным сообщением
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

		// Создаем билдер только с тегами и общим количеством
		builder := NewTagButtonsBuilder(tags, totalCount).
			WithCurrentPage(page)

		// Создаем сообщение с информацией о пагинации из билдера + специальное сообщение для редактирования
		message := fmt.Sprintf("%s\n\n%s", MSG_EDIT_TAG_BY_QUESTION, builder.GetPaginationInfo())

		return ctxBot.Send(message, &telebot.ReplyMarkup{
			InlineKeyboard: builder.BuildPageRows(),
		})
	}
}
