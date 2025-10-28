package question

import (
	"bot/internal/domain"
	"bot/internal/middleware"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/telebot.v3"
	"html"
	"log"
	"strconv"
	"strings"
	"time"
)

func QuestionByTag(ctx context.Context, data string, d domain.UseCases) telebot.HandlerFunc {
	return func(ctxBot telebot.Context) error {
		tag, tagPage, err := parsePageString(data)
		if err != nil {
			return err
		}
		userID := middleware.GetUserFromContext(ctxBot).TGUserID
		return showQuestionsPage(ctx, ctxBot, tag, 0, userID, d, tagPage)
	}
}

func showQuestionsPage(
	ctx context.Context, ctxBot telebot.Context, tag string, page int, userID int64, d domain.UseCases, tagPage int,
) error {
	// Получаем вопросы с пагинацией
	questions, totalCount, err := d.GetAllQuestionsWithPagination(ctx, userID, tag, QuestionsPerPage, page)
	if err != nil {
		return err
	}

	// Создаем билдер с опциями
	builder := NewQuestionButtonBuilder(
		WithQuestions(questions),
		WithTotalCount(totalCount),
		WithPage(page),
		WithTag(tag),
		WithTagPage(tagPage),
	)

	// Получаем сообщение и клавиатуру из билдера
	message, keyboard := builder.BuildQuestionsPage()

	if ctxBot.Callback() != nil {
		return ctxBot.Edit(message, &telebot.ReplyMarkup{
			InlineKeyboard: keyboard,
		})
	}

	return ctxBot.Send(message, &telebot.ReplyMarkup{
		InlineKeyboard: keyboard,
	})
}

// IsRepeat выбор учить или не учить вопрос.
func IsRepeat(ctx context.Context, d domain.UseCases) telebot.HandlerFunc {
	return func(ctxBot telebot.Context) error {
		userID := middleware.GetUserFromContext(ctxBot).TGUserID

		// Разбираем данные callback: "questionID_page_tag"
		parts := strings.Split(ctxBot.Data(), "_")
		if len(parts) < 3 {
			return errors.New("invalid command")
		}

		questionID, err := strconv.Atoi(parts[0])
		if err != nil {
			return err
		}

		page, err := strconv.Atoi(parts[1])
		if err != nil {
			return err
		}

		tag := strings.Join(parts[2:], "_")

		// Обновляем статус вопроса
		if err = d.UpdateIsEduUserQuestion(ctx, userID, int64(questionID)); err != nil {
			return err
		}

		// Получаем обновленный список вопросов с сохранением текущей страницы
		return showQuestionsPage(ctx, ctxBot, tag, page, userID, d, 0)
	}
}

// HandlePageNavigation обрабатывает навигацию по страницам
func HandlePageNavigation(ctx context.Context, d domain.UseCases) telebot.HandlerFunc {
	return func(ctxBot telebot.Context) error {
		userID := middleware.GetUserFromContext(ctxBot).TGUserID
		page, tag, tagPage, err := parsePageAndTag(ctxBot.Data())
		if err != nil {
			return err
		}
		return showQuestionsPage(ctx, ctxBot, tag, page, userID, d, tagPage)
	}
}

// parsePageAndTag парсит данные callback'а в формате "номер_тег_страницаТега" и возвращает номер страницы, тег и страницу тега
func parsePageAndTag(data string) (int, string, int, error) {
	dataParts := strings.Split(data, "_")
	if len(dataParts) != 3 {
		return 0, "", 0, fmt.Errorf("неверный формат данных: ожидается формат 'номер_тег_страницаТега'")
	}

	// Парсим основной номер страницы
	page, err := strconv.Atoi(dataParts[0])
	if err != nil {
		return 0, "", 0, fmt.Errorf("неверный номер страницы: %v", err)
	}

	// Получаем тег
	tag := dataParts[1]
	if tag == "" {
		return 0, "", 0, fmt.Errorf("не указан тег")
	}

	// Парсим номер страницы тега
	tagPage, err := strconv.Atoi(dataParts[2])
	if err != nil {
		return 0, "", 0, fmt.Errorf("неверный номер страницы тега: %v", err)
	}

	return page, tag, tagPage, nil
}

