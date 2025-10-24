package tag

import (
	"bot/internal/repo/edu"
	"context"
	"fmt"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
)

func (t Tag) DeleteQuestionsByTag(ctx context.Context, userID int64, tag string) error {
	uq, err := edu.UsersQuestions(
		qm.InnerJoin(
			fmt.Sprintf("%s ON %s = %s",
				edu.TableNames.Questions,
				edu.UsersQuestionTableColumns.QuestionID,
				edu.QuestionTableColumns.ID,
			),
		),
		qm.InnerJoin(
			fmt.Sprintf("%s ON %s = %s",
				edu.TableNames.Tags,
				edu.QuestionTableColumns.TagID,
				edu.TagTableColumns.ID,
			),
		),
		edu.TagWhere.Tag.EQ(tag),
		edu.UsersQuestionWhere.UserID.EQ(userID),
	).All(ctx, boil.GetContextDB())
	if err != nil {
		return err
	}

	if _, err = uq.DeleteAll(ctx, boil.GetContextDB(), false); err != nil {
		return err
	}

	return nil
}
