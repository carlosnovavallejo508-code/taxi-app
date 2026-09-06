package handlers

import (
    "context"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "golang.org/x/crypto/bcrypt"

    "taxi-app/internal/config"
    "taxi-app/internal/database"
    "taxi-app/internal/models"
    "taxi-app/pkg/auth"
    "taxi-app/pkg/response"
)

type AuthHandler struct {
    cfg *config.Config
    db  *database.Database
}

func NewAuthHandler(cfg *config.Config, db *database.Database) *AuthHandler {
    return &AuthHandler{cfg: cfg, db: db}
}

type RegisterRequest struct {
    Phone    string `json:"phone" binding:"required"`
    Password string `json:"password" binding:"required,min=6"`
    FullName string `json:"full_name" binding:"required"`
    Email    string `json:"email"`
    UserType string `json:"user_type" binding:"required,oneof=passenger driver business admin"`
}

type LoginRequest struct {
    Phone    string `json:"phone" binding:"required"`
    Password string `json:"password" binding:"required"`
}

type TokenResponse struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    ExpiresIn    int    `json:"expires_in"`
    TokenType    string `json:"token_type"`
    User         models.User `json:"user"`
}

// Register godoc
// @Summary Register new user
// @Description Register a new user in the system
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "User data"
// @Success 201 {object} TokenResponse
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
    var req RegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, http.StatusBadRequest, "Datos inválidos", "INVALID_REQUEST")
        return
    }

    // Verificar si el usuario ya existe
    var existingUser models.User
    query := `SELECT id FROM users WHERE phone = $1`
    err := h.db.Pool.QueryRow(context.Background(), query, req.Phone).Scan(&existingUser.ID)
    if err == nil {
        response.Error(c, http.StatusConflict, "El teléfono ya está registrado", "PHONE_EXISTS")
        return
    }

    // Hash password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        response.Error(c, http.StatusInternalServerError, "Error al procesar contraseña", "HASH_ERROR")
        return
    }

    // Crear usuario
    user := models.User{
        ID:           uuid.New(),
        Phone:        req.Phone,
        FullName:     req.FullName,
        PasswordHash: string(hashedPassword),
        UserType:     req.UserType,
        Rating:       5.0,
        IsActive:     true,
        CreatedAt:    time.Now(),
        UpdatedAt:    time.Now(),
    }

    if req.Email != "" {
        user.Email = &req.Email
    }

    // Insertar en base de datos
    insertQuery := `
        INSERT INTO users (id, phone, email, password_hash, full_name, user_type, rating, is_active, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
    `
    _, err = h.db.Pool.Exec(context.Background(), insertQuery,
        user.ID, user.Phone, user.Email, user.PasswordHash, user.FullName,
        user.UserType, user.Rating, user.IsActive, user.CreatedAt, user.UpdatedAt,
    )
    if err != nil {
        response.Error(c, http.StatusInternalServerError, "Error al crear usuario", "DB_ERROR")
        return
    }

    // Si es conductor, crear registro en tabla drivers
    if req.UserType == "driver" {
        driverQuery := `
            INSERT INTO drivers (id, user_id, vehicle_type, license_plate, license_number, is_available, is_online, created_at, updated_at)
            VALUES ($1, $2, 'standard', 'PENDIENTE', 'PENDIENTE', true, false, $3, $4)
        `
        _, err = h.db.Pool.Exec(context.Background(), driverQuery,
            uuid.New(), user.ID, time.Now(), time.Now(),
        )
        if err != nil {
            // No es crítico si falla, el conductor puede completar después
            c.Set("warning", "Perfil de conductor creado, complete sus datos")
        }
    }

    // Si es negocio, crear registro en tabla businesses
    if req.UserType == "business" {
        businessQuery := `
            INSERT INTO businesses (id, user_id, name, created_at, updated_at)
            VALUES ($1, $2, $3, $4, $5)
        `
        _, err = h.db.Pool.Exec(context.Background(), businessQuery,
            uuid.New(), user.ID, req.FullName, time.Now(), time.Now(),
        )
        if err != nil {
            c.Set("warning", "Perfil de negocio creado, complete sus datos")
        }
    }

    // Generar tokens
    accessToken, err := auth.GenerateToken(user.ID, user.Phone, user.UserType, h.cfg)
    if err != nil {
        response.Error(c, http.StatusInternalServerError, "Error al generar token", "TOKEN_ERROR")
        return
    }

    refreshToken, err := auth.GenerateRefreshToken(user.ID, h.cfg)
    if err != nil {
        response.Error(c, http.StatusInternalServerError, "Error al generar token", "TOKEN_ERROR")
        return
    }

    response.Created(c, TokenResponse{
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        ExpiresIn:    h.cfg.JWT.ExpirationHours * 3600,
        TokenType:    "Bearer",
        User:         user,
    })
}

