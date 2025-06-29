package ports

import (
	"errors"

	"github.com/volatiletech/sqlboiler/v4/boil"
	"gopkg.in/telebot.v3"

	"bot/internal/repo/edu"
)

var ErrUserNotFound = errors.New("Не зарегистрирован")

func GetUser(ctx telebot.Context) (*edu.User, error) {
	tgUser := ctx.Sender()
	userID := tgUser.ID

	u, err := edu.Users(edu.UserWhere.TGUserID.EQ(userID)).One(GetContext(ctx), boil.GetContextDB())
	if err != nil {
		return nil, ctx.Send(errors.Join(ErrUserNotFound, err).Error())
	}

	return u, nil
}
