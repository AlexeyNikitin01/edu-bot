package app

import (
	"context"
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
