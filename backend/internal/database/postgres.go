package database

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/jackc/pgx/v5/pgxpool"
    "taxi-app/internal/config"
)

type Database struct {
    Pool *pgxpool.Pool
}

func NewPostgresConnection(cfg *config.Config) (*Database, error) {
    dsn := fmt.Sprintf(
        "postgres://%s:%s@%s:%s/%s?sslmode=%s",
        cfg.Database.User,
        cfg.Database.Password,
        cfg.Database.Host,
        cfg.Database.Port,
        cfg.Database.DBName,
        cfg.Database.SSLMode,
    )

    poolConfig, err := pgxpool.ParseConfig(dsn)
    if err != nil {
        return nil, fmt.Errorf("error parsing database config: %w", err)
    }

    poolConfig.MaxConns = int32(cfg.Database.MaxConns)
    poolConfig.MinConns = int32(cfg.Database.MinConns)
    poolConfig.MaxConnLifetime = time.Hour
    poolConfig.MaxConnIdleTime = 30 * time.Minute

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
    if err != nil {
        return nil, fmt.Errorf("error creating connection pool: %w", err)
    }

    // Ping para verificar conexión
    if err := pool.Ping(ctx); err != nil {
        return nil, fmt.Errorf("error connecting to database: %w", err)
    }

    log.Println("✅ Conexión a PostgreSQL establecida")
    return &Database{Pool: pool}, nil
}

func (d *Database) Close() {
    if d.Pool != nil {
        d.Pool.Close()
    }
}