// Login godoc
// @Summary Login user
// @Description Authenticate user and get tokens
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} TokenResponse
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
    var req LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, http.StatusBadRequest, "Datos inválidos", "INVALID_REQUEST")
        return
    }

    // Buscar usuario
    var user models.User
    query := `
        SELECT id, phone, email, password_hash, full_name, user_type, 
               rating, total_rides, is_verified, is_active, created_at, updated_at
        FROM users 
        WHERE phone = $1 AND deleted_at IS NULL
    `
    err := h.db.Pool.QueryRow(context.Background(), query, req.Phone).Scan(
        &user.ID, &user.Phone, &user.Email, &user.PasswordHash, &user.FullName,
        &user.UserType, &user.Rating, &user.TotalRides, &user.IsVerified,
        &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
    )
    if err != nil {
        response.Error(c, http.StatusUnauthorized, "Credenciales inválidas", "INVALID_CREDENTIALS")
        return
    }

    // Verificar si el usuario está activo
    if !user.IsActive {
        response.Error(c, http.StatusForbidden, "Usuario desactivado", "USER_INACTIVE")
        return
    }

    // Verificar contraseña
    if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
        response.Error(c, http.StatusUnauthorized, "Credenciales inválidas", "INVALID_CREDENTIALS")
        return
    }

    // Actualizar último login
    updateQuery := `UPDATE users SET last_login = $1 WHERE id = $2`
    h.db.Pool.Exec(context.Background(), updateQuery, time.Now(), user.ID)

    // Generar tokens
    accessToken, err := auth.GenerateToken(user.ID, user.Phone, user.UserType, h.cfg)
    if err != nil {
        response.Error(c, http.StatusInternalServerError, "Error al generar token", "TOKEN_ERROR")
        return
    }

    refreshToken, err := auth.GenerateRefreshToken(user.ID, h.cfg)
    if err != nil {
        response.Error(c, http.StatusInternalServerError, "Error al generar token", "TOKEN_ERROR")
        return
    }

    // No devolver password hash
    user.PasswordHash = ""

    response.Success(c, http.StatusOK, TokenResponse{
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        ExpiresIn:    h.cfg.JWT.ExpirationHours * 3600,
        TokenType:    "Bearer",
        User:         user,
    })
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Get new access token using refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param refresh_token body string true "Refresh token"
// @Success 200 {object} TokenResponse
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
    var req struct {
        RefreshToken string `json:"refresh_token" binding:"required"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, http.StatusBadRequest, "Token requerido", "INVALID_REQUEST")
        return
    }

    // Validar refresh token
    claims, err := auth.ValidateToken(req.RefreshToken)
    if err != nil {
        response.Error(c, http.StatusUnauthorized, "Token inválido", "INVALID_TOKEN")
        return
    }

    // Buscar usuario
    var user models.User
    query := `SELECT id, phone, user_type, is_active FROM users WHERE id = $1`
    err = h.db.Pool.QueryRow(context.Background(), query, claims.UserID).Scan(
        &user.ID, &user.Phone, &user.UserType, &user.IsActive,
    )
    if err != nil || !user.IsActive {
        response.Error(c, http.StatusUnauthorized, "Usuario no encontrado", "USER_NOT_FOUND")
        return
    }

    // Generar nuevo access token
    accessToken, err := auth.GenerateToken(user.ID, user.Phone, user.UserType, h.cfg)
    if err != nil {
        response.Error(c, http.StatusInternalServerError, "Error al generar token", "TOKEN_ERROR")
        return
    }

    response.Success(c, http.StatusOK, gin.H{
        "access_token": accessToken,
        "expires_in":   h.cfg.JWT.ExpirationHours * 3600,
        "token_type":   "Bearer",
    })
}
