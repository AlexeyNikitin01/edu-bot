package ports

import (
	"fmt"
	"html"
	"strconv"
	"strings"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"gopkg.in/telebot.v3"

	"bot/internal/app"
	"bot/internal/repo/edu"
)

const (
	MSG_LIST_QUESTION = "ВОПРОСЫ: "
	MSG_LIST_TAGS     = "ТЭГИ: "
	MSG_EMPTY         = "У вас нет тэгов"
	MSG_BACK_TAGS     = "НАЗАД К ТЭГАМ"

	QuestionsPerPage = 10 // Оставляем место для кнопок пагинации и возврата
)

func showRepeatTagList(domain app.Apper) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {

		tagButtons, err := getButtonsTags(ctx, domain)
		if err != nil {
			return err
		}

		return ctx.Send(MSG_LIST_TAGS, &telebot.ReplyMarkup{
			InlineKeyboard: tagButtons,
		})
	}
}

func backTags(domain app.Apper) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {

		tagButtons, err := getButtonsTags(ctx, domain)
		if err != nil {
			return err
		}

		return ctx.Edit(MSG_LIST_TAGS, &telebot.ReplyMarkup{
			InlineKeyboard: tagButtons,
		})
	}
}

func getButtonsTags(ctx telebot.Context, domain app.Apper) ([][]telebot.InlineButton, error) {
	u := GetUserFromContext(ctx)

	tags, err := domain.GetUniqueTags(GetContext(ctx), u.TGUserID)
	if err != nil {
		return nil, sendErrorResponse(ctx, err.Error())
	}

	if len(tags) == 0 {
		return nil, sendErrorResponse(ctx, MSG_EMPTY)
	}

	var tagButtons [][]telebot.InlineButton

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

func questionByTag(tag string) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		return showQuestionsPage(ctx, tag, 0)
	}
}

func showQuestionsPage(ctx telebot.Context, tag string, page int) error {
	return ctx.Edit(fmt.Sprintf("%s %s (Стр. %d)", tag, MSG_LIST_QUESTION, page+1), &telebot.ReplyMarkup{
		InlineKeyboard: getQuestionBtns(ctx, tag, page),
	})
}

// handleToggleRepeat выбор учить или не учить вопрос.
func handleToggleRepeat(domain app.Apper) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		// Разбираем данные callback: "questionID_page_tag"
		parts := strings.Split(ctx.Data(), "_")
		if len(parts) < 3 {
			return sendErrorResponse(ctx, "Ошибка формата данных")
		}

		questionID, err := strconv.Atoi(parts[0])
		if err != nil {
			return sendErrorResponse(ctx, err.Error())
		}

		page, err := strconv.Atoi(parts[1])
		if err != nil {
			return sendErrorResponse(ctx, err.Error())
		}

		tag := strings.Join(parts[2:], "_")

		// Обновляем статус вопроса
		if err = domain.UpdateIsEduUserQuestion(GetContext(ctx), GetUserFromContext(ctx).TGUserID, int64(questionID)); err != nil {
			return sendErrorResponse(ctx, err.Error())
		}

		// Получаем обновленный список вопросов с сохранением текущей страницы
		return ctx.Edit(&telebot.ReplyMarkup{
			InlineKeyboard: getQuestionBtns(ctx, tag, page),
		})
	}
}

