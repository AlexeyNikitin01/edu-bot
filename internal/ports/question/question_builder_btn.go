package question

import (
	"bot/internal/repo/edu"
	"fmt"
	"gopkg.in/telebot.v3"
	"time"
)

// QuestionButtonBuilder отвечает за создание интерактивных кнопок для вопросов
type QuestionButtonBuilder struct {
	questions  edu.UsersQuestionSlice
	totalCount int
	page       int
	tag        string
	tagPage    int
}

// BuilderOptions опции для создания билдера
type BuilderOptions struct {
	Questions  edu.UsersQuestionSlice
	TotalCount int
	Page       int
	Tag        string
	TagPage    int
}

// BuilderOption функция для настройки опций
type BuilderOption func(*BuilderOptions)

// NewQuestionButtonBuilder создает новый экземпляр билдера кнопок с опциями
func NewQuestionButtonBuilder(opts ...BuilderOption) *QuestionButtonBuilder {
	// Устанавливаем значения по умолчанию
	options := &BuilderOptions{
		Questions:  edu.UsersQuestionSlice{},
		TotalCount: 0,
		Page:       0,
		Tag:        "",
		TagPage:    0,
	}

	// Применяем переданные опции
	for _, opt := range opts {
		opt(options)
	}

	return &QuestionButtonBuilder{
		questions:  options.Questions,
		totalCount: options.TotalCount,
		page:       options.Page,
		tag:        options.Tag,
		tagPage:    options.TagPage,
	}
}

// WithQuestions устанавливает вопросы
func WithQuestions(questions edu.UsersQuestionSlice) BuilderOption {
	return func(o *BuilderOptions) {
		o.Questions = questions
	}
}

// WithTotalCount устанавливает общее количество вопросов
func WithTotalCount(totalCount int) BuilderOption {
	return func(o *BuilderOptions) {
		o.TotalCount = totalCount
	}
}

// WithTagPage устанавливает номер страницы тэга
func WithTagPage(totalCount int) BuilderOption {
	return func(o *BuilderOptions) {
		o.TagPage = totalCount
	}
}

// WithPage устанавливает текущую страницу
func WithPage(page int) BuilderOption {
	return func(o *BuilderOptions) {
		o.Page = page
	}
}

// WithTag устанавливает тег
func WithTag(tag string) BuilderOption {
	return func(o *BuilderOptions) {
		o.Tag = tag
	}
}

// BuildQuestionsPage создает полную страницу с вопросами и пагинацией
func (b *QuestionButtonBuilder) BuildQuestionsPage() (string, [][]telebot.InlineButton) {
	// Проверяем есть ли вопросы
	if b.totalCount == 0 {
		message := fmt.Sprintf("📭 По тегу '%s' нет вопросов", b.tag)
		keyboard := b.BuildEmptyStateKeyboard()
		return message, keyboard
	}

	totalPages := b.totalPages()
	message := fmt.Sprintf("%s %s (Стр. %d/%d)", b.tag, MSG_LIST_QUESTION, b.page+1, totalPages)

	keyboard := b.BuildQuestionsKeyboard()

	return message, keyboard
}

// BuildQuestionsKeyboard создает только клавиатуру с вопросами и пагинацией
func (b *QuestionButtonBuilder) BuildQuestionsKeyboard() [][]telebot.InlineButton {
	var btns [][]telebot.InlineButton

	// Если нет вопросов, возвращаем клавиатуру с кнопкой возврата
	if b.totalCount == 0 {
		return b.BuildEmptyStateKeyboard()
	}

	// Добавляем кнопки вопросов
	for _, q := range b.questions {
		questionRows := b.BuildQuestionRow(q)
		btns = append(btns, questionRows...)
	}

	// Добавляем пагинацию
	if b.totalPages() > 1 {
		paginationRow := b.BuildPaginationButtons()
		btns = append(btns, paginationRow)
	} else {
		// Если всего одна страница, добавляем только кнопку возврата к тегам
		backRow := b.BuildBackToTagsButton()
		btns = append(btns, backRow)
	}

	return btns
}