func GetForUpdate(ctx context.Context, domain domain.UseCases) telebot.HandlerFunc {
	return func(ctxBot telebot.Context) error {
		qID := ctxBot.Data()
		id, err := strconv.Atoi(qID)
		if err != nil {
			return err
		}
		q, err := domain.GetQuestionAnswers(ctx, int64(id))
		if err != nil {
			return err
		}

		var btns [][]telebot.InlineButton

		editQuestion := telebot.InlineButton{
			Unique: INLINE_EDIT_NAME_QUESTION,
			Text:   "вопрос: " + q.Question,
			Data:   fmt.Sprintf("%d", id),
		}

		editTag := telebot.InlineButton{
			Unique: INLINE_EDIT_NAME_TAG_QUESTION,
			Text:   "тэг: " + q.R.GetTag().Tag,
			Data:   fmt.Sprintf("%d", id),
		}

		btns = append(btns, []telebot.InlineButton{editQuestion})
		btns = append(btns, []telebot.InlineButton{editTag})

		for _, a := range q.R.GetAnswers() {
			answer := telebot.InlineButton{
				Unique: INLINE_EDIT_ANSWER_QUESTION,
				Text:   "ответ: " + a.Answer,
				Data:   fmt.Sprintf("%d", a.ID),
			}
			btns = append(btns, []telebot.InlineButton{answer})
		}

		return ctxBot.Send("Выберите поле: ", &telebot.ReplyMarkup{
			InlineKeyboard: btns,
		})
	}
}

// ShowCurrentValue отображает текущее значение редактируемой сущности
func ShowCurrentValue(ctx context.Context, d domain.UseCases) telebot.HandlerFunc {
	return func(ctxBot telebot.Context) error {
		userID := middleware.GetUserFromContext(ctxBot).TGUserID

		draft, err := d.GetDraftQuestion(ctx, userID)
		if err != nil {
			return err
		}

		if draft == nil {
			return err
		}

		strID := ctxBot.Data()
		id, err := strconv.Atoi(strID)
		if err != nil {
			return err
		}

		var currentValue string
		var entityType string

		// Определяем, какая сущность редактируется и получаем ее текущее значение
		switch {
		case draft.TagID == int64(id):
			// Получаем тег
			tag, err := d.GetTagByID(ctx, int64(id))
			if err != nil {
				return ctxBot.Send("❌ Не удалось загрузить тег")
			}
			currentValue = tag.Tag
			entityType = "тег"

		case draft.QuestionIDByName == int64(id):
			// Получаем вопрос
			question, err := d.GetQuestionAnswers(ctx, int64(id))
			if err != nil {
				return ctxBot.Send("❌ Не удалось загрузить вопрос")
			}
			currentValue = question.Question
			entityType = "вопрос"

		case draft.QuestionIDByTag == int64(id):
			// Получаем вопрос для изменения тега
			q, err := d.GetQuestionAnswers(ctx, int64(id))
			if err != nil {
				return ctxBot.Send("❌ Не удалось загрузить вопрос")
			}
			tag, err := d.GetTagByID(ctx, q.TagID)
			if err != nil {
				currentValue = "Тег не найден"
			} else {
				currentValue = tag.Tag
			}
			entityType = "тег вопроса"

		case draft.AnswerID == int64(id):
			// Получаем ответ
			answer, err := d.GetAnswerByID(ctx, int64(id))
			if err != nil {
				return ctxBot.Send("❌ Не удалось загрузить ответ")
			}
			currentValue = answer.Answer
			entityType = "ответ"

		default:
			return ctxBot.Send("❌ Неизвестная сущность для редактирования")
		}

		message := fmt.Sprintf("<b>Введите новое значение для или нажмите /cancel для отмены:</b>\n\n 📋 Текущее значение %s:\n\n<code>%s</code>💡",
			entityType,
			html.EscapeString(currentValue))

		// Создаем клавиатуру с кнопкой "Свернуть"
		menu := &telebot.ReplyMarkup{}
		btnCollapse := menu.Data("📁 Свернуть", INLINE_COLLAPSE_VALUE, strID)
		menu.Inline(menu.Row(btnCollapse))

		// Редактируем сообщение, заменяя кнопку на значение
		if ctxBot.Callback() != nil {
			return ctxBot.Edit(message, menu, telebot.ModeHTML)
		}

		return ctxBot.Send(message, menu, telebot.ModeHTML)
	}
}

