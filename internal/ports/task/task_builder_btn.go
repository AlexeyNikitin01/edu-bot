package task

import (
	"bot/internal/ports/question"
	"fmt"
	"gopkg.in/telebot.v3"
)

type TaskButtonsBuilder struct {
	buttons [][]telebot.InlineButton
}

func NewTaskButtonsBuilder() *TaskButtonsBuilder {
	return &TaskButtonsBuilder{
		buttons: make([][]telebot.InlineButton, 0),
	}
}

func (b *TaskButtonsBuilder) AddShowAnswer(questionID int64, showAnswer bool) *TaskButtonsBuilder {
	if !showAnswer {
		btn := telebot.InlineButton{
			Unique: INLINE_TURN_ANSWER_TASK,
			Text:   BtnTurnAnswer,
			Data:   fmt.Sprintf("%d", questionID),
		}
		b.buttons = append(b.buttons, []telebot.InlineButton{btn})
		return b
	}

	btn := telebot.InlineButton{
		Unique: INLINE_SHOW_ANSWER_TASK,
		Text:   BtnShowAnswer,
		Data:   fmt.Sprintf("%d", questionID),
	}
	b.buttons = append(b.buttons, []telebot.InlineButton{btn})
	return b
}

func (b *TaskButtonsBuilder) AddDifficulty(questionID int64) *TaskButtonsBuilder {
	easy := telebot.InlineButton{
		Unique: INLINE_REMEMBER_HIGH_TASK,
		Text:   BtnEasy,
		Data:   fmt.Sprintf("%d", questionID),
	}
	skipTaskBtn := telebot.InlineButton{
		Unique: INLINE_SKIP_TASK,
		Text:   BtnSkipTask,
		Data:   fmt.Sprintf("%d", questionID),
	}
	forgot := telebot.InlineButton{
		Unique: INLINE_FORGOT_HIGH_TASK,
		Text:   BtnForgot,
		Data:   fmt.Sprintf("%d", questionID),
	}
	b.buttons = append(b.buttons, []telebot.InlineButton{easy, skipTaskBtn, forgot})
	return b
}

func (b *TaskButtonsBuilder) AddNavigation(tag string) *TaskButtonsBuilder {
	nextTaskBtn := telebot.InlineButton{
		Unique: INLINE_NEXT_TASK,
		Text:   BtnNextTask,
		Data:   fmt.Sprintf("%s", tag),
	}
	b.buttons = append(b.buttons, []telebot.InlineButton{nextTaskBtn})
	fmt.Println("количество кнопок", len(b.buttons))
	return b
}

func (b *TaskButtonsBuilder) AddActions(qID int64, isEdu bool) *TaskButtonsBuilder {
	label := BtnRepeatLabel
	if isEdu {
		label = BtnEduLabel
	}

	repeatBtn := telebot.InlineButton{
		Unique: INLINE_BTN_REPEAT_TASK_AFTER_POLL,
		Text:   label,
		Data:   fmt.Sprintf("%d", qID),
	}
	deleteBtn := telebot.InlineButton{
		Unique: INLINE_BTN_DELETE_TASK_AFTER_POLL,
		Text:   BtnDelete,
		Data:   fmt.Sprintf("%d", qID),
	}
	editBtn := telebot.InlineButton{
		Unique: question.INLINE_EDIT_QUESTION,
		Text:   BtnEdit,
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
