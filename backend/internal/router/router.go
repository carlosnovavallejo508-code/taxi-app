package router

import (
    "github.com/gin-gonic/gin"
    "taxi-app/internal/config"
    "taxi-app/internal/database"
    "taxi-app/internal/handlers"
    "taxi-app/internal/middleware"
    "taxi-app/internal/websocket"
)

func SetupRoutes(r *gin.Engine, cfg *config.Config, db *database.Database, cache *database.Cache) {
    // Crear Hub WebSocket
    hub := websocket.NewHub()
    go hub.Run()
    
    // Crear handler WebSocket
    wsHandler := websocket.NewWebSocketHandler(hub)
    
    // Health check
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "status":  "ok",
            "service": cfg.App.Name,
            "version": cfg.App.Version,
        })
    })
    
    // API v1
    v1 := r.Group("/api/v1")
    {
        // Rutas públicas
        public := v1.Group("/public")
        {
            public.GET("/ping", func(c *gin.Context) {
                c.JSON(200, gin.H{"message": "pong"})
            })
        }
        
        // Rutas de autenticación (públicas)
        auth := v1.Group("/auth")
        {
            authHandler := handlers.NewAuthHandler(cfg, db)
            auth.POST("/register", authHandler.Register)
            auth.POST("/login", authHandler.Login)
            auth.POST("/refresh", authHandler.RefreshToken)
        }
        
        // Rutas protegidas
        protected := v1.Group("")
        protected.Use(middleware.AuthMiddleware())
        {
            // Usuarios
            userHandler := handlers.NewUserHandler(db)
            protected.GET("/users/me", userHandler.GetProfile)
            protected.PUT("/users/me", userHandler.UpdateProfile)
            
            // Conductores
            driverHandler := handlers.NewDriverHandler(db, cache)
            protected.GET("/drivers/nearby", driverHandler.GetNearbyDrivers)
            protected.PUT("/drivers/location", driverHandler.UpdateLocation)
            protected.PUT("/drivers/availability", driverHandler.UpdateAvailability)
            
            // Viajes
            rideHandler := handlers.NewRideHandler(db, cache)
            protected.POST("/rides", rideHandler.CreateRide)
            protected.GET("/rides/:id", rideHandler.GetRide)
            protected.GET("/rides/active", rideHandler.GetActiveRide)
            protected.PUT("/rides/:id/cancel", rideHandler.CancelRide)
            protected.PUT("/rides/:id/status", rideHandler.UpdateRideStatus)
            protected.POST("/rides/:id/rate", rideHandler.RateRide)
            
            // Notificaciones
            notificationHandler := handlers.NewNotificationHandler(db)
            protected.GET("/notifications", notificationHandler.GetNotifications)
            protected.PUT("/notifications/:id/read", notificationHandler.MarkAsRead)
            
            // WebSocket
            protected.GET("/ws", wsHandler.HandleConnection)
        }
        
        // Rutas de administrador
        admin := v1.Group("/admin")
        admin.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware())
        {
            adminHandler := handlers.NewAdminHandler(db)
            admin.GET("/stats", adminHandler.GetStats)
            admin.GET("/users", adminHandler.GetUsers)
            admin.GET("/drivers", adminHandler.GetDrivers)
            admin.GET("/rides", adminHandler.GetRides)
        }
    }
}
