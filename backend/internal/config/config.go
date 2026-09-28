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
			MediaMaxBytes           int64         `env:"MEDIA_MAX_BYTES" envDefault:"10485760"`
		}

		AI struct {
			Provider                 string        `env:"AI_PROVIDER" envDefault:"ollama"`
			BaseURL                  string        `env:"AI_BASE_URL" envDefault:"http://10.8.1.6:11434/api/chat"`
			Timeout                  time.Duration `env:"AI_TIMEOUT" envDefault:"180s"`
			Think                    bool          `env:"AI_THINK" envDefault:"false"`
			Stream                   bool          `env:"AI_STREAM" envDefault:"false"`
			Format                   string        `env:"AI_FORMAT" envDefault:"json"`
			NumPredict               int           `env:"AI_NUM_PREDICT" envDefault:"32768"`
			NumCtx                   int           `env:"AI_NUM_CTX" envDefault:"65536"`
			PromptDir                string        `env:"AI_PROMPT_DIR" envDefault:"prompts"`
			MaxStructuredOutputRetry int           `env:"AI_MAX_STRUCTURED_OUTPUT_RETRY" envDefault:"1"`

			ScenarioArchitectTemperature     float64 `env:"AI_SCENARIO_ARCHITECT_TEMPERATURE" envDefault:"0.4"`
			BlockDesignerTemperature         float64 `env:"AI_BLOCK_DESIGNER_TEMPERATURE" envDefault:"0.4"`
			DialogueBuilderTemperature       float64 `env:"AI_DIALOGUE_BUILDER_TEMPERATURE" envDefault:"0.5"`
			ActionClassifierTemperature      float64 `env:"AI_ACTION_CLASSIFIER_TEMPERATURE" envDefault:"0.1"`
			VirtualPartnerTemperature        float64 `env:"AI_VIRTUAL_PARTNER_TEMPERATURE" envDefault:"0.5"`
			VirtualPartnerRuntimeTemperature float64 `env:"AI_VIRTUAL_PARTNER_RUNTIME_TEMPERATURE" envDefault:"0.4"`
			TrainingStateTemperature         float64 `env:"AI_TRAINING_STATE_TEMPERATURE" envDefault:"0.2"`
			TrainingAssessmentTemperature    float64 `env:"AI_TRAINING_ASSESSMENT_TEMPERATURE" envDefault:"0.2"`
			BlockDesignerThink               bool    `env:"AI_BLOCK_DESIGNER_THINK" envDefault:"false"`
			DialogueBuilderThink             bool    `env:"AI_DIALOGUE_BUILDER_THINK" envDefault:"false"`
			VirtualPartnerRuntimeThink       bool    `env:"AI_VIRTUAL_PARTNER_RUNTIME_THINK" envDefault:"false"`
			TrainingStateThink               bool    `env:"AI_TRAINING_STATE_THINK" envDefault:"true"`
			TrainingAssessmentThink          bool    `env:"AI_TRAINING_ASSESSMENT_THINK" envDefault:"true"`
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
