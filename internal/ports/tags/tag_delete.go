package tags

import (
	"bot/internal/domain"
	"bot/internal/middleware"
	"context"
	"fmt"
	"gopkg.in/telebot.v3"
)

func DeleteQuestionByTag(ctx context.Context, d domain.UseCases) telebot.HandlerFunc {
	return func(ctxBot telebot.Context) error {
		tag := ctxBot.Data()
		userID := middleware.GetUserFromContext(ctxBot).TGUserID

		// Получаем текущее сообщение для определения страницы
		message := ctxBot.Message()
		if message == nil {
			return nil
		}

		// Парсим номер страницы из текста сообщения
		currentPage := ExtractCurrentPage(message.Text)
		if currentPage == 0 {
			currentPage = 1
		}

		// Удаляем вопросы по тегу
		if err := d.DeleteQuestionsByTag(ctx, userID, tag); err != nil {
			return err
		}

		// Получаем обновленный список тегов
		pageSize := DEFAULT_PAGE_SIZE
		tags, totalCount, err := d.GetUniqueTags(ctx, userID, currentPage, pageSize)
		if err != nil {
			return err
		}

		// Если текущая страница пустая после удаления, переходим на предыдущую страницу
		if len(tags) == 0 && currentPage > 1 {
			currentPage--
			// Перезапрашиваем теги для предыдущей страницы
			tags, totalCount, err = d.GetUniqueTags(ctx, userID, currentPage, pageSize)
			if err != nil {
				return err
			}
		}

		// Создаем билдер
		builder := NewTagButtonsBuilder(tags, totalCount).
			WithCurrentPage(currentPage).
			WithPageSize(pageSize)

		// Создаем обновленное сообщение
		updatedMessage := fmt.Sprintf("%s\n\n%s", MSG_LIST_TAGS, builder.GetPaginationInfo())

		// Обновляем сообщение
		if err := ctxBot.Edit(updatedMessage, &telebot.ReplyMarkup{
			InlineKeyboard: builder.BuildPageRows(),
		}); err != nil {
			return err
		}

		// Показываем подтверждение удаления
		return ctxBot.Respond(&telebot.CallbackResponse{
			Text:      fmt.Sprintf("✅ Вопросы с тегом '%s' удалены", tag),
			ShowAlert: true,
		})
	}
}
