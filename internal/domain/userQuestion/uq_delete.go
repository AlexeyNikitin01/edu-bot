package userQuestion

import (
	"bot/internal/repo/edu"
	"context"
	"github.com/aarondl/sqlboiler/v4/boil"
)

func (u UserQuestion) DeleteQuestionUser(ctx context.Context, userID int64, qID int64) error {
	if _, err := edu.UsersQuestions(
		edu.UsersQuestionWhere.UserID.EQ(userID),
		edu.UsersQuestionWhere.QuestionID.EQ(qID),
	).DeleteAll(ctx, boil.GetContextDB(), false); err != nil {
		return err
	}

	return nil
}
