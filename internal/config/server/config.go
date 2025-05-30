package config

import (
	"errors"
	"log"
	"net"

	"github.com/caarlos0/env"
	"github.com/fatkulllin/metrilo/internal/logger"
	"github.com/spf13/pflag"
)

type Config struct {
	Address         string `env:"ADDRESS"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	WasIntervalSet  bool
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	WasPathSet      bool
	Restore         bool   `env:"RESTORE"`
	Database        string `env:"DATABASE_DSN"`
	WasDatabaseSet  bool
}

func validateAddress(s string) error {
	_, _, err := net.SplitHostPort(s)
	if err != nil {
		return errors.New("need address in the form host:port")
	}
	return nil
}

func LoadConfig() *Config {
	var config Config
	config.WasPathSet = false
	config.WasDatabaseSet = false
	config.WasIntervalSet = false
	pflag.StringVarP(&config.Address, "address", "a", "localhost:8080", "set host:port")
	pflag.IntVarP(&config.StoreInterval, "interval", "i", 300, "set interval")
	pflag.StringVarP(&config.FileStoragePath, "path", "f", ".temp", "set path/filename")
	pflag.BoolVarP(&config.Restore, "restore", "r", false, "set true/false")
	pflag.StringVarP(&config.Database, "database", "d", "", "set database dsn")
	pflag.Parse()

	err := env.Parse(&config)
	if err != nil {
		log.Printf("Error parsing environment variables:%v", err)
	}
	if err := validateAddress(config.Address); err != nil {
		log.Fatalf("Error parsing host %s", err)
	}
	pflag.Visit(func(f *pflag.Flag) {
		if f.Name == "path" {
			config.WasPathSet = true
		}
		if f.Name == "interval" {
			config.WasIntervalSet = true
		}
		if f.Name == "database" {
			config.WasDatabaseSet = true
			logger.Log.Info("Save metrics to db")
		}
	})

	if config.WasPathSet && config.WasIntervalSet {
		logger.Log.Info("Save metrics to file")
	}

	return &Config{
		Address:         config.Address,
		StoreInterval:   config.StoreInterval,
		WasIntervalSet:  config.WasIntervalSet,
		FileStoragePath: config.FileStoragePath,
		WasPathSet:      config.WasPathSet,
		Restore:         config.Restore,
		Database:        config.Database,
		WasDatabaseSet:  config.WasDatabaseSet,
	}
}
