package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env"
	"github.com/grishagavrin/link-shortener/internal/errs"
)

const (
	ServerAddress   = "ServerAddress"
	BaseURL         = "BaseURL"
	FileStoragePath = "FileStoragePath"
	DatabaseDSN     = "DatabaseDSN"
	LENHASH         = 16
)

type myConfig struct {
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:"127.0.0.1:8080"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" envDefault:"filedata"`
	DatabaseDSN     string `env:"DATABASE_DSN" envDefault:""`
}

var instance *myConfig

func Instance() *myConfig {
	if instance == nil {
		instance = new(myConfig)
		instance.initENV()
		instance.initFlags()
	}

	return instance
}

func (c *myConfig) initENV() {
	if err := env.Parse(c); err != nil {
		log.Fatalf("Can`t load ENV %+v\n", err)
	}
}

func (c *myConfig) initFlags() {
	aFlag := flag.String("a", "", "")
	bFlag := flag.String("b", "", "")
	fFlag := flag.String("f", "", "")
	dFlag := flag.String("d", "", "")
	flag.Parse()

	if *aFlag != "" {
		c.ServerAddress = *aFlag
	}
	if *bFlag != "" {
		c.BaseURL = *bFlag
	}
	if *fFlag != "" {
		c.FileStoragePath = *fFlag
	}
	if *dFlag != "" {
		c.DatabaseDSN = *dFlag
	}
}

func (c *myConfig) GetCfgValue(env string) (string, error) {
	switch env {
	case ServerAddress:
		return c.ServerAddress, nil
	case BaseURL:
		return c.BaseURL, nil
	case FileStoragePath:
		return c.FileStoragePath, nil
	case DatabaseDSN:
		return c.DatabaseDSN, nil
	}

	return "", errs.ErrUnknownEnvOrFlag
}
