package question

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"bot/internal/repo/edu"
	"github.com/aarondl/sqlboiler/v4/boil"
)

func (q Question) SaveQuestions(ctx context.Context, question, tag string, answers []string, userID int64) (err error) {
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

	newQuestion := &edu.Question{
		Question: question,
		TagID:    eduTag.ID,
	}
	if strings.HasPrefix(question, "ЗАДАЧА") {
		newQuestion.IsTask = true
	}
	if err = newQuestion.Insert(ctx, boil.GetContextDB(), boil.Infer()); err != nil {
		return err
	}

	if err = newQuestion.Reload(ctx, boil.GetContextDB()); err != nil {
		return err
	}

	for i, answer := range answers {
		answr := edu.Answer{
			QuestionID: newQuestion.ID,
			Answer:     answer,
			IsCorrect:  i == 0,
		}
		if err = answr.Insert(ctx, boil.GetContextDB(), boil.Infer()); err != nil {
			return err
		}
	}

	uq := edu.UsersQuestion{
		QuestionID: newQuestion.ID,
		UserID:     userID,
		IsEdu:      true,
		TimeRepeat: time.Now().Add(time.Minute * 5),
	}
	if err = uq.Insert(ctx, boil.GetContextDB(), boil.Infer()); err != nil {
		return
	}

	return nil
}
