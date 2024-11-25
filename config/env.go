package config

import "github.com/caarlos0/env/v6"

type Config struct {
	Debug    bool   `env:"DEBUG,required"`
	Db       string `env:"DB,required"`
	LogLever string `env:"LOG_LEVEL"`
}

func New() (Config, error) {
	var e Config
	return e, env.Parse(&e)
}
