package service

import (
	"github.com/icoder-new/avito-shop/internal/config"
	"github.com/icoder-new/avito-shop/internal/storage"
	"github.com/icoder-new/avito-shop/pkg/errors"
	"github.com/icoder-new/avito-shop/pkg/logger"
	"go.uber.org/zap"
)

type IInventory interface {
	BuyItem(userID int64, itemName string) error
}

type inventory struct {
	cfg     *config.Config
	log     *logger.Logger
	storage storage.IStorage
	items   map[string]itemInfo
}

type itemInfo struct {
	id    int64
	price int64
}

func newInventory(cfg *config.Config, log *logger.Logger, storage storage.IStorage) IInventory {
	items := map[string]itemInfo{
		"t-shirt":    {id: 1, price: 80},
		"cup":        {id: 2, price: 20},
		"book":       {id: 3, price: 50},
		"pen":        {id: 4, price: 10},
		"powerbank":  {id: 5, price: 200},
		"hoody":      {id: 6, price: 300},
		"umbrella":   {id: 7, price: 200},
		"socks":      {id: 8, price: 10},
		"wallet":     {id: 9, price: 50},
		"pink-hoody": {id: 10, price: 500},
	}

	return &inventory{
		cfg:     cfg,
		log:     log,
		storage: storage,
		items:   items,
	}
}

func (i *inventory) BuyItem(userID int64, itemName string) error {
	const op = "service.inventory.BuyItem"

	item, exists := i.items[itemName]
	if !exists {
		return errors.ErrBadRequest("invalid item name")
	}

	user, err := i.storage.User().GetUserByID(userID)
	if err != nil {
		i.log.Error("failed to get user:",
			zap.String("method", op),
			zap.Error(err),
		)
		return errors.ErrNotFound("user not found")
	}

	if user.Coins < item.price {
		return errors.ErrBadRequest("insufficient funds")
	}

	err = i.storage.Inventory().BuyItem(userID, item.id)
	if err != nil {
		i.log.Error("failed to buy item:",
			zap.String("method", op),
			zap.Error(err),
		)
		return errors.ErrInternal(err)
	}

	return nil
}
