package redis

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	_ "github.com/redis/go-redis/v9"
	"strconv"
)

func (r *RedisCache) waitingUsersSetKey() string {
	return "bot:waiting_users"
}

func (r *RedisCache) userWaitingKey(userID int64) string {
	return fmt.Sprintf("user:%d:waiting", userID)
}

func (r *RedisCache) SetUserWaiting(ctx context.Context, userID int64, waiting bool) error {
	userKey := r.userWaitingKey(userID)
	setKey := r.waitingUsersSetKey()

	if waiting {
		// Добавляем пользователя в множество ожидающих и устанавливаем флаг
		if err := r.client.Set(ctx, userKey, true, r.ttl).Err(); err != nil {
			return err
		}
		// Добавляем пользователя в множество
		if err := r.client.SAdd(ctx, setKey, userID).Err(); err != nil {
			return err
		}
		// Устанавливаем TTL для множества
		return r.client.Expire(ctx, setKey, r.ttl).Err()
	} else {
		// Удаляем пользователя из множества ожидающих и сбрасываем флаг
		if err := r.client.Del(ctx, userKey).Err(); err != nil {
			return err
		}
		return r.client.SRem(ctx, setKey, userID).Err()
	}
}

func (r *RedisCache) GetUserWaiting(ctx context.Context, userID int64) (bool, error) {
	key := r.userWaitingKey(userID)
	result, err := r.client.Get(ctx, key).Bool()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	return result, err
}

// GetAllWaitingUsers возвращает ID всех пользователей, ожидающих ответ
func (r *RedisCache) GetAllWaitingUsers(ctx context.Context) ([]int64, error) {
	setKey := r.waitingUsersSetKey()

	// Получаем все элементы множества как строки
	members, err := r.client.SMembers(ctx, setKey).Result()
	if err != nil {
		return nil, err
	}

	// Конвертируем строки в int64
	userIDs := make([]int64, 0, len(members))
	for _, member := range members {
		if userID, err := strconv.ParseInt(member, 10, 64); err == nil {
			userIDs = append(userIDs, userID)
		}
		// Можно добавить логирование для ошибок конвертации, если нужно
	}

	return userIDs, nil
}
