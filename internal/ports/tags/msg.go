package tags

const (
	INLINE_BTN_TAGS                    = "tags"
	INLINE_BTN_DELETE_QUESTIONS_BY_TAG = "delete_tag"
	INLINE_BTN_QUESTION_BY_TAG         = "question_by_tag"
	INLINE_EDIT_TAG                    = "edit_tag"
	INLINE_BACK_TAGS                   = "back_to_tags"
	INLINE_PAUSE_TAG                   = "pause_tag"
	INLINE_SELECT_TAG                  = "select_tag" // Для выбора тега при создании вопроса

	// Пагинация
	INLINE_PAGINATION_PREV = "pagination_prev"
	INLINE_PAGINATION_NEXT = "pagination_next"
	INLINE_PAGINATION_INFO = "pagination_info"
	INLINE_NO_TAGS         = "no_tags"

	// Константы пагинации
	DEFAULT_PAGE_SIZE = 10
	MAX_PAGE_SIZE     = 50

	// Тексты для кнопок и сообщений пагинации
	PAGINATION_PREV_TEXT   = "⬅️ Назад"
	PAGINATION_NEXT_TEXT   = "Вперед ➡️"
	PAGINATION_INFO_TEXT   = "info"
	NO_TAGS_TEXT           = "📭 Нет тегов"
	PAGINATION_INFO_FORMAT = "%d/%d"
	INLINE_NAME_DELETE     = "🗑️"

	// Эмодзи и символы
	EMOJI_BELL      = "🔔"
	EMOJI_SLEEP     = "💤"
	EMOJI_EDIT      = "✏️"
	EMOJI_TRASH     = "🗑️"
	EMOJI_ENVELOPE  = "📭"
	EMOJI_PAGE      = "📄"
	EMOJI_BAR_CHART = "📊"

	// Форматы сообщений
	PAGINATION_MESSAGE_FORMAT     = "%s\n\n%s"
	PAGINATION_INFO_FULL_FORMAT   = "📄 Страница %d из %d | Всего тегов: %d"
	PAGINATION_INFO_SIMPLE_FORMAT = "📊 Всего тегов: %d"

	// Сообщения по умолчанию
	MSG_LIST_TAGS = "📚 Список тегов:"

	// Сообщения для добавления/редактирования тегов
	MSG_ADD_TAG              = "🏷 Введите свой тэг или выберите из списка, или нажмите /cancel для отмены: "
	MSG_EDIT_TAG_BY_QUESTION = "Выберите или введите свой тэг или нажмите /cancel для отмены: "
	MSG_SUCCESS_UPDATE_TAG   = "Тэг обновлен"
)
