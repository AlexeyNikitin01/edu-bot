package adapters

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Face interface {
	GetQuestions(ctx context.Context) (any, error)
}

type Config struct {
	Host   string `json:"host,omitempty" yaml:"host"`
	Port   string `json:"port,omitempty" yaml:"port"`
	User   string `json:"user,omitempty" yaml:"user"`
	Dbname string `json:"dbname,omitempty" yaml:"dbname"`
	Pass   string `json:"pass,omitempty" yaml:"pass"`
	SSL    string `json:"ssl,omitempty" yaml:"ssl"`
}

func OpenConnectPostgres(cfg *Config) (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Dbname, cfg.Pass, cfg.SSL))
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

type Postgres struct {
	DB *sqlx.DB
}

func NewPostgres(DB *sqlx.DB) Face {
	return &Postgres{DB: DB}
}
