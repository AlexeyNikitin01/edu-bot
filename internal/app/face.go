package app

import (
	"context"

	"bot/internal/repo/edu"
)

type Apper interface {
	GetQuestionsAnswers(ctx context.Context, userID int64) (edu.UsersQuestionSlice, error)
	UpdateRepeatTime(ctx context.Context, question *edu.UsersQuestion) error
}

type App struct {
}

func NewApp() *App {
	return &App{}
}
