package question

import (
	"fmt"
	"strings"
	"time"
)

func EscapeMarkdown(text string) string {
	specialChars := []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}
	for _, char := range specialChars {
		text = strings.ReplaceAll(text, char, "\\"+char)
	}
	return text
}

func timeLeftMsg(duration time.Duration) string {
	var timeParts []string

	// pluralize возвращает правильную форму слова для числа
	pluralize := func(n int, forms []string) string {
		n = n % 100
		if n > 10 && n < 20 {
			return forms[2]
		}
		n = n % 10
		if n == 1 {
			return forms[0]
		}
		if n >= 2 && n <= 4 {
			return forms[1]
		}
		return forms[2]
	}

	// Разбиваем duration на составляющие
	days := int(duration.Hours() / 24)
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60

	// Добавляем дни если есть
	if days > 0 {
		timeParts = append(timeParts, fmt.Sprintf("%d %s", days, pluralize(days, []string{"день", "дня", "дней"})))
	}

	// Добавляем часы если есть
	if hours > 0 {
		timeParts = append(timeParts, fmt.Sprintf("%d %s", hours, pluralize(hours, []string{"час", "часа", "часов"})))
	}

	// Добавляем минуты только если нет дней (для краткости)
	if minutes > 0 && days == 0 {
		timeParts = append(timeParts, fmt.Sprintf("%d %s", minutes, pluralize(minutes, []string{"минуту", "минуты", "минут"})))
	}

	// Собираем итоговую строку
	t := strings.Join(timeParts, " ")
	if t == "" {
		t = "менее минуты"
	}

	return t
}
