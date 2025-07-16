package ports

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"gopkg.in/telebot.v3"

	"bot/internal/app"
)

const (
	// CSV обработка
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
)

func setQuestionsByCSV(domain app.Apper) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		var records [][]string
		var err error
		var isFile bool

		// Обработка входящих данных
		if ctx.Message().Document != nil {
			if !strings.HasSuffix(ctx.Message().Document.FileName, ".csv") {
				return ctx.Send(MSG_CSV_INVALID_FILE)
			}

			file, err := ctx.Bot().File(&ctx.Message().Document.File)
			if err != nil {
				return ctx.Send(MSG_CSV_FILE_LOAD_ERROR + err.Error())
			}
			defer file.Close()

			records, err = parseCSV(file)
			if err != nil {
				return ctx.Send(fmt.Sprintf(MSG_CSV_PARSE_ERROR, err))
			}
			isFile = true
		} else {
			text := strings.TrimSpace(ctx.Text())
			if text == "" {
				return ctx.Send(MSG_CSV_INVALID_FORMAT + "\n\n" + MSG_CSV_FORMAT_EXAMPLE)
			}

			records, err = parseCSV(strings.NewReader(text))
			if err != nil {
				return ctx.Send(fmt.Sprintf(MSG_CSV_PARSE_ERROR, err) + "\n\n" + MSG_CSV_FORMAT_EXAMPLE)
			}
		}

		// Обработка записей
		userID := ctx.Sender().ID
		var successCount, errorCount int
		var errorLines []string

		for i, record := range records {
			lineNum := i + 1

			// Проверка минимального формата
			if len(record) < 3 {
				errorCount++
				errorLines = append(errorLines, fmt.Sprintf(MSG_CSV_MIN_FIELDS_ERROR, lineNum))
				continue
			}

			question := strings.TrimSpace(record[0])
			tag := strings.TrimSpace(record[1])
			correctAnswer := strings.TrimSpace(record[2])

			// Проверка пустых полей
			if question == "" || tag == "" || correctAnswer == "" {
				errorCount++
				errorLines = append(errorLines, fmt.Sprintf(MSG_CSV_EMPTY_FIELDS_ERROR, lineNum))
				continue
			}

			// Сбор неправильных ответов
			var wrongAnswers []string
			for j := 3; j < len(record); j++ {
				if ans := strings.TrimSpace(record[j]); ans != "" {
					wrongAnswers = append(wrongAnswers, ans)
				}
			}

			// Сохранение вопроса
			allAnswers := append([]string{correctAnswer}, wrongAnswers...)
			if err := domain.SaveQuestions(
				GetContext(ctx), question, tag, allAnswers, userID,
			); err != nil {
				errorCount++
				errorLines = append(errorLines, fmt.Sprintf("• Строка %d: %v", lineNum, err))
				continue
			}

			successCount++
		}

		// Формирование результата
		msg := fmt.Sprintf(MSG_CSV_SUCCESS_TEMPLATE, successCount)
		if errorCount > 0 {
			msg += fmt.Sprintf(MSG_CSV_ERRORS_TEMPLATE, errorCount)

			if len(errorLines) <= 5 {
				msg += MSG_CSV_ERRORS_LIST_HEADER + strings.Join(errorLines, "\n")
			} else {
				msg += fmt.Sprintf(MSG_CSV_ERRORS_TRUNCATED, len(errorLines), strings.Join(errorLines[:5], "\n"))
			}

			if !isFile {
				msg += MSG_CSV_ADVICE_TEXT
			}
		}

		if successCount == 0 && errorCount > 0 {
			msg = MSG_CSV_ALL_FAILED + "\n\n" + MSG_CSV_FORMAT_EXAMPLE
		}

		return ctx.Send(msg, telebot.ModeHTML)
	}
}

// parseCSV корректно обрабатывает CSV с точками с запятой внутри полей
func parseCSV(r io.Reader) ([][]string, error) {
	reader := csv.NewReader(r)
	reader.Comma = ';'
	reader.LazyQuotes = true // Разрешает неэкранированные кавычки в полях
	reader.TrimLeadingSpace = true

	var records [][]string
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("ошибка чтения строки: %v", err)
		}
		records = append(records, record)
	}
	return records, nil
}
