package main

import (
	initRedis "bot/internal/adapters/cache/redis"
	"bot/internal/domain/distpatcher"
	"bot/internal/domain/user"
	"bot/internal/domain/userQuestion"
	"context"
	"gopkg.in/telebot.v3"
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/golang-migrate/migrate/v4/source/file"

	"bot/cmd/cfg"
	"bot/internal/domain"
	"bot/internal/ports"
)

func main() {
	log.Println("Starting bot...")
	ctx := context.Background()

	// инициализация подключений к сторонним модулям
	ConnectDB(cfg.GetConfig().PSQL)
	c := initRedis.NewCache(ConnectRedis(ctx, cfg.GetConfig().CACHE))
	bot := ConnectBot(cfg.GetConfig().Token)

	// инициализация сервисов
	a := domain.NewDomain(
		domain.WithDefaultUserService(),
		domain.WithDefaultQuestionService(),
		domain.WithDefaultUserQuestionService(),
		domain.WithDefaultTagService(),
		domain.WithDefaultAnswerService(),
		domain.WithDefaultDispatcher(distpatcher.NewDispatcher(user.NewUser(), userQuestion.NewUserQuestion(), c)),
	)

	// запуск транспортного слоя
	ports.StartBot(ctx, bot, a)

	// gracefully shutdown
	waitForShutdown(bot, a)
}

func waitForShutdown(bot *telebot.Bot, d domain.UseCases) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	<-sigChan

	d.Stop()
	bot.Stop()

	log.Println("Приложение завершено")
	os.Exit(0)
}
