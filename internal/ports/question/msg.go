package question

const (
	MSG_SUCESS_DELETE_QUESTION = "🤫Вопрос удален👁"

	INLINE_BTN_TAGS                            = "tags"
	INLINE_BTN_REPEAT_QUESTION                 = "toggle_repeat"
	INLINE_BTN_DELETE_QUESTION                 = "delete_question"
	INLINE_BTN_DELETE_QUESTIONS_BY_TAG         = "delete_tag"
	INLINE_BTN_DELETE_QUESTION_AFTER_POLL      = "delete_question_after_poll"
	INLINE_BTN_DELETE_QUESTION_AFTER_POLL_HIGH = "delete_question_after_poll_high"
	INLINE_BTN_REPEAT_QUESTION_AFTER_POLL      = "repeat_question_after_poll"
	INLINE_BTN_REPEAT_QUESTION_AFTER_POLL_HIGH = "repeat_question_after_poll_high"
	INLINE_BTN_QUESTION_BY_TAG                 = "question_by_tag"
	INLINE_FORGOT_HIGH_QUESTION                = "forgot_high_question"
	INLINE_REMEMBER_HIGH_QUESTION              = "remember_high_question"
	INLINE_NEXT_QUESTION                       = "next_question"
	INLINE_EDIT_TAG                            = "edit_tag"
	INLINE_EDIT_QUESTION                       = "edit_question"
	INLINE_EDIT_NAME_QUESTION                  = "inline_edit_name_question"
	INLINE_EDIT_ANSWER_QUESTION                = "inline_edit_answer_question"
	INLINE_EDIT_NAME_TAG_QUESTION              = "inline_edit_name_tag_question"
	INLINE_BACK_TAGS                           = "back_to_tags"
	INLINE_PAUSE_TAG                           = "pause_tag"
	INLINE_BTN_QUESTION_PAGE                   = "inline_btn_page"
	INLINE_SHOW_ANSWER                         = "show_answer"
	INLINE_TURN_ANSWER                         = "turn_answer"
	INLINE_BTN_TASK_BY_TAG                     = "inline_btn_task_by_tag"
	INLINE_SHOW_CURRENT_VALUE                  = "show_current_value"
	INLINE_COLLAPSE_VALUE                      = "collapse_value"

	INLINE_NAME_DELETE_AFTER_POLL = "🗑️"
	INLINE_NAME_REPEAT_AFTER_POLL = "️ПОВТОРЕНИЕ"
	INLINE_NAME_DELETE            = "🗑️"

	BTN_ADD_QUESTION       = "➕ Вопрос"
	BTN_MANAGMENT_QUESTION = "📚 Управление"
	BTN_ADD_CSV            = "➕ Вопросы CSV"
	BTN_NEXT_QUESTION      = "🌀 Дальше"
	BTN_NEXT_TASK          = "🧐Получить задачу"

	MSG_WRONG_BTN = "⚠️ Неизвестная команда. Используйте меню ниже."

	MSG_GRETING = "Добро пожаловать!\n\n" +
		"🤖 Этот бот предназначен для интервального повторения собственной базы вопросов. Бот сам отправляет периодически вопросы!\n\n" +
		"🤖 Напиши при создании вопроса 'ЗАДАЧА' перед ним и ты можешь получать случайную задачу по нажатию на кнопку. Удобно для leetCode😇\n\n" +
		"✨ Выберите действие с помощью кнопок ниже:\n\n" +
		"🔹 \"➕ Вопрос\" — Создать новый вопрос вручную.\n\n" +
		"🔹 \"➕ Вопросы CSV\" — Массовая загрузка вопросов из файла CSV.\n\n" +
		"🔹 \"📚 Управление\" — Просмотр и редактирование существующих вопросов.\n\n" +
		"🔹 \"🧐 Получить задачу\" — Получить случайную задачу для немедленного решения. Удобно для работы с конкретными задачами!\n\n" +
		"🔹 \"🌀 Дальше\" — Получить случайный вопрос для проверки знаний!"

	MSG_CSV = `📤 Отправьте CSV данные (файл или текст):

	Формат:
	Вопрос;Тег;Правильный ответ[;Другие ответы...]
	
	Если в вопросе/ответе есть ";", заключите его в кавычки:
	"Вопрос с ; внутри";Тег;"Ответ с ;"
	
	Примеры:
	1. Простой: Что такое GPT?;AI;Generative Pre-trained Transformer
	2. Сложный: "Что выведет: x++; y--;?";Программирование;"1; 2; 3"`

	CMD_START         = "/start"
	CMD_CANCEL string = "/cancel"

	MSG_FORGOT        = "СЛОХЖНО" // Текст кнопки "Забыл" - сложный вопрос
	MSG_REMEMBER      = "ЛЕГКО"   // Текст кнопки "Помню" - легкий вопрос
	MSG_NEXT_QUESTION = "😎"       // Сообщение при запросе следующего вопроса

	BtnShowAnswer = "📝 Показать ответ" // Кнопка показа ответа на вопрос
	BtnRepeat     = "🔔"                // Кнопка повторения обычного вопроса
	BtnRepeatEdu  = "💤"                // Кнопка повторения обучающего вопроса
	BtnDelete     = "🗑️"               // Кнопка удаления вопроса
	BtnEdit       = "✏️"               // Кнопка редактирования вопроса
)
