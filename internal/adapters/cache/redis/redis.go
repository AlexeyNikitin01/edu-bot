package redis

import (
	"bot/internal/config"
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

func NewClientRedis(ctx context.Context, cfg *config.Redis) (*redis.Client, error) {
	db := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		Username:     cfg.User,
		MaxRetries:   cfg.MaxRetries,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	})

	if err := db.Ping(ctx).Err(); err != nil {
		fmt.Printf("failed to connect to redis server: %s\n", err.Error())
		return nil, err
	}

	return db, nil
}

type RedisCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewCache(client *redis.Client) *RedisCache {
	return &RedisCache{
		client: client,
		ttl:    config.GetConfig().CACHE.GetTTL(),
	}
}