// BuildEmptyStateKeyboard создает клавиатуру для пустого состояния
func (b *QuestionButtonBuilder) BuildEmptyStateKeyboard() [][]telebot.InlineButton {
	return [][]telebot.InlineButton{
		b.BuildBackToTagsButton(),
	}
}

// BuildBackToTagsButton создает кнопку возврата к тегам
func (b *QuestionButtonBuilder) BuildBackToTagsButton() []telebot.InlineButton {
	return []telebot.InlineButton{
		{
			Unique: "back_to_tags",
			Text:   MSG_BACK_TAGS,
			Data:   fmt.Sprintf("%d", b.tagPage),
		},
	}
}

// BuildQuestionRow создает ряд кнопок для одного вопроса
func (b *QuestionButtonBuilder) BuildQuestionRow(uq *edu.UsersQuestion) [][]telebot.InlineButton {
	return [][]telebot.InlineButton{
		{b.BuildQuestionTextButton(uq)},
		{
			b.BuildRepeatButtonList(uq),
			b.BuildDeleteButtonList(uq),
			b.BuildEditButton(uq),
			b.BuildTimeButton(uq),
		},
	}
}

// BuildInList создает полную клавиатуру для вопроса (для ViewAnswer)
func (b *QuestionButtonBuilder) BuildInList(uq *edu.UsersQuestion, showAnswer bool) [][]telebot.InlineButton {
	return [][]telebot.InlineButton{
		b.BuildAnswerRow(uq, showAnswer),
		b.BuildDifficultyRow(uq),
		b.BuildActionsRowList(uq),
		b.BuildBackToTagsButton(),
	}
}

func (b *QuestionButtonBuilder) BuildAfterSend(
	uq *edu.UsersQuestion, showAnswer bool,
) [][]telebot.InlineButton {
	return [][]telebot.InlineButton{
		b.BuildAnswerRow(uq, showAnswer),
		b.BuildDifficultyRow(uq),
		b.BuildActionsRow(uq),
	}
}

// BuildAnswerRow создает ряд с кнопкой ответа
func (b *QuestionButtonBuilder) BuildAnswerRow(uq *edu.UsersQuestion, showAnswer bool) []telebot.InlineButton {
	return []telebot.InlineButton{
		b.BuildAnswerButton(uq, showAnswer),
	}
}

// BuildDifficultyRow создает ряд с кнопками оценки сложности
func (b *QuestionButtonBuilder) BuildDifficultyRow(uq *edu.UsersQuestion) []telebot.InlineButton {
	return []telebot.InlineButton{
		b.WithPrefixEmoji("😎", b.BuildEasyButton(uq)),
		b.WithPrefixEmoji("😵", b.BuildForgotButton(uq)),
	}
}

// BuildActionsRowList создает ряд с кнопками действий (повторение, удаление, редактирование) в списке
func (b *QuestionButtonBuilder) BuildActionsRowList(uq *edu.UsersQuestion) []telebot.InlineButton {
	return []telebot.InlineButton{
		b.BuildRepeatButton(uq),
		b.BuildDeleteButton(uq),
		b.BuildEditButton(uq),
	}
}

// BuildRepeatButtonList создает кнопку повторения вопроса в списке
func (b *QuestionButtonBuilder) BuildRepeatButtonList(uq *edu.UsersQuestion) telebot.InlineButton {
	label := "🔔"
	if uq.IsEdu {
		label = "💤"
	}

	return telebot.InlineButton{
		Unique: INLINE_BTN_REPEAT_QUESTION,
		Text:   label,
		Data:   b.makeData(uq.QuestionID),
	}
}

