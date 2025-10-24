package domain

import (
	"bot/internal/domain/answer"
	"bot/internal/domain/question"
	"bot/internal/domain/tag"
	"bot/internal/domain/user"
	"bot/internal/domain/userQuestion"
)

func WithDefaultUserService() OptFunc {
	return func(app *App) {
		app.UserService = user.NewUser()
	}
}

func WithDefaultQuestionService() OptFunc {
	return func(app *App) {
		app.QuestionService = question.NewQuestion()
	}
}

func WithDefaultUserQuestionService() OptFunc {
	return func(app *App) {
		app.UserQuestionService = userQuestion.NewUserQuestion()
	}
}

func WithDefaultTagService() OptFunc {
	return func(app *App) {
		app.TagService = tag.NewTag()
	}
}

func WithDefaultAnswerService() OptFunc {
	return func(app *App) {
		app.AnswerService = answer.NewAnswer()
	}
}

func WithDefaultDispatcher(d Dispatcher) OptFunc {
	return func(app *App) {
		app.Dispatcher = d
	}
}
