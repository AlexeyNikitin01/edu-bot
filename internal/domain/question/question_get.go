package question

import (
	"bot/internal/repo/edu"
	"context"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"log"
)

func (q Question) GetQuestionAnswers(ctx context.Context, qID int64) (*edu.Question, error) {
	question, err := edu.Questions(
		qm.Load(qm.Rels(edu.QuestionRels.Answers)),
		qm.Load(qm.Rels(edu.QuestionRels.Tag)),
		edu.QuestionWhere.ID.EQ(qID),
	).One(ctx, boil.GetContextDB())
	if err != nil {
		log.Println("Ошибка при выборке вопроса:", err)
		return nil, err
	}

	return question, nil
}