// BuildDeleteButtonList создает кнопку удаления вопроса
func (b *QuestionButtonBuilder) BuildDeleteButtonList(uq *edu.UsersQuestion) telebot.InlineButton {
	return telebot.InlineButton{
		Unique: INLINE_BTN_DELETE_QUESTION,
		Text:   INLINE_NAME_DELETE,
		Data:   b.makeData(uq.QuestionID),
	}
}

// BuildEditButtonList создает кнопку редактирования в списке
func (b *QuestionButtonBuilder) BuildEditButtonList(uq *edu.UsersQuestion) telebot.InlineButton {
	return telebot.InlineButton{
		Unique: INLINE_EDIT_QUESTION,
		Text:   "✏️",
		Data:   b.makeData(uq.QuestionID),
	}
}

// BuildActionsRow создает ряд с кнопками действий (повторение, удаление, редактирование)
func (b *QuestionButtonBuilder) BuildActionsRow(uq *edu.UsersQuestion) []telebot.InlineButton {
	return []telebot.InlineButton{
		b.BuildRepeatButton(uq),
		b.BuildDeleteButton(uq),
		b.BuildEditButton(uq),
	}
}

// BuildAnswerButton создает кнопку показа/скрытия ответа
func (b *QuestionButtonBuilder) BuildAnswerButton(uq *edu.UsersQuestion, showAnswer bool) telebot.InlineButton {
	if showAnswer {
		return telebot.InlineButton{
			Unique: INLINE_TURN_ANSWER,
			Text:   "📝 Свернуть ответ",
			Data:   b.makeData(uq.QuestionID),
		}
	}

	return telebot.InlineButton{
		Unique: INLINE_SHOW_ANSWER,
		Text:   BtnShowAnswer,
		Data:   b.makeData(uq.QuestionID),
	}
}

// BuildEasyButton создает кнопку "ЛЕГКО" для оценки сложности вопроса
func (b *QuestionButtonBuilder) BuildEasyButton(uq *edu.UsersQuestion) telebot.InlineButton {
	return telebot.InlineButton{
		Unique: INLINE_REMEMBER_HIGH_QUESTION,
		Text:   MSG_REMEMBER,
		Data:   fmt.Sprintf("%d", uq.QuestionID),
	}
}

// BuildForgotButton создает кнопку "СЛОЖНО" для оценки сложности вопроса
func (b *QuestionButtonBuilder) BuildForgotButton(uq *edu.UsersQuestion) telebot.InlineButton {
	return telebot.InlineButton{
		Unique: INLINE_FORGOT_HIGH_QUESTION,
		Text:   MSG_FORGOT,
		Data:   fmt.Sprintf("%d", uq.QuestionID),
	}
}

// BuildRepeatButton создает кнопку повторения вопроса
func (b *QuestionButtonBuilder) BuildRepeatButton(uq *edu.UsersQuestion) telebot.InlineButton {
	label := "🔔"
	if uq.IsEdu {
		label = "💤"
	}

	return telebot.InlineButton{
		Unique: INLINE_BTN_REPEAT_QUESTION_AFTER_POLL_HIGH,
		Text:   label,
		Data:   fmt.Sprintf("%d", uq.QuestionID),
	}
}

// BuildDeleteButton создает кнопку удаления вопроса
func (b *QuestionButtonBuilder) BuildDeleteButton(uq *edu.UsersQuestion) telebot.InlineButton {
	return telebot.InlineButton{
		Unique: INLINE_BTN_DELETE_QUESTION_AFTER_POLL_HIGH,
		Text:   INLINE_NAME_DELETE,
		Data:   fmt.Sprintf("%d", uq.QuestionID),
	}
}

// BuildEditButton создает кнопку редактирования вопроса
func (b *QuestionButtonBuilder) BuildEditButton(uq *edu.UsersQuestion) telebot.InlineButton {
	return telebot.InlineButton{
		Unique: INLINE_EDIT_QUESTION,
		Text:   "✏️",
		Data:   fmt.Sprintf("%d", uq.QuestionID),
	}
}

