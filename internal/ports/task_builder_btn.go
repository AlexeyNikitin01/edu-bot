package ports

import (
	"fmt"
	"gopkg.in/telebot.v3"
)

const (
	INLINE_REMEMBER_HIGH_TASK = "high_task"
	INLINE_FORGOT_HIGH_TASK   = "fogot_task"

	INLINE_NEXT_TASK = "next_task"

	INLINE_SKIP_TASK = "skip_task"
)

// TaskButtonsBuilder предоставляет fluent-интерфейс для построения клавиатур с кнопками задач.
// Позволяет легко создавать различные комбинации кнопок для взаимодействия с вопросами и задачами.
//
// Пример использования:
//
//	keyboard := NewTaskButtonsBuilder().
//	    AddShowAnswer(questionID).
//	    AddDifficulty(questionID).
//	    AddNavigation(questionID).
//	    Build()
type TaskButtonsBuilder struct {
	buttons [][]telebot.InlineButton
}

// NewTaskButtonsBuilder создает новый экземпляр билдера кнопок задач.
// Возвращает инициализированный билдер с пустым списком кнопок.
//
// Returns:
//   - *TaskButtonsBuilder: указатель на новый билдер кнопок
func NewTaskButtonsBuilder() *TaskButtonsBuilder {
	return &TaskButtonsBuilder{
		buttons: make([][]telebot.InlineButton, 0),
	}
}

// AddShowAnswer добавляет кнопку для отображения ответа/решения к вопросу.
// Создает отдельный ряд с одной кнопкой "📝 Показать ответ".
//
// Parameters:
//   - questionID: идентификатор вопроса для которого показывается ответ
//
// Returns:
//   - *TaskButtonsBuilder: текущий билдер для цепочки вызовов
func (b *TaskButtonsBuilder) AddShowAnswer(questionID int64) *TaskButtonsBuilder {
	btn := telebot.InlineButton{
		Unique: INLINE_SHOW_ANSWER,
		Text:   "📝 Показать ответ",
		Data:   fmt.Sprintf("%d", questionID),
	}
	b.buttons = append(b.buttons, []telebot.InlineButton{btn})
	return b
}

// AddDifficulty добавляет кнопки оценки сложности вопроса.
// Создает ряд с двумя кнопками:
//   - "✅" - вопрос был легким
//   - "❌" - вопрос был сложным
//
// Parameters:
//   - questionID: идентификатор вопроса для оценки сложности
//
// Returns:
//   - *TaskButtonsBuilder: текущий билдер для цепочки вызовов
func (b *TaskButtonsBuilder) AddDifficulty(questionID int64) *TaskButtonsBuilder {
	easy := telebot.InlineButton{
		Unique: INLINE_REMEMBER_HIGH_TASK,
		Text:   "✅",
		Data:   fmt.Sprintf("%d", questionID),
	}
	forgot := telebot.InlineButton{
		Unique: INLINE_FORGOT_HIGH_TASK,
		Text:   "❌",
		Data:   fmt.Sprintf("%d", questionID),
	}
	b.buttons = append(b.buttons, []telebot.InlineButton{easy, forgot})
	return b
}

// AddNavigation добавляет кнопки навигации между задачами.
// Создает два ряда кнопок:
//   - Первый ряд: "➡️ Следующая" - переход к следующей задаче
//   - Второй ряд: "⏩ Пропустить" и "🔁 Продолжить" - дополнительные опции навигации
//
// Parameters:
//   - qID: идентификатор текущего вопроса для передачи в данные кнопок
//
// Returns:
//   - *TaskButtonsBuilder: текущий билдер для цепочки вызовов
func (b *TaskButtonsBuilder) AddNavigation(qID int64) *TaskButtonsBuilder {
	nextTaskBtn := telebot.InlineButton{
		Unique: INLINE_NEXT_TASK,
		Text:   "➡️ Следующая",
		Data:   fmt.Sprintf("%d", qID),
	}
	skipTaskBtn := telebot.InlineButton{
		Unique: INLINE_SKIP_TASK,
		Text:   "⏩ Пропустить",
		Data:   fmt.Sprintf("%d", qID),
	}
	continueQuestionsBtn := telebot.InlineButton{
		Unique: BTN_NEXT_QUESTION,
		Text:   BTN_NEXT_QUESTION,
		Data:   fmt.Sprintf("%d", qID),
	}
	b.buttons = append(b.buttons, []telebot.InlineButton{nextTaskBtn})
	b.buttons = append(b.buttons, []telebot.InlineButton{skipTaskBtn, continueQuestionsBtn})
	return b
}

// AddActions добавляет кнопки дополнительных действий с вопросом.
// Создает ряд с тремя кнопками:
//   - Повторить (с кастомной меткой) - настройка повторения вопроса
//   - Удалить - удаление вопроса из системы
//   - ✏️ - редактирование вопроса
//
// Parameters:
//   - qID: идентификатор вопроса для действий
//   - label: метка для кнопки повтора (обычно "🔔" или "💤")
//
// Returns:
//   - *TaskButtonsBuilder: текущий билдер для цепочки вызовов
//
// Note:
//   - Метка "🔔" вопрос на паузке
//   - Метка "💤" вопрос в выборке
func (b *TaskButtonsBuilder) AddActions(qID int64, label string) *TaskButtonsBuilder {
	repeatBtn := telebot.InlineButton{
		Unique: INLINE_BTN_REPEAT_QUESTION_AFTER_POLL_HIGH,
		Text:   label,
		Data:   fmt.Sprintf("%d", qID),
	}
	deleteBtn := telebot.InlineButton{
		Unique: INLINE_BTN_DELETE_QUESTION_AFTER_POLL_HIGH,
		Text:   INLINE_NAME_DELETE_AFTER_POLL,
		Data:   fmt.Sprintf("%d", qID),
	}
	editBtn := telebot.InlineButton{
		Unique: INLINE_EDIT_QUESTION,
		Text:   "✏️",
		Data:   fmt.Sprintf("%d", qID),
	}
	b.buttons = append(b.buttons, []telebot.InlineButton{repeatBtn, deleteBtn, editBtn})
	return b
}

// AddCustomRow добавляет произвольный ряд кнопок к клавиатуре.
// Полезно для добавления специализированных кнопок, не охваченных стандартными методами.
//
// Parameters:
//   - buttons: variadic параметр с кнопками для добавления в один ряд
//
// Returns:
//   - *TaskButtonsBuilder: текущий билдер для цепочки вызовов
//
// Example:
//
//	customBtn := telebot.InlineButton{Text: "Custom", Unique: "custom", Data: "data"}
//	builder.AddCustomRow(customBtn, anotherBtn)
func (b *TaskButtonsBuilder) AddCustomRow(buttons ...telebot.InlineButton) *TaskButtonsBuilder {
	if len(buttons) > 0 {
		b.buttons = append(b.buttons, buttons)
	}
	return b
}

// Build завершает построение клавиатуры и возвращает готовую разметку для Telegram.
// Этот метод должен вызываться последним в цепочке вызовов билдера.
//
// Returns:
//   - *telebot.ReplyMarkup: готовая клавиатура с собранными кнопками
//
// Note:
//   - После вызова Build() билдер не должен использоваться повторно
//   - Для новой клавиатуры создавайте новый билдер через NewTaskButtonsBuilder()
func (b *TaskButtonsBuilder) Build() *telebot.ReplyMarkup {
	return &telebot.ReplyMarkup{
		InlineKeyboard: b.buttons,
	}
}
