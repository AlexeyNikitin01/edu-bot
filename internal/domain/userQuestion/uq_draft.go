package userQuestion

import (
	"bot/internal/repo/dto"
	"context"
)

func (u UserQuestion) SetDraftQuestion(ctx context.Context, userID int64, draftQuestion *dto.QuestionDraft) error {
	return u.c.SaveDraft(ctx, userID, draftQuestion)
}

func (u UserQuestion) GetDraftQuestion(ctx context.Context, userID int64) (*dto.QuestionDraft, error) {
	return u.c.GetDraft(ctx, userID)
}

func (u UserQuestion) DeleteDraftQuestion(ctx context.Context, userID int64) error {
	return u.c.DeleteDraft(ctx, userID)
}
