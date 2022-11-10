package config

import (
	"context"
	"fmt"

	"github.com/sethvargo/go-envconfig"
	"go.uber.org/zap/zapcore"
)

type Backend struct {
	S3       *BackendS3 `env:",prefix=S3_"`
	Type     string     `env:"TYPE,required"`
	RootPath string     `env:"ROOT_PATH,default=/tmp"`
}

func (b *Backend) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("type", b.Type)
	return nil
}

type BackendS3 struct {
	AccessKey    string `env:"ACCESS_KEY"`
	Bucket       string `env:"BUCKET"`
	Endpoint     string `env:"ENDPOINT"`
	SecretKey    string `env:"SECRET_KEY"`
	UsePathStyle bool   `env:"USE_PATH_STYLE, default=false"`
}

type Config struct {
	Backend        *Backend `env:",prefix=BACKEND_"`
	EnableModule   bool     `env:"ENABLE_MODULE_REGISTRY,default=false"`
	EnableProvider bool     `env:"ENABLE_PROVIDER_REGISTRY,default=false"`
	Log            *Log     `env:",prefix=LOG_"`
	Port           int      `env:"PORT,default=5000"`
	Trace          *Trace   `env:",prefix=TRACE_"`
}

type Trace struct {
	Enable bool   `env:"ENABLE,default=false"`
	Type   string `env:"TYPE,default=console"`
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
