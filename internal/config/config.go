package config

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Server   HTTPConfig     `yaml:"server"`
	Logger   LoggerConfig   `yaml:"logger"`
	Postgres PostgresConfig `yaml:"postgres"`
}

type HTTPConfig struct {
	Host    string `yaml:"host"`
	Port    string `yaml:"port"`
	Timeout struct {
		Server time.Duration `yaml:"server"`
		Write  time.Duration `yaml:"write"`
		Read   time.Duration `yaml:"read"`
		Idle   time.Duration `yaml:"idle"`
	} `yaml:"timeout"`
}

type LoggerConfig struct {
	Path string `yaml:"path" env-default:"config/logger.json"`
}

type PostgresConfig struct {
	Host     string `env:"POSTGRES_HOST" env-default:"localhost"`
	Port     string `env:"POSTGRES_PORT" env-default:"5432"`
	Username string `env:"POSTGRES_USER" env-default:"postgres"`
	Password string `env:"POSTGRES_PASSWORD" env-default:"postgres"`
	Database string `env:"POSTGRES_DB" env-default:"postgres"`
	SSLMode  string `env:"POSTGRES_SSL_MODE" env-default:"disable"`

	MaxOpenConns    int32         `yaml:"max_open_conns" env-default:"25"`
	MaxIdleConns    int32         `yaml:"max_idle_conns" env-default:"5"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" env-default:"1h"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time" env-default:"30m"`
	Timeout         time.Duration `yaml:"timeout" env-default:"5s"`
	MigrationsPath  string        `yaml:"migrations_path" env-default:"./migrations"`
}

const (
	defaultConfigPath = "config/config.yaml"
	Path              = "CONFIG_PATH"
)

func MustLoad() *Config {
	cfg, err := load()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}
	return cfg
}

func load() (*Config, error) {
	if _, err := os.Stat(".env"); err == nil {
		if loadErr := godotenv.Load(); loadErr != nil {
			return nil, fmt.Errorf("failed to load .env: %w", loadErr)
		}
	}

	configPath := getConfigPath()
	if configPath != "" {
		fileInfo, err := os.Stat(configPath)

		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file %s does not exist", configPath)
		}

		if err != nil {
			return nil, fmt.Errorf("failed to access config file: %w", err)
		}

		if fileInfo.IsDir() {
			return nil, fmt.Errorf("config path is a directory, not a file: %s", configPath)
		}
	}

	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		return nil, fmt.Errorf("failed to read env vars: %w", err)
	}

	return &cfg, nil
}

func getConfigPath() string {
	configPath := os.Getenv(Path)
	if configPath == "" {
		configPath = defaultConfigPath
	}

	return configPath
}

func (h *HTTPConfig) GetAddr() string {
	return net.JoinHostPort(h.Host, h.Port)
}
