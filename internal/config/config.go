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
	Server HTTPConfig   `yaml:"server"`
	Logger LoggerConfig `yaml:"logger"`
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
			return nil, fmt.Errorf("failed to load .env: %w", err)
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
