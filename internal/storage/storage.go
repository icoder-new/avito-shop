package storage

import (
	"github.com/icoder-new/avito-shop/internal/models"
)

type IStorage interface {
	CloseDB()

	User() IUser
	Coin() ICoin
	Inventory() IInventory
	TransactionHistory() ITransactionHistory
}

type IUser interface {
	CreateUser(username, passwordHash string) (models.User, error)
	GetUserByID(userID int64) (models.User, error)
	GetUserByUsername(username string) (models.User, error)
	UpdateUserCoins(userID int64, coins int64) error
}

type ICoin interface {
	TransferCoins(fromUserID, toUserID int64, amount int64) error
	GetUserTransactions(userID int64) ([]models.Transaction, error)
}

type IInventory interface {
	BuyItem(userID, merchID int64) error
	GetUserInventory(userID int64) ([]models.UserInventory, error)
}

type ITransactionHistory interface {
	GetTransactions(userID int64) ([]models.Transaction, error)
}
