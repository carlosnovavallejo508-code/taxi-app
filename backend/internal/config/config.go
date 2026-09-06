package config

import (
    "os"
    "strconv"
    "time"
)

type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Redis    RedisConfig
    JWT      JWTConfig
    App      AppConfig
}

type ServerConfig struct {
    Port         string
    Environment  string
    AllowOrigins []string
}

type DatabaseConfig struct {
    Host     string
    Port     string
    User     string
    Password string
    DBName   string
    SSLMode  string
    MaxConns int
    MinConns int
}

type RedisConfig struct {
    Host     string
    Port     string
    Password string
    DB       int
}

type JWTConfig struct {
    Secret           string
    ExpirationHours  int
    RefreshHours     int
}

type AppConfig struct {
    Name          string
    Version       string
    MaxRideRadius float64
    DriverTimeout time.Duration
}

func Load() *Config {
    return &Config{
        Server: ServerConfig{
            Port:         getEnv("SERVER_PORT", "8080"),
            Environment:  getEnv("ENVIRONMENT", "development"),
            AllowOrigins: []string{"*"},
        },
        Database: DatabaseConfig{
            Host:     getEnv("DB_HOST", "localhost"),
            Port:     getEnv("DB_PORT", "5432"),
            User:     getEnv("DB_USER", "taxi_user"),
            Password: getEnv("DB_PASSWORD", "taxi_pass"),
            DBName:   getEnv("DB_NAME", "taxi_db"),
            SSLMode:  getEnv("DB_SSLMODE", "disable"),
            MaxConns: getEnvInt("DB_MAX_CONNS", 20),
            MinConns: getEnvInt("DB_MIN_CONNS", 5),
        },
        Redis: RedisConfig{
            Host:     getEnv("REDIS_HOST", "localhost"),
            Port:     getEnv("REDIS_PORT", "6379"),
            Password: getEnv("REDIS_PASSWORD", ""),
            DB:       getEnvInt("REDIS_DB", 0),
        },
        JWT: JWTConfig{
            Secret:          getEnv("JWT_SECRET", "tu-secreto-super-seguro"),
            ExpirationHours: getEnvInt("JWT_EXPIRATION_HOURS", 24),
            RefreshHours:    getEnvInt("JWT_REFRESH_HOURS", 168),
        },
        App: AppConfig{
            Name:          "Taxi App",
            Version:       "1.0.0",
            MaxRideRadius: 10.0,
            DriverTimeout: 30 * time.Second,
        },
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if intVal, err := strconv.Atoi(value); err == nil {
            return intVal
        }
    }
    return defaultValue
}
