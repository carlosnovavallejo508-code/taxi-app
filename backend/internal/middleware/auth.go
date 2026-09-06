package middleware

import (
    "net/http"
    "strings"
    
    "github.com/gin-gonic/gin"
    "taxi-app/pkg/auth"
    "taxi-app/pkg/response"
)

func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := extractToken(c)
        if token == "" {
            response.Error(c, http.StatusUnauthorized, "No token provided", "AUTH_REQUIRED")
            c.Abort()
            return
        }
        
        claims, err := auth.ValidateToken(token)
        if err != nil {
            response.Error(c, http.StatusUnauthorized, "Invalid token", "INVALID_TOKEN")
            c.Abort()
            return
        }
        
        c.Set("user_id", claims.UserID)
        c.Set("user_type", claims.UserType)
        c.Set("phone", claims.Phone)
        c.Next()
    }
}

func AdminMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        userType, exists := c.Get("user_type")
        if !exists || userType != "admin" {
            response.Error(c, http.StatusForbidden, "Admin access required", "FORBIDDEN")
            c.Abort()
            return
        }
        c.Next()
    }
}

func DriverMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        userType, exists := c.Get("user_type")
        if !exists || userType != "driver" {
            response.Error(c, http.StatusForbidden, "Driver access required", "FORBIDDEN")
            c.Abort()
            return
        }
        c.Next()
    }
}

func extractToken(c *gin.Context) string {
    bearerToken := c.GetHeader("Authorization")
    if len(strings.Split(bearerToken, " ")) == 2 {
        return strings.Split(bearerToken, " ")[1]
    }
    return ""
}
