package tag

import (
	"bot/internal/repo/edu"
	"context"
	"database/sql"
	"fmt"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/pkg/errors"
)

type Tag struct{}

func NewTag() *Tag {
	return &Tag{}
}

func (t Tag) GetUniqueTags(ctx context.Context, userID int64, page, pageSize int) ([]*edu.Tag, int, error) {
	// Если pageSize = 0, получаем все теги (без пагинации)
	if pageSize <= 0 {
		// Получаем все теги без ограничений
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
			qm.OrderBy(edu.TagTableColumns.ID),
		).All(ctx, boil.GetContextDB())
		if err != nil {
			return nil, 0, err
		}

		// Получаем общее количество
		totalCount, err := edu.Tags(
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
		).Count(ctx, boil.GetContextDB())
		if err != nil {
			return nil, 0, err
		}

		return ts, int(totalCount), nil
	}

	// Вычисляем offset для пагинации
	offset := (page - 1) * pageSize

	// Получаем теги для текущей страницы
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
		qm.OrderBy(edu.TagTableColumns.ID),
		qm.Limit(pageSize),
		qm.Offset(offset),
	).All(ctx, boil.GetContextDB())
	if err != nil {
		return nil, 0, err
	}

	// Получаем общее количество тегов для пагинации
	totalCount, err := edu.Tags(
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
	).Count(ctx, boil.GetContextDB())
	if err != nil {
		return nil, 0, err
	}

	return ts, int(totalCount), nil
}

func (t Tag) UpdateTag(ctx context.Context, tagID int64, s string) error {
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

func (t Tag) UpdateTagByQuestion(ctx context.Context, qID int64, tagText string) error {
	getTag, err := edu.Tags(
		edu.TagWhere.Tag.EQ(tagText)).One(ctx, boil.GetContextDB())
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	} else if errors.Is(err, sql.ErrNoRows) {
		getTag = &edu.Tag{
			Tag: tagText,
		}
		if err = getTag.Insert(ctx, boil.GetContextDB(), boil.Infer()); err != nil {
			return err
		}
		if err = getTag.Reload(ctx, boil.GetContextDB()); err != nil {
			return err
		}
	}

	q, err := edu.FindQuestion(ctx, boil.GetContextDB(), qID)
	if err != nil {
		return err
	}

	q.TagID = getTag.ID

	if _, err = q.Update(ctx, boil.GetContextDB(), boil.Whitelist(edu.QuestionColumns.TagID)); err != nil {
		return err
	}

	return nil
}

func (t Tag) GetUniqueTagsByTask(ctx context.Context, userID int64) ([]*edu.Tag, error) {
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

func (t Tag) GetTagByID(ctx context.Context, tagID int64) (*edu.Tag, error) {
	return edu.FindTag(ctx, boil.GetContextDB(), tagID)
}
