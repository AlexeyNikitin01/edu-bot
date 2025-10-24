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

func (q Question) GetAllQuestions(ctx context.Context, userID int64, tag string) (edu.QuestionSlice, error) {
	qs, err := edu.Questions(
		qm.InnerJoin(fmt.Sprintf("%s ON %s = %s", edu.TableNames.UsersQuestions,
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
	).All(ctx, boil.GetContextDB())
	if err != nil {
		return nil, err
	}

	return qs, nil
}
