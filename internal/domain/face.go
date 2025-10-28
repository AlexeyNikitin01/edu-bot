package domain

import (
	"bot/internal/repo/dto"
	"bot/internal/repo/edu"
	"context"
	"time"
)

type UseCases interface {
	UserService
	QuestionService
	UserQuestionService
	TagService
	AnswerService
	Dispatcher
}

type Dispatcher interface {
	StartPollingLoop(ctx context.Context, ch chan *edu.UsersQuestion)
	Stop()
	SetUserWaiting(ctx context.Context, userID int64, wait bool) error
}

type UserService interface {
	GetUser(context.Context, int64) (*edu.User, error)
	CreateUser(context.Context, int64, int64, string) (*edu.User, error)
	GetUsersForSend(ctx context.Context, activityUsers []int64) (edu.UserSlice, error)
}

type QuestionService interface {
	GetQuestionAnswers(ctx context.Context, qID int64) (*edu.Question, error)
	UpdateQuestionName(context.Context, int64, string) error
	SaveQuestions(context.Context, string, string, []string, int64) error
	GetAllQuestions(context.Context, int64, string) (edu.QuestionSlice, error)
}

type UserQuestionService interface {
	GetRandomNearestQuestionWithAnswer(ctx context.Context, userID int64) (*edu.UsersQuestion, error)
	GetUserQuestion(ctx context.Context, userID, qID int64) (*edu.UsersQuestion, error)
	UpdateRepeatTime(context.Context, *edu.UsersQuestion, bool) error
	UpdateIsEduUserQuestion(context.Context, int64, int64) error
	GetNearestTimeRepeat(ctx context.Context, userID int64) (time.Time, error)
	GetTask(ctx context.Context, userID int64, tag string) (*edu.UsersQuestion, error)
	DeleteQuestionUser(ctx context.Context, userID int64, qID int64) error
	SetDraftQuestion(ctx context.Context, userID int64, draftQuestion *dto.QuestionDraft) error
	GetDraftQuestion(ctx context.Context, userID int64) (*dto.QuestionDraft, error)
	DeleteDraftQuestion(ctx context.Context, userID int64) error
}

type TagService interface {
	GetUniqueTags(ctx context.Context, userID int64, page, pageSize int) ([]*edu.Tag, int, error)
	UpdateTag(context.Context, int64, string) error
	UpdateTagByQuestion(ctx context.Context, qID int64, newTag string) error
	GetUniqueTagsByTask(ctx context.Context, userID int64) ([]*edu.Tag, error)
	GetTagByID(ctx context.Context, tagID int64) (*edu.Tag, error)
	DeleteQuestionsByTag(ctx context.Context, userID int64, tag string) error
}

type AnswerService interface {
	GetAnswerByID(ctx context.Context, answerID int64) (*edu.Answer, error)
	UpdateAnswer(ctx context.Context, aID int64, answer string) error
}

type App struct {
	UserService
	QuestionService
	UserQuestionService
	TagService
	AnswerService
	Dispatcher
}

type OptFunc = func(*App)

func NewDomain(opts ...OptFunc) *App {
	app := &App{}

	for _, opt := range opts {
		opt(app)
	}

	return app
}
