package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/gin-contrib/cors"
    
    "taxi-app/internal/config"
    "taxi-app/internal/database"
    "taxi-app/internal/router"
)

func main() {
    // Cargar configuración
    cfg := config.Load()
    
    // Conectar a base de datos
    db, err := database.NewPostgresConnection(cfg)
    if err != nil {
        log.Printf("⚠️  No se pudo conectar a PostgreSQL: %v", err)
        log.Println("⚠️  Iniciando sin base de datos (modo desarrollo)")
    } else {
        defer db.Close()
    }
    
    // Conectar a Redis
    cache, err := database.NewRedisConnection(cfg)
    if err != nil {
        log.Printf("⚠️  No se pudo conectar a Redis: %v", err)
        log.Println("⚠️  Iniciando sin cache (modo desarrollo)")
    } else {
        defer cache.Close()
    }
    
    // Configurar Gin
    if cfg.Server.Environment == "production" {
        gin.SetMode(gin.ReleaseMode)
    }
    
    r := gin.New()
    r.Use(gin.Logger())
    r.Use(gin.Recovery())
    
    // Configurar CORS
    r.Use(cors.New(cors.Config{
        AllowOrigins:     cfg.Server.AllowOrigins,
        AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-API-Key"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    }))
    
    // Configurar rutas
    router.SetupRoutes(r, cfg, db, cache)
    
    // Crear servidor HTTP
    srv := &http.Server{
        Addr:    ":" + cfg.Server.Port,
        Handler: r,
    }
    
    // Iniciar servidor en goroutine
    go func() {
        log.Printf("🚀 %s v%s iniciado en puerto %s", cfg.App.Name, cfg.App.Version, cfg.Server.Port)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Error al iniciar servidor: %v", err)
        }
    }()
    
    // Esperar señal de interrupción
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    log.Println("🛑 Apagando servidor...")
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    if err := srv.Shutdown(ctx); err != nil {
        log.Fatal("Error al apagar servidor:", err)
    }
    
    log.Println("✅ Servidor apagado correctamente")
}
