package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"

	"bot/internal/repo/edu"
)

func (a *App) GetQuestionsAnswers(ctx context.Context, userID int64) (edu.UsersQuestionSlice, error) {
	now := time.Now().UTC()

	questions, err := edu.UsersQuestions(
		qm.Load(qm.Rels(edu.UsersQuestionRels.Question, edu.QuestionRels.Answers)),
		qm.Load(qm.Rels(edu.UsersQuestionRels.Question, edu.QuestionRels.Tag)),
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
		qm.InnerJoin(
			fmt.Sprintf("%s ON %s = %s",
				edu.TableNames.Tags,
				edu.TagTableColumns.ID,
				edu.QuestionTableColumns.TagID)),
		edu.UsersQuestionWhere.TimeRepeat.LTE(now),
		edu.UsersQuestionWhere.UserID.EQ(userID),
		edu.UsersQuestionWhere.IsEdu.EQ(true),
		edu.UsersQuestionWhere.DeletedAt.IsNull(),
		edu.QuestionWhere.DeletedAt.IsNull(),
		edu.QuestionWhere.IsTask.EQ(false),
		edu.TagWhere.IsPause.EQ(false),
		edu.AnswerWhere.DeletedAt.IsNull(),
		qm.OrderBy("RANDOM()"),
	).All(ctx, boil.GetContextDB())
	if err != nil {
		log.Println("Ошибка при выборке вопросов:", err)
		return nil, err
	}

	return questions, nil
}

// GetTask todo: привязка будет по tag_id
func (a *App) GetTask(ctx context.Context, userID int64, tag string) (*edu.UsersQuestion, error) {
	// Сначала находим минимальный totalSerial для пользователя
	minSerial, err := edu.UsersQuestions(
		qm.InnerJoin(
			fmt.Sprintf("%s ON %s = %s",
				edu.TableNames.Questions,
				edu.QuestionTableColumns.ID,
				edu.UsersQuestionTableColumns.QuestionID),
		),
		qm.InnerJoin(
			fmt.Sprintf("%s ON %s = %s",
				edu.TableNames.Tags,
				edu.TagTableColumns.ID,
				edu.QuestionTableColumns.TagID),
		),
		edu.QuestionWhere.DeletedAt.IsNull(),
		edu.QuestionWhere.IsTask.EQ(true),
		edu.TagWhere.IsPause.EQ(false),
		edu.TagWhere.Tag.EQ(tag),
		qm.Select(edu.UsersQuestionColumns.TotalSerial),
		edu.UsersQuestionWhere.UserID.EQ(userID),
		edu.UsersQuestionWhere.IsEdu.EQ(true),
		edu.UsersQuestionWhere.DeletedAt.IsNull(),
		qm.OrderBy(edu.UsersQuestionColumns.TotalSerial),
	).One(ctx, boil.GetContextDB())
	if err != nil {
		log.Println("Ошибка при поиске минимального totalSerial:", err)
		return nil, err
	}

	t, err := edu.UsersQuestions(
		qm.Load(qm.Rels(edu.UsersQuestionRels.Question, edu.QuestionRels.Answers)),
		qm.Load(qm.Rels(edu.UsersQuestionRels.Question, edu.QuestionRels.Tag)),
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
		qm.InnerJoin(
			fmt.Sprintf("%s ON %s = %s",
				edu.TableNames.Tags,
				edu.TagTableColumns.ID,
				edu.QuestionTableColumns.TagID)),
		edu.UsersQuestionWhere.UserID.EQ(userID),
		edu.UsersQuestionWhere.IsEdu.EQ(true),
		edu.UsersQuestionWhere.DeletedAt.IsNull(),
		edu.UsersQuestionWhere.TotalSerial.EQ(minSerial.TotalSerial), // Фильтр по минимальному serial
		edu.QuestionWhere.DeletedAt.IsNull(),
		edu.QuestionWhere.IsTask.EQ(true),
		edu.TagWhere.IsPause.EQ(false),
		edu.TagWhere.Tag.EQ(tag),
		edu.AnswerWhere.DeletedAt.IsNull(),
	).One(ctx, boil.GetContextDB())
	if err != nil {
		log.Println("Ошибка при выборке вопроса:", err)
		return nil, err
	}

	return t, nil
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

// GetUniqueTags Функция для получения уникальных тегов по задачам
func (a *App) GetUniqueTagsByTask(ctx context.Context, userID int64) ([]*edu.Tag, error) {
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
		edu.QuestionWhere.IsTask.EQ(true),
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
		edu.UsersQuestionWhere.IsEdu.EQ(true),
		edu.UsersQuestionWhere.IsPause.EQ(false),
		qm.InnerJoin(
			fmt.Sprintf("%s ON %s = %s",
				edu.TableNames.Questions,
				edu.QuestionTableColumns.ID,
				edu.UsersQuestionTableColumns.QuestionID)),
		qm.InnerJoin(
			fmt.Sprintf("%s ON %s = %s",
				edu.TableNames.Tags,
				edu.TagTableColumns.ID,
				edu.QuestionTableColumns.TagID)),
		edu.TagWhere.IsPause.EQ(false),
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
		edu.UsersQuestionWhere.IsEdu.EQ(true),
		edu.UsersQuestionWhere.IsPause.EQ(false),
		qm.InnerJoin(
			fmt.Sprintf("%s ON %s = %s",
				edu.TableNames.Questions,
				edu.QuestionTableColumns.ID,
				edu.UsersQuestionTableColumns.QuestionID)),
		qm.InnerJoin(
			fmt.Sprintf("%s ON %s = %s",
				edu.TableNames.Tags,
				edu.TagTableColumns.ID,
				edu.QuestionTableColumns.TagID)),
		edu.TagWhere.IsPause.EQ(false),
	).One(ctx, boil.GetContextDB())
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return time.Time{}, err
	} else if errors.Is(err, sql.ErrNoRows) {
		return time.Now().UTC(), nil
	}

	return usersQuestion.TimeRepeat, nil
}
