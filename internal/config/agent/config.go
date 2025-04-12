package config

import (
	"errors"
	"log"
	"net"

	"github.com/caarlos0/env"
	"github.com/spf13/pflag"
)

type Config struct {
	ServerAddress  string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
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

	pflag.StringVarP(&config.ServerAddress, "address", "a", "localhost:8080", "set host:port for server")
	pflag.IntVarP(&config.ReportInterval, "reportInterval", "r", 10, "frequency send")
	pflag.IntVarP(&config.PollInterval, "pollInterval", "p", 2, "refresh metric")
	pflag.Parse()

	err := env.Parse(&config)
	if err != nil {
		log.Printf("Error parsing environment variables:%v", err)
	}

	if err := validateAddress(config.ServerAddress); err != nil {
		log.Fatalf("Error parsing host %s", err)
	}

	return &Config{
		ServerAddress:  config.ServerAddress,
		ReportInterval: config.ReportInterval,
		PollInterval:   config.PollInterval,
	}
}
