package config

import (
	"fmt"

	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Logging  LoggingConfig
}

type ServerConfig struct {
	Port string `env:"SERVER_PORT" envDefault:"3030"`
}

type DatabaseConfig struct {
	Driver   string `env:"DB_DRIVER" envDefault:"postgres"`
	Host     string `env:"DB_HOST"`
	Port     string `env:"DB_PORT"`
	User     string `env:"DB_USER"`
	Password string `env:"DB_PASSWORD"`
	Name     string `env:"DB_NAME"`
}

type LoggingConfig struct {
	Level string `env:"LOG_LEVEL" envDefault:"error"`
}

func Load() (*Config, error) {
	_ = godotenv.Load()
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	if err := cfg.Database.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (d DatabaseConfig) validate() error {
	if d.Driver == "sqlite" {
		if d.Name == "" {
			return fmt.Errorf("DB_NAME is required for sqlite driver (database file path)")
		}
		return nil
	}

	missing := make([]string, 0, 5)
	if d.Host == "" {
		missing = append(missing, "DB_HOST")
	}
	if d.Port == "" {
		missing = append(missing, "DB_PORT")
	}
	if d.User == "" {
		missing = append(missing, "DB_USER")
	}
	if d.Password == "" {
		missing = append(missing, "DB_PASSWORD")
	}
	if d.Name == "" {
		missing = append(missing, "DB_NAME")
	}

	if len(missing) > 0 {
		return fmt.Errorf("required environment variables missing for driver %q: %v", d.Driver, missing)
	}

	return nil
}
