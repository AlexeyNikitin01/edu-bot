package user

import (
	"bot/internal/repo/edu"
	"context"
	"database/sql"
	"fmt"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/pkg/errors"
	"time"
)

func (User) GetUser(ctx context.Context, userID int64) (*edu.User, error) {
	u, err := edu.Users(edu.UserWhere.TGUserID.EQ(userID)).One(ctx, boil.GetContextDB())
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	} else if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	return u, nil
}

func (User) GetUsersForSend(ctx context.Context, activityUsers []int64) (edu.UserSlice, error) {
	users, err := edu.Users(
		qm.Select(edu.UserColumns.TGUserID),
		edu.UserWhere.TGUserID.NIN(activityUsers),
		edu.UserWhere.Block.EQ(false),
		qm.InnerJoin(
			fmt.Sprintf(
				"%s on %s = %s",
				edu.TableNames.UsersQuestions,
				edu.UsersQuestionTableColumns.UserID,
				edu.UserTableColumns.TGUserID),
		),
		edu.UsersQuestionWhere.IsEdu.EQ(true),
		edu.UsersQuestionWhere.IsPause.EQ(false),
		edu.UsersQuestionWhere.TimeRepeat.LTE(time.Now().UTC()),
		qm.GroupBy(edu.UserTableColumns.TGUserID),
	).All(ctx, boil.GetContextDB())
	if err != nil {
		return nil, err
	}

	return users, nil
}
