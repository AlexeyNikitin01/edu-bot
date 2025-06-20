package main

import (
	"context"
	"log"
	"time"

	"gopkg.in/telebot.v3"

	"github.com/volatiletech/sqlboiler/v4/boil"

	"bot/internal/adapters"
	"bot/internal/app"
	"bot/internal/ports"
)

func main() {
	log.Println("Starting bot...")

	ctx := context.Background()

	db, err := adapters.OpenConnectPostgres(&adapters.Config{
		Host:   "localhost",
		Port:   "7878",
		User:   "postgres",
		Dbname: "edu",
		Pass:   "pass",
		SSL:    "disable",
	})
	if err != nil {
		log.Fatal(err)
	}
	boil.SetDB(db)

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
