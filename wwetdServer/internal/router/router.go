package router

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"wwetd-server/internal/config"
	"wwetd-server/internal/handler"
)

type Dependencies struct {
	Auth        *handler.AuthHandler
	Health      *handler.HealthHandler
	Users       *handler.UserHandler
	Restaurants *handler.RestaurantHandler
	WebSocket   *handler.WebSocketHandler
}

func New(cfg config.Config, deps Dependencies) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery(), CORSMiddleware(cfg.CORS.AllowedOrigins))

	engine.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "wwetd-server", "status": "ok"})
	})

	api := engine.Group("/api/v1")
	{
		api.GET("/health", deps.Health.Check)

		auth := api.Group("/auth")
		{
			auth.POST("/register", deps.Auth.Register)
			auth.POST("/login", deps.Auth.Login)
			auth.POST("/refresh", deps.Auth.Refresh)
			auth.POST("/otp/request", deps.Auth.RequestOTP)
			auth.POST("/otp/verify", deps.Auth.VerifyOTP)
		}

		users := api.Group("/users")
		{
			users.POST("", deps.Users.Create)
			users.GET("", deps.Users.List)
			users.GET("/:id", deps.Users.GetByID)
		}

		restaurants := api.Group("/restaurants")
		{
			restaurants.GET("/nearby", deps.Restaurants.SearchNearby)
			restaurants.GET("/today", deps.Restaurants.PickNearby)
			restaurants.GET("/viewed", deps.Restaurants.ListViewed)
			restaurants.POST("/:data_id/viewed", deps.Restaurants.RecordViewed)
			restaurants.GET("/saved", deps.Restaurants.ListSaved)
			restaurants.POST("/:data_id/saved", deps.Restaurants.Save)
			restaurants.GET("/:data_id", deps.Restaurants.GetDetail)
		}

		api.GET("/ws", deps.WebSocket.Handle)
	}

	engine.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"message": "route not found"}})
	})

	return engine
}
