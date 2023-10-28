// Package config implement functions for env and project configs
package config

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"

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
	Config          = "CONFIG"
)

// JSONConfig for json config
type JSONConfig struct {
	BaseURL         string `json:"base_url"`
	ServerAddress   string `json:"server_address"`
	FileStoragePath string `json:"file_storage_path"`
	DatabaseDsn     string `json:"database_dsn"`
	EnableHTTPS     bool   `json:"enable_https"`
}

// Config base struct with default initialize
type MyConfig struct {
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:"127.0.0.1:8080"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" envDefault:"../../filedata"`
	DatabaseDSN     string `env:"DATABASE_DSN" envDefault:"postgresql://postgres:220098@127.0.0.1:5432/golangDB"`
	EnableHTTPS     string `env:"ENABLE_HTTPS" envDefault:""`
	Config          string `env:"CONFIG" envDefault:""`
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
		instance.initJSON()
	}

	return instance, nil
}

// Parse JSON
func (c *MyConfig) initJSON() {
	// Init from json evn config
	if c.Config == "" {
		return
	}

	// Get path directory of service
	pwd, _ := os.Getwd()
	path := pwd + "/" + c.Config
	fmt.Println(path)

	// Read path to config file
	byteValue, err := os.ReadFile(path)
	// If we os.Open returns an error then handle it
	if err != nil {
		// Nothing to do
		return
	}

	// initialize JSON config file
	var config JSONConfig

	// jsonFile's content into 'config' which we defined above
	err = json.Unmarshal(byteValue, &config)
	if err != nil {
		return
	}
	if c.BaseURL == "" {
		c.BaseURL = config.BaseURL
	}
	if c.ServerAddress == "" {
		c.ServerAddress = config.ServerAddress
	}
	if c.FileStoragePath == "" {
		if _, err := os.Stat(config.FileStoragePath); !errors.Is(err, os.ErrNotExist) {
			c.FileStoragePath = config.FileStoragePath
		}
	}
	if c.DatabaseDSN == "" {
		c.DatabaseDSN = config.DatabaseDsn
	}
	if c.EnableHTTPS == "" {
		c.EnableHTTPS = strconv.FormatBool(config.EnableHTTPS)
	}

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
	cFlag := flag.String("c", "", "")
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
	if *cFlag != "" {
		c.Config = *cFlag
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
	case Config:
		return c.Config, nil
	}

	return "", errs.ErrUnknownEnvOrFlag
}
