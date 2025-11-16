package logger

import (
	"encoding/json"
	"fmt"
	"os"

	"go.uber.org/zap"
)

func MustLoad(path string) *zap.Logger {
	log, err := load(path)
	if err != nil {
		panic(fmt.Sprintf("logger load error: %s", err))
	}

	return log
}

func load(path string) (*zap.Logger, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log, logErr := zap.NewProduction(zap.AddStacktrace(zap.ErrorLevel), zap.AddCaller())
		if logErr != nil {
			return nil, fmt.Errorf("failed to create default logger: %w", logErr)
		}

		log.Warn("using default logger because config file not found",
			zap.String("path", path))

		return log, nil
	}

	configData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg zap.Config
	if err = json.Unmarshal(configData, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	logger, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build logger from config %q: %w", path, err)
	}

	level := getLevel()
	logger = logger.WithOptions(zap.IncreaseLevel(level))
	zap.ReplaceGlobals(logger)

	return logger, nil
}

func getLevel() zap.AtomicLevel {
	mode := os.Getenv("GIN_MODE")
	if mode == "release" {
		return zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	return zap.NewAtomicLevelAt(zap.DebugLevel)
}