func getQuestionBtns(ctx telebot.Context, tag string, page int) [][]telebot.InlineButton {
	qs, err := edu.Questions(
		qm.InnerJoin(fmt.Sprintf("%s ON %s = %s", edu.TableNames.UsersQuestions,
			edu.QuestionTableColumns.ID,
			edu.UsersQuestionTableColumns.QuestionID,
		)),
		qm.InnerJoin(fmt.Sprintf("%s ON %s = %s", edu.TableNames.Tags,
			edu.TagTableColumns.ID,
			edu.QuestionTableColumns.TagID,
		)),
		edu.UsersQuestionWhere.UserID.EQ(GetUserFromContext(ctx).TGUserID),
		edu.TagWhere.Tag.EQ(tag),
		edu.UsersQuestionWhere.DeletedAt.IsNull(),
	).All(GetContext(ctx), boil.GetContextDB())
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

	var btns [][]telebot.InlineButton

	for _, q := range pageQuestions {
		questionButtons := getQuestionBtn(
			ctx,
			q.ID,
			INLINE_BTN_REPEAT_QUESTION,
			q.Question,
			INLINE_NAME_DELETE,
			INLINE_BTN_DELETE_QUESTION,
			page,
			tag,
		)
		btns = append(btns, []telebot.InlineButton{questionButtons[0]},
			[]telebot.InlineButton{questionButtons[1], questionButtons[2], questionButtons[3], questionButtons[4]})
	}

	// Добавляем кнопки пагинации, если нужно
	var paginationRow []telebot.InlineButton

	if page > 0 {
		paginationRow = append(paginationRow, telebot.InlineButton{
			Unique: INLINE_BTN_QUESTION_PAGE + "_prev",
			Text:   "⬅️ Назад",
			Data:   fmt.Sprintf("%d_%s", page, tag),
		})
	}

	// Кнопка возврата к тегам всегда в центре
	paginationRow = append(paginationRow, telebot.InlineButton{
		Unique: INLINE_BACK_TAGS,
		Text:   MSG_BACK_TAGS,
	})

	if page < totalPages-1 {
		paginationRow = append(paginationRow, telebot.InlineButton{
			Unique: INLINE_BTN_QUESTION_PAGE + "_next",
			Text:   "Вперед ➡️",
			Data:   fmt.Sprintf("%d_%s", page, tag),
		})
	}

	if len(paginationRow) > 0 {
		btns = append(btns, paginationRow)
	}

	return btns
}

func getQuestionBtn(
	ctx telebot.Context, qID int64, repeat, repeatMSG, deleteMSG, delete string, page int, tag string,
) []telebot.InlineButton {
	uq, err := edu.UsersQuestions(
		edu.UsersQuestionWhere.UserID.EQ(GetUserFromContext(ctx).TGUserID),
		edu.UsersQuestionWhere.QuestionID.EQ(qID),
		edu.UsersQuestionWhere.DeletedAt.IsNull(),
	).One(GetContext(ctx), boil.GetContextDB())
	if err != nil {
		return nil
	}

	makeData := func(qID int64, page int, tag string) string {
		if page == -1 && tag == "" {
			return fmt.Sprintf("%d", qID)
		}
		if page == -1 {
			return fmt.Sprintf("%d_%s", qID, tag)
		}
		if tag == "" {
			return fmt.Sprintf("%d_%d", qID, page)
		}
		return fmt.Sprintf("%d_%d_%s", qID, page, tag)
	}

	now := time.Now().UTC()
	duration := uq.TimeRepeat.Sub(now)

	questionText := telebot.InlineButton{
		Text: repeatMSG,
		Data: makeData(qID, page, tag),
	}

	label := "🔔"
	if uq.IsEdu {
		label = "💤"
	}

	repeatBtn := telebot.InlineButton{
		Unique: repeat,
		Text:   label,
		Data:   makeData(qID, page, tag),
	}

	deleteBtn := telebot.InlineButton{
		Unique: delete,
		Text:   deleteMSG,
		Data:   makeData(qID, page, tag),
	}

	editBtn := telebot.InlineButton{
		Unique: INLINE_EDIT_QUESTION,
		Text:   "✏️",
		Data:   fmt.Sprintf("%d", qID),
	}

	timeInline := telebot.InlineButton{
		Text: "⏳" + timeLeftMsg(duration),
	}

	return []telebot.InlineButton{questionText, repeatBtn, deleteBtn, editBtn, timeInline}
}

