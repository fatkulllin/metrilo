package config

import (
	"encoding/json"
	"errors"
	"log"
	"net"
	"os"
	"strconv"

	"github.com/caarlos0/env"
	"github.com/spf13/pflag"

	"github.com/fatkulllin/metrilo/internal/logger"
)

type Config struct {
	Address         string `json:"address" env:"ADDRESS"`
	StoreInterval   int    `json:"store_interval" env:"STORE_INTERVAL"`
	WasIntervalSet  bool
	FileStoragePath string `json:"store_file" env:"FILE_STORAGE_PATH"`
	WasPathSet      bool
	Restore         bool   `json:"restore" env:"RESTORE"`
	Database        string `json:"database_dsn" env:"DATABASE_DSN"`
	WasDatabaseSet  bool
	Key             string `json:"key" env:"KEY"`
	WasKeySet       bool
	CryptoKey       string `json:"crypto_key" env:"CRYPTO_KEY"`
	ConfigFile      string
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
	config.WasKeySet = false
	pflag.StringVarP(&config.ConfigFile, "config", "c", "", "path to config file (json)")
	pflag.StringVarP(&config.Address, "address", "a", "localhost:8080", "set host:port")
	pflag.IntVarP(&config.StoreInterval, "store_interval", "i", 300, "set interval")
	pflag.StringVarP(&config.FileStoragePath, "store_file", "f", ".temp", "set path/filename")
	pflag.BoolVarP(&config.Restore, "restore", "r", false, "set true/false")
	pflag.StringVarP(&config.Database, "database_dsn", "d", "", "set database dsn")
	pflag.StringVarP(&config.Key, "key", "k", "", "key secret")
	pflag.StringVarP(&config.CryptoKey, "crypto_key", "", "./keys/private.pem", "set private key")
	pflag.Parse()

	if config.ConfigFile == "" {
		config.ConfigFile = os.Getenv("CONFIG")
	}

	if config.ConfigFile != "" {
		file, err := os.Open(config.ConfigFile)
		if err != nil {
			log.Fatalf("failed to open config file: %v", err)
		}
		defer file.Close()

		dec := json.NewDecoder(file)
		if err := dec.Decode(&config); err != nil {
			log.Fatalf("failed to decode config file: %v", err)
		}
	}

	pflag.Visit(func(f *pflag.Flag) {
		switch f.Name {
		case "address":
			config.Address = f.Value.String()
		case "store_interval":
			config.StoreInterval, _ = strconv.Atoi(f.Value.String())
			config.WasIntervalSet = true
		case "store_file":
			config.FileStoragePath = f.Value.String()
			config.WasPathSet = true
		case "restore":
			config.Restore, _ = strconv.ParseBool(f.Value.String())
		case "database_dsn":
			config.Database = f.Value.String()
			config.WasDatabaseSet = true
		case "key":
			config.Key = f.Value.String()
			config.WasKeySet = true
		case "crypto_key":
			config.CryptoKey = f.Value.String()
		}
	})

	err := env.Parse(&config)
	if err != nil {
		log.Printf("Error parsing environment variables:%v", err)
	}
	if err := validateAddress(config.Address); err != nil {
		log.Fatalf("Error parsing host %s", err)
	}

	if config.WasPathSet && config.WasIntervalSet {
		logger.Log.Info("Save metrics to file")
	}
	if config.WasDatabaseSet || config.Database != "" {
		logger.Log.Info("Save metrics to db")
		config.WasDatabaseSet = true
	}
	return &config
}
