package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/icoder-new/avito-shop/internal/config"
	"github.com/icoder-new/avito-shop/internal/dto"
	"github.com/icoder-new/avito-shop/internal/service"
	"github.com/icoder-new/avito-shop/pkg/errors"
	"github.com/icoder-new/avito-shop/pkg/jwt"
	"github.com/icoder-new/avito-shop/pkg/logger"
	"github.com/icoder-new/avito-shop/pkg/validator"
	"go.uber.org/zap"
	"net/http"
)

type Handler struct {
	Cfg       *config.Config
	Manager   *jwt.TokenManager
	log       *logger.Logger
	svc       service.IService
	validator *validator.Validator
}

func NewHandler(cfg *config.Config, log *logger.Logger, svc service.IService, manager *jwt.TokenManager) *Handler {
	return &Handler{
		Cfg:       cfg,
		Manager:   manager,
		log:       log,
		svc:       svc,
		validator: validator.New(),
	}
}

func (h *Handler) Login(c *gin.Context) {
	const op = "handler.Login"

	var req dto.AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("failed to bind request",
			zap.String("method", op),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, errors.AppError{
			Code:    errors.BadRequest,
			Message: "Invalid request body",
		})
		return
	}

	if err := h.validator.Validate(req); err != nil {
		h.log.Error("validation failed",
			zap.String("method", op),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, errors.AppError{
			Code:    errors.ValidationError,
			Message: err.Error(),
		})
		return
	}

	response, err := h.svc.Auth().Login(req)
	if err != nil {
		h.log.Error("login failed",
			zap.String("method", op),
			zap.Error(err),
		)
		if errors.IsAppError(err) {
			appErr := err.(*errors.AppError)
			c.JSON(appErr.Code, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, errors.AppError{
			Code:    errors.InternalServerError,
			Message: "Internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) GetUserInfo(c *gin.Context) {
	const op = "handler.GetUserInfo"

	userID, exists := c.Get("user_id")
	if !exists {
		h.log.Error("user_id not found in context",
			zap.String("method", op),
		)
		c.JSON(http.StatusUnauthorized, errors.AppError{
			Code:    errors.UnauthorizedError,
			Message: "Unauthorized",
		})
		return
	}

	info, err := h.svc.User().GetInfo(userID.(int64))
	if err != nil {
		h.log.Error("failed to get user info",
			zap.String("method", op),
			zap.Error(err),
		)
		if errors.IsAppError(err) {
			appErr := err.(*errors.AppError)
			c.JSON(appErr.Code, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, errors.AppError{
			Code:    errors.InternalServerError,
			Message: "Internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, info)
}

func (h *Handler) SendCoin(c *gin.Context) {
	const op = "handler.SendCoin"

	userID, exists := c.Get("user_id")
	if !exists {
		h.log.Error("user_id not found in context",
			zap.String("method", op),
		)
		c.JSON(http.StatusUnauthorized, errors.AppError{
			Code:    errors.UnauthorizedError,
			Message: "Unauthorized",
		})
		return
	}

	var req dto.SendCoinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("failed to bind request",
			zap.String("method", op),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, errors.AppError{
			Code:    errors.BadRequest,
			Message: "Invalid request body",
		})
		return
	}

	if err := h.validator.Validate(req); err != nil {
		h.log.Error("validation failed",
			zap.String("method", op),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, errors.AppError{
			Code:    errors.ValidationError,
			Message: err.Error(),
		})
		return
	}

	err := h.svc.Coin().Send(userID.(int64), req)
	if err != nil {
		h.log.Error("failed to send coins",
			zap.String("method", op),
			zap.Error(err),
		)
		if errors.IsAppError(err) {
			appErr := err.(*errors.AppError)
			c.JSON(appErr.Code, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, errors.AppError{
			Code:    errors.InternalServerError,
			Message: "Internal server error",
		})
		return
	}

	c.Status(http.StatusOK)
}

func (h *Handler) BuyItem(c *gin.Context) {
	const op = "handler.BuyItem"

	userID, exists := c.Get("user_id")
	if !exists {
		h.log.Error("user_id not found in context",
			zap.String("method", op),
		)
		c.JSON(http.StatusUnauthorized, errors.AppError{
			Code:    errors.UnauthorizedError,
			Message: "Unauthorized",
		})
		return
	}

	itemName := c.Param("item")
	if itemName == "" {
		h.log.Error("item name is empty",
			zap.String("method", op),
		)
		c.JSON(http.StatusBadRequest, errors.AppError{
			Code:    errors.BadRequest,
			Message: "Item name is required",
		})
		return
	}

	err := h.svc.Inventory().BuyItem(userID.(int64), itemName)
	if err != nil {
		h.log.Error("failed to buy item",
			zap.String("method", op),
			zap.Error(err),
		)
		if errors.IsAppError(err) {
			appErr := err.(*errors.AppError)
			c.JSON(appErr.Code, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, errors.AppError{
			Code:    errors.InternalServerError,
			Message: "Internal server error",
		})
		return
	}

	c.Status(http.StatusOK)
}
