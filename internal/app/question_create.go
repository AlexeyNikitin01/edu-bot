package app

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"bot/internal/repo/edu"
	"github.com/aarondl/sqlboiler/v4/boil"
)

func (a *App) SaveQuestions(ctx context.Context, question, tag string, answers []string, userID int64) (err error) {
	eduTag, err := edu.Tags(
		edu.TagWhere.Tag.EQ(tag),
	).One(ctx, boil.GetContextDB())
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	} else if errors.Is(err, sql.ErrNoRows) {
		eduTag = &edu.Tag{
			Tag: tag,
		}
		if err = eduTag.Insert(ctx, boil.GetContextDB(), boil.Infer()); err != nil {
			return err
		}
		if err = eduTag.Reload(ctx, boil.GetContextDB()); err != nil {
			return err
		}
	}

	q := &edu.Question{
		Question: question,
		TagID:    eduTag.ID,
	}
	if strings.HasPrefix(question, "ЗАДАЧА") {
		q.IsTask = true
	}
	if err = q.Insert(ctx, boil.GetContextDB(), boil.Infer()); err != nil {
		return err
	}

	if err = q.Reload(ctx, boil.GetContextDB()); err != nil {
		return err
	}

	for i, answer := range answers {
		answr := edu.Answer{
			QuestionID: q.ID,
			Answer:     answer,
			IsCorrect:  i == 0,
		}
		if err = answr.Insert(ctx, boil.GetContextDB(), boil.Infer()); err != nil {
			return err
		}
	}

	uq := edu.UsersQuestion{
		QuestionID: q.ID,
		UserID:     userID,
		IsEdu:      true,
		TimeRepeat: time.Now().Add(time.Minute * 5),
	}
	if err = uq.Insert(ctx, boil.GetContextDB(), boil.Infer()); err != nil {
		return
	}

	return nil
}
