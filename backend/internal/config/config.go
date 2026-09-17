package config

import (
	"fmt"
	"net"
	"time"

	"github.com/caarlos0/env/v10"
)

type (
	Config struct {
		HTTP struct {
			Port string `env:"HTTP_PORT" envDefault:"8080"`
		}

		PG struct {
			Host     string `env:"POSTGRES_HOST" envDefault:"localhost"`
			Port     string `env:"POSTGRES_PORT" envDefault:"5432"`
			DB       string `env:"POSTGRES_DB" envDefault:"db"`
			User     string `env:"POSTGRES_USER" envDefault:"db_user"`
			Password string `env:"POSTGRES_PASSWORD" envDefault:"12345"`
			MaxConns int32  `env:"POSTGRES_MAX_CONNS" envDefault:"8"`
			MinConns int32  `env:"POSTGRES_MIN_CONNS" envDefault:"1"`
		}

		Settings struct {
			AccessTokenTTL          time.Duration `env:"ACCESS_TOKEN_TTL" envDefault:"15m"`
			RefreshTokenTTL         time.Duration `env:"REFRESH_TOKEN_TTL" envDefault:"720h"`
			ShutdownTimeout         time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"5s"`
			JWTSecret               string        `env:"JWT_SECRET,required"`
			PGHealthCheckPeriod     time.Duration `env:"POSTGRES_HEALTH_CHECK_PERIOD" envDefault:"30s"`
			PGMaxConnLifetime       time.Duration `env:"POSTGRES_MAX_CONN_LIFETIME" envDefault:"1h"`
			PGMaxConnIdleTime       time.Duration `env:"POSTGRES_MAX_CONN_IDLE_TIME" envDefault:"5m"`
			ServerReadHeaderTimeout time.Duration `env:"SERVER_READ_HEADER_TIMEOUT" envDefault:"5s"`
			HTTPClientTimeout       time.Duration `env:"HTTP_CLIENT_TIMEOUT" envDefault:"5s"`
		}
	}
)

func New() (*Config, error) {
	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) ConstructPostgresURL() string {
	hostPort := net.JoinHostPort(c.PG.Host, c.PG.Port)
	return fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		c.PG.User,
		c.PG.Password,
		hostPort,
		c.PG.DB,
	)
}
