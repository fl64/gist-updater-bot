package cfg

import (
	env "github.com/caarlos0/env/v6"
)

type Cfg struct {
	TgToken string `env:"BOT_TOKEN"`
	Debug   bool   `env:"BOT_DEBUG" envDefault:"false"`
	Timeout int    `env:"BOT_TIMEOUT" envDefault:"30"`
	DbFile  string `env:"BOT_DB_FILE" envDefault:"./db"`
}

func GetConfig() (*Cfg, error) {
	c := &Cfg{}
	if err := env.Parse(c); err != nil {
		return nil, err
	}
	return c, nil
}
