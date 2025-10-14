package app

import (
	"bot/internal/repo/edu"
	"context"
	"database/sql"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/pkg/errors"
	"math"
	"math/rand"
	"time"
)

func (a *App) UpdateRepeatTime(ctx context.Context, question *edu.UsersQuestion, correct bool) error {
	var serial int64

	if correct {
		serial = question.TotalSerial + 1
		question.TotalSerial++
		question.TotalCorrect++
	} else {
		question.TotalWrong++
		question.TotalSerial = 0
	}

	switch serial {
	case 0:
		question.TimeRepeat = time.Now().Add(time.Second)
	case 1:
		randomOffset := time.Duration(rand.Intn(11)-5) * time.Minute
		question.TimeRepeat = time.Now().Add(60*time.Minute + randomOffset)
	case 2:
		randomOffset := time.Duration(rand.Intn(21)-10) * time.Minute
		question.TimeRepeat = time.Now().Add(120*time.Minute + randomOffset)
	case 3:
		randomOffset := time.Duration(rand.Intn(41)-20) * time.Minute
		question.TimeRepeat = time.Now().Add(240*time.Minute + randomOffset)
	case 4:
		randomOffset := time.Duration(rand.Intn(121)-60) * time.Minute
		question.TimeRepeat = time.Now().Add(12*time.Hour + randomOffset)
	case 5:
		randomOffset := time.Duration(rand.Intn(13)-6) * time.Hour
		question.TimeRepeat = time.Now().Add(24*time.Hour*3 + randomOffset)
	case 6:
		randomOffset := time.Duration(rand.Intn(25)-12) * time.Hour
		question.TimeRepeat = time.Now().Add(24*time.Hour*7 + randomOffset)
	default:
		// Для default увеличиваем интервал с каждым разом и добавляем случайность
		baseInterval := 24 * time.Hour * 7 * time.Duration(math.Pow(1.5, float64(serial-6)))
		randomOffset := time.Duration(rand.Intn(25)-12) * time.Hour
		question.TimeRepeat = time.Now().Add(baseInterval + randomOffset)
	}

	_, err := question.Update(ctx, boil.GetContextDB(),
		boil.Whitelist(
			edu.UsersQuestionColumns.TimeRepeat,
			edu.UsersQuestionColumns.TotalWrong,
			edu.UsersQuestionColumns.TotalCorrect,
			edu.UsersQuestionColumns.TotalSerial,
		))
	if err != nil {
		return err
	}

	return nil
}

func (a *App) UpdateIsEduUserQuestion(ctx context.Context, userID, questionID int64) error {
	uq, err := edu.UsersQuestions(
		edu.UsersQuestionWhere.UserID.EQ(userID),
		edu.UsersQuestionWhere.QuestionID.EQ(questionID),
		qm.Load(edu.UsersQuestionRels.Question),
	).One(ctx, boil.GetContextDB())
	if err != nil {
		return err
	}

	uq.IsEdu = !uq.IsEdu
	_, err = uq.Update(ctx, boil.GetContextDB(), boil.Infer())
	if err != nil {
		return err
	}

	return nil
}

func (a *App) UpdateTag(ctx context.Context, tagID int64, s string) error {
	tag, err := edu.FindTag(ctx, boil.GetContextDB(), tagID)
	if err != nil {
		return err
	}

	tag.Tag = s

	if _, err = tag.Update(ctx, boil.GetContextDB(), boil.Whitelist(edu.TagColumns.Tag)); err != nil {
		return err
	}

	return nil
}

func (a *App) UpdateQuestionName(ctx context.Context, qID int64, question string) error {
	q, err := edu.FindQuestion(ctx, boil.GetContextDB(), qID)
	if err != nil {
		return err
	}

	q.Question = question

	if _, err = q.Update(ctx, boil.GetContextDB(), boil.Whitelist(edu.QuestionColumns.Question)); err != nil {
		return err
	}

	return nil
}

func (a *App) UpdateAnswer(ctx context.Context, aID int64, answerText string) error {
	answer, err := edu.FindAnswer(ctx, boil.GetContextDB(), aID)
	if err != nil {
		return err
	}

	answer.Answer = answerText

	if _, err = answer.Update(ctx, boil.GetContextDB(), boil.Whitelist(edu.AnswerColumns.Answer)); err != nil {
		return err
	}

	return nil
}

func (a *App) UpdateTagByQuestion(ctx context.Context, qID int64, newTag string) error {
	t, err := edu.Tags(
		edu.TagWhere.Tag.EQ(newTag)).One(ctx, boil.GetContextDB())
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	} else if errors.Is(err, sql.ErrNoRows) {
		t = &edu.Tag{
			Tag: newTag,
		}
		if err = t.Insert(ctx, boil.GetContextDB(), boil.Infer()); err != nil {
			return err
		}
		if err = t.Reload(ctx, boil.GetContextDB()); err != nil {
			return err
		}
	}

	q, err := edu.FindQuestion(ctx, boil.GetContextDB(), qID)
	if err != nil {
		return err
	}

	q.TagID = t.ID

	if _, err = q.Update(ctx, boil.GetContextDB(), boil.Whitelist(edu.QuestionColumns.TagID)); err != nil {
		return err
	}

	return nil
}
