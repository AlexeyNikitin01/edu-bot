package main

import (
	"context"
	"github.com/redis/go-redis/v9"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/pkg/errors"
	"gopkg.in/telebot.v3"

	"github.com/aarondl/sqlboiler/v4/boil"

	postgresMigrate "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"bot/cmd/cfg"
	"bot/internal/adapters"
	"bot/internal/app"
	"bot/internal/ports"
)

func main() {
	log.Println("Starting bot...")
	ctx := context.Background()

	connectDB(cfg.GetConfig().PSQL)
	bot := connectBot(cfg.GetConfig().Token)
	domain := app.NewApp()
	cache := app.NewRedisUserCache(connectRedis(ctx, cfg.GetConfig().CACHE))
	dispatcher := ports.NewDispatcher(ctx, domain, bot, cache)

	ports.StartBot(
		ctx,
		bot,
		domain,
		cache,
		dispatcher,
	)
	waitForShutdown(bot, dispatcher)
}

func waitForShutdown(bot *telebot.Bot, d *ports.QuestionDispatcher) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	<-sigChan

	d.Stop()
	bot.Stop()

	log.Println("Приложение завершено")
	os.Exit(0)
}

func connectDB(cfg *cfg.PG) {
	db, err := adapters.OpenConnectPostgres(cfg)
	if err != nil {
		log.Fatal(err)
	}
	boil.SetDB(db)

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

func connectBot(token string) *telebot.Bot {
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

func connectRedis(ctx context.Context, cfg *cfg.Redis) *redis.Client {
	r, err := adapters.NewClientRedis(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}
	return r
}
