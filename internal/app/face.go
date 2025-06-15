package app

import (
	"bot/internal/adapters"
)

type Apper interface {
	questions()
}

type App struct {
	DB adapters.Face
}

func NewApp(DB adapters.Face) *App {
	return &App{DB: DB}
}
