package task

const (
	// Button texts
	BtnShowAnswer  = "📝 Показать ответ"
	BtnTurnAnswer  = "📝 Скрыть ответ"
	BtnNextTask    = "➡️ Следующая"
	BtnSkipTask    = "⏩"
	BtnRepeatLabel = "🔔"
	BtnEduLabel    = "💤"
	BtnEdit        = "✏️"
	BtnDelete      = "🗑️"
	BtnEasy        = "✅"
	BtnForgot      = "❌"

	// Messages
	MsgChoiceSaved             = "Выбор сохранён!\n\n"
	MsgTaskSkipped             = "⏩ Вопрос пропущен!\n\n"
	MsgAllTasksCompleted       = "🎉 Все вопросы завершены! Вы великолепны!"
	MsgNoTagsAvailable         = "📝 У вас пока нет тегов. Создайте первый вопрос!"
	MsgTagQuestion             = "%s: %s"
	MSG_SUCESS_DELETE_QUESTION = "Задача удалена"

	// Inline button unique identifiers
	INLINE_REMEMBER_HIGH_TASK = "high_task"
	INLINE_FORGOT_HIGH_TASK   = "fogot_task"
	INLINE_NEXT_TASK          = "next_task"
	INLINE_SKIP_TASK          = "skip_task"
	INLINE_SHOW_ANSWER_TASK   = "show_answer_task"
	INLINE_TURN_ANSWER_TASK   = "turn_answer_task"

	INLINE_BTN_REPEAT_TASK_AFTER_POLL = "repaet_task_after_poll"
	INLINE_BTN_DELETE_TASK_AFTER_POLL = "delete_task_after_poll"
	INLINE_BTN_EDIT_TASK_AFTER_POLL   = "edit_task_after_poll"
)
