package dto

// QuestionDraft представляет черновик вопроса для создания или редактирования
type QuestionDraft struct {
	Step             int      // Текущий шаг в процессе создания
	Question         string   // Текст вопроса
	Tag              string   // Тег вопроса
	Answers          []string // Список ответов
	TagID            int64    // ID тега для редактирования
	QuestionIDByTag  int64    // ID вопроса для изменения тега
	QuestionIDByName int64    // ID вопроса для изменения названия
	AnswerID         int64    // ID ответа для редактирования
}
