package app

import (
	"context"

	"github.com/volatiletech/sqlboiler/v4/boil"

	"bot/internal/repo/edu"
)

func (a *App) GetOrCreate(ctx context.Context, userID, chatID int64, name string) (*edu.User, error) {
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

	return u, nil
}
