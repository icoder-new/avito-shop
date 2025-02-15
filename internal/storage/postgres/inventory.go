package postgres

import (
	"context"
	"errors"

	"github.com/icoder-new/avito-shop/internal/models"
	"github.com/icoder-new/avito-shop/internal/storage"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type inventoryRepo struct {
	ctx  context.Context
	pool *pgxpool.Pool
}

func newInventoryRepo(ctx context.Context, pool *pgxpool.Pool) *inventoryRepo {
	return &inventoryRepo{
		ctx:  ctx,
		pool: pool,
	}
}

func (s *store) Inventory() storage.IInventory {
	return s.inventory
}

func (i *inventoryRepo) BuyItem(userID, merchID int64) error {
	tx, err := i.pool.Begin(i.ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(i.ctx)

	var price int64
	err = tx.QueryRow(i.ctx, "SELECT price FROM merch WHERE id = $1", merchID).Scan(&price)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("merch not found")
		}
		return err
	}

	var userCoins int64
	err = tx.QueryRow(i.ctx, "SELECT coins FROM users WHERE id = $1", userID).Scan(&userCoins)
	if err != nil {
		return err
	}
	if userCoins < price {
		return errors.New("not enough coins")
	}

	_, err = tx.Exec(i.ctx, "UPDATE users SET coins = coins - $1 WHERE id = $2", price, userID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(i.ctx, `
		INSERT INTO user_inventory (user_id, merch_id, quantity)
		VALUES ($1, $2, 1)
		ON CONFLICT (user_id, merch_id) DO UPDATE SET quantity = user_inventory.quantity + 1
	`, userID, merchID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(i.ctx, `
		INSERT INTO transactions (from_user_id, to_user_id, amount, type, merch_id)
		VALUES ($1, NULL, $2, 'purchase', $3)
	`, userID, price, merchID)
	if err != nil {
		return err
	}

	return tx.Commit(i.ctx)
}

func (i *inventoryRepo) GetUserInventory(userID int64) ([]models.UserInventory, error) {
	rows, err := i.pool.Query(i.ctx, `
		SELECT id, user_id, merch_id, quantity, created_at 
		FROM user_inventory 
		WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var inventory []models.UserInventory
	for rows.Next() {
		var item models.UserInventory
		if err := rows.Scan(&item.ID, &item.UserID, &item.MerchID, &item.Quantity, &item.CreatedAt); err != nil {
			return nil, err
		}
		inventory = append(inventory, item)
	}

	return inventory, nil
}
