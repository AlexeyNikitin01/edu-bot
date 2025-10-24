package ports

import (
	"bot/internal/repo/edu"
	"fmt"
	"gopkg.in/telebot.v3"
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
func (b *QuestionButtonBuilder) BuildRepeatButton(uq *edu.UsersQuestion) telebot.InlineButton {
	label := BtnRepeat
	if uq.IsEdu {
		label = BtnRepeatEdu
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
		Text:   BtnDelete,
		Data:   fmt.Sprintf("%d", uq.QuestionID),
	}
}

// BuildEditButton создает кнопку редактирования вопроса
func (b *QuestionButtonBuilder) BuildEditButton(uq *edu.UsersQuestion) telebot.InlineButton {
	return telebot.InlineButton{
		Unique: INLINE_EDIT_QUESTION,
		Text:   BtnEdit,
		Data:   fmt.Sprintf("%d", uq.QuestionID),
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
func (b *QuestionButtonBuilder) BuildActionsRow(uq *edu.UsersQuestion) []telebot.InlineButton {
	return []telebot.InlineButton{
		b.BuildRepeatButton(uq),
		b.BuildDeleteButton(uq),
		b.BuildEditButton(uq),
	}
}

// BuildFullKeyboard создает полную клавиатуру для вопроса
func (b *QuestionButtonBuilder) BuildFullKeyboard(uq *edu.UsersQuestion, showAnswer bool) [][]telebot.InlineButton {
	return [][]telebot.InlineButton{
		b.BuildAnswerRow(uq, showAnswer),
		b.BuildDifficultyRow(uq),
		b.BuildActionsRow(uq),
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
func (b *QuestionButtonBuilder) BuildActionsOnlyKeyboard(uq *edu.UsersQuestion) [][]telebot.InlineButton {
	return [][]telebot.InlineButton{
		b.BuildActionsRow(uq),
	}
}
