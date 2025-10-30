package main

import (
	initRedis "bot/internal/adapters/cache/redis"
	"bot/internal/adapters/db/postres"
	"bot/internal/domain/distpatcher"
	"bot/internal/domain/user"
	"bot/internal/domain/userQuestion"
	"context"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/golang-migrate/migrate/v4"
	postgresMigrate "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"gopkg.in/telebot.v3"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

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
		domain.WithDefaultTagService(),
		domain.WithDefaultUserService(),
		domain.WithDefaultAnswerService(),
		domain.WithDefaultQuestionService(),
		domain.WithUserQuestionService(userQuestion.NewUserQuestion(userQuestion.WithCacheUQ(c))),
		domain.WithDispatcher(distpatcher.NewDispatcher(user.NewUser(), userQuestion.NewUserQuestion(), c)),
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

func ConnectDB(cfg *cfg.PG) {
	db, err := postres.OpenConnectPostgres(cfg)
	if err != nil {
		log.Fatal(err)
	}
	boil.SetDB(db)
	boil.DebugMode = cfg.DebugPG

	// авто-миграции
	driver, err := postgresMigrate.WithInstance(db.DB, &postgresMigrate.Config{
		MigrationsTable: "schema_migrations",
	})
	if err != nil {
		log.Fatal(err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations/postgres",
		"postgres",
		driver,
	)
	if err != nil {
		log.Fatal(err)
	}

	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatal(err)
	}

	log.Println("Migrations applied successfully")
}

func ConnectBot(token string) *telebot.Bot {
	pref := telebot.Settings{
		Token:  token,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := telebot.NewBot(pref)
	if err != nil {
		log.Fatal(err)
	}

	return b
}

func ConnectRedis(ctx context.Context, cfg *cfg.Redis) *redis.Client {
	r, err := initRedis.NewClientRedis(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}
	return r
}
