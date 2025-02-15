package api

import (
	"net/http"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/icoder-new/avito-shop/api/handler"
	"github.com/icoder-new/avito-shop/pkg/logger"
)

func SetUpRoutes(h *handler.Handler, log *logger.Logger) *gin.Engine {
	router := gin.New()
	router.HandleMethodNotAllowed = true

	router.Use(
		requestid.New(
			requestid.WithGenerator(func() string {
				return "avito-shop-" + uuid.New().String()
			}),
			requestid.WithCustomHeaderStrKey("X-Request-ID"),
		),
	)

	router.Use(gin.Recovery())
	router.Use(handler.NewMWLogger(log))
	router.Use(handler.CorsMiddleware(h.Cfg.Settings.CORS))

	router.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"error": "Method not allowed",
		})
	})

	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Route not found",
		})
	})

	route := router.Group("/api")

	route.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	route.POST("/auth", h.Login)

	protected := route.Group("")
	protected.Use(handler.AuthMiddleware(log, h.Manager))
	protected.GET("/info", h.GetUserInfo)
	protected.POST("/sendCoin", h.SendCoin)
	protected.GET("/buy/:item", h.BuyItem)

	return router
}
