package question

import (
	"bot/internal/domain"
	"bot/internal/middleware"
	"bot/internal/repo/edu"
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

const (
	MSG_LIST_QUESTION = "ВОПРОСЫ: "
	MSG_LIST_TAGS     = "ТЭГИ: "
	MSG_EMPTY         = "У вас нет тэгов"
	MSG_BACK_TAGS     = "НАЗАД К ТЭГАМ"

	QuestionsPerPage = 10 // Оставляем место для кнопок пагинации и возврата

)

func ShowRepeatTagList(ctx context.Context, domain domain.UseCases) telebot.HandlerFunc {
	return func(ctxBot telebot.Context) error {

		tagButtons, err := getButtonsTags(ctx, ctxBot, domain)
		if err != nil {
			return err
		}

		return ctxBot.Send(MSG_LIST_TAGS, &telebot.ReplyMarkup{
			InlineKeyboard: tagButtons,
		})
	}
}

func BackTags(ctx context.Context, d domain.UseCases) telebot.HandlerFunc {
	return func(ctxBot telebot.Context) error {

		tagButtons, err := getButtonsTags(ctx, ctxBot, d)
		if err != nil {
			return err
		}

		return ctxBot.Edit(MSG_LIST_TAGS, &telebot.ReplyMarkup{
			InlineKeyboard: tagButtons,
		})
	}
}

func getButtonsTags(ctx context.Context, ctxBot telebot.Context, d domain.UseCases) ([][]telebot.InlineButton, error) {
	userID := middleware.GetUserFromContext(ctxBot).TGUserID

	tags, err := d.GetUniqueTags(ctx, userID)
	if err != nil {
		return nil, err
	}

	if len(tags) == 0 {
		return nil, nil
	}

	var tagButtons [][]telebot.InlineButton

	// todo кнопки
	for _, tag := range tags {
		tagBtn := telebot.InlineButton{
			Unique: INLINE_BTN_QUESTION_BY_TAG,
			Text:   tag.Tag,
			Data:   tag.Tag,
		}
		deleteBtn := telebot.InlineButton{
			Unique: INLINE_BTN_DELETE_QUESTIONS_BY_TAG,
			Text:   INLINE_NAME_DELETE,
			Data:   tag.Tag,
		}
		editBtn := telebot.InlineButton{
			Unique: INLINE_EDIT_TAG,
			Text:   "✏️",
			Data:   fmt.Sprintf("%d", tag.ID),
		}

		label := "🔔"
		if !tag.IsPause {
			label = "💤"
		}

		pauseTag := telebot.InlineButton{
			Unique: INLINE_PAUSE_TAG,
			Text:   label,
			Data:   fmt.Sprintf("%d", tag.ID),
		}

		tagButtons = append(tagButtons, []telebot.InlineButton{tagBtn, deleteBtn, editBtn, pauseTag})
	}

	return tagButtons, nil
}

func QuestionByTag(ctx context.Context, tag string, d domain.UseCases) telebot.HandlerFunc {
	return func(ctxBot telebot.Context) error {
		userID := middleware.GetUserFromContext(ctxBot).TGUserID
		return showQuestionsPage(ctx, ctxBot, tag, 0, userID, d)
	}
}

func showQuestionsPage(
	ctx context.Context, ctxBot telebot.Context, tag string, page int, userID int64, d domain.UseCases,
) error {
	return ctxBot.Edit(fmt.Sprintf("%s %s (Стр. %d)", tag, MSG_LIST_QUESTION, page+1), &telebot.ReplyMarkup{
		InlineKeyboard: getQuestionBtns(ctx, ctxBot, d, tag, page, userID),
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
		return ctxBot.Edit(&telebot.ReplyMarkup{
			InlineKeyboard: getQuestionBtns(ctx, ctxBot, d, tag, page, userID),
		})
	}
}

func getQuestionBtns(
	ctx context.Context, ctxBot telebot.Context, d domain.UseCases, tag string, page int, userID int64,
) [][]telebot.InlineButton {
	qs, err := d.GetAllQuestions(ctx, userID, tag)
	if err != nil || len(qs) == 0 {
		return nil
	}

	totalPages := (len(qs) + QuestionsPerPage - 1) / QuestionsPerPage
	if page >= totalPages {
		page = totalPages - 1
	}
	if page < 0 {
		page = 0
	}

	start := page * QuestionsPerPage
	end := start + QuestionsPerPage
	if end > len(qs) {
		end = len(qs)
	}
	pageQuestions := qs[start:end]

	// Получаем UsersQuestion для каждого вопроса
	userQuestions := make(map[int64]*edu.UsersQuestion)
	for _, q := range pageQuestions {
		uq, err := d.GetUserQuestion(ctx, userID, q.ID)
		if err == nil {
			userQuestions[q.ID] = uq
		}
	}

	builder := NewQuestionButtonBuilder()

	// Создаем клавиатуру с вопросами
	btns := builder.BuildQuestionsKeyboard(pageQuestions, userQuestions, page, tag)

	// Добавляем кнопки пагинации, если нужно
	if totalPages > 1 {
		paginationRow := builder.BuildPaginationButtons(page, totalPages, tag)
		if len(paginationRow) > 0 {
			btns = append(btns, paginationRow)
		}
	}

	return btns
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

// HandlePageNavigation обрабатывает навигацию по страницам
func HandlePageNavigation(ctx context.Context, ctxBot telebot.Context, pageOffset int, d domain.UseCases) error {
	userID := middleware.GetUserFromContext(ctxBot).TGUserID
	page, tag, err := parsePageAndTag(ctxBot.Data())
	if err != nil {
		return err
	}
	return showQuestionsPage(ctx, ctxBot, tag, page+pageOffset, userID, d)
}

// parsePageAndTag парсит данные callback'а и возвращает номер страницы и тег
func parsePageAndTag(data string) (int, string, error) {
	dataParts := strings.Split(data, "_")
	if len(dataParts) != 2 {
		return 0, "", fmt.Errorf("Ошибка: неверный формат данных")
	}

	page, err := strconv.Atoi(dataParts[0])
	if err != nil {
		return 0, "", fmt.Errorf("Ошибка: неверный номер страницы")
	}

	tag := dataParts[1]
	if tag == "" {
		return 0, "", fmt.Errorf("Ошибка: не указан тег")
	}

	return page, tag, nil
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
		return ctx.Edit(MSG_EDIT, menu, telebot.ModeHTML)
	}
}

func ViewAnswer(ctx context.Context, d domain.UseCases, showAnswer bool) telebot.HandlerFunc {
	return func(ctxBot telebot.Context) error {
		data := ctxBot.Data()
		qID, err := strconv.Atoi(data)
		if err != nil {
			return err
		}

		userID := middleware.GetUserFromContext(ctxBot).TGUserID

		uq, err := d.GetUserQuestion(ctx, userID, int64(qID))
		if err != nil {
			return err
		}

		question := uq.GetQuestion().Question
		tag := uq.R.GetQuestion().R.GetTag().Tag
		answer := uq.R.GetQuestion().R.GetAnswers()[0]

		result := EscapeMarkdown(tag) + ": " + EscapeMarkdown(question)
		if showAnswer {
			result += "\n\n" + EscapeMarkdown(answer.Answer)
		}

		return ctxBot.Edit(
			result,
			telebot.ModeMarkdownV2,
			&telebot.ReplyMarkup{
				InlineKeyboard: NewQuestionButtonBuilder().BuildFullKeyboard(uq, showAnswer),
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
