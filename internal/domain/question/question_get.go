package question

import (
	"bot/internal/repo/edu"
	"context"
	"fmt"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"log"
)

func (q Question) GetQuestionAnswers(ctx context.Context, qID int64) (*edu.Question, error) {
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

func (q Question) GetAllQuestionsWithPagination(
	ctx context.Context, userID int64, tag string, limit, offset int,
) (edu.UsersQuestionSlice, int, error) {
	qs, err := edu.UsersQuestions(
		qm.Load(qm.Rels(edu.UsersQuestionRels.Question)),
		qm.Load(qm.Rels(edu.UsersQuestionRels.User)),
		qm.Load(qm.Rels(edu.UsersQuestionRels.Question, edu.QuestionRels.Tag)),

		qm.InnerJoin(fmt.Sprintf("%s ON %s = %s", edu.TableNames.Questions,
			edu.QuestionTableColumns.ID,
			edu.UsersQuestionTableColumns.QuestionID,
		)),
		qm.InnerJoin(fmt.Sprintf("%s ON %s = %s", edu.TableNames.Tags,
			edu.TagTableColumns.ID,
			edu.QuestionTableColumns.TagID,
		)),
		edu.UsersQuestionWhere.UserID.EQ(userID),
		edu.TagWhere.Tag.EQ(tag),
		edu.UsersQuestionWhere.DeletedAt.IsNull(),
		qm.Limit(limit),
		qm.Offset(offset),
		qm.OrderBy(fmt.Sprintf("%s DESC, %s DESC", edu.QuestionTableColumns.CreatedAt, edu.QuestionTableColumns.ID)),
	).All(ctx, boil.GetContextDB())
	if err != nil {
		return nil, 0, err
	}

	// Получаем общее количество вопросов
	totalCount, err := edu.UsersQuestions(
		qm.InnerJoin(fmt.Sprintf("%s ON %s = %s", edu.TableNames.Questions,
			edu.QuestionTableColumns.ID,
			edu.UsersQuestionTableColumns.QuestionID,
		)),
		qm.InnerJoin(fmt.Sprintf("%s ON %s = %s", edu.TableNames.Tags,
			edu.TagTableColumns.ID,
			edu.QuestionTableColumns.TagID,
		)),
		edu.UsersQuestionWhere.UserID.EQ(userID),
		edu.TagWhere.Tag.EQ(tag),
		edu.UsersQuestionWhere.DeletedAt.IsNull(),
	).Count(ctx, boil.GetContextDB())
	if err != nil {
		return nil, 0, err
	}

	return qs, int(totalCount), nil
}
