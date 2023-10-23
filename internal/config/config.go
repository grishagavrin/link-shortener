// Package config implement functions for env and project configs
package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env"
	"github.com/grishagavrin/link-shortener/internal/errs"
)

// Config consts for param config func
const (
	ServerAddress   = "ServerAddress"
	BaseURL         = "BaseURL"
	FileStoragePath = "FileStoragePath"
	DatabaseDSN     = "DatabaseDSN"
	EnableHTTPS     = "EnableHTTPS"
	LENHASH         = 16
)

// Config base struct with default initialize
type MyConfig struct {
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:"127.0.0.1:8080"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" envDefault:"../../filedata"`
	DatabaseDSN     string `env:"DATABASE_DSN" envDefault:""`
	EnableHTTPS     string `env:"ENABLE_HTTPS" envDefault:""`
}

// Instance variable of config
var instance *MyConfig

// First instance in main func
func Instance() (*MyConfig, error) {
	if instance == nil {
		instance = new(MyConfig)
		err := instance.initENV()
		if err != nil {
			return nil, err
		}
		instance.initFlags()
	}

	return instance, nil
}

// Parse env
func (c *MyConfig) initENV() error {
	if err := env.Parse(c); err != nil {
		return fmt.Errorf("%w: %v", errs.ErrENVLoading, err)
	}
	return nil
}

// Flag initialize for start app
func (c *MyConfig) initFlags() {
	aFlag := flag.String("a", "", "")
	bFlag := flag.String("b", "", "")
	fFlag := flag.String("f", "", "")
	dFlag := flag.String("d", "", "")
	sFlag := flag.String("s", "", "")
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
	if *sFlag != "" {
		c.EnableHTTPS = *sFlag
	}
}

// Get param config
func (c *MyConfig) GetCfgValue(env string) (string, error) {
	switch env {
	case ServerAddress:
		return c.ServerAddress, nil
	case BaseURL:
		return c.BaseURL, nil
	case FileStoragePath:
		return c.FileStoragePath, nil
	case DatabaseDSN:
		return c.DatabaseDSN, nil
	case EnableHTTPS:
		return c.EnableHTTPS, nil
	}

	return "", errs.ErrUnknownEnvOrFlag
}
