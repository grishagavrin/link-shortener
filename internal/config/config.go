package config

import (
	"log"
	"sync"

	"github.com/caarlos0/env"
)

type ConfigENV struct {
	MU            sync.RWMutex
	ServerAddress string `env:"SERVER_ADDRESS" envDefault:"127.0.0.1:8080"`
	BaseURL       string `env:"BASE_URL" envDefault:"http://localhost:8080"`
}

const (
	LENHASH = 16
)

func (cfg *ConfigENV) GetENVServer() string {
	cfg.MU.RLock()
	defer cfg.MU.RUnlock()

	if err := env.Parse(cfg); err != nil {
		log.Fatalf("can`t load ENV %+v\n", err)
	}

	return cfg.ServerAddress
}

func (cfg *ConfigENV) GetENVBaseUrl() string {
	cfg.MU.RLock()
	defer cfg.MU.RUnlock()

	if err := env.Parse(cfg); err != nil {
		log.Fatalf("can`t load ENV %+v\n", err)
	}

	return cfg.BaseURL
}
