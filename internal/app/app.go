package app

import (
	"github.com/Killazius/avito-pr/internal/config"
	"github.com/Killazius/avito-pr/internal/repository"
	"github.com/Killazius/avito-pr/internal/server"
	"github.com/Killazius/avito-pr/internal/service"
	"go.uber.org/zap"
)

type App struct {
	log  *zap.Logger
	api  *server.Server
	cfg  *config.Config
	repo *repository.Repository
}

func New(log *zap.Logger, cfg *config.Config) *App {
	pool, err := repository.CreatePool(cfg.Postgres)
	if err != nil {
		log.Panic("failed to create db pool", zap.Error(err))
	}
	repo := repository.New(pool)
	serv := service.New()
	_ = serv

	api := server.New(log, cfg.Server)

	return &App{
		log:  log,
		api:  api,
		cfg:  cfg,
		repo: repo,
	}
}

func (a *App) Run() {
	defer func() {
		if r := recover(); r != nil {
			a.log.Error("application panicked and recovered", zap.Error(r.(error)))
			a.Stop()
		}
	}()
	a.api.MustRun()
}

func (a *App) Stop() {
	a.log.Info("closing HTTP server")
	if err := a.api.Close(); err != nil {
		a.log.Error("failed to close HTTP server", zap.Error(err))
	}
	a.log.Info("closing database connection pool")
	a.repo.Close()
	a.log.Info("application stopped")
}
