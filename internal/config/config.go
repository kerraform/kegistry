package config

import (
	"context"
	"fmt"

	"github.com/sethvargo/go-envconfig"
	"go.uber.org/zap/zapcore"
)

type Backend struct {
	GCS  *BackendGCS `env:",prefix=GCS_"`
	S3   *BackendS3  `env:",prefix=S3_"`
	Type string      `env:"TYPE,required"`
}

func (b *Backend) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("type", b.Type)
	return nil
}

type BackendGCS struct {
	Bucket string `env:"BUCKET"`
}

type BackendS3 struct {
	AccessKey    string `env:"ACCESS_KEY"`
	Bucket       string `env:"BUCKET"`
	Endpoint     string `env:"ENDPOINT"`
	SecretKey    string `env:"SECRET_KEY"`
	UsePathStyle bool   `env:"USE_PATH_STYLE, default=false"`
}

type Config struct {
	Backend *Backend `env:",prefix=BACKEND_"`
	Log     *Log     `env:",prefix=LOG_"`
	Port    int      `env:"PORT,default=5000"`
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
