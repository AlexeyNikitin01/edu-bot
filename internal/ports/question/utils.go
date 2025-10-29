package question

import (
	"fmt"
	"strconv"
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
	if duration < 0 {
		return "готов"
	}

	var timeParts []string

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

	days := int(duration.Hours() / 24)
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60

	if days > 0 {
		timeParts = append(timeParts, fmt.Sprintf("%d %s", days, pluralize(days, []string{"день", "дня", "дней"})))
	}
	if hours > 0 {
		timeParts = append(timeParts, fmt.Sprintf("%d %s", hours, pluralize(hours, []string{"час", "часа", "часов"})))
	}
	if minutes > 0 && days == 0 { // Минуты показываем только если нет дней
		timeParts = append(timeParts, fmt.Sprintf("%d %s", minutes, pluralize(minutes, []string{"минуту", "минуты", "минут"})))
	}

	t := strings.Join(timeParts, " ")
	if t == "" {
		t = "менее минуты"
	}

	return t
}

func parsePageString(s string) (tag string, page int, err error) {
	parts := strings.Split(s, "_page_")
	if len(parts) != 2 {
		return "", 0, fmt.Errorf("неверный формат строки")
	}

	tag = parts[0]
	page, err = strconv.Atoi(parts[1])
	if err != nil {
		return "", 0, fmt.Errorf("ошибка парсинга номера страницы: %v", err)
	}

	return tag, page, nil
}
