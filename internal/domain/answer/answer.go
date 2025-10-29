package answer

import (
	"bot/internal/repo/edu"
	"context"
	"github.com/aarondl/sqlboiler/v4/boil"
)

type Answer struct{}

func NewAnswer() *Answer {
	return &Answer{}
}

func (a Answer) GetAnswerByID(ctx context.Context, answerID int64) (*edu.Answer, error) {
	return edu.FindAnswer(ctx, boil.GetContextDB(), answerID)
}

func (a Answer) UpdateAnswer(ctx context.Context, aID int64, answerText string) error {
	answer, err := edu.FindAnswer(ctx, boil.GetContextDB(), aID)
	if err != nil {
		return err
	}

	answer.Answer = answerText

	if _, err = answer.Update(ctx, boil.GetContextDB(), boil.Whitelist(edu.AnswerColumns.Answer)); err != nil {
		return err
	}

	return nil
}