func getForUpdate(domain app.Apper) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		qID := ctx.Data()
		id, err := strconv.Atoi(qID)
		if err != nil {
			return sendErrorResponse(ctx, err.Error())
		}
		q, err := domain.GetQuestionAnswers(GetContext(ctx), int64(id))
		if err != nil {
			return sendErrorResponse(ctx, err.Error())
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

		return ctx.Send("Выберите поле: ", &telebot.ReplyMarkup{
			InlineKeyboard: btns,
		})
	}
}

// handlePageNavigation обрабатывает навигацию по страницам
func handlePageNavigation(ctx telebot.Context, pageOffset int) error {
	page, tag, err := parsePageAndTag(ctx.Data())
	if err != nil {
		return sendErrorResponse(ctx, err.Error())
	}
	return showQuestionsPage(ctx, tag, page+pageOffset)
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

// sendErrorResponse отправляет ответ с ошибкой
func sendErrorResponse(ctx telebot.Context, text string) error {
	return ctx.Respond(&telebot.CallbackResponse{
		Text: text,
	})
}

// showCurrentValue отображает текущее значение редактируемой сущности
func showCurrentValue(domain app.Apper, cache app.DraftCacher) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		user := GetUserFromContext(ctx)
		if user == nil {
			return ctx.Send("❌ Пользователь не найден")
		}

		// Получаем черновик из кэша
		draft, err := cache.GetDraft(GetContext(ctx), user.TGUserID)
		if err != nil {
			return ctx.Send("❌ Ошибка при получении черновика")
		}

		if draft == nil {
			return ctx.Send("❌ Черновик не найден. Начните редактирование заново.")
		}

		strID := ctx.Data()
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
			tag, err := domain.GetTagByID(GetContext(ctx), int64(id))
			if err != nil {
				return ctx.Send("❌ Не удалось загрузить тег")
			}
			currentValue = tag.Tag
			entityType = "тег"

		case draft.QuestionIDByName == int64(id):
			// Получаем вопрос
			question, err := domain.GetQuestionAnswers(GetContext(ctx), int64(id))
			if err != nil {
				return ctx.Send("❌ Не удалось загрузить вопрос")
			}
			currentValue = question.Question
			entityType = "вопрос"

		case draft.QuestionIDByTag == int64(id):
			// Получаем вопрос для изменения тега
			q, err := domain.GetQuestionAnswers(GetContext(ctx), int64(id))
			if err != nil {
				return ctx.Send("❌ Не удалось загрузить вопрос")
			}
			tag, err := domain.GetTagByID(GetContext(ctx), q.TagID)
			if err != nil {
				currentValue = "Тег не найден"
			} else {
				currentValue = tag.Tag
			}
			entityType = "тег вопроса"

		case draft.AnswerID == int64(id):
			// Получаем ответ
			answer, err := domain.GetAnswerByID(GetContext(ctx), int64(id))
			if err != nil {
				return ctx.Send("❌ Не удалось загрузить ответ")
			}
			currentValue = answer.Answer
			entityType = "ответ"

		default:
			return ctx.Send("❌ Неизвестная сущность для редактирования")
		}

		message := fmt.Sprintf("<b>Введите новое значение для или нажмите /cancel для отмены:</b>\n\n 📋 Текущее значение %s:\n\n<code>%s</code>💡",
			entityType,
			html.EscapeString(currentValue))

		// Создаем клавиатуру с кнопкой "Свернуть"
		menu := &telebot.ReplyMarkup{}
		btnCollapse := menu.Data("📁 Свернуть", INLINE_COLLAPSE_VALUE, strID)
		menu.Inline(menu.Row(btnCollapse))

		// Редактируем сообщение, заменяя кнопку на значение
		if ctx.Callback() != nil {
			return ctx.Edit(message, menu, telebot.ModeHTML)
		}

		return ctx.Send(message, menu, telebot.ModeHTML)
	}
}

// collapseValue скрывает значение и возвращает кнопку просмотра
func collapseValue(domain app.Apper) telebot.HandlerFunc {
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
