package cfg

import (
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type Config struct {
	PSQL PG `yaml:"psql"`
}

type PG struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	DBName   string `yaml:"dbname"`
	Password string `yaml:"pass"`
	SSLmode  string `yaml:"sslmode"`
}

func NewCfgPostgres() (*Config, error) {
	yamlFile, err := os.ReadFile("./etc/config.yml")
	if err != nil {
		return nil, errors.Wrap(err, "read file config")
	}

	var config *Config

	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal")
	}

	return config, nil
}
