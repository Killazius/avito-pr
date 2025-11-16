package repository

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Killazius/avito-pr/internal/config"
	"github.com/Killazius/avito-pr/internal/repository/postgres"
	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/jackc/pgx/v5/pgxpool"
)

func New(cfg config.PostgresConfig) (*postgres.Repository, *manager.Manager, error) {
	pool, err := CreatePool(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create pool: %w", err)
	}

	trManager := manager.Must(trmpgx.NewDefaultFactory(pool))
	repo := postgres.New(pool, trmpgx.DefaultCtxGetter)

	return repo, trManager, nil
}

func CreatePool(cfg config.PostgresConfig) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig("")

	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}
	port, err := strconv.ParseUint(cfg.Port, 10, 16)

	if err != nil {
		return nil, fmt.Errorf("invalid port %s: %w", cfg.Port, err)
	}
	poolConfig.ConnConfig.Host = cfg.Host
	poolConfig.ConnConfig.Port = uint16(port)
	poolConfig.ConnConfig.Database = cfg.Database
	poolConfig.ConnConfig.User = cfg.Username
	poolConfig.ConnConfig.Password = cfg.Password

	if cfg.SSLMode == "disable" {
		poolConfig.ConnConfig.TLSConfig = nil
	}

	poolConfig.MaxConns = cfg.MaxOpenConns
	poolConfig.MinConns = cfg.MaxIdleConns
	poolConfig.MaxConnLifetime = cfg.ConnMaxLifetime
	poolConfig.MaxConnIdleTime = cfg.ConnMaxIdleTime
	poolConfig.ConnConfig.ConnectTimeout = cfg.Timeout

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)

	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	if err = pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}
