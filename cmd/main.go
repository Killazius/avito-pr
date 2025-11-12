package main

import (
	"github.com/Killazius/avito-pr/internal/config"
	"github.com/Killazius/avito-pr/internal/logger"
)

func main() {
	cfg := config.MustLoad()
	log := logger.MustLoad(cfg.Logger.Path)
	log.Info("configuration loaded")

	//application := app.New(log, cfg)
	//go application.Run()
	//stop := make(chan os.Signal, 1)
	//signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	//<-stop
	//application.Stop()
}
