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

func WithUserQuestionService(uq *userQuestion.UserQuestion) OptFunc {
	return func(app *App) {
		app.UserQuestionService = uq
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

func WithDispatcher(d Dispatcher) OptFunc {
	return func(app *App) {
		app.Dispatcher = d
	}
}
