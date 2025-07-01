package app

import (
	"context"

	"bot/internal/repo/edu"
)

type Apper interface {
	GetQuestionsAnswers(context.Context, int64) (edu.UsersQuestionSlice, error)
	UpdateRepeatTime(context.Context, *edu.UsersQuestion, bool) error
	GetUniqueTags(context.Context, int64) ([]string, error)
	SaveQuestions(context.Context, string, string, []string, int64) error
	GetUser(context.Context, int64) (*edu.User, error)
	CreateUser(context.Context, int64, int64, string) (*edu.User, error)
}

type App struct {
}

func NewApp() *App {
	return &App{}
}
