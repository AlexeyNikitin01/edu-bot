package config

import (
	"os"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

var Cfg *Config = nil

type Config struct {
	PSQL  *PG    `yaml:"psql"`
	CACHE *Redis `yaml:"cache"`
	Token string `yaml:"token"`
}

type Redis struct {
	Addr            string        `yaml:"addr" json:"addr" env:"REDIS_ADDR" env-default:"localhost:6379"`
	Password        string        `yaml:"pass" json:"pass" env:"REDIS_PASSWORD" env-default:""`
	DB              int           `yaml:"db" json:"db" env:"REDIS_DB" env-default:"0"`
	User            string        `yaml:"user" json:"user" env:"REDIS_USER" env-default:""`
	MaxRetries      int           `yaml:"max_retries" json:"max_retries" env:"REDIS_MAX_RETRIES" env-default:"3"`
	DialTimeout     time.Duration `yaml:"dial_timeout" json:"dial_timeout" env:"REDIS_DIAL_TIMEOUT" env-default:"5s"`
	ReadTimeout     time.Duration `yaml:"read_timeout" json:"read_timeout" env:"REDIS_READ_TIMEOUT" env-default:"3s"`
	WriteTimeout    time.Duration `yaml:"write_timeout" json:"write_timeout" env:"REDIS_WRITE_TIMEOUT" env-default:"3s"`
	PoolSize        int           `yaml:"pool_size" json:"pool_size" env:"REDIS_POOL_SIZE" env-default:"10"`
	MinIdleConns    int           `yaml:"min_idle_conns" json:"min_idle_conns" env:"REDIS_MIN_IDLE_CONNS" env-default:"5"`
	MaxIdleConns    int           `yaml:"max_idle_conns" json:"max_idle_conns" env:"REDIS_MAX_IDLE_CONNS" env-default:"10"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" json:"conn_max_lifetime" env:"REDIS_CONN_MAX_LIFETIME" env-default:"30m"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time" json:"conn_max_idle_time" env:"REDIS_CONN_MAX_IDLE_TIME" env-default:"5m"`

	TTL *RedisTTL `yaml:"ttl" json:"ttl"`
}

type RedisTTL struct {
	Day time.Duration `yaml:"day" json:"day"`
}

// GetTTL возвращает TTL в днях
func (r *Redis) GetTTL() time.Duration {
	if r.TTL != nil && r.TTL.Day > 0 {
		return r.TTL.Day
	}
	return 7 * 24 * time.Hour // значение по умолчанию: 7 дней
}

type PG struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	DBName   string `yaml:"dbname"`
	Password string `yaml:"pass"`
	SSLmode  string `yaml:"sslmode"`
	DebugPG  bool   `yaml:"debug_pg"`
}

var PATH_TO_CFG = "./etc/config.yml"

func NewCfg() (*Config, error) {
	yamlFile, err := os.ReadFile(PATH_TO_CFG)
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

func GetConfig() *Config {
	if Cfg == nil {
		c, err := NewCfg()
		if err != nil {
			panic(err)
		}
		Cfg = c
	}
	return Cfg
}
