package config

import (
	"errors"
	"log"
	"net"

	"github.com/caarlos0/env"
	"github.com/spf13/pflag"
)

type Config struct {
	Address         string `env:"ADDRESS"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
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
	pflag.StringVarP(&config.Address, "address", "a", "localhost:8080", "set host:port")
	pflag.IntVarP(&config.StoreInterval, "interval", "i", 300, "set interval")
	pflag.StringVarP(&config.FileStoragePath, "path", "f", ".temp", "set path/filename")
	pflag.BoolVarP(&config.Restore, "restore", "r", false, "set true/false")
	pflag.Parse()

	err := env.Parse(&config)
	if err != nil {
		log.Printf("Error parsing environment variables:%v", err)
	}
	if err := validateAddress(config.Address); err != nil {
		log.Fatalf("Error parsing host %s", err)
	}
	return &Config{
		Address:         config.Address,
		StoreInterval:   config.StoreInterval,
		FileStoragePath: config.FileStoragePath,
		Restore:         config.Restore,
	}
}
