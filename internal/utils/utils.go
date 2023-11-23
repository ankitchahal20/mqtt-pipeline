package utils

import (
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/mqtt-pipeline/internal/config"
	"go.uber.org/zap"
)

var Logger *zap.Logger

func InitRedis() *redis.Client {
	cfg := config.GetConfig()
	client := redis.NewClient(&redis.Options{
		Addr:        cfg.RedisConfig.URL,
		Password:    "",
		DB:          cfg.RedisConfig.DBNum,
		IdleTimeout: time.Duration(cfg.RedisConfig.IdleTimeout),
	})

	return client
}

func InitLogClient() {
	Logger, _ = zap.NewDevelopment()
}
