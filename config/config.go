package config

import (
	"crypto"
	"flag"
	"runtime"
	"time"

	"github.com/caarlos0/env/v6"
)

const (
	DefaultRunAddress              = ":8080"
	DefaultDatabaseURI             = "postgresql://user:secret@localhost:5432/gophermart?sslmode=disable"
	DefaultAccrualSystemAddress    = "localhost:8081"
	DefaultJWTSecretKey            = "Very Secret Key"
	DefaultJWTTTL                  = 24 * time.Hour
	DefaultPasswordCryptoHash      = crypto.SHA256
	DefaultAccrualsPollingInterval = 2 * time.Second
)

type Config struct {
	RunAddress              string `env:"RUN_ADDRESS"`
	DatabaseURI             string `env:"DATABASE_URI"`
	AccrualSystemAddress    string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	JWTSecret               string
	JWTTTL                  time.Duration
	PasswordCryptoHash      crypto.Hash
	AccrualsPollingInterval time.Duration
	AccrualsWorkersCount    int
}

type CLIExport func(*Config, *flag.FlagSet)

func Load(cli CLIExport) (*Config, error) {
	cfg := &Config{
		JWTSecret:               DefaultJWTSecretKey,
		JWTTTL:                  DefaultJWTTTL,
		PasswordCryptoHash:      DefaultPasswordCryptoHash,
		AccrualsPollingInterval: DefaultAccrualsPollingInterval,
		AccrualsWorkersCount:    runtime.NumCPU(),
	}

	if cli != nil {
		cli(cfg, flag.CommandLine)
		flag.Parse()
	}

	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
