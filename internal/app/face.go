package app

import (
	"context"
	"time"

	"bot/internal/repo/edu"
)

type Apper interface {
	GetQuestionsAnswers(context.Context, int64) (edu.UsersQuestionSlice, error)
	UpdateRepeatTime(context.Context, *edu.UsersQuestion, bool) error
	GetUniqueTags(context.Context, int64) ([]*edu.Tag, error)
	SaveQuestions(context.Context, string, string, []string, int64) error
	GetUser(context.Context, int64) (*edu.User, error)
	CreateUser(context.Context, int64, int64, string) (*edu.User, error)
	UpdateIsEduUserQuestion(context.Context, int64, int64) error
	UpdateTag(context.Context, int64, string) error
	GetQuestionAnswers(ctx context.Context, qID int64) (*edu.Question, error)
	UpdateQuestionName(context.Context, int64, string) error
	UpdateAnswer(ctx context.Context, aID int64, answer string) error
	UpdateTagByQuestion(ctx context.Context, qID int64, newTag string) error
	GetNearestTimeRepeat(ctx context.Context, userID int64) (time.Time, error)
}

type App struct {
}

func NewApp() *App {
	return &App{}
}
