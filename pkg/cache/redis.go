package cache

import (
	"context"

	"github.com/ecommerce-platform/pkg/config"
	"github.com/redis/go-redis/v9"
)

// 根据传入的 Redis 配置创建并初始化一个 Redis 客户端 ， 同时会先执行 Ping 操作验证连接是否可用
func NewRedis(cfg *config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	return client, nil
}
