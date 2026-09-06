package database

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/go-redis/redis/v9"
    "taxi-app/internal/config"
)

type Cache struct {
    Client *redis.Client
}

func NewRedisConnection(cfg *config.Config) (*Cache, error) {
    client := redis.NewClient(&redis.Options{
        Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
        Password: cfg.Redis.Password,
        DB:       cfg.Redis.DB,
    })

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := client.Ping(ctx).Err(); err != nil {
        return nil, fmt.Errorf("error connecting to Redis: %w", err)
    }

    log.Println("✅ Conexión a Redis establecida")
    return &Cache{Client: client}, nil
}

func (c *Cache) Close() {
    if c.Client != nil {
        c.Client.Close()
    }
}
