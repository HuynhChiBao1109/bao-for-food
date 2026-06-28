package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	App       AppConfig
	Server    ServerConfig
	CORS      CORSConfig
	Mongo     MongoConfig
	Redis     RedisConfig
	Cache     CacheConfig
	Piso      PisoConfig
	WebSocket WebSocketConfig
}

type AppConfig struct {
	Env   string
	Debug bool
}

type ServerConfig struct {
	Host string
	Port int
}

func (c ServerConfig) Address() string {
	return net.JoinHostPort(c.Host, strconv.Itoa(c.Port))
}

type CORSConfig struct {
	AllowedOrigins []string
}

type MongoConfig struct {
	URI      string
	Database string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type CacheConfig struct {
	UserTTL     time.Duration
	LocationTTL time.Duration
}

type PisoConfig struct {
	APIKey  string
	BaseURL string
	Timeout time.Duration
}

type WebSocketConfig struct {
	AllowedOrigins []string
	RedisChannel   string
}

func Load() (Config, error) {
	_ = godotenv.Load()

	serverPort, err := getEnvAsInt("SERVER_PORT", 8080)
	if err != nil {
		return Config{}, err
	}

	redisDB, err := getEnvAsInt("REDIS_DB", 0)
	if err != nil {
		return Config{}, err
	}

	userTTL, err := getEnvAsDuration("USER_CACHE_TTL", 5*time.Minute)
	if err != nil {
		return Config{}, err
	}

	locationTTL, err := getEnvAsDuration("LOCATION_CACHE_TTL", 24*time.Hour)
	if err != nil {
		return Config{}, err
	}

	pisoTimeout, err := getEnvAsDuration("PISO_TIMEOUT", 8*time.Second)
	if err != nil {
		return Config{}, err
	}

	appEnv := getEnv("APP_ENV", "development")
	debug, err := getEnvAsBool("DEBUG", appEnv != "production")
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		App: AppConfig{
			Env:   appEnv,
			Debug: debug,
		},
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Port: serverPort,
		},
		CORS: CORSConfig{
			AllowedOrigins: getEnvAsCSV("CORS_ALLOWED_ORIGINS", []string{"*"}),
		},
		Mongo: MongoConfig{
			URI:      getEnv("MONGO_URI", "mongodb://localhost:27017"),
			Database: getEnv("MONGO_DATABASE", "wwetd"),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       redisDB,
		},
		Cache: CacheConfig{
			UserTTL:     userTTL,
			LocationTTL: locationTTL,
		},
		Piso: PisoConfig{
			APIKey:  getEnv("PISO_API_KEY", ""),
			BaseURL: getEnv("PISO_BASE_URL", "https://api.pisomap.tech"),
			Timeout: pisoTimeout,
		},
		WebSocket: WebSocketConfig{
			AllowedOrigins: getEnvAsCSV("WS_ALLOWED_ORIGINS", []string{"*"}),
			RedisChannel:   getEnv("WS_REDIS_CHANNEL", "realtime.broadcast"),
		},
	}

	return cfg, nil
}

func getEnv(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func getEnvAsInt(key string, fallback int) (int, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s as int: %w", key, err)
	}
	return parsed, nil
}

func getEnvAsBool(key string, fallback bool) (bool, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("parse %s as bool: %w", key, err)
	}
	return parsed, nil
}

func getEnvAsDuration(key string, fallback time.Duration) (time.Duration, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s as duration: %w", key, err)
	}
	return parsed, nil
}

func getEnvAsCSV(key string, fallback []string) []string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parts := strings.Split(value, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item != "" {
			items = append(items, item)
		}
	}

	if len(items) == 0 {
		return fallback
	}
	return items
}
