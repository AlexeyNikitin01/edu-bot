package user

import (
	"bot/internal/repo/edu"
	"context"
	"github.com/aarondl/sqlboiler/v4/boil"
)

func (User) CreateUser(ctx context.Context, userID int64, chatID int64, name string) (*edu.User, error) {
	u := &edu.User{
		TGUserID:  userID,
		ChatID:    chatID,
		FirstName: name,
	}

	if err := u.Upsert(
		ctx,
		boil.GetContextDB(),
		true,
		[]string{edu.UserColumns.TGUserID},
		boil.Infer(),
		boil.Infer(),
	); err != nil {
		return nil, err
	}

	if err := u.Reload(ctx, boil.GetContextDB()); err != nil {
		return nil, err
	}

	return u, nil
}
