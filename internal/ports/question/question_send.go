package question

import (
	"bot/internal/domain"
	"bot/internal/repo/edu"
	"context"
	"gopkg.in/telebot.v3"
	"log"
)

func SendQuestion(ctx context.Context, b *telebot.Bot, dis domain.Dispatcher) {
	ch := make(chan *edu.UsersQuestion)
	// для каждого пользователя, каждые 2 секунды рассылка вопросов для пользователей
	dis.StartPollingLoop(ctx, ch)

	for {
		select {
		case uq := <-ch:
			// Форматируем текст вопроса
			tag := EscapeMarkdown(uq.GetQuestion().R.GetTag().Tag)
			questionText := EscapeMarkdown(uq.GetQuestion().Question)

			// Создаем интерактивные кнопки

			// Отправляем сообщение пользователю
			rec := &telebot.User{ID: uq.UserID}
			_, err := b.Send(
				rec,
				tag+": "+questionText,
				telebot.ModeMarkdownV2,
			)
			log.Println(err)
		}
	}
}
