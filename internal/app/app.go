package app

import (
	"github.com/Killazius/avito-pr/internal/config"
	"github.com/Killazius/avito-pr/internal/repository"
	"github.com/Killazius/avito-pr/internal/server"
	"github.com/Killazius/avito-pr/internal/server/handler/pr"
	"github.com/Killazius/avito-pr/internal/server/handler/team"
	"github.com/Killazius/avito-pr/internal/server/handler/user"
	prservice "github.com/Killazius/avito-pr/internal/service/pr"
	teamservice "github.com/Killazius/avito-pr/internal/service/team"
	userservice "github.com/Killazius/avito-pr/internal/service/user"

	"go.uber.org/zap"
)

type Repository interface {
	Close()
}
type App struct {
	log  *zap.Logger
	api  *server.Server
	cfg  *config.Config
	repo Repository
}

func New(log *zap.Logger, cfg *config.Config) *App {
	repo, trManager, err := repository.New(cfg.Postgres)
	if err != nil {
		log.Fatal("failed to create repository", zap.Error(err))
	}
	teamService := teamservice.NewService(repo, repo, trManager)
	teamHandler := team.NewHandler(teamService)

	userService := userservice.NewService(repo, trManager)
	userHandler := user.NewHandler(userService)

	prService := prservice.NewService(repo, repo, repo, trManager)
	prHandler := pr.NewHandler(prService)
	api := server.New(log, cfg.Server, teamHandler, userHandler, prHandler)

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
