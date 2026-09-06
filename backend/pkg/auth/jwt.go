package auth

import (
    "errors"
    "time"
    
    "github.com/golang-jwt/jwt/v5"
    "github.com/google/uuid"
    "taxi-app/internal/config"
)

type Claims struct {
    UserID   uuid.UUID `json:"user_id"`
    Phone    string    `json:"phone"`
    UserType string    `json:"user_type"`
    jwt.RegisteredClaims
}

func GenerateToken(userID uuid.UUID, phone, userType string, cfg *config.Config) (string, error) {
    claims := Claims{
        UserID:   userID,
        Phone:    phone,
        UserType: userType,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(cfg.JWT.ExpirationHours) * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Issuer:    cfg.App.Name,
        },
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(cfg.JWT.Secret))
}

func GenerateRefreshToken(userID uuid.UUID, cfg *config.Config) (string, error) {
    claims := Claims{
        UserID: userID,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(cfg.JWT.RefreshHours) * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Issuer:    cfg.App.Name,
        },
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(cfg.JWT.Secret))
}

func ValidateToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        return []byte(config.Load().JWT.Secret), nil
    })
    
    if err != nil {
        return nil, err
    }
    
    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }
    
    return nil, errors.New("invalid token")
}
