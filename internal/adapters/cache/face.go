package cache

import (
	"bot/internal/repo/dto"
	"context"
)

type Cache interface {
	UserCache
	QuestionCache
}

type UserCache interface {
	SetUserWaiting(ctx context.Context, userID int64, waiting bool) error
	GetUserWaiting(ctx context.Context, userID int64) (bool, error)
	GetAllWaitingUsers(ctx context.Context) ([]int64, error) // Новый метод
}

type QuestionCache interface {
	SaveDraft(ctx context.Context, userID int64, draft *dto.QuestionDraft) error
	GetDraft(ctx context.Context, userID int64) (*dto.QuestionDraft, error)
	DeleteDraft(ctx context.Context, userID int64) error
}

type InstanceCache struct {
	Cache
}

func NewCache(c Cache) InstanceCache {
	return InstanceCache{
		Cache: c,
	}
}
