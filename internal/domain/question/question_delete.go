package question

import (
	"bot/internal/repo/edu"
	"context"
	"github.com/aarondl/sqlboiler/v4/boil"
)

func (Question) DeleteQuestion(ctx context.Context, qID int64) error {
	if _, err := edu.Questions(edu.QuestionWhere.ID.EQ(qID)).DeleteAll(ctx, boil.GetContextDB(), false); err != nil {
		return err
	}
	return nil
}
