package question

import (
	"bot/internal/repo/edu"
	"context"
	"github.com/aarondl/sqlboiler/v4/boil"
)

func (q Question) UpdateQuestionName(ctx context.Context, qID int64, question string) error {
	getQuestion, err := edu.FindQuestion(ctx, boil.GetContextDB(), qID)
	if err != nil {
		return err
	}

	getQuestion.Question = question

	if _, err = getQuestion.Update(ctx, boil.GetContextDB(), boil.Whitelist(edu.QuestionColumns.Question)); err != nil {
		return err
	}

	return nil
}
