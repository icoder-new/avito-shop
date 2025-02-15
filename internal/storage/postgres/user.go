package postgres

import (
	"context"
	"github.com/icoder-new/avito-shop/internal/models"
	"github.com/icoder-new/avito-shop/internal/storage"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userRepo struct {
	ctx  context.Context
	pool *pgxpool.Pool
}

func newUserRepo(ctx context.Context, pool *pgxpool.Pool) *userRepo {
	return &userRepo{
		ctx:  ctx,
		pool: pool,
	}
}

func (s *store) User() storage.IUser {
	return s.user
}

func (u *userRepo) CreateUser(username, passwordHash string) (models.User, error) {
	query := `
		INSERT INTO users (username, password_hash, coins, created_at, updated_at)
		VALUES ($1, $2, 0, NOW(), NOW())
		RETURNING id, username, password_hash, coins, created_at, updated_at;
	`

	var user models.User
	err := u.pool.QueryRow(u.ctx, query, username, passwordHash).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.Coins, &user.CreatedAt, &user.UpdatedAt,
	)

	return user, err
}

func (u *userRepo) GetUserByID(userID int64) (models.User, error) {
	var (
		user  models.User
		query = "SELECT id, username, password_hash, coins, created_at, updated_at FROM users WHERE id = $1"
	)

	err := u.pool.QueryRow(u.ctx, query, userID).
		Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Coins, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (u *userRepo) GetUserByUsername(username string) (models.User, error) {
	var (
		user  models.User
		query = "SELECT id, username, password_hash, coins, created_at, updated_at FROM users WHERE username = $1"
	)

	err := u.pool.QueryRow(u.ctx, query, username).
		Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Coins, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (u *userRepo) UpdateUserCoins(userID int64, coins int64) error {
	_, err := u.pool.Exec(u.ctx, "UPDATE users SET coins = $1, updated_at = NOW() WHERE id = $2", coins, userID)
	return err
}
