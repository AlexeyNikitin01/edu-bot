package question

const (
	MSG_SUCESS_DELETE_QUESTION = "🤫Вопрос удален👁"

	INLINE_BTN_REPEAT_QUESTION                 = "toggle_repeat"
	INLINE_BTN_DELETE_QUESTION                 = "delete_question"
	INLINE_BTN_DELETE_QUESTION_AFTER_POLL      = "delete_question_after_poll"
	INLINE_BTN_DELETE_QUESTION_AFTER_POLL_HIGH = "delete_question_after_poll_high"
	INLINE_BTN_REPEAT_QUESTION_AFTER_POLL_HIGH = "repeat_question_after_poll_high"
	INLINE_FORGOT_HIGH_QUESTION                = "forgot_high_question"
	INLINE_REMEMBER_HIGH_QUESTION              = "remember_high_question"
	INLINE_NEXT_QUESTION                       = "next_question"
	INLINE_EDIT_QUESTION                       = "edit_question"
	INLINE_EDIT_NAME_QUESTION                  = "inline_edit_name_question"
	INLINE_EDIT_ANSWER_QUESTION                = "inline_edit_answer_question"
	INLINE_EDIT_NAME_TAG_QUESTION              = "inline_edit_name_tag_question"
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

	MSG_FORGOT        = "СЛОЖНО" // Текст кнопки "Забыл" - сложный вопрос
	MSG_REMEMBER      = "ЛЕГКО"  // Текст кнопки "Помню" - легкий вопрос
	MSG_NEXT_QUESTION = "😎"      // Сообщение при запросе следующего вопроса

	BtnShowAnswer = "📝 Показать ответ" // Кнопка показа ответа на вопрос
	BtnRepeat     = "🔔"                // Кнопка повторения обычного вопроса
	BtnRepeatEdu  = "💤"                // Кнопка повторения обучающего вопроса
	BtnDelete     = "🗑️"               // Кнопка удаления вопроса
	BtnEdit       = "✏️"               // Кнопка редактирования вопроса

	// Сообщения для добавления вопросов
	MSG_ADD_QUESTION                   = "✍️ Напишите вопрос или нажмите /cancel для отмены"
	MSG_ADD_CORRECT_ANSWER             = "✍✅ Введите правильный ответ или нажмите /cancel для отмены: "
	MSG_CANCEL                         = "Вы отменили действие👊!"
	MSG_SUCCESS                        = "✅ Успех!"
	MSG_EDIT                           = "<b>Введите новое значение для или нажмите /cancel для отмены:</b>\n\n "
	MSG_SUCCESS_UPDATE_NAME_QUESTION   = "Вопрос обновлен"
	MSG_SUCCESS_UPDATE_ANSWER          = "Ответ обновлен"
	MSG_SUCCESS_UPDATE_TAG_BY_QUESTION = "Тэг для вопроса обновлен"
	MSG_ADD_TAG                        = "🏷 Введите свой тэг или выберите из списка, или нажмите /cancel для отмены: "
	MSG_SUCCESS_UPDATE_TAG             = "Тэг обновлен"
	MSG_EDIT_TAG_BY_QUESTION           = "Выберите или введите свой тэг или нажмите /cancel для отмены: "

	// Сообщения для CSV
	MSG_CSV_INVALID_FILE       = "📛 Пожалуйста, отправьте файл с расширением .csv"
	MSG_CSV_FILE_LOAD_ERROR    = "❌ Не удалось загрузить файл: "
	MSG_CSV_INVALID_FORMAT     = "ℹ️ Пожалуйста, отправьте CSV файл или текст в формате CSV"
	MSG_CSV_PARSE_ERROR        = "❌ Ошибка разбора CSV: %v"
	MSG_CSV_SUCCESS_TEMPLATE   = "✅ Успешно добавлено: %d"
	MSG_CSV_ERRORS_TEMPLATE    = "\n❌ Ошибок: %d"
	MSG_CSV_ERRORS_LIST_HEADER = "\n\nСписок ошибок:\n"
	MSG_CSV_ERRORS_TRUNCATED   = "\n\nПервые 5 ошибок из %d:\n%s"
	MSG_CSV_ADVICE_TEXT        = "\n\nℹ️ Совет: Для текста с ; используйте кавычки: \"Текст с ; внутри\""
	MSG_CSV_ALL_FAILED         = "❌ Не удалось добавить ни одного вопроса. Проверьте формат данных"
	MSG_CSV_MIN_FIELDS_ERROR   = "• Строка %d: требуется минимум 3 поля (вопрос;тег;правильный ответ)"
	MSG_CSV_EMPTY_FIELDS_ERROR = "• Строка %d: вопрос, тег и правильный ответ не могут быть пустыми"
	MSG_CSV_FORMAT_EXAMPLE     = "Пример правильного формата:\n\"Вопрос с ; внутри\";Тег;\"Ответ с ;\"\nОбычный вопрос;Тег;Ответ"

	// Сообщения для списка вопросов
	MSG_LIST_QUESTION = "ВОПРОСЫ: "
	MSG_BACK_TAGS     = "НАЗАД К ТЭГАМ"

	// Константы пагинации вопросов
	QuestionsPerPage = 10
)
