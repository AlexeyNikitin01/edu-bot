package app

import (
	"context"
	"log"
	"time"

	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"bot/internal/repo/edu"
)

func (a *App) GetQuestionsAnswers(ctx context.Context, userID int64) (edu.UsersQuestionSlice, error) {
	now := time.Now().UTC()

	questions, err := edu.UsersQuestions(
		qm.Load(qm.Rels(edu.UsersQuestionRels.Question, edu.QuestionRels.Answers)),
		edu.UsersQuestionWhere.TimeRepeat.LTE(now),
		edu.UsersQuestionWhere.UserID.EQ(userID),
	).All(ctx, boil.GetContextDB())

	if err != nil {
		log.Println("Ошибка при выборке вопросов:", err)
		return nil, err
	}

	return questions, nil
}

func (a *App) UpdateRepeatTime(ctx context.Context, question *edu.UsersQuestion) error {
	question.TimeRepeat = time.Now().Add(10 * time.Minute)
	_, err := question.Update(ctx, boil.GetContextDB(), boil.Whitelist("time_repeat"))
	if err != nil {
		return err
	}

	return nil
}
