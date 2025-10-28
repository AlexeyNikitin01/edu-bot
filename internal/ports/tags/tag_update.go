package tags

import (
	"bot/internal/domain"
	"bot/internal/middleware"
	"bot/internal/repo/edu"
	"context"
	"fmt"
	"github.com/aarondl/sqlboiler/v4/boil"
	"gopkg.in/telebot.v3"
	"strconv"
)

func PauseTag(ctx context.Context, d domain.UseCases) telebot.HandlerFunc {
	return func(ctxBot telebot.Context) error {
		tagIDStr := ctxBot.Data()
		tagID, err := strconv.Atoi(tagIDStr)
		if err != nil {
			return err
		}

		tag, err := edu.Tags(
			edu.TagWhere.ID.EQ(int64(tagID)),
		).One(ctx, boil.GetContextDB())
		if err != nil {
			return err
		}

		tag.IsPause = !tag.IsPause
		if _, err = tag.Update(ctx, boil.GetContextDB(), boil.Whitelist(
			edu.TagColumns.IsPause,
		)); err != nil {
			return err
		}

		// Получаем текущее сообщение для определения страницы
		message := ctxBot.Message()
		if message == nil {
			return err
		}

		// Парсим номер страницы из текста сообщения
		currentPage := ExtractCurrentPage(message.Text)
		if currentPage == 0 {
			currentPage = 1
		}

		// Получаем теги для текущей страницы
		userID := middleware.GetUserFromContext(ctxBot).TGUserID
		pageSize := DEFAULT_PAGE_SIZE
		tags, totalCount, err := d.GetUniqueTags(ctx, userID, currentPage, pageSize)
		if err != nil {
			return err
		}

		// Создаем билдер с текущей страницей
		builder := NewTagButtonsBuilder(tags, totalCount).
			WithCurrentPage(currentPage)

		// Создаем обновленное сообщение
		updatedMessage := fmt.Sprintf("%s\n\n%s", MSG_LIST_TAGS, builder.GetPaginationInfo())

		return ctxBot.Edit(updatedMessage, &telebot.ReplyMarkup{
			InlineKeyboard: builder.BuildPageRows(),
		})
	}
}
