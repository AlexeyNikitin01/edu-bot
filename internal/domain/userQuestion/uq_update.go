package userQuestion

import (
	"bot/internal/repo/edu"
	"context"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"math"
	"math/rand"
	"time"
)

func (UserQuestion) UpdateRepeatTime(ctx context.Context, question *edu.UsersQuestion, correct bool) error {
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

func (UserQuestion) UpdateIsEduUserQuestion(ctx context.Context, userID, questionID int64) error {
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
