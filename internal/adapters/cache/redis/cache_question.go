package redis

import (
	"bot/internal/repo/dto"
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

func (r *RedisCache) key(userID int64) string {
	return fmt.Sprintf("user:%d:draft", userID)
}

func (r *RedisCache) SaveDraft(ctx context.Context, userID int64, draft *dto.QuestionDraft) error {
	data, err := json.Marshal(draft)
	if err != nil {
		return err
	}
	return r.client.SetEx(ctx, r.key(userID), data, r.ttl).Err()
}

func (r *RedisCache) GetDraft(ctx context.Context, userID int64) (*dto.QuestionDraft, error) {
	data, err := r.client.Get(ctx, r.key(userID)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, nil // Черновик не найден - это нормально
	}
	if err != nil {
		return nil, err
	}

	var draft dto.QuestionDraft
	if err := json.Unmarshal(data, &draft); err != nil {
		return nil, err
	}

	return &draft, nil
}

func (r *RedisCache) DeleteDraft(ctx context.Context, userID int64) error {
	return r.client.Del(ctx, r.key(userID)).Err()
}
