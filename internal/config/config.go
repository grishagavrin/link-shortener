package config

import (
	"log"

	"github.com/caarlos0/env"
)

type ConfigENV struct {
	ServerAddress string `env:"SERVER_ADDRESS" envDefault:"127.0.0.1:8080"`
	BaseURL       string `env:"BASE_URL" envDefault:"/api/shorten"`
}

const (
	LENHASH = 16
)

func GetENV() ConfigENV {
	cfgENV := ConfigENV{}
	if err := env.Parse(&cfgENV); err != nil {
		log.Fatalf("can`t load ENV %+v\n", err)
	}

	return cfgENV
}
