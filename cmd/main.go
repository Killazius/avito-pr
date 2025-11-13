package main

import (
	"github.com/Killazius/avito-pr/internal/config"
	"github.com/Killazius/avito-pr/internal/logger"
	"github.com/Killazius/avito-pr/internal/repository"
	"go.uber.org/zap"
)

func main() {
	cfg := config.MustLoad()
	log := logger.MustLoad(cfg.Logger.Path)
	log.Info("configuration loaded")
	pool, err := repository.CreatePool(cfg.Postgres)
	if err != nil {
		log.Fatal("failed to create db pool", zap.Error(err))
	}
	repo := repository.New(pool)
	_ = repo
	log.Info("database connection pool created")
	//application := app.New(log, cfg)
	//go application.Run()
	//stop := make(chan os.Signal, 1)
	//signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	//<-stop
	//application.Stop()
}
