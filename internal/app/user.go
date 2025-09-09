package app

import (
	"context"
	"database/sql"
	"errors"

	"github.com/aarondl/sqlboiler/v4/boil"

	"bot/internal/repo/edu"
)

func (a *App) GetUser(ctx context.Context, userID int64) (*edu.User, error) {
	u, err := edu.Users(edu.UserWhere.TGUserID.EQ(userID)).One(ctx, boil.GetContextDB())
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	} else if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	return u, nil
}

func (a *App) CreateUser(ctx context.Context, userID int64, chatID int64, name string) (*edu.User, error) {
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
