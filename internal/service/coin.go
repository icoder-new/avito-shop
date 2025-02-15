package service

import (
	"github.com/icoder-new/avito-shop/internal/config"
	"github.com/icoder-new/avito-shop/internal/dto"
	"github.com/icoder-new/avito-shop/internal/storage"
	"github.com/icoder-new/avito-shop/pkg/errors"
	"github.com/icoder-new/avito-shop/pkg/logger"
	"go.uber.org/zap"
)

type ICoin interface {
	Send(fromUserID int64, req dto.SendCoinRequest) error
}

type coin struct {
	cfg     *config.Config
	log     *logger.Logger
	storage storage.IStorage
}

func newCoin(cfg *config.Config, log *logger.Logger, storage storage.IStorage) ICoin {
	return &coin{
		cfg:     cfg,
		log:     log,
		storage: storage,
	}
}

func (c *coin) Send(fromUserID int64, req dto.SendCoinRequest) error {
	const op = "service.coin.Send"

	toUser, err := c.storage.User().GetUserByUsername(req.ToUser)
	if err != nil {
		c.log.Error("recipient not found:",
			zap.String("method", op),
			zap.Error(err),
		)
		return errors.ErrNotFound("recipient not found")
	}

	if fromUserID == toUser.ID {
		return errors.ErrBadRequest("cannot send coins to yourself")
	}

	if req.Amount <= 0 {
		return errors.ErrBadRequest("amount must be positive")
	}

	sender, err := c.storage.User().GetUserByID(fromUserID)
	if err != nil {
		c.log.Error("failed to get sender:",
			zap.String("method", op),
			zap.Error(err),
		)
		return errors.ErrInternal(err)
	}

	if sender.Coins < req.Amount {
		return errors.ErrBadRequest("insufficient funds")
	}

	err = c.storage.Coin().TransferCoins(fromUserID, toUser.ID, req.Amount)
	if err != nil {
		c.log.Error("failed to transfer coins:",
			zap.String("method", op),
			zap.Error(err),
		)
		return errors.ErrInternal(err)
	}

	return nil
}
