package config

import (
	"flag"
	"log"
	"reflect"

	"github.com/caarlos0/env"
)

type ConfigENV struct {
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:"127.0.0.1:8080"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	// ServerAddress   string `env:"SERVER_ADDRESS" envDefault:""`
	// BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080"`
}

const (
	ServerAddress   = "ServerAddress"
	BaseURL         = "BaseURL"
	FileStoragePath = "FileStoragePath"
	LENHASH         = 16
	HashSymbols     = "1234567890qwertyuiopasdfghjklzxcvbnm"
)

var (
	aFlag string
	bFlag string
	fFlag string
)

func SetFlag() {
	flag.StringVar(&aFlag, "a", "127.0.0.1:8080", "default host and port")
	flag.StringVar(&bFlag, "b", "http://localhost:8080", "base url for response query")
	flag.StringVar(&fFlag, "f", "", "file storage location")
	flag.Parse()
}

func (cfg ConfigENV) GetEnvValue(fieldName string) (string, bool) {
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("can`t load ENV %+v\n", err)
	}

	values := reflect.ValueOf(cfg)
	typesOf := values.Type()
	var valForEqual string
	for i := 0; i < values.NumField(); i++ {
		if typesOf.Field(i).Name != fieldName {
			continue
		} else {
			valForEqual = values.Field(i).String()
			if valForEqual != "" {
				return valForEqual, true
			}
		}
	}

	switch fieldName {
	case ServerAddress:
		valForEqual = aFlag
	case BaseURL:
		valForEqual = bFlag
	case FileStoragePath:
		valForEqual = fFlag
	}

	if valForEqual != "" {
		return valForEqual, true
	}

	return "", false
}
