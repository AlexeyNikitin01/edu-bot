package userQuestion

import (
	"bot/internal/adapters/cache"
)

type UserQuestion struct {
	c cache.Cache
}

type OptUserQuestion = func(*UserQuestion)

func NewUserQuestion(opts ...OptUserQuestion) *UserQuestion {
	uq := &UserQuestion{}

	for _, opt := range opts {
		opt(uq)
	}

	return uq
}

func WithCacheUQ(c cache.Cache) OptUserQuestion {
	return func(u *UserQuestion) {
		u.c = c
	}
}
