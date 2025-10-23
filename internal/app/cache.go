package app

import (
	"bot/internal/repo/dto"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

type UserCacher interface {
	SetUserWaiting(ctx context.Context, userID int64, waiting bool) error
	GetUserWaiting(ctx context.Context, userID int64) (bool, error)

	AddWorker(ctx context.Context, userID int64) error
	RemoveWorker(ctx context.Context, userID int64) error
	GetActiveWorkers(ctx context.Context) ([]int64, error)

	DraftCacher
}

type RedisUserCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisUserCache(client *redis.Client) *RedisUserCache {
	return &RedisUserCache{
		client: client,
		ttl:    24 * time.Hour,
	}
}

func (r *RedisUserCache) SetUserWaiting(ctx context.Context, userID int64, waiting bool) error {
	key := fmt.Sprintf("user:%d:waiting", userID)
	return r.client.Set(ctx, key, waiting, r.ttl).Err()
}

func (r *RedisUserCache) GetUserWaiting(ctx context.Context, userID int64) (bool, error) {
	key := fmt.Sprintf("user:%d:waiting", userID)
	result, err := r.client.Get(ctx, key).Bool()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	return result, err
}

func (r *RedisUserCache) AddToSet(ctx context.Context, key string, value string) error {
	return r.client.SAdd(ctx, key, value).Err()
}

func (r *RedisUserCache) RemoveFromSet(ctx context.Context, key string, value string) error {
	return r.client.SRem(ctx, key, value).Err()
}

func (r *RedisUserCache) GetSetMembers(ctx context.Context, key string) ([]string, error) {
	return r.client.SMembers(ctx, key).Result()
}

func (r *RedisUserCache) AddWorker(ctx context.Context, userID int64) error {
	return r.AddToSet(ctx, "bot:active_workers", strconv.FormatInt(userID, 10))
}

func (r *RedisUserCache) RemoveWorker(ctx context.Context, userID int64) error {
	return r.RemoveFromSet(ctx, "bot:active_workers", strconv.FormatInt(userID, 10))
}

func (r *RedisUserCache) GetActiveWorkers(ctx context.Context) ([]int64, error) {
	members, err := r.GetSetMembers(ctx, "bot:active_workers")
	if err != nil {
		return nil, err
	}

	userIDs := make([]int64, 0, len(members))
	for _, member := range members {
		if userID, err := strconv.ParseInt(member, 10, 64); err == nil {
			userIDs = append(userIDs, userID)
		}
	}

	return userIDs, nil
}

type DraftCacher interface {
	SaveDraft(ctx context.Context, userID int64, draft *dto.QuestionDraft) error
	GetDraft(ctx context.Context, userID int64) (*dto.QuestionDraft, error)
	DeleteDraft(ctx context.Context, userID int64) error
}

func (r *RedisUserCache) key(userID int64) string {
	return fmt.Sprintf("user:%d:draft", userID)
}

func (r *RedisUserCache) SaveDraft(ctx context.Context, userID int64, draft *dto.QuestionDraft) error {
	data, err := json.Marshal(draft)
	if err != nil {
		return err
	}
	return r.client.SetEx(ctx, r.key(userID), data, r.ttl).Err()
}

func (r *RedisUserCache) GetDraft(ctx context.Context, userID int64) (*dto.QuestionDraft, error) {
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

func (r *RedisUserCache) DeleteDraft(ctx context.Context, userID int64) error {
	return r.client.Del(ctx, r.key(userID)).Err()
}
