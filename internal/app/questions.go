package app

import (
	"context"
	"log"
	"time"

	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"bot/internal/repo/edu"
)

func (a *App) GetQuestionsAnswers(ctx context.Context, userID int64) (edu.UsersQuestionSlice, error) {
	now := time.Now().UTC()

	questions, err := edu.UsersQuestions(
		qm.Load(qm.Rels(edu.UsersQuestionRels.Question, edu.QuestionRels.Answers)),
		edu.UsersQuestionWhere.TimeRepeat.LTE(now),
		edu.UsersQuestionWhere.UserID.EQ(userID),
		edu.UsersQuestionWhere.IsEdu.EQ(true),
	).All(ctx, boil.GetContextDB())

	if err != nil {
		log.Println("Ошибка при выборке вопросов:", err)
		return nil, err
	}

	return questions, nil
}

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
		question.TimeRepeat = time.Now().Add(10 * time.Minute)
	case 1:
		question.TimeRepeat = time.Now().Add(60 * time.Minute)
	case 2:
		question.TimeRepeat = time.Now().Add(120 * time.Minute)
	case 3:
		question.TimeRepeat = time.Now().Add(240 * time.Minute)
	case 4:
		question.TimeRepeat = time.Now().Add(12 * time.Hour)
	case 5:
		question.TimeRepeat = time.Now().Add(24 * time.Hour * 3)
	case 6:
		question.TimeRepeat = time.Now().Add(24 * time.Hour * 7)
	default:
		question.TimeRepeat = time.Now().Add(24 * time.Hour * 7)
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
