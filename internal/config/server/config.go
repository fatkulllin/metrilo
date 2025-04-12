package config

import (
	"errors"
	"log"
	"net"

	"github.com/caarlos0/env"
	"github.com/spf13/pflag"
)

type Config struct {
	Address string `env:"ADDRESS" envDefault:"localhost:8080"`
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
	err := env.Parse(&config)
	if err != nil {
		log.Printf("Error parsing environment variables:%v", err)
	}
	if config.Address == "localhost:8080" {
		pflag.StringVarP(&config.Address, "address", "a", "localhost:8080", "set host:port")
		pflag.Parse()
	}

	if err := validateAddress(config.Address); err != nil {
		log.Fatalf("Error parsing host %s", err)
	}
	return &Config{
		Address: config.Address,
	}
}
