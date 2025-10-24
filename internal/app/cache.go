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
	GetAllWaitingUsers(ctx context.Context) ([]int64, error) // Новый метод

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

// Ключ для множества всех ожидающих пользователей
func (r *RedisUserCache) waitingUsersSetKey() string {
	return "bot:waiting_users"
}

// Ключ для отдельного пользователя
func (r *RedisUserCache) userWaitingKey(userID int64) string {
	return fmt.Sprintf("user:%d:waiting", userID)
}

func (r *RedisUserCache) SetUserWaiting(ctx context.Context, userID int64, waiting bool) error {
	userKey := r.userWaitingKey(userID)
	setKey := r.waitingUsersSetKey()

	if waiting {
		// Добавляем пользователя в множество ожидающих и устанавливаем флаг
		if err := r.client.Set(ctx, userKey, true, r.ttl).Err(); err != nil {
			return err
		}
		return r.client.SAdd(ctx, setKey, userID).Err()
	} else {
		// Удаляем пользователя из множества ожидающих и сбрасываем флаг
		if err := r.client.Del(ctx, userKey).Err(); err != nil {
			return err
		}
		return r.client.SRem(ctx, setKey, userID).Err()
	}
}

func (r *RedisUserCache) GetUserWaiting(ctx context.Context, userID int64) (bool, error) {
	key := r.userWaitingKey(userID)
	result, err := r.client.Get(ctx, key).Bool()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	return result, err
}

// GetAllWaitingUsers возвращает ID всех пользователей, ожидающих ответ
func (r *RedisUserCache) GetAllWaitingUsers(ctx context.Context) ([]int64, error) {
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

// Альтернативная реализация через сканирование ключей (если предпочтительнее)
func (r *RedisUserCache) GetAllWaitingUsersScan(ctx context.Context) ([]int64, error) {
	var waitingUsers []int64
	var cursor uint64
	var keys []string
	var err error

	// Ищем все ключи по шаблону user:*:waiting
	pattern := r.userWaitingKey(0) // "user:*:waiting"

	for {
		keys, cursor, err = r.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return nil, err
		}

		for _, key := range keys {
			// Извлекаем userID из ключа "user:{id}:waiting"
			var userID int64
			_, err := fmt.Sscanf(key, "user:%d:waiting", &userID)
			if err == nil {
				// Проверяем, что пользователь действительно ожидает
				if waiting, err := r.GetUserWaiting(ctx, userID); err == nil && waiting {
					waitingUsers = append(waitingUsers, userID)
				}
			}
		}

		if cursor == 0 {
			break
		}
	}

	return waitingUsers, nil
}

// Остальные методы без изменений
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
