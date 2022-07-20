package config

import (
	"context"
	"fmt"

	"github.com/sethvargo/go-envconfig"
	"go.uber.org/zap/zapcore"
)

type Backend struct {
	Type string     `env:"TYPE,required"`
	S3   *BackendS3 `env:",prefix=S3_"`
}

func (b *Backend) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("type", b.Type)
	return nil
}

type BackendS3 struct {
	AccessKey string `env:"ACCESS_KEY"`
	Bucket    string `env:"BUCKET"`
	SecretKey string `env:"SECRET_KEY"`
}

type Config struct {
	BaseURL string   `env:"BASE_URL,required"`
	Backend *Backend `env:",prefix=BACKEND_"`
	Port    int      `env:"PORT,default=5000"`
	Log     *Log     `env:",prefix=LOG_"`
}

func (cfg *Config) Address() string {
	return fmt.Sprintf(":%d", cfg.Port)
}

type Log struct {
	Format string `env:"FORMAT,default=json"`
	Level  string `env:"LEVEL,default=info"`
}

func Load(ctx context.Context) (*Config, error) {
	var cfg Config
	if err := envconfig.Process(ctx, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
