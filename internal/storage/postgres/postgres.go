package postgres

import (
	"context"
	"fmt"
	"github.com/icoder-new/avito-shop/internal/config"
	"go.uber.org/zap"

	"github.com/icoder-new/avito-shop/internal/storage"
	"github.com/icoder-new/avito-shop/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

type store struct {
	pool *pgxpool.Pool
	log  *logger.Logger

	user                   *userRepo
	coin                   *coinRepo
	inventory              *inventoryRepo
	transactionHistoryRepo *transactionHistoryRepo
}

func newStore(ctx context.Context, pool *pgxpool.Pool, log *logger.Logger) *store {
	return &store{
		pool: pool,
		log:  log,

		user:                   newUserRepo(ctx, pool),
		coin:                   newCoinRepo(ctx, pool),
		inventory:              newInventoryRepo(ctx, pool),
		transactionHistoryRepo: newTransactionHistoryRepo(ctx, pool),
	}
}

func NewStorage(ctx context.Context, log *logger.Logger, dsn string, cfg config.DBSettings) (storage.IStorage, error) {
	pgxCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Error("failed to parse DSN: ", zap.Error(err))
		return nil, fmt.Errorf("postgres.NewStorage: %w", err)
	}

	pgxCfg.MaxConns = int32(cfg.MaxOpenConns)
	pgxCfg.MinConns = int32(cfg.MaxIdleConns)
	pgxCfg.MaxConnLifetime = cfg.ConnMaxLifetime

	pool, err := pgxpool.NewWithConfig(ctx, pgxCfg)
	if err != nil {
		log.Error("failed to create DB pool: ", zap.Error(err))
		return nil, fmt.Errorf("postgres.NewStorage: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		log.Error("failed to ping DB: ", zap.Error(err))
		return nil, fmt.Errorf("postgres.NewStorage: %w", err)
	}

	return newStore(ctx, pool, log), nil
}

func (s *store) CloseDB() {
	s.pool.Close()
}
