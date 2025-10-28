package task

import (
	"bot/internal/ports/menu"
	"bot/internal/ports/question"
	"fmt"
	"gopkg.in/telebot.v3"
)

const (
	INLINE_REMEMBER_HIGH_TASK = "high_task"
	INLINE_FORGOT_HIGH_TASK   = "fogot_task"

	INLINE_NEXT_TASK = "next_task"

	INLINE_SKIP_TASK = "skip_task"
)

type TaskButtonsBuilder struct {
	buttons [][]telebot.InlineButton
}

func NewTaskButtonsBuilder() *TaskButtonsBuilder {
	return &TaskButtonsBuilder{
		buttons: make([][]telebot.InlineButton, 0),
	}
}

func (b *TaskButtonsBuilder) AddShowAnswer(questionID int64) *TaskButtonsBuilder {
	btn := telebot.InlineButton{
		Unique: question.INLINE_SHOW_ANSWER,
		Text:   "📝 Показать ответ",
		Data:   fmt.Sprintf("%d", questionID),
	}
	b.buttons = append(b.buttons, []telebot.InlineButton{btn})
	return b
}

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
		Unique: menu.BTN_NEXT_QUESTION,
		Text:   menu.BTN_NEXT_QUESTION,
		Data:   fmt.Sprintf("%d", qID),
	}
	b.buttons = append(b.buttons, []telebot.InlineButton{nextTaskBtn})
	b.buttons = append(b.buttons, []telebot.InlineButton{skipTaskBtn, continueQuestionsBtn})
	return b
}

func (b *TaskButtonsBuilder) AddActions(qID int64, label string) *TaskButtonsBuilder {
	repeatBtn := telebot.InlineButton{
		Unique: question.INLINE_BTN_REPEAT_QUESTION_AFTER_POLL_HIGH,
		Text:   label,
		Data:   fmt.Sprintf("%d", qID),
	}
	deleteBtn := telebot.InlineButton{
		Unique: question.INLINE_BTN_DELETE_QUESTION_AFTER_POLL_HIGH,
		Text:   question.INLINE_NAME_DELETE_AFTER_POLL,
		Data:   fmt.Sprintf("%d", qID),
	}
	editBtn := telebot.InlineButton{
		Unique: question.INLINE_EDIT_QUESTION,
		Text:   "✏️",
		Data:   fmt.Sprintf("%d", qID),
	}
	b.buttons = append(b.buttons, []telebot.InlineButton{repeatBtn, deleteBtn, editBtn})
	return b
}

func (b *TaskButtonsBuilder) AddCustomRow(buttons ...telebot.InlineButton) *TaskButtonsBuilder {
	if len(buttons) > 0 {
		b.buttons = append(b.buttons, buttons)
	}
	return b
}

func (b *TaskButtonsBuilder) Build() *telebot.ReplyMarkup {
	return &telebot.ReplyMarkup{
		InlineKeyboard: b.buttons,
	}
}
