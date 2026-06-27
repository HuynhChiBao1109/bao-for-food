package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"wwetd-server/internal/config"
	"wwetd-server/internal/handler"
	"wwetd-server/internal/infrastructure/mongodb"
	"wwetd-server/internal/infrastructure/realtime"
	redisinfra "wwetd-server/internal/infrastructure/redis"
	"wwetd-server/internal/integrations/piso"
	"wwetd-server/internal/interfaces"
	"wwetd-server/internal/repository"
	"wwetd-server/internal/router"
	"wwetd-server/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	appCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	initCtx, cancel := context.WithTimeout(appCtx, 10*time.Second)
	defer cancel()

	mongoClient, err := mongodb.NewClient(initCtx, cfg.Mongo)
	if err != nil {
		log.Fatalf("connect mongodb: %v", err)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := mongoClient.Close(shutdownCtx); err != nil {
			log.Printf("close mongodb: %v", err)
		}
	}()

	redisClient, err := redisinfra.NewClient(initCtx, cfg.Redis)
	if err != nil {
		log.Fatalf("connect redis: %v", err)
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Printf("close redis: %v", err)
		}
	}()

	userRepo := repository.NewMongoUserRepository(mongoClient.Collection("users"))
	if err := userRepo.EnsureIndexes(initCtx); err != nil {
		log.Fatalf("ensure user indexes: %v", err)
	}

	authRepo := repository.NewMongoAuthRepository(mongoClient.Collection("auth_users"))
	if err := authRepo.EnsureIndexes(initCtx); err != nil {
		log.Fatalf("ensure auth indexes: %v", err)
	}

	restaurantRepo := repository.NewMongoRestaurantRepository(mongoClient.Collection("restaurant_details"))
	if err := restaurantRepo.EnsureIndexes(initCtx); err != nil {
		log.Fatalf("ensure restaurant indexes: %v", err)
	}

	userRestaurantRepo := repository.NewMongoUserRestaurantRepository(mongoClient.Collection("user_saved_restaurants"))
	if err := userRestaurantRepo.EnsureIndexes(initCtx); err != nil {
		log.Fatalf("ensure user restaurant indexes: %v", err)
	}

	hub := realtime.NewHub(cfg.WebSocket.AllowedOrigins)
	go hub.Run(appCtx)

	redisBridge := realtime.NewRedisBridge(redisClient, hub, []string{cfg.WebSocket.RedisChannel})
	if err := redisBridge.Start(appCtx); err != nil {
		log.Fatalf("start redis websocket bridge: %v", err)
	}

	pisoClient := piso.NewClient(cfg.Piso)

	authService := service.NewAuthService(authRepo, redisClient)
	userService := service.NewUserService(userRepo, redisClient, cfg.Cache.UserTTL, cfg.WebSocket.RedisChannel)
	restaurantService := service.NewRestaurantService(redisClient, pisoClient, restaurantRepo, userRestaurantRepo, cfg.Cache.LocationTTL)
	healthService := service.NewHealthService(map[string]interfaces.Pinger{
		"mongodb": mongoClient,
		"redis":   redisClient,
	})

	authHandler := handler.NewAuthHandler(authService)
	healthHandler := handler.NewHealthHandler(healthService)
	userHandler := handler.NewUserHandler(userService)
	restaurantHandler := handler.NewRestaurantHandler(restaurantService, authService)
	webSocketHandler := handler.NewWebSocketHandler(hub)

	engine := router.New(cfg, router.Dependencies{
		Auth:        authHandler,
		Health:      healthHandler,
		Users:       userHandler,
		Restaurants: restaurantHandler,
		WebSocket:   webSocketHandler,
	})

	server := &http.Server{
		Addr:         cfg.Server.Address(),
		Handler:      engine,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("server listening on %s", cfg.Server.Address())
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server listen: %v", err)
		}
	}()

	<-appCtx.Done()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown: %v", err)
	}
}
