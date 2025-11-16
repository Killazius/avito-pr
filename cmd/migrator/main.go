package main

import (
	"flag"
	"os"

	"github.com/Killazius/avito-pr/internal/config"
	"github.com/Killazius/avito-pr/internal/logger"
	"github.com/Killazius/avito-pr/internal/repository"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose"
	"go.uber.org/zap"
)

var command = flag.String("command", "up", "goose command (up, down, status, etc.)")

func main() {
	cfg := config.MustLoad()
	log := logger.MustLoad(cfg.Logger.Path)

	if _, err := os.Stat(cfg.Postgres.MigrationsPath); os.IsNotExist(err) {
		log.Fatal("migrations directory does not exist", zap.String("path", cfg.Postgres.MigrationsPath))
	}
	if *command == "" {
		if len(flag.Args()) > 0 {
			*command = flag.Args()[0]
		} else {
			log.Info("no goose command provided. Usage: -command <command> or provide command as argument")
		}
	}

	pool, err := repository.CreatePool(cfg.Postgres)
	if err != nil {
		log.Fatal("error creating postgres pool", zap.Error(err))
	}
	defer pool.Close()
	sqlDB := stdlib.OpenDBFromPool(pool)
	defer func() {
		if errClose := sqlDB.Close(); errClose != nil {
			log.Error("failed to close sqlDB", zap.Error(errClose))
		}
	}()
	if err = goose.Run(*command, sqlDB, cfg.Postgres.MigrationsPath); err != nil {
		log.Error("failed to run goose command", zap.Error(err), zap.String("command", *command))
		return
	}

	log.Info("goose command executed successfully", zap.String("command", *command))
}
