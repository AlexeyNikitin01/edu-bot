package main

import (
	"bot/cmd/cfg"
	initRedis "bot/internal/adapters/cache/redis"
	"bot/internal/adapters/db/postres"
	"context"
	"errors"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/golang-migrate/migrate/v4"
	postgresMigrate "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/redis/go-redis/v9"
	"gopkg.in/telebot.v3"
	"log"
	"time"
)

func ConnectDB(cfg *cfg.PG) {
	db, err := postres.OpenConnectPostgres(cfg)
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
