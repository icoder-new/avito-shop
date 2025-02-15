package service

import (
	"fmt"
	"github.com/icoder-new/avito-shop/internal/config"
	"github.com/icoder-new/avito-shop/internal/dto"
	"github.com/icoder-new/avito-shop/internal/models"
	"github.com/icoder-new/avito-shop/internal/storage"
	"github.com/icoder-new/avito-shop/pkg/errors"
	"github.com/icoder-new/avito-shop/pkg/logger"
	"go.uber.org/zap"
)

type IUser interface {
	GetInfo(userID int64) (dto.UserInfo, error)
}

type user struct {
	cfg     *config.Config
	log     *logger.Logger
	storage storage.IStorage
}

func newUser(cfg *config.Config, log *logger.Logger, storage storage.IStorage) IUser {
	return &user{
		cfg:     cfg,
		log:     log,
		storage: storage,
	}
}

func (u *user) GetInfo(userID int64) (dto.UserInfo, error) {
	const op = "service.user.GetInfo"

	user, err := u.storage.User().GetUserByID(userID)
	if err != nil {
		u.log.Error("failed to get user:",
			zap.String("method", op),
			zap.Error(err),
		)
		return dto.UserInfo{}, errors.ErrNotFound("user not found")
	}

	inventory, err := u.storage.Inventory().GetUserInventory(userID)
	if err != nil {
		u.log.Error("failed to get inventory:",
			zap.String("method", op),
			zap.Error(err),
		)
		return dto.UserInfo{}, errors.ErrInternal(err)
	}

	transactions, err := u.storage.Coin().GetUserTransactions(userID)
	if err != nil {
		u.log.Error("failed to get transactions:",
			zap.String("method", op),
			zap.Error(err),
		)
		return dto.UserInfo{}, errors.ErrInternal(err)
	}

	response := dto.UserInfo{
		Coins:     user.Coins,
		Inventory: u.convertInventory(inventory),
		CoinHistory: dto.CoinHistory{
			Received: make([]dto.CoinTransfer, 0),
			Sent:     make([]dto.CoinTransfer, 0),
		},
	}

	for _, tx := range transactions {
		if tx.Type == models.TransactionTypeTransfer {
			transfer := u.processCoinTransfer(tx, userID)
			if tx.ToUserID == userID {
				response.CoinHistory.Received = append(response.CoinHistory.Received, transfer)
			} else {
				response.CoinHistory.Sent = append(response.CoinHistory.Sent, transfer)
			}
		}
	}

	return response, nil
}

func (u *user) convertInventory(inventory []models.UserInventory) []dto.InventoryItem {
	items := make([]dto.InventoryItem, len(inventory))
	merchTypes := map[int64]string{
		1:  "t-shirt",
		2:  "cup",
		3:  "book",
		4:  "pen",
		5:  "powerbank",
		6:  "hoody",
		7:  "umbrella",
		8:  "socks",
		9:  "wallet",
		10: "pink-hoody",
	}

	for i, item := range inventory {
		itemType := merchTypes[item.MerchID]
		if itemType == "" {
			itemType = fmt.Sprintf("item_%d", item.MerchID)
		}

		items[i] = dto.InventoryItem{
			Type:     itemType,
			Quantity: item.Quantity,
		}
	}

	return items
}

func (u *user) processCoinTransfer(tx models.Transaction, userID int64) dto.CoinTransfer {
	if tx.ToUserID == userID {
		fromUser, err := u.storage.User().GetUserByID(tx.FromUserID)
		if err != nil {
			u.log.Error("failed to get sender info:", zap.Error(err))
			return dto.CoinTransfer{
				FromUser: "unknown",
				Amount:   tx.Amount,
			}
		}
		return dto.CoinTransfer{
			FromUser: fromUser.Username,
			Amount:   tx.Amount,
		}
	} else {
		toUser, err := u.storage.User().GetUserByID(tx.ToUserID)
		if err != nil {
			u.log.Error("failed to get receiver info:", zap.Error(err))
			return dto.CoinTransfer{
				ToUser: "unknown",
				Amount: tx.Amount,
			}
		}
		return dto.CoinTransfer{
			ToUser: toUser.Username,
			Amount: tx.Amount,
		}
	}
}
