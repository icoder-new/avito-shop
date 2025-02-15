package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/icoder-new/avito-shop/internal/models"
	"github.com/icoder-new/avito-shop/internal/storage"
	"github.com/jackc/pgx/v5/pgxpool"
)

type coinRepo struct {
	ctx  context.Context
	pool *pgxpool.Pool
}

func newCoinRepo(ctx context.Context, pool *pgxpool.Pool) *coinRepo {
	return &coinRepo{
		ctx:  ctx,
		pool: pool,
	}
}

func (s *store) Coin() storage.ICoin {
	return s.coin
}

func (c *coinRepo) TransferCoins(fromUserID, toUserID int64, amount int64) error {
	tx, err := c.pool.Begin(c.ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(c.ctx)

	_, err = tx.Exec(c.ctx, `UPDATE users SET coins = coins - $1 WHERE id = $2 AND coins >= $1`, amount, fromUserID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(c.ctx, `UPDATE users SET coins = coins + $1 WHERE id = $2`, amount, toUserID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(c.ctx, `
		INSERT INTO transactions (from_user_id, to_user_id, amount, type, created_at)
		VALUES ($1, $2, $3, $4, NOW())
	`, fromUserID, toUserID, amount, models.TransactionTypeTransfer)
	if err != nil {
		return err
	}

	return tx.Commit(c.ctx)
}

func (c *coinRepo) GetUserTransactions(userID int64) ([]models.Transaction, error) {
	rows, err := c.pool.Query(c.ctx, `
        SELECT id, from_user_id, to_user_id, amount, type, merch_id, created_at 
        FROM transactions 
        WHERE from_user_id = $1 OR to_user_id = $1 
        ORDER BY created_at DESC
    `, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var t models.Transaction
		var nullToUserID sql.NullInt64 // Промежуточная переменная для сканирования NULL

		err := rows.Scan(
			&t.ID,
			&t.FromUserID,
			&nullToUserID, // Сканируем в NullInt64
			&t.Amount,
			&t.Type,
			&t.MerchID,
			&t.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan transaction: %w", err)
		}

		// Если значение валидно, присваиваем его, иначе оставляем 0
		if nullToUserID.Valid {
			t.ToUserID = nullToUserID.Int64
		}

		transactions = append(transactions, t)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate transactions: %w", err)
	}

	return transactions, nil
}