// BuildTimeButton создает кнопку с временем до следующего повторения
func (b *QuestionButtonBuilder) BuildTimeButton(uq *edu.UsersQuestion) telebot.InlineButton {
	now := time.Now().UTC()
	duration := uq.TimeRepeat.Sub(now)

	return telebot.InlineButton{
		Text: "⏳" + b.timeLeftMsg(duration),
	}
}

// BuildQuestionTextButton создает кнопку с текстом вопроса
func (b *QuestionButtonBuilder) BuildQuestionTextButton(uq *edu.UsersQuestion) telebot.InlineButton {
	// Получаем текст вопроса
	questionText := uq.R.Question.Question
	if len(questionText) > 50 {
		questionText = questionText[:47] + "..."
	}

	return telebot.InlineButton{
		Unique: INLINE_SHOW_ANSWER,
		Text:   questionText,
		Data:   b.makeData(uq.QuestionID),
	}
}

// BuildPaginationButtons создает кнопки пагинации
func (b *QuestionButtonBuilder) BuildPaginationButtons() []telebot.InlineButton {
	var paginationRow []telebot.InlineButton
	totalPages := b.totalPages()

	if b.page > 0 {
		paginationRow = append(paginationRow, telebot.InlineButton{
			Unique: INLINE_BTN_QUESTION_PAGE + "_prev",
			Text:   "⬅️ Назад",
			Data:   fmt.Sprintf("%d_%s_%d", b.page-1, b.tag, b.tagPage),
		})
	}

	// Кнопка возврата к тегам всегда в центре
	paginationRow = append(paginationRow, telebot.InlineButton{
		Unique: "back_to_tags",
		Text:   MSG_BACK_TAGS,
		Data:   fmt.Sprintf("%d", b.tagPage), // Сохраняем номер страницы тегов
	})

	if b.page < totalPages-1 {
		paginationRow = append(paginationRow, telebot.InlineButton{
			Unique: INLINE_BTN_QUESTION_PAGE + "_next",
			Text:   "Вперед ➡️",
			Data:   fmt.Sprintf("%d_%s_%d", b.page+1, b.tag, b.tagPage),
		})
	}

	return paginationRow
}

// WithPrefixEmoji добавляет эмодзи к тексту кнопки
func (b *QuestionButtonBuilder) WithPrefixEmoji(emoji string, button telebot.InlineButton) telebot.InlineButton {
	button.Text = emoji + " " + button.Text
	return button
}

// WithSuffixEmoji добавляет эмодзи после текста кнопки
func (b *QuestionButtonBuilder) WithSuffixEmoji(button telebot.InlineButton, emoji string) telebot.InlineButton {
	button.Text = button.Text + " " + emoji
	return button
}

// Вспомогательные методы
func (b *QuestionButtonBuilder) makeData(qID int64) string {
	return fmt.Sprintf("%d_%d_%s", qID, b.page, b.tag)
}

func (b *QuestionButtonBuilder) timeLeftMsg(duration time.Duration) string {
	if duration < 0 {
		return "готов"
	}

	if duration < time.Hour {
		return fmt.Sprintf("%.0fm", duration.Minutes())
	}
	if duration < 24*time.Hour {
		return fmt.Sprintf("%.0fh", duration.Hours())
	}
	return fmt.Sprintf("%.0fd", duration.Hours()/24)
}

func (b *QuestionButtonBuilder) totalPages() int {
	if b.totalCount == 0 {
		return 0
	}
	return (b.totalCount + QuestionsPerPage - 1) / QuestionsPerPage
}

// WithPrefixEmoji добавляет эмодзи к тексту кнопки (статическая функция)
func WithPrefixEmoji(emoji string, button telebot.InlineButton) telebot.InlineButton {
	button.Text = emoji + " " + button.Text
	return button
}