// CollapseValue скрывает значение и возвращает кнопку просмотра
func CollapseValue(ctx context.Context, domain domain.UseCases) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		strID := ctx.Data()

		// Создаем клавиатуру с кнопкой для просмотра текущего значения
		menu := &telebot.ReplyMarkup{}
		btnShowCurrent := menu.Data("👀 Посмотреть текущее значение", INLINE_SHOW_CURRENT_VALUE, strID)
		menu.Inline(menu.Row(btnShowCurrent))

		// Редактируем сообщение, возвращая исходное состояние
		return ctx.Edit("Действие", menu, telebot.ModeHTML)
	}
}

func ViewAnswer(ctx context.Context, d domain.UseCases, showAnswer bool) telebot.HandlerFunc {
	return func(ctxBot telebot.Context) error {
		// Разбираем данные callback: "questionID_page_tag"
		parts := strings.Split(ctxBot.Data(), "_")
		if len(parts) < 3 {
			return errors.New("invalid command")
		}

		questionID, err := strconv.Atoi(parts[0])
		if err != nil {
			return err
		}

		page, err := strconv.Atoi(parts[1])
		if err != nil {
			return err
		}

		tag := strings.Join(parts[2:], "_")

		userID := middleware.GetUserFromContext(ctxBot).TGUserID

		uq, err := d.GetUserQuestion(ctx, userID, int64(questionID))
		if err != nil {
			return err
		}

		question := uq.GetQuestion().Question
		tagName := uq.R.GetQuestion().R.GetTag().Tag
		answer := uq.R.GetQuestion().R.GetAnswers()[0]

		result := EscapeMarkdown(tagName) + ": " + EscapeMarkdown(question)
		if showAnswer {
			result += "\n\n" + EscapeMarkdown(answer.Answer)
		}

		// Создаем билдер для кнопок ответа с опциями
		builder := NewQuestionButtonBuilder(
			WithPage(page),
			WithTag(tag),
		)

		return ctxBot.Edit(
			result,
			telebot.ModeMarkdownV2,
			&telebot.ReplyMarkup{
				InlineKeyboard: builder.BuildFullKeyboard(uq, showAnswer),
			},
		)
	}
}

func NextQuestion(ctx context.Context, d domain.UseCases) telebot.HandlerFunc {
	return func(ctxBot telebot.Context) error {
		if err := ctxBot.Send(MSG_NEXT_QUESTION); err != nil {
			return err
		}

		userID := middleware.GetUserFromContext(ctxBot).TGUserID

		t, err := d.GetNearestTimeRepeat(ctx, userID)
		if err != nil {
			return ctxBot.Respond(&telebot.CallbackResponse{Text: err.Error()})
		}

		now := time.Now().UTC()
		if !now.After(t) {
			duration := t.Sub(now)
			msg := fmt.Sprintf("⏳ Следующий вопрос будет доступен через: %s", timeLeftMsg(duration))

			if err = ctxBot.Send(msg, telebot.ModeMarkdown); err != nil {
				return ctxBot.Respond(&telebot.CallbackResponse{Text: err.Error()})
			}
		}

		if err = d.SetUserWaiting(ctx, userID, false); err != nil {
			log.Printf("Ошибка сброса статуса waiting в Redis для пользователя %d: %v", userID, err)
		}

		return nil
	}
}
