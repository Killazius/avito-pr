package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/Killazius/avito-pr/internal/app"
	"github.com/Killazius/avito-pr/internal/config"
	"github.com/Killazius/avito-pr/internal/logger"
	"go.uber.org/zap"
)

func main() {
	cfg := config.MustLoad()
	log := logger.MustLoad(cfg.Logger.Path)

	application := app.New(log, cfg)
	go application.Run()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	sig := <-stop
	log.Info("revieving shutdown signal", zap.String("signal", sig.String()))
	application.Stop()
}
