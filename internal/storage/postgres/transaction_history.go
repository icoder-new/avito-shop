package postgres

import (
	"context"
	"github.com/icoder-new/avito-shop/internal/storage"

	"github.com/icoder-new/avito-shop/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type transactionHistoryRepo struct {
	ctx  context.Context
	pool *pgxpool.Pool
}

func newTransactionHistoryRepo(ctx context.Context, pool *pgxpool.Pool) *transactionHistoryRepo {
	return &transactionHistoryRepo{
		ctx:  ctx,
		pool: pool,
	}
}

func (s *store) TransactionHistory() storage.ITransactionHistory {
	return s.transactionHistoryRepo
}

func (t *transactionHistoryRepo) GetTransactions(userID int64) ([]models.Transaction, error) {
	rows, err := t.pool.Query(t.ctx, `
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
		var tx models.Transaction
		if err := rows.Scan(&tx.ID, &tx.FromUserID, &tx.ToUserID, &tx.Amount, &tx.Type, &tx.MerchID, &tx.CreatedAt); err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}

	return transactions, nil
}
