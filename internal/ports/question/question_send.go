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
	dis.StartPollingLoop(ctx, ch)

	for {
		select {
		case uq, ok := <-ch:
			if !ok || uq == nil {
				continue
			}
			tag := EscapeMarkdown(uq.GetQuestion().R.GetTag().Tag)
			questionText := EscapeMarkdown(uq.GetQuestion().Question)

			rec := &telebot.User{ID: uq.UserID}
			_, err := b.Send(
				rec,
				tag+": "+questionText,
				telebot.ModeMarkdownV2,
				&telebot.ReplyMarkup{InlineKeyboard: NewQuestionButtonBuilder(WithTag(uq.GetQuestion().R.GetTag().Tag)).
					BuildAfterSend(uq, false),
				},
			)
			log.Println(err)
		}
	}
}
