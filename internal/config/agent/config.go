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
)

type Config struct {
	ServerAddress  string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	Key            string `env:"KEY"`
	RateLimit      int    `env:"RATE_LIMIT"`
	WasKeySet      bool
	CryptoKey      string `env:"CRYPTO_KEY"`
	ConfigFile     string
	AgentHostIP    string `env:"AGENT_HOST_IP"`
	GRPCAddress    string `json:"grpc_address" env:"GRPC_ADDRESS"`
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
	config.WasKeySet = false
	pflag.StringVarP(&config.ConfigFile, "config", "c", "", "path to config file (json)")
	pflag.StringVarP(&config.ServerAddress, "address", "a", "localhost:8080", "set host:port for server")
	pflag.IntVarP(&config.ReportInterval, "report_interval", "r", 10, "frequency send")
	pflag.IntVarP(&config.PollInterval, "poll_interval", "p", 2, "refresh metric")
	pflag.StringVarP(&config.Key, "key", "k", "", "key secret")
	pflag.IntVarP(&config.RateLimit, "limit", "l", 1, "send worker rate limit")
	pflag.StringVarP(&config.CryptoKey, "crypto_key", "", "./keys/public.pem", "set public key")
	pflag.StringVarP(&config.AgentHostIP, "agent_host_ip", "i", "", "set agent ip")
	pflag.StringVarP(&config.GRPCAddress, "grpc_address", "g", "", "set gprc address for server")
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
			config.ServerAddress = f.Value.String()
		case "report_interval":
			config.ReportInterval, _ = strconv.Atoi(f.Value.String())
		case "poll_interval":
			config.PollInterval, _ = strconv.Atoi(f.Value.String())
		case "crypto_key":
			config.CryptoKey = f.Value.String()
		case "key":
			config.WasKeySet = true
		case "agent_host_ip":
			config.AgentHostIP = f.Value.String()
		case "grpc_address":
			config.GRPCAddress = f.Value.String()
		}
	})

	err := env.Parse(&config)
	if err != nil {
		log.Printf("Error parsing environment variables:%v", err)
	}

	if err := validateAddress(config.ServerAddress); err != nil {
		log.Fatalf("Error parsing host %s", err)
	}
	pflag.Visit(func(f *pflag.Flag) {

	})

	return &config
}
