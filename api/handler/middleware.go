package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/icoder-new/avito-shop/internal/config"
	"github.com/icoder-new/avito-shop/pkg/errors"
	"github.com/icoder-new/avito-shop/pkg/jwt"
	"github.com/icoder-new/avito-shop/pkg/logger"
	"go.uber.org/zap"
)

func NewMWLogger(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		c.Next()

		log.Info("incoming request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("remote_addr", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.String("request_id", c.GetHeader("X-Request-Id")),
			zap.Int("status", c.Writer.Status()),
			zap.Int("bytes", c.Writer.Size()),
			zap.Duration("latency", time.Since(startTime)),
		)
	}
}

func AuthMiddleware(log *logger.Logger, manager *jwt.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errors.AppError{
				Code:    errors.UnauthorizedError,
				Message: "Authorization header is required",
			})
			return
		}

		log.Info("authorization header:", zap.String("header", authHeader))

		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errors.AppError{
				Code:    errors.UnauthorizedError,
				Message: "Invalid authorization header format",
			})
			return
		}

		tokenString := headerParts[1]

		log.Info("token:", zap.String("token", tokenString))

		claims, err := manager.Parse(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errors.AppError{
				Code:    errors.UnauthorizedError,
				Message: "Invalid token",
			})
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)

		log.Info("user_id:", zap.Int64("user_id", claims.UserID))
		log.Info("username:", zap.String("username", claims.Username))

		c.Next()
	}
}

func CorsMiddleware(cfg config.CORSSettings) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", strings.Join(cfg.AllowedOrigins, ", "))
		c.Writer.Header().Set("Access-Control-Allow-Methods", strings.Join(cfg.AllowedMethods, ", "))
		c.Writer.Header().Set("Access-Control-Allow-Headers", strings.Join(cfg.AllowedHeaders, ", "))
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Max-Age", "300")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	}
}
