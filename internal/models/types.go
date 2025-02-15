package models

import "time"

type User struct {
	ID           int64     `db:"id" json:"id"`
	Username     string    `db:"username" json:"username"`
	PasswordHash string    `db:"password_hash" json:"-"`
	Coins        int64     `db:"coins" json:"coins"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

type Merch struct {
	ID    int64  `db:"id" json:"id"`
	Name  string `db:"name" json:"name"`
	Price int64  `db:"price" json:"price"`
}

type UserInventory struct {
	ID        int64     `db:"id" json:"id"`
	UserID    int64     `db:"user_id" json:"user_id"`
	MerchID   int64     `db:"merch_id" json:"merch_id"`
	Quantity  int64     `db:"quantity" json:"quantity"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type TransactionType string

const (
	TransactionTypeTransfer TransactionType = "transfer"
	TransactionTypePurchase TransactionType = "purchase"
)

type Transaction struct {
	ID         int64           `db:"id" json:"id"`
	FromUserID int64           `db:"from_user_id" json:"from_user_id"`
	ToUserID   int64           `db:"to_user_id" json:"to_user_id,omitempty"`
	Amount     int64           `db:"amount" json:"amount"`
	Type       TransactionType `db:"type" json:"type"`
	MerchID    *int64          `db:"merch_id" json:"merch_id,omitempty"`
	CreatedAt  time.Time       `db:"created_at" json:"created_at"`
}
