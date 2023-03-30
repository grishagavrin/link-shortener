package config

import (
	"errors"
	"flag"
	"log"

	"github.com/caarlos0/env"
)

var errUnknownEnvOrFlag = errors.New("unknown env or flag param")

type myConfig struct {
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:"127.0.0.1:8080"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" envDefault:""`
}

const (
	ServerAddress   = "ServerAddress"
	BaseURL         = "BaseURL"
	FileStoragePath = "FileStoragePath"
	LENHASH         = 16
	HashSymbols     = "1234567890qwertyuiopasdfghjklzxcvbnm"
)

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
		log.Fatalf("can`t load ENV %+v\n", err)
	}
}

func (c *myConfig) initFlags() {
	aFlag := flag.String("a", "", "")
	bFlag := flag.String("b", "", "")
	fFlag := flag.String("f", "", "")
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
}

func (c *myConfig) GetCfgValue(env string) (string, error) {
	switch env {
	case ServerAddress:
		return c.ServerAddress, nil
	case BaseURL:
		return c.BaseURL, nil
	case FileStoragePath:
		return c.FileStoragePath, nil
	}

	return "", errUnknownEnvOrFlag
}
