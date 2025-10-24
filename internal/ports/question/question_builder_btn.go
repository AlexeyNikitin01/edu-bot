package question

import (
	"bot/internal/repo/edu"
	"fmt"
	"gopkg.in/telebot.v3"
	"time"
)

// QuestionButtonBuilder отвечает за создание интерактивных кнопок для вопросов
type QuestionButtonBuilder struct{}

// NewQuestionButtonBuilder создает новый экземпляр билдера кнопок
func NewQuestionButtonBuilder() *QuestionButtonBuilder {
	return &QuestionButtonBuilder{}
}

// BuildAnswerButton создает кнопку показа/скрытия ответа
func (b *QuestionButtonBuilder) BuildAnswerButton(uq *edu.UsersQuestion, showAnswer bool) telebot.InlineButton {
	if showAnswer {
		return telebot.InlineButton{
			Unique: INLINE_TURN_ANSWER,
			Text:   "📝 Свернуть ответ",
			Data:   fmt.Sprintf("%d", uq.QuestionID),
		}
	}

	return telebot.InlineButton{
		Unique: INLINE_SHOW_ANSWER,
		Text:   BtnShowAnswer,
		Data:   fmt.Sprintf("%d", uq.QuestionID),
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
func (b *QuestionButtonBuilder) BuildRepeatButton(uq *edu.UsersQuestion, page int, tag string) telebot.InlineButton {
	label := "🔔"
	if uq.IsEdu {
		label = "💤"
	}

	return telebot.InlineButton{
		Unique: INLINE_BTN_REPEAT_QUESTION_AFTER_POLL_HIGH,
		Text:   label,
		Data:   b.makeData(uq.QuestionID, page, tag),
	}
}

// BuildDeleteButton создает кнопку удаления вопроса
func (b *QuestionButtonBuilder) BuildDeleteButton(uq *edu.UsersQuestion, page int, tag string) telebot.InlineButton {
	return telebot.InlineButton{
		Unique: INLINE_BTN_DELETE_QUESTION,
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
func (b *QuestionButtonBuilder) BuildQuestionTextButton(q *edu.Question, page int, tag string) telebot.InlineButton {
	return telebot.InlineButton{
		Text: q.Question,
		Data: b.makeData(q.ID, page, tag),
	}
}

// BuildPaginationButtons создает кнопки пагинации
func (b *QuestionButtonBuilder) BuildPaginationButtons(page int, totalPages int, tag string) []telebot.InlineButton {
	var paginationRow []telebot.InlineButton

	if page > 0 {
		paginationRow = append(paginationRow, telebot.InlineButton{
			Unique: INLINE_BTN_QUESTION_PAGE + "_prev",
			Text:   "⬅️ Назад",
			Data:   fmt.Sprintf("%d_%s", page-1, tag),
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
			Data:   fmt.Sprintf("%d_%s", page+1, tag),
		})
	}

	return paginationRow
}

// BuildQuestionRow создает ряд кнопок для одного вопроса
func (b *QuestionButtonBuilder) BuildQuestionRow(q *edu.Question, uq *edu.UsersQuestion, page int, tag string) [][]telebot.InlineButton {
	return [][]telebot.InlineButton{
		{b.BuildQuestionTextButton(q, page, tag)},
		{
			b.BuildRepeatButton(uq, page, tag),
			b.BuildDeleteButton(uq, page, tag),
			b.BuildEditButton(uq),
			b.BuildTimeButton(uq),
		},
	}
}

// BuildQuestionsKeyboard создает полную клавиатуру со списком вопросов и пагинацией
func (b *QuestionButtonBuilder) BuildQuestionsKeyboard(questions []*edu.Question, userQuestions map[int64]*edu.UsersQuestion, page int, tag string) [][]telebot.InlineButton {
	var btns [][]telebot.InlineButton

	for _, q := range questions {
		if uq, exists := userQuestions[q.ID]; exists {
			questionRows := b.BuildQuestionRow(q, uq, page, tag)
			btns = append(btns, questionRows...)
		}
	}

	return btns
}

// BuildFullKeyboard создает полную клавиатуру для вопроса (альтернативный вариант)
func (b *QuestionButtonBuilder) BuildFullKeyboard(uq *edu.UsersQuestion, showAnswer bool) [][]telebot.InlineButton {
	return [][]telebot.InlineButton{
		b.BuildAnswerRow(uq, showAnswer),
		b.BuildDifficultyRow(uq),
		b.BuildActionsRow(uq, -1, ""),
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
		b.BuildEasyButton(uq),
		b.BuildForgotButton(uq),
	}
}

// BuildActionsRow создает ряд с кнопками действий (повторение, удаление, редактирование)
func (b *QuestionButtonBuilder) BuildActionsRow(uq *edu.UsersQuestion, page int, tag string) []telebot.InlineButton {
	return []telebot.InlineButton{
		b.BuildRepeatButton(uq, page, tag),
		b.BuildDeleteButton(uq, page, tag),
		b.BuildEditButton(uq),
	}
}

// BuildMinimalKeyboard создает минимальную клавиатуру (только ответ и сложность)
func (b *QuestionButtonBuilder) BuildMinimalKeyboard(uq *edu.UsersQuestion, showAnswer bool) [][]telebot.InlineButton {
	return [][]telebot.InlineButton{
		b.BuildAnswerRow(uq, showAnswer),
		b.BuildDifficultyRow(uq),
	}
}

// BuildActionsOnlyKeyboard создает клавиатуру только с кнопками действий
func (b *QuestionButtonBuilder) BuildActionsOnlyKeyboard(uq *edu.UsersQuestion, page int, tag string) [][]telebot.InlineButton {
	return [][]telebot.InlineButton{
		b.BuildActionsRow(uq, page, tag),
	}
}

// Вспомогательные методы
func (b *QuestionButtonBuilder) makeData(qID int64, page int, tag string) string {
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

func (b *QuestionButtonBuilder) timeLeftMsg(duration time.Duration) string {
	if duration < time.Hour {
		return fmt.Sprintf("%.0fm", duration.Minutes())
	}
	if duration < 24*time.Hour {
		return fmt.Sprintf("%.0fh", duration.Hours())
	}
	return fmt.Sprintf("%.0fd", duration.Hours()/24)
}

func WithPrefixEmoji(text string, t telebot.InlineButton) telebot.InlineButton {
	t.Text = text + " " + t.Text
	return t
}
