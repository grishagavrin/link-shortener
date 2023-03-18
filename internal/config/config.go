package config

import (
	"flag"
	"log"
	"reflect"

	"github.com/caarlos0/env"
)

type ConfigENV struct {
	ServerAddress string `env:"SERVER_ADDRESS" envDefault:"127.0.0.1:8080"`
	BaseURL       string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	// FileStoragePath string `env:"FILE_STORAGE_PATH"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" envDefault:"internal/storage/FileDB.log"`
}

const (
	LENHASH         = 16
	ServerAddress   = "ServerAddress"
	BaseURL         = "BaseURL"
	FileStoragePath = "FileStoragePath"
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
	flag.StringVar(&fFlag, "f", "internal/storage/FileDB.log", "file storage location")
	flag.Parse()
}

func (cfg ConfigENV) GetEnvValue(fieldName string) (string, bool) {
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("can`t load ENV %+v\n", err)
	}

	values := reflect.ValueOf(cfg)
	typesOf := values.Type()
	valForEq := ""
	for i := 0; i < values.NumField(); i++ {
		if typesOf.Field(i).Name == fieldName {
			valForEq = values.Field(i).String()
			// return values.Field(i).String(), true
		}
	}

	switch fieldName {
	case ServerAddress:
		if valForEq == "" {
			valForEq = aFlag
		}
	case BaseURL:
		if valForEq == "" {
			valForEq = bFlag
		}
	case FileStoragePath:
		if valForEq == "" {
			valForEq = fFlag
		}
	}

	if valForEq != "" {
		return valForEq, true
	}

	return valForEq, false
}
