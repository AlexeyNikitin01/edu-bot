package tags

import (
	"regexp"
	"strconv"
)

// ExtractCurrentPage извлекает номер текущей страницы из текста сообщения
func ExtractCurrentPage(messageText string) int {
	// Ищем паттерн "Страница X из Y" в тексте
	re := regexp.MustCompile(`Страница (\d+) из \d+`)
	matches := re.FindStringSubmatch(messageText)
	if len(matches) > 1 {
		if page, err := strconv.Atoi(matches[1]); err == nil {
			return page
		}
	}
	return 1
}
