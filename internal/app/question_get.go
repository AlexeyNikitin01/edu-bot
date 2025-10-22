package app

import (
	"bot/internal/repo/edu"
	"context"
	"database/sql"
	"fmt"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/pkg/errors"
	"log"
	"time"
)

func (a *App) GetRandomNearestQuestionWithAnswer(ctx context.Context, userID int64) (*edu.UsersQuestion, error) {
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
		edu.AnswerWhere.IsCorrect.EQ(true),
		qm.OrderBy("RANDOM()"),
	).One(ctx, boil.GetContextDB())
	if err != nil {
		log.Println("Ошибка при выборке вопросов:", err)
		return nil, err
	}

	return questions, nil
}

func (a *App) GetUserQuestion(ctx context.Context, userID, qID int64) (*edu.UsersQuestion, error) {
	questions, err := edu.UsersQuestions(
		qm.Load(qm.Rels(edu.UsersQuestionRels.Question, edu.QuestionRels.Answers)),
		qm.Load(qm.Rels(edu.UsersQuestionRels.Question, edu.QuestionRels.Tag)),
		qm.InnerJoin(
			fmt.Sprintf("%s ON %s = %s",
				edu.TableNames.Questions,
				edu.QuestionTableColumns.ID,
				edu.UsersQuestionTableColumns.QuestionID,
			),
		),
		edu.UsersQuestionWhere.UserID.EQ(userID),
		edu.QuestionWhere.ID.EQ(qID),
	).One(ctx, boil.GetContextDB())
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
		qm.OrderBy("RANDOM()"),
	).One(ctx, boil.GetContextDB())
	if err != nil {
		log.Println("Ошибка при выборке вопроса:", err)
		return nil, err
	}

	return t, nil
}

// GetUniqueTagsByTask Функция для получения уникальных тегов по задачам
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

func (a *App) GetTagByID(ctx context.Context, tagID int64) (*edu.Tag, error) {
	return edu.FindTag(ctx, boil.GetContextDB(), tagID)
}

func (a *App) GetAnswerByID(ctx context.Context, answerID int64) (*edu.Answer, error) {
	return edu.FindAnswer(ctx, boil.GetContextDB(), answerID)
}
