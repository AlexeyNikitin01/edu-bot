package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gopkg.in/telebot.v3"

	"github.com/volatiletech/sqlboiler/v4/boil"

	postgresMigrate "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"bot/internal/adapters"
	"bot/internal/app"
	"bot/internal/ports"
)

func main() {
	log.Println("Starting bot...")

	ctx := context.Background()

	db, err := adapters.OpenConnectPostgres(&adapters.Config{
		Host:   "db",
		Port:   "5432",
		User:   "postgres",
		Dbname: "edu",
		Pass:   "pass",
		SSL:    "disable",
	})
	if err != nil {
		log.Fatal(err)
	}
	boil.SetDB(db)

	if err = runMigrations(db); err != nil {
		log.Fatal(err)
	}

	domain := app.NewApp()

	// TODO config
	pref := telebot.Settings{
		Token:  "7700250115:AAFnBqR2zs7yHqBIhxVHwfgQiFv-33iHY8g",
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := telebot.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}

	ports.StartBot(ctx, b, domain)
}

func runMigrations(db *sqlx.DB) error {
	driver, err := postgresMigrate.WithInstance(db.DB, &postgresMigrate.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations/postgres",
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	log.Println("Migrations applied successfully")
	return nil
}
