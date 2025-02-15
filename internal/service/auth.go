package service

import (
	"github.com/icoder-new/avito-shop/internal/config"
	"github.com/icoder-new/avito-shop/internal/dto"
	"github.com/icoder-new/avito-shop/internal/storage"
	"github.com/icoder-new/avito-shop/pkg/errors"
	"github.com/icoder-new/avito-shop/pkg/hash"
	"github.com/icoder-new/avito-shop/pkg/jwt"
	"github.com/icoder-new/avito-shop/pkg/logger"
	"github.com/spf13/cast"
	"go.uber.org/zap"
)

type IAuth interface {
	Login(req dto.AuthRequest) (dto.AuthResponse, error)
}

type auth struct {
	cfg     *config.Config
	log     *logger.Logger
	storage storage.IStorage
	hasher  *hash.Hasher
	jwt     *jwt.TokenManager
}

func newAuth(cfg *config.Config, log *logger.Logger, storage storage.IStorage, jwt *jwt.TokenManager) IAuth {

	hasher := hash.NewHasher(hash.Config{
		Time:    1,
		Memory:  64 * 1024,
		Threads: 4,
		KeyLen:  32,
	})

	return &auth{
		cfg:     cfg,
		log:     log,
		storage: storage,
		hasher:  hasher,
		jwt:     jwt,
	}
}

func (a *auth) Login(req dto.AuthRequest) (dto.AuthResponse, error) {
	const op = "service.auth.Login"

	user, err := a.storage.User().GetUserByUsername(req.Username)
	if err != nil {
		hashedPassword, err := a.hasher.Hash(req.Password)
		if err != nil {
			a.log.Error("failed to hash password:",
				zap.String("method", op),
				zap.Error(err),
			)
			return dto.AuthResponse{}, errors.ErrInternal(err)
		}

		user, err = a.storage.User().CreateUser(req.Username, hashedPassword)
		if err != nil {
			a.log.Error("failed to create user:",
				zap.String("method", op),
				zap.Error(err),
			)
			return dto.AuthResponse{}, errors.ErrInternal(err)
		}

		err = a.storage.User().UpdateUserCoins(user.ID, cast.ToInt64(a.cfg.Settings.Service.InitialCoins))
		if err != nil {
			a.log.Error("failed to set initial coins:",
				zap.String("method", op),
				zap.Error(err),
			)
			return dto.AuthResponse{}, errors.ErrInternal(err)
		}
	} else {
		valid, err := a.hasher.Verify(req.Password, user.PasswordHash)
		if err != nil {
			a.log.Error("failed to verify password:",
				zap.String("method", op),
				zap.Error(err),
			)
			return dto.AuthResponse{}, errors.ErrInternal(err)
		}
		if !valid {
			return dto.AuthResponse{}, errors.ErrUnauthorized("invalid credentials")
		}
	}

	token, err := a.jwt.NewJWT(user.ID, user.Username)
	if err != nil {
		a.log.Error("failed to generate token:",
			zap.String("method", op),
			zap.Error(err),
		)
		return dto.AuthResponse{}, errors.ErrInternal(err)
	}

	return dto.AuthResponse{Token: token}, nil
}
