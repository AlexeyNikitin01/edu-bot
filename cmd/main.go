package main

import (
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
	pg := adapters.NewPostgres(db)
	domain := app.NewApp(pg)

	boil.SetDB(db)

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

	ports.StartBot(b, domain)
}
