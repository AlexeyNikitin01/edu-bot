package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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
		qm.InnerJoin(
			fmt.Sprintf("%s ON %s = %s",
				edu.TableNames.Questions,
				edu.QuestionTableColumns.ID,
				edu.UsersQuestionTableColumns.QuestionID)),
		qm.InnerJoin(
			fmt.Sprintf("%s ON %s = %s",
				edu.TableNames.Answers,
				edu.QuestionTableColumns.ID,
				edu.AnswerTableColumns.QuestionID)),
		edu.UsersQuestionWhere.TimeRepeat.LTE(now),
		edu.UsersQuestionWhere.UserID.EQ(userID),
		edu.UsersQuestionWhere.IsEdu.EQ(true),
		edu.UsersQuestionWhere.DeletedAt.IsNull(),
		edu.QuestionWhere.DeletedAt.IsNull(),
		edu.AnswerWhere.DeletedAt.IsNull(),
		qm.OrderBy("RANDOM()"),
	).All(ctx, boil.GetContextDB())
	if err != nil {
		log.Println("Ошибка при выборке вопросов:", err)
		return nil, err
	}

	return questions, nil
}

func (a *App) GetQuestionAnswers(ctx context.Context, qID int64) (*edu.Question, error) {
	question, err := edu.Questions(
		qm.Load(qm.Rels(edu.QuestionRels.Answers)),
		qm.Load(qm.Rels(edu.QuestionRels.Tag)),
		edu.QuestionWhere.ID.EQ(qID),
	).One(ctx, boil.GetContextDB())
	if err != nil {
		log.Println("Ошибка при выборке вопроса:", err)
		return nil, err
	}

	return question, nil
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

// GetUniqueTags Функция для получения уникальных тегов
func (a *App) GetUniqueTags(ctx context.Context, userID int64) ([]*edu.Tag, error) {
	ts, err := edu.Tags(
		qm.InnerJoin(
			fmt.Sprintf("%s ON %s = %s",
				edu.TableNames.Questions,
				edu.TagTableColumns.ID,
				edu.QuestionTableColumns.TagID),
		),
		qm.InnerJoin(
			fmt.Sprintf("%s ON %s = %s",
				edu.TableNames.UsersQuestions,
				edu.UsersQuestionTableColumns.QuestionID,
				edu.QuestionTableColumns.ID),
		),
		edu.UsersQuestionWhere.UserID.EQ(userID),
		edu.UsersQuestionWhere.DeletedAt.IsNull(),
		qm.GroupBy(edu.TagTableColumns.ID),
	).All(ctx, boil.GetContextDB())
	if err != nil {
		return nil, err
	}

	return ts, nil
}

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

func (a *App) GetNearestTimeRepeat(ctx context.Context, userID int64) (time.Time, error) {
	exist, err := edu.UsersQuestions(
		edu.UsersQuestionWhere.UserID.EQ(userID),
		edu.UsersQuestionWhere.TimeRepeat.LTE(time.Now().UTC()),
		edu.UsersQuestionWhere.DeletedAt.IsNull(),
	).Exists(ctx, boil.GetContextDB())
	if err != nil {
		return time.Time{}, err
	}
	if exist {
		return time.Now().UTC(), nil
	}

	// Получаем ближайший вопрос по времени повторения
	usersQuestion, err := edu.UsersQuestions(
		edu.UsersQuestionWhere.UserID.EQ(userID),
		edu.UsersQuestionWhere.DeletedAt.IsNull(),
		qm.OrderBy(edu.UsersQuestionColumns.TimeRepeat+" ASC"),
	).One(ctx, boil.GetContextDB())
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return time.Time{}, err
	} else if errors.Is(err, sql.ErrNoRows) {
		return time.Now().UTC(), nil
	}

	return usersQuestion.TimeRepeat, nil
}
